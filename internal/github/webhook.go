package github

// GitHub webhook handler will be implemented here
// This will receive and process GitHub webhook events

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"deploy-platform/internal/build"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"deploy-platform/internal/queue"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var (
	webhookSecret string
	buildService  *build.Service
	buildQueue    queue.BuildQueue
)

// InitWebhook initializes webhook secret from config
func InitWebhook(cfg *config.Config) {
	webhookSecret = cfg.WebhookSecret
	if webhookSecret == "" {
		webhookSecret = "nncfebvjhebhjvrevjejrvhjelv" // Default for development
	}
}

// InitBuildService initializes the build service for webhook handlers
func InitBuildService() error {
	bs, err := build.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize build service: %w", err)
	}
	buildService = bs
	return nil
}

// InitBuildServiceWithService sets the build service instance directly
func InitBuildServiceWithService(bs *build.Service) {
	buildService = bs
}

// InitBuildQueue sets the build queue instance
func InitBuildQueue(q queue.BuildQueue) {
	buildQueue = q
}

func HandleWebhook(c *gin.Context) {
	// Verify webhook signature
	signature := c.GetHeader("X-Hub-Signature-256")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	if !verifySignature(signature, body) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	event := c.GetHeader("X-GitHub-Event")

	switch event {
	case "push":
		handlePushEvent(c, body)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
	}
}

func handlePushEvent(c *gin.Context, body []byte) {
	event, err := github.ParseWebHook("push", body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook: " + err.Error()})
		return
	}

	// Type assert to PushEvent
	pushEvent, ok := event.(*github.PushEvent)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unexpected event type"})
		return
	}

	// Handle nil pointers safely
	if pushEvent.Repo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository information missing"})
		return
	}

	if pushEvent.Repo.Owner == nil || pushEvent.Repo.Owner.Login == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository owner information missing"})
		return
	}

	if pushEvent.Repo.Name == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Repository name missing"})
		return
	}

	if pushEvent.HeadCommit == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Head commit information missing"})
		return
	}

	if pushEvent.HeadCommit.ID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Commit SHA missing"})
		return
	}

	// Find project by repo
	var project models.Project
	result := database.DB.Where("repo_owner = ? AND repo_name = ?",
		*pushEvent.Repo.Owner.Login, *pushEvent.Repo.Name).First(&project)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found for repository"})
		return
	}

	// Parse branch from ref (e.g., "refs/heads/main" -> "main")
	branch := ""
	if pushEvent.Ref != nil {
		parts := strings.Split(*pushEvent.Ref, "/")
		if len(parts) > 0 {
			branch = parts[len(parts)-1]
		}
	}
	if branch == "" {
		branch = "main" // Default branch
	}

	// Get commit message safely
	commitMsg := ""
	if pushEvent.HeadCommit.Message != nil {
		commitMsg = *pushEvent.HeadCommit.Message
	}

	// Hostname will be assigned during deployment by hostname manager
	// For now, leave it empty - it will be set when deployment is processed
	hostname := ""

	// Create deployment
	deployment := &models.Deployment{
		ProjectID: project.ID,
		Status:    "pending",
		CommitSHA: *pushEvent.HeadCommit.ID,
		CommitMsg: commitMsg,
		Branch:    branch,
		Hostname:  hostname,
	}

	if err := database.DB.Create(deployment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment: " + err.Error()})
		return
	}

	// Enqueue build job (will be processed by worker pool)
	if buildQueue != nil {
		if err := buildQueue.Enqueue(deployment.ID); err != nil {
			log.Printf("❌ Failed to enqueue deployment %d: %v", deployment.ID, err)
			database.DB.Model(&models.Deployment{}).Where("id = ?", deployment.ID).Update("status", "failed")
		} else {
			log.Printf("✅ Deployment %d enqueued for build", deployment.ID)
		}
	} else if buildService != nil {
		// Fallback to direct build if queue not available
		go func(deploymentID uint) {
			ctx := context.Background()
			if err := buildService.BuildDeployment(ctx, deploymentID); err != nil {
				log.Printf("❌ Build failed for deployment %d: %v", deploymentID, err)
				database.DB.Model(&models.Deployment{}).Where("id = ?", deploymentID).Update("status", "failed")
			} else {
				log.Printf("✅ Build completed successfully for deployment %d", deploymentID)
			}
		}(deployment.ID)
	} else {
		log.Println("⚠️  Build service not initialized, skipping build")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Deployment triggered",
		"deployment": deployment,
	})
}

func verifySignature(signature string, body []byte) bool {
	if signature == "" {
		return false
	}

	if webhookSecret == "" {
		// In development, allow requests without secret
		return true
	}

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}

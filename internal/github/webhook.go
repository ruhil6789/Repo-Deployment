package github

// GitHub webhook handler will be implemented here
// This will receive and process GitHub webhook events

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

var webhookSecret string

// InitWebhook initializes webhook secret from config
func InitWebhook(cfg *config.Config) {
	webhookSecret = cfg.WebhookSecret
	if webhookSecret == "" {
		webhookSecret = "nncfebvjhebhjvrevjejrvhjelv" // Default for development
	}
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

	// Generate unique hostname (temporary, will be assigned properly later)
	// Format: project-slug-commit-short.localhost
	commitSHA := *pushEvent.HeadCommit.ID
	shortCommit := commitSHA
	if len(commitSHA) > 7 {
		shortCommit = commitSHA[:7] // First 7 chars of commit SHA
	}

	// Use project slug or fallback to repo name
	projectSlug := project.Slug
	if projectSlug == "" {
		projectSlug = strings.ToLower(project.Name)
		if projectSlug == "" {
			projectSlug = "deploy"
		}
	}

	hostname := fmt.Sprintf("%s-%s.localhost", projectSlug, shortCommit)

	// Ensure uniqueness by checking database
	var existingDeployment models.Deployment
	for database.DB.Where("hostname = ?", hostname).First(&existingDeployment).Error == nil {
		// Hostname exists, generate a new one with random suffix
		randomBytes := make([]byte, 2)
		rand.Read(randomBytes)
		randomSuffix := hex.EncodeToString(randomBytes)
		hostname = fmt.Sprintf("%s-%s-%s.localhost", projectSlug, shortCommit, randomSuffix)
	}

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

	// TODO: Trigger build (we'll implement this in Phase 4)

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

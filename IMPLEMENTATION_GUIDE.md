# Hands-On Implementation Guide: Vercel-like Platform in Go

This guide will walk you through building a complete deployment platform step by step. Follow each section in order, and you'll learn Go while building a real-world project!

## Prerequisites

Before starting, make sure you have:
- Go 1.21+ installed (`go version`)
- Docker installed (`docker --version`)
- Kubernetes cluster (Minikube for local: `minikube start`)
- Git installed
- Basic understanding of REST APIs
- A GitHub account (for testing)

---

## Phase 1: Project Setup and Foundation

### Step 1.1: Initialize Go Module

**What you'll learn**: Go modules, project structure

1. Create the project structure:
```bash
mkdir -p cmd/api cmd/worker cmd/web
mkdir -p internal/api internal/build internal/kubernetes internal/hostname internal/github internal/models internal/database internal/config
mkdir -p web/ui
mkdir -p pkg/docker pkg/k8s
mkdir -p k8s/manifests
mkdir -p migrations
```

2. Initialize Go module:
```bash
cd /data/Go/deploy-platform
go mod init deploy-platform
```

3. Create `go.mod` and add initial dependencies:
```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get gorm.io/driver/postgres
go get github.com/google/go-github/v56/github
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
```

**Your task**: Run these commands and verify `go.mod` was created.

---

### Step 1.2: Create Basic Project Structure

**What you'll learn**: Go package organization, main functions

Create these files with basic structure:

**File: `cmd/api/main.go`**
```go
package main

import (
	"fmt"
	"log"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	fmt.Println("Starting API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
```

**File: `cmd/worker/main.go`**
```go
package main

import "fmt"

func main() {
	fmt.Println("Build worker starting...")
	// We'll implement this later
}
```

**Your task**: 
1. Create both files
2. Run `go run cmd/api/main.go`
3. Test: `curl http://localhost:8080/health`
4. You should see: `{"status":"ok"}`

**Learning check**: 
- What does `gin.Default()` do?
- What is `:8080` in `r.Run(":8080")`?

---

## Phase 2: Database Layer

### Step 2.1: Define Data Models

**What you'll learn**: Structs, GORM tags, relationships

**File: `internal/models/models.go`**
```go
package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GitHubID  int64     `gorm:"uniqueIndex" json:"github_id"`
	Username  string    `gorm:"uniqueIndex" json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Projects []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"`
}

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"`
	Name        string    `json:"name"`
	Slug        string    `gorm:"uniqueIndex" json:"slug"`
	RepoURL     string    `json:"repo_url"`
	RepoOwner   string    `json:"repo_owner"`
	RepoName    string    `json:"repo_name"`
	Branch      string    `gorm:"default:main" json:"branch"`
	GitHubToken string    `gorm:"type:text" json:"-"` // Don't expose in JSON
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Deployments []Deployment `gorm:"foreignKey:ProjectID" json:"deployments,omitempty"`
	Environments []Environment `gorm:"foreignKey:ProjectID" json:"environments,omitempty"`
}

type Deployment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"index" json:"project_id"`
	Status      string    `gorm:"default:pending" json:"status"` // pending, building, deploying, live, failed
	CommitSHA   string    `json:"commit_sha"`
	CommitMsg   string    `json:"commit_msg"`
	Branch      string    `json:"branch"`
	Hostname    string    `gorm:"uniqueIndex" json:"hostname"`
	ImageTag    string    `json:"image_tag"`
	K8sNamespace string   `json:"k8s_namespace"`
	K8sDeploymentName string `json:"k8s_deployment_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Build   Build   `gorm:"foreignKey:DeploymentID" json:"build,omitempty"`
}

type Build struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DeploymentID uint      `gorm:"index" json:"deployment_id"`
	Status       string    `gorm:"default:pending" json:"status"` // pending, building, success, failed
	Logs         string    `gorm:"type:text" json:"logs"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Environment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"index" json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `gorm:"type:text" json:"value"` // In production, encrypt this!
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

type Hostname struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Hostname  string    `gorm:"uniqueIndex" json:"hostname"`
	ProjectID uint      `gorm:"index" json:"project_id"`
	DeploymentID uint   `gorm:"index" json:"deployment_id"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

**Your task**:
1. Create the file
2. Understand each struct field
3. Research: What do GORM tags like `gorm:"primaryKey"` do?
4. Why do we use `json:"-"` for GitHubToken?

---

### Step 2.2: Database Connection

**What you'll learn**: Database connections, migrations

**File: `internal/database/db.go`**
```go
package database

import (
	"log"
	"deploy-platform/internal/models"
	
	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(databaseURL string) error {
	var err error
	var dialector gorm.Dialector
	
	// Use SQLite for development, PostgreSQL for production
	if databaseURL == "" {
		databaseURL = "deployments.db"
		dialector = sqlite.Open(databaseURL)
	} else {
		dialector = postgres.Open(databaseURL)
	}
	
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	if err != nil {
		return err
	}
	
	// Auto-migrate all models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Deployment{},
		&models.Build{},
		&models.Environment{},
		&models.Hostname{},
	)
	
	if err != nil {
		return err
	}
	
	log.Println("Database connected and migrated successfully")
	return nil
}
```

**Your task**:
1. Create the file
2. Update `cmd/api/main.go` to initialize the database:
```go
import (
	"deploy-platform/internal/database"
)

func main() {
	// Initialize database
	if err := database.InitDB(""); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// ... rest of your code
}
```
3. Run `go run cmd/api/main.go`
4. Check: You should see a `deployments.db` file created
5. Research: What does `AutoMigrate` do?

---

## Phase 3: GitHub Integration

### Step 3.1: GitHub OAuth Setup

**What you'll learn**: OAuth flow, HTTP handlers

First, create a GitHub OAuth App:
1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Create new OAuth App
3. Note: Client ID and Client Secret
4. Set Authorization callback URL: `http://localhost:8080/auth/github/callback`

**File: `internal/config/config.go`**
```go
package config

import "os"

type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	BaseURL            string
	BaseDomain         string // e.g., "deploy.example.com"
	DatabaseURL        string
	KubernetesConfig   string // Path to kubeconfig
}

func Load() *Config {
	return &Config{
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubCallbackURL:  getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/auth/github/callback"),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		BaseDomain:         getEnv("BASE_DOMAIN", "localhost"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		KubernetesConfig:   getEnv("KUBECONFIG", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

**File: `internal/github/oauth.go`**
```go
package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

var oauthConfig *oauth2.Config

func InitOAuth(cfg *config.Config) {
	oauthConfig = &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubCallbackURL,
		Scopes:       []string{"repo", "user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}
}

// HandleGitHubLogin initiates OAuth flow
func HandleGitHubLogin(c *gin.Context) {
	state := generateState()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	
	url := oauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGitHubCallback handles OAuth callback
func HandleGitHubCallback(c *gin.Context) {
	state := c.Query("state")
	cookieState, _ := c.Cookie("oauth_state")
	
	if state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}
	
	code := c.Query("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Get user info from GitHub
	client := github.NewClient(oauthConfig.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Create or update user in database
	dbUser := &models.User{
		GitHubID:  int64(*user.ID),
		Username:  *user.Login,
		Email:     user.Email,
		AvatarURL: *user.AvatarURL,
	}
	
	result := database.DB.Where("github_id = ?", dbUser.GitHubID).FirstOrCreate(dbUser, models.User{GitHubID: dbUser.GitHubID})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	
	// Update token (in production, encrypt this!)
	database.DB.Model(dbUser).Update("github_token", token.AccessToken)
	
	c.JSON(http.StatusOK, gin.H{
		"user": dbUser,
		"token": token.AccessToken, // In production, use JWT instead
	})
}

func generateState() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}
```

**Your task**:
1. Install OAuth dependency: `go get golang.org/x/oauth2`
2. Create both files
3. Add routes to `cmd/api/main.go`:
```go
import (
	"deploy-platform/internal/config"
	"deploy-platform/internal/github"
)

func main() {
	cfg := config.Load()
	github.InitOAuth(cfg)
	
	r := gin.Default()
	r.GET("/auth/github", github.HandleGitHubLogin)
	r.GET("/auth/github/callback", github.HandleGitHubCallback)
	// ... rest
}
```
4. Set environment variables:
```bash
export GITHUB_CLIENT_ID="your_client_id"
export GITHUB_CLIENT_SECRET="your_client_secret"
```
5. Test: Visit `http://localhost:8080/auth/github`

**Learning check**: 
- What is OAuth state and why do we use it?
- What does `FirstOrCreate` do in GORM?

---

### Step 3.2: GitHub Webhook Handler

**What you'll learn**: Webhook validation, event handling

**File: `internal/github/webhook.go`**
```go
package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v56/github"
)

const webhookSecret = "your_webhook_secret" // Store in config!

func HandleWebhook(c *gin.Context) {
	// Verify webhook signature
	signature := c.GetHeader("X-Hub-Signature-256")
	body, _ := io.ReadAll(c.Request.Body)
	
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
	var event github.PushEvent
	if err := github.ParseWebHook("push", body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find project by repo
	var project models.Project
	result := database.DB.Where("repo_owner = ? AND repo_name = ?", 
		*event.Repo.Owner.Login, *event.Repo.Name).First(&project)
	
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Create deployment
	deployment := &models.Deployment{
		ProjectID: project.ID,
		Status:    "pending",
		CommitSHA: *event.HeadCommit.ID,
		CommitMsg: *event.HeadCommit.Message,
		Branch:    *event.Ref,
	}
	
	database.DB.Create(deployment)
	
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
	
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(signature), []byte(expected))
}
```

**Your task**:
1. Create the file
2. Add webhook route:
```go
r.POST("/webhooks/github", github.HandleWebhook)
```
3. Setup webhook in GitHub:
   - Go to your repo → Settings → Webhooks
   - Add webhook: `http://your-server/webhooks/github`
   - Content type: `application/json`
   - Secret: (use same as in code)
   - Events: Just the push event
4. Test by pushing to your repo

**Learning check**:
- Why do we verify webhook signatures?
- What is HMAC?

---

## Phase 4: Build Service

### Step 4.1: Docker Client Wrapper

**What you'll learn**: Docker API, container operations

**File: `pkg/docker/client.go`**
```go
package docker

import (
	"context"
	"fmt"
	"io"
	
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	
	return &Client{cli: cli}, nil
}

func (c *Client) BuildImage(ctx context.Context, buildContext io.Reader, imageTag string, dockerfile string) error {
	buildOptions := types.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: dockerfile,
		Remove:     true,
	}
	
	response, err := c.cli.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	
	// Read build output (logs)
	_, err = io.Copy(io.Discard, response.Body)
	return err
}

func (c *Client) PushImage(ctx context.Context, imageTag string) error {
	// TODO: Implement image push to registry
	return nil
}
```

**Your task**:
1. Install Docker client: `go get github.com/docker/docker/client github.com/docker/docker/api/types`
2. Create the file
3. Research: What does `client.FromEnv` do?
4. Why do we use `io.Discard`?

---

### Step 4.2: Build Service Implementation

**What you'll learn**: Git operations, build orchestration

**File: `internal/build/service.go`**
```go
package build

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"deploy-platform/pkg/docker"
	"time"
)

type Service struct {
	dockerClient *docker.Client
}

func NewService() (*Service, error) {
	dc, err := docker.NewClient()
	if err != nil {
		return nil, err
	}
	
	return &Service{dockerClient: dc}, nil
}

func (s *Service) BuildDeployment(ctx context.Context, deploymentID uint) error {
	var deployment models.Deployment
	if err := database.DB.Preload("Project").First(&deployment, deploymentID).Error; err != nil {
		return err
	}
	
	// Create build record
	build := &models.Build{
		DeploymentID: deploymentID,
		Status:       "building",
		StartedAt:    &[]time.Time{time.Now()}[0],
	}
	database.DB.Create(build)
	
	// Clone repository
	repoPath := fmt.Sprintf("/tmp/builds/%d", deploymentID)
	if err := s.cloneRepo(deployment.Project.RepoURL, repoPath, deployment.Branch); err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}
	
	// Detect build type and create Dockerfile if needed
	dockerfile, err := s.detectAndCreateDockerfile(repoPath)
	if err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}
	
	// Build Docker image
	imageTag := fmt.Sprintf("deploy-%d:%s", deploymentID, deployment.CommitSHA[:7])
	buildContext, err := s.createBuildContext(repoPath)
	if err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}
	
	if err := s.dockerClient.BuildImage(ctx, buildContext, imageTag, dockerfile); err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}
	
	// Update build and deployment
	completed := time.Now()
	build.CompletedAt = &completed
	build.Status = "success"
	database.DB.Save(build)
	
	deployment.Status = "deploying"
	deployment.ImageTag = imageTag
	database.DB.Save(deployment)
	
	return nil
}

func (s *Service) cloneRepo(repoURL, path, branch string) error {
	os.MkdirAll(path, 0755)
	cmd := exec.Command("git", "clone", "-b", branch, repoURL, path)
	return cmd.Run()
}

func (s *Service) detectAndCreateDockerfile(repoPath string) (string, error) {
	// Check if Dockerfile exists
	if _, err := os.Stat(filepath.Join(repoPath, "Dockerfile")); err == nil {
		return "Dockerfile", nil
	}
	
	// Auto-generate Dockerfile based on detected language
	// This is simplified - you can expand this
	if _, err := os.Stat(filepath.Join(repoPath, "package.json")); err == nil {
		return s.createNodeDockerfile(repoPath)
	}
	
	if _, err := os.Stat(filepath.Join(repoPath, "requirements.txt")); err == nil {
		return s.createPythonDockerfile(repoPath)
	}
	
	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		return s.createGoDockerfile(repoPath)
	}
	
	return "", fmt.Errorf("could not detect project type")
}

func (s *Service) createNodeDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`
	
	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createPythonDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]`
	
	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createGoDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]`
	
	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createBuildContext(repoPath string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		
		relPath, _ := filepath.Rel(repoPath, path)
		header.Name = relPath
		
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		
		if !info.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()
			io.Copy(tw, data)
		}
		
		return nil
	})
	
	tw.Close()
	return &buf, err
}

func (s *Service) updateBuildStatus(buildID uint, status, logs string) {
	database.DB.Model(&models.Build{}).Where("id = ?", buildID).Updates(map[string]interface{}{
		"status": status,
		"logs":   logs,
	})
}
```

**Your task**:
1. Install git: `go get github.com/go-git/go-git/v5` (or use exec.Command as shown)
2. Create the file
3. Understand the build flow
4. Research: What is a tar archive and why do we use it for Docker builds?

**Learning check**:
- Why do we create Dockerfiles automatically?
- What does `Preload` do in GORM?

---

## Phase 5: Kubernetes Integration

### Step 5.1: Kubernetes Client Setup

**What you'll learn**: Kubernetes API, client-go

**File: `internal/kubernetes/client.go`**
```go
package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error
	
	if kubeconfigPath == "" {
		// In-cluster config
		config, err = rest.InClusterConfig()
	} else {
		// Out-of-cluster config
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	
	if err != nil {
		return nil, err
	}
	
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	
	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}
```

**Your task**:
1. Create the file
2. Test connection:
```go
// In cmd/api/main.go
k8sClient, err := kubernetes.NewClient("")
if err != nil {
	log.Println("K8s client error:", err)
} else {
	log.Println("K8s client connected!")
}
```
3. Research: What's the difference between in-cluster and out-of-cluster config?

---

### Step 5.2: Hostname Manager

**What you'll learn**: String manipulation, unique ID generation

**File: `internal/hostname/manager.go`**
```go
package hostname

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"strings"
)

type Manager struct {
	baseDomain string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		baseDomain: cfg.BaseDomain,
	}
}

func (m *Manager) GenerateHostname(projectSlug string) string {
	// Generate short hash
	hash := generateShortHash()
	
	// Create slug-safe hostname
	slug := strings.ToLower(strings.ReplaceAll(projectSlug, " ", "-"))
	slug = strings.ReplaceAll(slug, "_", "-")
	
	hostname := fmt.Sprintf("%s-%s.%s", slug, hash, m.baseDomain)
	return hostname
}

func (m *Manager) AssignHostname(projectID uint, deploymentID uint) (string, error) {
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		return "", err
	}
	
	// Check if project already has a hostname
	var existingHostname models.Hostname
	result := database.DB.Where("project_id = ? AND is_active = ?", projectID, true).First(&existingHostname)
	
	if result.Error == nil {
		// Reuse existing hostname
		return existingHostname.Hostname, nil
	}
	
	// Generate new hostname
	hostname := m.GenerateHostname(project.Slug)
	
	// Ensure uniqueness
	for {
		var check models.Hostname
		if database.DB.Where("hostname = ?", hostname).First(&check).Error != nil {
			break // Hostname is unique
		}
		hostname = m.GenerateHostname(project.Slug) // Regenerate
	}
	
	// Save hostname
	hostnameRecord := &models.Hostname{
		Hostname:     hostname,
		ProjectID:    projectID,
		DeploymentID: deploymentID,
		IsActive:     true,
	}
	database.DB.Create(hostnameRecord)
	
	return hostname, nil
}

func generateShortHash() string {
	b := make([]byte, 3) // 6 hex characters
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

**Your task**:
1. Create the file
2. Test hostname generation
3. Research: Why do we check for uniqueness?

---

### Step 5.3: Kubernetes Deployment Creation

**What you'll learn**: Kubernetes resources, YAML generation

**File: `internal/kubernetes/deployment.go`**
```go
package kubernetes

import (
	"context"
	"fmt"
	"deploy-platform/internal/models"
	
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (c *Client) CreateDeployment(ctx context.Context, deployment *models.Deployment, hostname string, envVars map[string]string) error {
	namespace := "default" // Or create per-project namespace
	deploymentName := fmt.Sprintf("deploy-%d", deployment.ID)
	
	// Create Deployment
	k8sDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "app",
							Image: deployment.ImageTag,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: convertEnvVars(envVars),
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
	
	_, err := c.clientset.AppsV1().Deployments(namespace).Create(ctx, k8sDeployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	
	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": deploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
	
	_, err = c.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	
	// Create Ingress
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: hostname,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: func() *networkingv1.PathType { p := networkingv1.PathTypePrefix; return &p }(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: deploymentName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	_, err = c.clientset.NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	return err
}

func convertEnvVars(envVars map[string]string) []corev1.EnvVar {
	var env []corev1.EnvVar
	for k, v := range envVars {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return env
}

func int32Ptr(i int32) *int32 { return &i }
```

**Your task**:
1. Create the file
2. Understand each Kubernetes resource
3. Research: What are Deployments, Services, and Ingress?
4. Why do we set resource limits?

---

## Phase 6: API Endpoints

### Step 6.1: Project Management API

**File: `internal/api/handlers.go`**
```go
package api

import (
	"net/http"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	
	"github.com/gin-gonic/gin"
)

func ListProjects(c *gin.Context) {
	var projects []models.Project
	database.DB.Find(&projects)
	c.JSON(http.StatusOK, projects)
}

func GetProject(c *gin.Context) {
	var project models.Project
	id := c.Param("id")
	
	if err := database.DB.Preload("Deployments").Preload("Environments").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	c.JSON(http.StatusOK, project)
}

func CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	database.DB.Create(&project)
	c.JSON(http.StatusCreated, project)
}

func ListDeployments(c *gin.Context) {
	var deployments []models.Deployment
	projectID := c.Param("projectId")
	
	database.DB.Where("project_id = ?", projectID).Preload("Build").Find(&deployments)
	c.JSON(http.StatusOK, deployments)
}

func GetDeployment(c *gin.Context) {
	var deployment models.Deployment
	id := c.Param("id")
	
	if err := database.DB.Preload("Project").Preload("Build").First(&deployment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}
	
	c.JSON(http.StatusOK, deployment)
}
```

**Your task**:
1. Create the file
2. Add routes to `cmd/api/main.go`:
```go
import "deploy-platform/internal/api"

apiGroup := r.Group("/api")
{
	apiGroup.GET("/projects", api.ListProjects)
	apiGroup.GET("/projects/:id", api.GetProject)
	apiGroup.POST("/projects", api.CreateProject)
	apiGroup.GET("/projects/:projectId/deployments", api.ListDeployments)
	apiGroup.GET("/deployments/:id", api.GetDeployment)
}
```
3. Test with curl or Postman

---

## Phase 7: Web UI (Optional but Recommended)

Create a simple HTML/JS dashboard. This is a good exercise in frontend-backend integration!

**File: `web/ui/index.html`**
```html
<!DOCTYPE html>
<html>
<head>
	<title>Deployment Platform</title>
	<script src="https://unpkg.com/axios/dist/axios.min.js"></script>
</head>
<body>
	<h1>My Projects</h1>
	<div id="projects"></div>
	
	<script>
		axios.get('/api/projects')
			.then(response => {
				const projects = response.data;
				const div = document.getElementById('projects');
				projects.forEach(project => {
					div.innerHTML += `<div>
						<h3>${project.name}</h3>
						<p>${project.repo_url}</p>
					</div>`;
				});
			});
	</script>
</body>
</html>
```

**Your task**: 
1. Create the file
2. Serve static files in `cmd/api/main.go`:
```go
r.Static("/ui", "./web/ui")
r.StaticFile("/", "./web/ui/index.html")
```

---

## Phase 8: Integration and Testing

### Step 8.1: Connect Everything

Update your main API to trigger builds and deployments:

```go
// In cmd/api/main.go or a new handler
func TriggerDeployment(c *gin.Context) {
	var deployment models.Deployment
	id := c.Param("id")
	
	database.DB.First(&deployment, id)
	
	// Start build
	go func() {
		buildService, _ := build.NewService()
		buildService.BuildDeployment(context.Background(), deployment.ID)
		
		// After build, deploy to K8s
		hostnameManager := hostname.NewManager(config.Load())
		hostname, _ := hostnameManager.AssignHostname(deployment.ProjectID, deployment.ID)
		
		k8sClient, _ := kubernetes.NewClient("")
		k8sClient.CreateDeployment(context.Background(), &deployment, hostname, map[string]string{})
	}()
	
	c.JSON(http.StatusOK, gin.H{"message": "Deployment started"})
}
```

---

## Next Steps

1. **Error Handling**: Add proper error handling everywhere
2. **Logging**: Use structured logging (logrus or zap)
3. **Testing**: Write unit tests for each component
4. **Security**: Encrypt sensitive data, add authentication middleware
5. **Monitoring**: Add health checks, metrics
6. **Documentation**: Document your API

---

## Learning Resources

- [Go by Example](https://gobyexample.com/)
- [GORM Documentation](https://gorm.io/docs/)
- [Kubernetes Concepts](https://kubernetes.io/docs/concepts/)
- [Docker API](https://docs.docker.com/engine/api/)

---

## Troubleshooting

**Database errors**: Make sure SQLite file has write permissions
**Docker errors**: Ensure Docker daemon is running
**Kubernetes errors**: Check `kubectl` access and cluster connectivity
**GitHub OAuth**: Verify callback URL matches exactly

---

Good luck! Build step by step, test each phase, and don't hesitate to experiment!

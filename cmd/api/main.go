package main

import (
	"fmt"
	"log"
	"net/http"

	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/github"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	// Validate OAuth config before initializing
	if cfg.GitHubClientID == "" {
		log.Fatal("❌ GITHUB_CLIENT_ID is not set! Please check your .env file")
	}
	if cfg.GitHubClientSecret == "" {
		log.Fatal("❌ GITHUB_CLIENT_SECRET is not set! Please check your .env file")
	}

	log.Printf("✅ OAuth Config loaded - Client ID: %s...", cfg.GitHubClientID[:10])

	github.InitOAuth(cfg)
    github.InitWebhook(cfg)
	// Initialize database
	if err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	r := gin.Default()
	r.GET("/auth/github", github.HandleGitHubLogin)
	r.GET("/auth/github/callback", github.HandleGitHubCallback)

	r.POST("/webhooks/github", github.HandleWebhook)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fmt.Println("Starting API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

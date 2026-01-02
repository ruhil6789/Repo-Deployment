package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"deploy-platform/internal/api"
	"deploy-platform/internal/auth"
	"deploy-platform/internal/build"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/github"
	"deploy-platform/internal/hostname"
	"deploy-platform/internal/kubernetes"
	"deploy-platform/internal/oauth"
	"deploy-platform/internal/queue"
	"deploy-platform/internal/ratelimit"
	"deploy-platform/pkg/docker"

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
	oauth.InitGoogleOAuth(cfg)

	// Initialize database
	if err := database.InitDB(cfg.DatabaseURL); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize Docker client: %v", err)
		log.Println("   Builds will be skipped. Make sure Docker is running.")
		dockerClient = nil
	} else {
		log.Println("✅ Docker client initialized")
	}

	// Initialize Kubernetes client (optional)
	// Try to initialize even if config is empty (will use in-cluster or default kubeconfig)
	var k8sClient *kubernetes.Client
	k8s, err := kubernetes.NewClient(cfg.KubernetesConfig)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to initialize Kubernetes client: %v", err)
		log.Println("   Kubernetes deployments will be skipped.")
	} else {
		k8sClient = k8s
		log.Println("✅ Kubernetes client initialized")
	}

	// Initialize hostname manager
	hostnameMgr := hostname.NewManager(cfg)

	// Initialize JWT
	auth.InitJWT(cfg)

	// Initialize build service for webhook handlers
	var buildService *build.Service
	if dockerClient != nil {
		if k8sClient != nil {
			// Use build service with Kubernetes support
			buildService = build.NewServiceWithK8s(dockerClient, k8sClient, hostnameMgr)
			log.Println("✅ Build service initialized with Kubernetes support")
		} else {
			// Use build service without Kubernetes
			buildService, err = build.NewService()
			if err != nil {
				log.Printf("⚠️  Warning: Failed to initialize build service: %v", err)
			} else {
				log.Println("✅ Build service initialized (without Kubernetes)")
			}
		}
		github.InitBuildServiceWithService(buildService)
	} else {
		log.Println("⚠️  Build service not initialized (Docker client unavailable)")
	}

	// Initialize build queue and worker pool
	var workerPool *queue.WorkerPool
	if buildService != nil {
		buildQueue := queue.NewInMemoryQueue()
		github.InitBuildQueue(buildQueue)

		// Start worker pool with 3 workers (configurable)
		workerPool = queue.NewWorkerPool(buildQueue, buildService, 3)
		workerPool.Start()
		log.Println("✅ Build queue and worker pool initialized")
	}

	// Initialize rate limiter (10 requests per minute per IP)
	rateLimiter := ratelimit.NewLimiter(10, 60*time.Second)

	// Setup Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Public routes
	r.GET("/", api.ServeIndex)
	r.GET("/login", api.ServeLogin)
	r.GET("/dashboard", func(c *gin.Context) {
		// Try to get from query parameter (OAuth redirect)
		if queryToken := c.Query("token"); queryToken != "" {
			// Store token in localStorage via JavaScript redirect
			c.HTML(http.StatusOK, "dashboard_redirect.html", gin.H{
				"Token": queryToken,
			})
			return
		}
		// For regular access, serve the dashboard page
		// Client-side JavaScript will handle authentication via localStorage
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Auth routes
	r.GET("/auth/github", github.HandleGitHubLogin)
	r.GET("/auth/github/callback", github.HandleGitHubCallback)
	r.GET("/auth/google", oauth.HandleGoogleLogin)
	r.GET("/auth/google/callback", oauth.HandleGoogleCallback)

	// API routes
	apiGroup := r.Group("/api")
	{
		// Public auth endpoints
		apiGroup.POST("/auth/register", api.Register)
		apiGroup.POST("/auth/login", api.Login)

		// Protected endpoints
		protected := apiGroup.Group("")
		protected.Use(auth.AuthMiddleware())
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID := c.GetUint("user_id")
				username := c.GetString("username")
				c.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
				})
			})
			protected.GET("/projects", api.GetProjects)
			protected.POST("/projects", api.CreateProject)
			protected.POST("/projects/:id/link", api.LinkProject)
			protected.GET("/deployments", api.GetDeployments)
			protected.GET("/deployments/:id", api.GetDeployment)
		}
	}

	// Webhook with rate limiting
	r.POST("/webhooks/github", func(c *gin.Context) {
		// Simple rate limiting (in production, use a per-IP limiter map)
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		github.HandleWebhook(c)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Graceful shutdown
	defer func() {
		if workerPool != nil {
			workerPool.Stop()
		}
	}()

	fmt.Println("🚀 Starting API server on :8080")
	fmt.Println("📊 Dashboard: http://localhost:8080")
	fmt.Println("🔐 Login: http://localhost:8080/login")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

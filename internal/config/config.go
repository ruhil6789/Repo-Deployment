package config

// Configuration management will be here
// This will load environment variables and application config

import "os"

type Config struct {
	GitHubClientID     string
	GitHubClientSecret string
	GitHubCallbackURL  string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	BaseURL            string
	BaseDomain         string // e.g., "deploy.example.com" or "localhost" for development
	PublicURL          string // Public URL prefix, e.g., "https://" or "http://"
	DatabaseURL        string
	KubernetesConfig   string // Path to kubeconfig
	JWTSecret          string // Add this
	WebhookSecret      string // Add this
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Load() *Config {
	return &Config{
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubCallbackURL:  getEnv("GITHUB_CALLBACK_URL", "http://localhost:8080/auth/github/callback"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleCallbackURL:  getEnv("GOOGLE_CALLBACK_URL", "http://localhost:8080/auth/google/callback"),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		BaseDomain:         getEnv("BASE_DOMAIN", "localhost"),
		PublicURL:          getEnv("PUBLIC_URL", "http://"), // http:// for localhost, https:// for production
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		KubernetesConfig:   getEnv("KUBECONFIG", ""),
		JWTSecret:          getEnv("JWT_SECRET", "bbdjvcbjfebvjebvjbejvhbejbvjfnvkj"),
		WebhookSecret:      getEnv("WEBHOOK_SECRET", ""), // Add this
	}
}

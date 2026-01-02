package oauth

import (
	"context"
	"crypto/rand"
	"deploy-platform/internal/auth"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOAuth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

var googleOAuthConfig *oauth2.Config

// InitGoogleOAuth initializes Google OAuth configuration
func InitGoogleOAuth(cfg *config.Config) {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		log.Println("⚠️  Google OAuth not configured (missing GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET)")
		return
	}

	googleOAuthConfig = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleCallbackURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
	log.Println("✅ Google OAuth initialized")
}

// HandleGoogleLogin initiates Google OAuth flow
func HandleGoogleLogin(c *gin.Context) {
	if googleOAuthConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Google OAuth not configured",
		})
		return
	}

	state := generateState()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url := googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGoogleCallback handles Google OAuth callback
func HandleGoogleCallback(c *gin.Context) {
	state := c.Query("state")
	cookieState, _ := c.Cookie("oauth_state")

	if state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token: " + err.Error()})
		return
	}

	// Get user info from Google
	client := googleOAuthConfig.Client(context.Background(), token)
	service, err := googleOAuth2.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google service: " + err.Error()})
		return
	}

	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Create or update user
	email := userInfo.Email
	if email == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email not provided by Google"})
		return
	}

	username := userInfo.Name
	if username == "" {
		username = email // Fallback to email if name not available
	}

	dbUser := &models.User{
		Username:  username,
		Email:     email,
		AvatarURL: userInfo.Picture,
	}

	// Check if user exists by email
	var existingUser models.User
	result := database.DB.Where("email = ?", email).First(&existingUser)

	if result.Error != nil {
		// User doesn't exist, create new
		if err := database.DB.Create(dbUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
			return
		}
	} else {
		// User exists, update
		dbUser = &existingUser
		if userInfo.Picture != "" {
			dbUser.AvatarURL = userInfo.Picture
		}
		if username != "" {
			dbUser.Username = username
		}
		database.DB.Save(dbUser)
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(dbUser.ID, dbUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token: " + err.Error()})
		return
	}

	// Redirect to dashboard with token
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard?token="+jwtToken)
}

func generateState() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

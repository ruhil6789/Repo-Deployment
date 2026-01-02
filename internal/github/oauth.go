package github

// GitHub OAuth implementation will be here
// This will handle OAuth flow and user authentication

import (
	"context"
	"crypto/rand"
	"deploy-platform/internal/auth"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"encoding/base64"
	"io"
	"net/http"

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

		Scopes:   []string{"repo", "user:email"},
		Endpoint: githubOAuth.Endpoint,
	}
}

// HandleGitHubLogin initiates OAuth flow
func HandleGitHubLogin(c *gin.Context) {
	state := generateState()
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url := oauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleGitHubCallback handles OAuth callback (fixed function name)
func HandleGitHubCallback(c *gin.Context) {
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

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token: " + err.Error()})
		return
	}

	// Get user info from GitHub
	client := github.NewClient(oauthConfig.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Handle nil pointers safely
	if user.ID == nil || user.Login == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data from GitHub"})
		return
	}

	// Get user email (might need separate API call)
	email := ""
	if user.Email != nil {
		email = *user.Email
	} else {
		// Try to get email from user emails endpoint
		emails, _, err := client.Users.ListEmails(context.Background(), nil)
		if err == nil && len(emails) > 0 {
			for _, e := range emails {
				if e.Primary != nil && *e.Primary {
					email = *e.Email
					break
				}
			}
			if email == "" && len(emails) > 0 {
				email = *emails[0].Email
			}
		}
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	// Create or update user in database
	githubID := int64(*user.ID)
	dbUser := &models.User{
		GitHubID:  &githubID,
		Username:  *user.Login,
		Email:     email,
		AvatarURL: avatarURL,
	}

	result := database.DB.Where("github_id = ?", *dbUser.GitHubID).FirstOrCreate(dbUser, models.User{GitHubID: dbUser.GitHubID})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + result.Error.Error()})
		return
	}

	// Update GitHub token (store encrypted in production!)
	if err := database.DB.Model(dbUser).Update("github_token", token.AccessToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update token: " + err.Error()})
		return
	}

	// Generate JWT token instead of returning GitHub token
	jwtToken, err := auth.GenerateToken(dbUser.ID, dbUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT token: " + err.Error()})
		return
	}

	// Redirect to dashboard with token (same as Google OAuth)
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard?token="+jwtToken)
}

func generateState() string {
	b := make([]byte, 32)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

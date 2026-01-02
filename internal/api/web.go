package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServeLogin serves the login page
func ServeLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// ServeDashboard serves the dashboard page
func ServeDashboard(c *gin.Context) {
	// Check if user is authenticated
	token := c.GetHeader("Authorization")
	if token == "" {
		// Try to get from cookie or redirect
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	c.HTML(http.StatusOK, "index.html", nil)
}

// ServeIndex redirects to dashboard or login
func ServeIndex(c *gin.Context) {
	// Check authentication via middleware or cookie
	c.Redirect(http.StatusTemporaryRedirect, "/dashboard")
}

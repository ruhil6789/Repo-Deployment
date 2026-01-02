package api

import (
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CreateProjectRequest represents a project creation request
type CreateProjectRequest struct {
	Name      string `json:"name" binding:"required"`
	RepoURL   string `json:"repo_url" binding:"required"`
	RepoOwner string `json:"repo_owner" binding:"required"`
	RepoName  string `json:"repo_name" binding:"required"`
	Branch    string `json:"branch"`
}

// CreateProject creates a new project
func CreateProject(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if project already exists
	var existingProject models.Project
	if err := database.DB.Where("repo_owner = ? AND repo_name = ?", req.RepoOwner, req.RepoName).First(&existingProject).Error; err == nil {
		// Project exists, link it to current user if not already linked
		if existingProject.UserID != userID {
			existingProject.UserID = userID
			database.DB.Save(&existingProject)
		}
		c.JSON(http.StatusOK, existingProject)
		return
	}

	// Generate slug from name
	slug := generateSlug(req.Name)

	// Create new project
	project := &models.Project{
		UserID:    userID,
		Name:      req.Name,
		Slug:      slug,
		RepoURL:   req.RepoURL,
		RepoOwner: req.RepoOwner,
		RepoName:  req.RepoName,
		Branch:    req.Branch,
	}

	if req.Branch == "" {
		project.Branch = "main"
	}

	if err := database.DB.Create(project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// LinkProject links an existing project to the current user
func LinkProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	project.UserID = userID
	if err := database.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to link project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func generateSlug(name string) string {
	slug := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			slug += string(char)
		} else if char == ' ' || char == '-' || char == '_' {
			slug += "-"
		}
	}
	if slug == "" {
		slug = "project"
	}
	return slug
}

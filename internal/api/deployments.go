package api

import (
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDeployments returns all deployments for the authenticated user
func GetDeployments(c *gin.Context) {
	userID := c.GetUint("user_id")

	var deployments []models.Deployment
	if err := database.DB.Where("project_id IN (SELECT id FROM projects WHERE user_id = ?)", userID).
		Preload("Project").
		Preload("Build").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, deployments)
}

// GetDeployment returns a specific deployment
func GetDeployment(c *gin.Context) {
	userID := c.GetUint("user_id")
	deploymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var deployment models.Deployment
	if err := database.DB.Preload("Project").Preload("Build").First(&deployment, deploymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Check if user owns this deployment
	var project models.Project
	if err := database.DB.First(&project, deployment.ProjectID).Error; err != nil || project.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// GetProjects returns all projects for the authenticated user
func GetProjects(c *gin.Context) {
	userID := c.GetUint("user_id")

	var projects []models.Project
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Deployments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(10)
		}).
		Order("created_at DESC").
		Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	// Keep only the latest deployment with hostname for each project (for "Live" link)
	for i := range projects {
		// Find latest deployment with hostname
		var latestDeployment models.Deployment
		result := database.DB.Where("project_id = ? AND hostname != ? AND hostname != ''", projects[i].ID, "").
			Order("created_at DESC").
			First(&latestDeployment)

		// Replace deployments array with just the latest one (if found)
		if result.Error == nil && latestDeployment.ID > 0 {
			projects[i].Deployments = []models.Deployment{latestDeployment}
		} else {
			// If no deployment with hostname, keep the latest deployment from preload
			if len(projects[i].Deployments) > 0 {
				projects[i].Deployments = []models.Deployment{projects[i].Deployments[0]}
			} else {
				projects[i].Deployments = []models.Deployment{} // Empty array instead of nil
			}
		}
	}

	c.JSON(http.StatusOK, projects)
}

package hostname

import (
	"crypto/rand"
	"deploy-platform/internal/config"
	"deploy-platform/internal/database"
	"deploy-platform/internal/models"
	"encoding/hex"
	"fmt"
	"strings"
)

type Manager struct {
	baseDomain string
	publicURL  string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		baseDomain: cfg.BaseDomain,
		publicURL:  cfg.PublicURL,
	}
}

// GenerateProjectHostname generates a persistent hostname for a project (Vercel-style)
// Format: project-slug.base-domain (no commit SHA - persistent per project)
func (m *Manager) GenerateProjectHostname(projectSlug string) string {
	// Create slug-safe hostname
	slug := strings.ToLower(strings.ReplaceAll(projectSlug, " ", "-"))
	slug = strings.ReplaceAll(slug, "_", "-")

	// Remove special characters
	slug = strings.ReplaceAll(slug, ".", "-")
	slug = strings.ReplaceAll(slug, "/", "-")

	// Format: project-slug.base-domain (persistent, like Vercel)
	hostname := fmt.Sprintf("%s.%s", slug, m.baseDomain)
	return hostname
}

// GetFullURL returns the full accessible URL for a hostname
func (m *Manager) GetFullURL(hostname string) string {
	if hostname == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", m.publicURL, hostname)
}

// AssignHostname assigns a persistent hostname to a project (Vercel-style)
// Reuses the same hostname for the project, updating it to point to the latest deployment
func (m *Manager) AssignHostname(projectID uint, deploymentID uint, commitSHA string) (string, error) {
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		return "", err
	}

	// Generate project slug
	projectSlug := project.Slug
	if projectSlug == "" {
		projectSlug = strings.ToLower(project.Name)
		if projectSlug == "" {
			// Use repo name as fallback
			projectSlug = strings.ToLower(project.RepoName)
			if projectSlug == "" {
				projectSlug = "deploy"
			}
		}
	}

	// Generate persistent hostname for project (no commit SHA)
	hostname := m.GenerateProjectHostname(projectSlug)

	// Check if project already has an active hostname
	var existingHostname models.Hostname
	result := database.DB.Where("project_id = ? AND is_active = ?", projectID, true).First(&existingHostname)

	if result.Error == nil {
		// Project already has a hostname - reuse it and update to point to new deployment
		// Mark old deployment's hostname as inactive
		database.DB.Model(&models.Hostname{}).
			Where("project_id = ? AND deployment_id != ? AND is_active = ?", projectID, deploymentID, true).
			Update("is_active", false)

		// Update existing hostname to point to new deployment
		existingHostname.DeploymentID = deploymentID
		existingHostname.IsActive = true
		database.DB.Save(&existingHostname)

		// Also update the deployment record
		database.DB.Model(&models.Deployment{}).Where("id = ?", deploymentID).Update("hostname", hostname)

		return hostname, nil
	}

	// New project - create hostname
	// Ensure uniqueness across all projects
	originalHostname := hostname
	counter := 0
	for {
		var check models.Hostname
		if database.DB.Where("hostname = ?", hostname).First(&check).Error != nil {
			break // Hostname is unique
		}
		// Add counter suffix if hostname exists (for different projects)
		counter++
		hostname = fmt.Sprintf("%s-%d.%s", strings.Split(originalHostname, ".")[0], counter, m.baseDomain)
	}

	// Mark any old hostnames for this project as inactive
	database.DB.Model(&models.Hostname{}).
		Where("project_id = ?", projectID).
		Update("is_active", false)

	// Create new hostname record
	hostnameRecord := &models.Hostname{
		Hostname:     hostname,
		ProjectID:    projectID,
		DeploymentID: deploymentID,
		IsActive:     true,
	}
	database.DB.Create(hostnameRecord)

	// Update deployment record with hostname
	database.DB.Model(&models.Deployment{}).Where("id = ?", deploymentID).Update("hostname", hostname)

	return hostname, nil
}

func generateShortHash() string {
	b := make([]byte, 3) // 6 hex characters
	rand.Read(b)
	return hex.EncodeToString(b)
}

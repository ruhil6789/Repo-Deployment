package models

// Data models will be defined here
// This will contain User, Project, Deployment, Build, Environment, and Hostname models

import (
	"time"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`                          // Primary key, auto-increments
	GitHubID    int64     `gorm:"column:github_id;uniqueIndex" json:"github_id"` // Unique GitHub user ID
	Username    string    `gorm:"uniqueIndex" json:"username"`                   // Unique GitHub username
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	GitHubToken string    `gorm:"column:github_token;type:text" json:"-"` // GitHub access token (hidden from JSON)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Projects []Project `gorm:"foreignKey:UserID" json:"projects,omitempty"` // One-to-many: User has many Projects
}

type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index" json:"user_id"` // Foreign key to User
	Name        string    `json:"name"`
	Slug        string    `gorm:"uniqueIndex" json:"slug"`    // Unique project slug
	RepoURL     string    `json:"repo_url"`                   // Repository URL
	RepoOwner   string    `json:"repo_owner"`                 // Repository owner
	RepoName    string    `json:"repo_name"`                  // Repository name
	Branch      string    `gorm:"default:main" json:"branch"` // Default branch
	GitHubToken string    `gorm:"type:text" json:"-"`         // Don't expose in JSON
	CreatedAt   time.Time `json:"created_at"`                 // Creation timestamp
	UpdatedAt   time.Time `json:"updated_at"`                 // Last update timestamp

	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`            // One-to-one: Project belongs to User
	Deployments  []Deployment  `gorm:"foreignKey:ProjectID" json:"deployments,omitempty"`  // One-to-many: Project has many Deployments
	Environments []Environment `gorm:"foreignKey:ProjectID" json:"environments,omitempty"` // One-to-many: Project has many Environments
}
type Deployment struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	ProjectID         uint      `gorm:"index" json:"project_id"`       // Foreign key to Project
	Status            string    `gorm:"default:pending" json:"status"` // pending, building, deploying, live, failed
	CommitSHA         string    `json:"commit_sha"`
	CommitMsg         string    `json:"commit_msg"`
	Branch            string    `json:"branch"`
	Hostname          string    `gorm:"uniqueIndex" json:"hostname"` // Unique hostname
	ImageTag          string    `json:"image_tag"`
	K8sNamespace      string    `json:"k8s_namespace"`
	K8sDeploymentName string    `json:"k8s_deployment_name"` // Kubernetes deployment name
	CreatedAt         time.Time `json:"created_at"`          // Creation timestamp
	UpdatedAt         time.Time `json:"updated_at"`          // Last update timestamp

	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Build   Build   `gorm:"foreignKey:DeploymentID" json:"build,omitempty"`
}

type Build struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	DeploymentID uint       `gorm:"index" json:"deployment_id"`    // Foreign key to Deployment
	Status       string     `gorm:"default:pending" json:"status"` // pending, building, success, failed
	Logs         string     `gorm:"type:text" json:"logs"`         // Build logs
	StartedAt    *time.Time `json:"started_at"`                    // Start time
	CompletedAt  *time.Time `json:"completed_at"`                  // Completion time
	CreatedAt    time.Time  `json:"created_at"`                    // Creation timestamp
	UpdatedAt    time.Time  `json:"updated_at"`                    // Last update timestamp
}

type Environment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"index" json:"project_id"` // Foreign key to Project
	Key       string    `json:"key"`
	Value     string    `gorm:"type:text" json:"value"` // In production, encrypt this!
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

type Hostname struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Hostname     string    `gorm:"uniqueIndex" json:"hostname"` // Unique hostname
	ProjectID    uint      `gorm:"index" json:"project_id"`
	DeploymentID uint      `gorm:"index" json:"deployment_id"`
	IsActive     bool      `gorm:"default:true" json:"is_active"` // Default: true
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Project    Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Deployment Deployment `gorm:"foreignKey:DeploymentID" json:"deployment,omitempty"`
}

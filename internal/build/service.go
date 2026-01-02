package build

// Build service will be implemented here
// This will handle Docker builds, build detection, and build orchestration

import (
	"archive/tar"
	"bytes"
	"context"
	"deploy-platform/internal/database"
	"deploy-platform/internal/hostname"
	"deploy-platform/internal/kubernetes"
	"deploy-platform/internal/models"
	"deploy-platform/pkg/docker"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type Service struct {
	dockerClient *docker.Client
	k8sClient    *kubernetes.Client
	hostnameMgr  *hostname.Manager
}

func NewService() (*Service, error) {
	dc, err := docker.NewClient()
	if err != nil {
		return nil, err
	}

	return &Service{dockerClient: dc}, nil
}

func NewServiceWithK8s(dockerClient *docker.Client, k8sClient *kubernetes.Client, hostnameMgr *hostname.Manager) *Service {
	return &Service{
		dockerClient: dockerClient,
		k8sClient:    k8sClient,
		hostnameMgr:  hostnameMgr,
	}
}

func (s *Service) BuildDeployment(ctx context.Context, deploymentID uint) error {
	var deployment models.Deployment
	if err := database.DB.Preload("Project").First(&deployment, deploymentID).Error; err != nil {
		return err
	}

	// Create build record
	build := &models.Build{
		DeploymentID: deploymentID,
		Status:       "building",
		StartedAt:    &[]time.Time{time.Now()}[0],
	}
	database.DB.Create(build)

	// Clone repository
	repoPath := fmt.Sprintf("/tmp/builds/%d", deploymentID)
	if err := s.cloneRepo(deployment.Project.RepoURL, repoPath, deployment.Branch); err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}

	// Detect build type and create Dockerfile if needed
	dockerfile, err := s.detectAndCreateDockerfile(repoPath)
	if err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}

	// Build Docker image
	imageTag := fmt.Sprintf("deploy-%d:%s", deploymentID, deployment.CommitSHA[:7])
	buildContext, err := s.createBuildContext(repoPath)
	if err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}

	if err := s.dockerClient.BuildImage(ctx, buildContext, imageTag, dockerfile); err != nil {
		s.updateBuildStatus(build.ID, "failed", err.Error())
		return err
	}

	// Update build and deployment
	completed := time.Now()
	build.CompletedAt = &completed
	build.Status = "success"
	database.DB.Save(build)

	deployment.Status = "deploying"
	deployment.ImageTag = imageTag
	database.DB.Save(deployment)

	// Deploy to Kubernetes if client is available
	if s.k8sClient != nil && s.hostnameMgr != nil {
		if err := s.deployToKubernetes(ctx, &deployment); err != nil {
			log.Printf("❌ Kubernetes deployment failed for deployment %d: %v", deploymentID, err)
			deployment.Status = "failed"
			database.DB.Save(deployment)
			return fmt.Errorf("kubernetes deployment failed: %w", err)
		}
		log.Printf("✅ Successfully deployed to Kubernetes: %s", deployment.Hostname)
		deployment.Status = "deployed"
		database.DB.Save(deployment)
	} else {
		log.Println("⚠️  Kubernetes client not available, skipping deployment")
	}

	return nil
}

func (s *Service) deployToKubernetes(ctx context.Context, deployment *models.Deployment) error {
	// Always assign/update hostname (Vercel-style: persistent per project)
	hostname, err := s.hostnameMgr.AssignHostname(deployment.ProjectID, deployment.ID, deployment.CommitSHA)
	if err != nil {
		return fmt.Errorf("failed to assign hostname: %w", err)
	}
	deployment.Hostname = hostname
	database.DB.Save(deployment)

	// Prepare environment variables (can be extended to load from project settings)
	envVars := map[string]string{
		"PORT": "8080",
	}

	// Update Kubernetes deployment (or create if doesn't exist)
	// This will update the existing deployment to point to the new image
	if err := s.k8sClient.CreateOrUpdateDeployment(ctx, deployment, hostname, envVars); err != nil {
		return fmt.Errorf("failed to create/update kubernetes resources: %w", err)
	}

	return nil
}

func (s *Service) cloneRepo(repoURL, path, branch string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Clone repository using go-git
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:           repoURL,
		SingleBranch:  true,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Progress:      os.Stdout, // Optional: show clone progress
	})

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return nil
}

func (s *Service) detectAndCreateDockerfile(repoPath string) (string, error) {
	// Check if Dockerfile exists
	if _, err := os.Stat(filepath.Join(repoPath, "Dockerfile")); err == nil {
		return "Dockerfile", nil
	}

	// Auto-generate Dockerfile based on detected language
	// This is simplified - you can expand this
	if _, err := os.Stat(filepath.Join(repoPath, "package.json")); err == nil {
		return s.createNodeDockerfile(repoPath)
	}

	if _, err := os.Stat(filepath.Join(repoPath, "requirements.txt")); err == nil {
		return s.createPythonDockerfile(repoPath)
	}

	if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err == nil {
		return s.createGoDockerfile(repoPath)
	}

	return "", fmt.Errorf("could not detect project type")
}

func (s *Service) createNodeDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]`

	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createPythonDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "app.py"]`

	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createGoDockerfile(repoPath string) (string, error) {
	dockerfile := `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]`

	path := filepath.Join(repoPath, "Dockerfile")
	return "Dockerfile", os.WriteFile(path, []byte(dockerfile), 0644)
}

func (s *Service) createBuildContext(repoPath string) (io.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(repoPath, path)
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()
			io.Copy(tw, data)
		}

		return nil
	})

	tw.Close()
	return &buf, err
}

func (s *Service) updateBuildStatus(buildID uint, status, logs string) {
	database.DB.Model(&models.Build{}).Where("id = ?", buildID).Updates(map[string]interface{}{
		"status": status,
		"logs":   logs,
	})
}

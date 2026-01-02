# Quick Start Checklist

Follow this checklist as you implement the platform. Check off each item as you complete it!

## Setup (Day 1)

- [ ] Install Go 1.21+ (`go version`)
- [ ] Install Docker (`docker --version`)
- [ ] Install Kubernetes/Minikube (`minikube start`)
- [ ] Create GitHub OAuth App (get Client ID and Secret)
- [ ] Clone or create a test repository for deployment

## Phase 1: Foundation

- [ ] Initialize Go module (`go mod init`)
- [ ] Create project directory structure
- [ ] Install dependencies (Gin, GORM, etc.)
- [ ] Create `cmd/api/main.go` with health endpoint
- [ ] Test: `curl http://localhost:8080/health` returns `{"status":"ok"}`

## Phase 2: Database

- [ ] Create `internal/models/models.go` with all models
- [ ] Create `internal/database/db.go` with connection logic
- [ ] Test: Run API and verify `deployments.db` is created
- [ ] Verify tables are created (use SQLite browser or `sqlite3 deployments.db .tables`)

## Phase 3: GitHub Integration

- [ ] Create `internal/config/config.go`
- [ ] Set environment variables (GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET)
- [ ] Create `internal/github/oauth.go`
- [ ] Test: Visit `http://localhost:8080/auth/github` and complete OAuth flow
- [ ] Create `internal/github/webhook.go`
- [ ] Setup webhook in GitHub repository
- [ ] Test: Push to repo and verify webhook is received

## Phase 4: Build Service

- [ ] Create `pkg/docker/client.go`
- [ ] Test Docker client connection
- [ ] Create `internal/build/service.go`
- [ ] Test: Manually trigger a build and verify Docker image is created
- [ ] Verify build logs are stored in database

## Phase 5: Kubernetes

- [ ] Setup Kubernetes cluster (Minikube or cloud)
- [ ] Create `internal/kubernetes/client.go`
- [ ] Test: Verify K8s client can connect to cluster
- [ ] Create `internal/hostname/manager.go`
- [ ] Test: Generate hostnames and verify uniqueness
- [ ] Create `internal/kubernetes/deployment.go`
- [ ] Test: Create a deployment manually and verify it appears in `kubectl get deployments`

## Phase 6: API Endpoints

- [ ] Create `internal/api/handlers.go`
- [ ] Add all routes to main.go
- [ ] Test: Create a project via API
- [ ] Test: List projects
- [ ] Test: Trigger deployment
- [ ] Test: View deployment status

## Phase 7: Integration

- [ ] Connect webhook → build → deploy flow
- [ ] Test: Push to GitHub and verify:
  - Webhook received
  - Build triggered
  - Image built
  - K8s deployment created
  - Hostname assigned
  - Application accessible via hostname

## Phase 8: Web UI (Optional)

- [ ] Create basic HTML dashboard
- [ ] Display projects list
- [ ] Display deployments
- [ ] Show build logs

## Testing Checklist

- [ ] Can create a project
- [ ] Can connect GitHub repository
- [ ] Webhook triggers on push
- [ ] Build completes successfully
- [ ] Docker image is created
- [ ] Kubernetes deployment is created
- [ ] Service is created
- [ ] Ingress is created
- [ ] Application is accessible via hostname
- [ ] Multiple projects can be deployed
- [ ] Multiple deployments per project work

## Common Issues & Solutions

**Issue**: Database connection fails
- **Solution**: Check file permissions, ensure directory exists

**Issue**: Docker build fails
- **Solution**: Ensure Docker daemon is running (`docker ps`)

**Issue**: Kubernetes connection fails
- **Solution**: Check `kubectl` works, verify kubeconfig path

**Issue**: GitHub OAuth fails
- **Solution**: Verify callback URL matches exactly, check Client ID/Secret

**Issue**: Webhook not received
- **Solution**: Use ngrok for local testing, verify webhook secret

## Next Steps After Basic Implementation

1. Add authentication middleware
2. Add error handling and logging
3. Add build log streaming
4. Add environment variable management
5. Add custom domains
6. Add preview deployments (branch-based)
7. Add rollback functionality
8. Add monitoring and metrics

## Learning Goals Achieved

By completing this project, you will have learned:

- [x] Go project structure and modules
- [x] REST API development with Gin
- [x] Database operations with GORM
- [x] OAuth authentication flow
- [x] Webhook handling
- [x] Docker API usage
- [x] Kubernetes client-go
- [x] Container orchestration
- [x] Full-stack development

---

**Remember**: It's okay to get stuck! Use the learning resources, experiment, and ask questions. The goal is to learn, not to build perfectly on the first try.

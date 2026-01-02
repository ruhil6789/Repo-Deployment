# Vercel-like Deployment Platform in Go

A self-hosted deployment platform that automatically builds and deploys GitHub repositories to Kubernetes with automatic hostname assignment. **Completely free with no limits!**

## 🎯 What You'll Build

- **GitHub Integration**: Connect repositories via OAuth
- **Automatic Builds**: Docker-based builds on every push
- **Kubernetes Deployment**: Automatic deployment to K8s cluster
- **Free Hostnames**: Automatic subdomain assignment (e.g., `myapp-abc123.yourdomain.com`)
- **Web Dashboard**: View projects, deployments, and logs
- **No Limits**: Deploy unlimited repositories for free

## 📚 Learning Path

This project is designed for **beginners learning Go**. Follow the implementation guide step by step:

1. **Start Here**: Read [IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)
2. **Track Progress**: Use [QUICK_START.md](./QUICK_START.md) as your checklist
3. **Build Phase by Phase**: Don't skip ahead - each phase builds on the previous

## 🚀 Quick Start

### Prerequisites

```bash
# Check installations
go version      # Should be 1.21+
docker --version
kubectl version
minikube version  # Optional, for local K8s
```

### Initial Setup

```bash
# Initialize the project
go mod init deploy-platform

# Install dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get github.com/google/go-github/v56/github
go get k8s.io/client-go@latest

# Start local Kubernetes (if using Minikube)
minikube start
```

### Run the API Server

```bash
# Set environment variables
export GITHUB_CLIENT_ID="your_client_id"
export GITHUB_CLIENT_SECRET="your_client_secret"

# Run the server
go run cmd/api/main.go
```

Visit: `http://localhost:8080/health`

## 📖 Documentation

- **[IMPLEMENTATION_GUIDE.md](./IMPLEMENTATION_GUIDE.md)** - Step-by-step implementation instructions
- **[QUICK_START.md](./QUICK_START.md)** - Progress checklist and testing guide

## 🏗️ Architecture

```
GitHub Push → Webhook → Build Service → Docker Image → Kubernetes → Live App
```

1. **GitHub Webhook** receives push events
2. **Build Service** clones repo and builds Docker image
3. **Kubernetes Service** creates Deployment, Service, and Ingress
4. **Hostname Manager** assigns unique subdomain
5. **Application** is live and accessible

## 🧩 Project Structure

```
deploy-platform/
├── cmd/
│   ├── api/          # API server
│   └── worker/       # Build worker
├── internal/
│   ├── api/          # HTTP handlers
│   ├── build/        # Build service
│   ├── kubernetes/   # K8s client
│   ├── hostname/     # Hostname management
│   ├── github/       # GitHub integration
│   ├── models/       # Data models
│   └── database/     # Database layer
├── web/ui/           # Web dashboard
└── k8s/manifests/    # K8s templates
```

## 🎓 Learning Goals

By building this project, you'll learn:

- ✅ Go fundamentals (structs, interfaces, goroutines)
- ✅ REST API development
- ✅ Database operations (GORM)
- ✅ OAuth authentication
- ✅ Docker API
- ✅ Kubernetes client-go
- ✅ Container orchestration
- ✅ Full-stack development

## 🔧 Development Workflow

1. **Read** the implementation guide section
2. **Understand** the concepts explained
3. **Implement** the code yourself
4. **Test** each component
5. **Research** the "Learning check" questions
6. **Move** to the next phase

## 🐛 Troubleshooting

### Database Issues
```bash
# Check if database file exists
ls -la deployments.db

# View database contents
sqlite3 deployments.db .tables
```

### Docker Issues
```bash
# Check Docker daemon
docker ps

# Test Docker build
docker build -t test-image .
```

### Kubernetes Issues
```bash
# Check cluster connection
kubectl cluster-info

# View resources
kubectl get deployments
kubectl get services
kubectl get ingress
```

### GitHub OAuth Issues
- Verify callback URL matches exactly
- Check Client ID and Secret are correct
- Ensure OAuth app has correct scopes

## 📝 Testing Checklist

After each phase, test:

- [ ] API server starts without errors
- [ ] Database connection works
- [ ] GitHub OAuth flow completes
- [ ] Webhook receives events
- [ ] Docker builds succeed
- [ ] Kubernetes deployments are created
- [ ] Applications are accessible

## 🚧 Current Status

This is a **learning project**. Build it step by step following the guide.

## 📚 Resources

- [Go Documentation](https://go.dev/doc/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Guide](https://gorm.io/docs/)
- [Kubernetes Concepts](https://kubernetes.io/docs/concepts/)
- [Docker API](https://docs.docker.com/engine/api/)

## 🤝 Contributing

This is a learning project. Focus on understanding each component before moving forward.

## 📄 License

This is an educational project. Use it to learn and build your own platform!

---

**Happy Coding! 🎉**

Remember: The goal is to **learn**, not to build perfectly. Take your time, experiment, and enjoy the journey!
Test deployment

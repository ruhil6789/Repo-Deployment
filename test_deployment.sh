#!/bin/bash

echo "🧪 Testing Vercel-like Deployment System"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if server is running
echo "1. Checking if server is running..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ Server is running${NC}"
else
    echo -e "${RED}❌ Server is not running${NC}"
    echo "   Start it with: go run cmd/api/main.go"
    exit 1
fi

# Check database
echo ""
echo "2. Checking database..."
if [ -f "deployments.db" ]; then
    echo -e "${GREEN}✅ Database exists${NC}"
    
    # Check projects
    PROJECT_COUNT=$(sqlite3 deployments.db "SELECT COUNT(*) FROM projects WHERE user_id = 2;" 2>/dev/null)
    echo "   Projects for user: $PROJECT_COUNT"
    
    # Check deployments
    DEPLOYMENT_COUNT=$(sqlite3 deployments.db "SELECT COUNT(*) FROM deployments WHERE project_id IN (SELECT id FROM projects WHERE user_id = 2);" 2>/dev/null)
    echo "   Deployments: $DEPLOYMENT_COUNT"
    
    # Check hostnames
    echo ""
    echo "3. Checking hostnames (should be persistent per project)..."
    sqlite3 deployments.db "SELECT p.name, d.hostname, d.status, d.commit_sha FROM deployments d JOIN projects p ON d.project_id = p.id WHERE p.user_id = 2 ORDER BY d.id DESC LIMIT 5;" 2>/dev/null | while IFS='|' read -r name hostname status commit; do
        if [ -n "$hostname" ]; then
            echo -e "   ${GREEN}✅${NC} $name: $hostname ($status)"
        else
            echo -e "   ${YELLOW}⚠️${NC}  $name: No hostname yet ($status)"
        fi
    done
else
    echo -e "${RED}❌ Database not found${NC}"
fi

# Check Docker
echo ""
echo "4. Checking Docker..."
if docker ps > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker is running${NC}"
else
    echo -e "${YELLOW}⚠️  Docker is not running (builds will fail)${NC}"
fi

# Check Kubernetes (optional)
echo ""
echo "5. Checking Kubernetes (optional)..."
if kubectl cluster-info > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Kubernetes is accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Kubernetes not configured (deployments will skip K8s)${NC}"
fi

echo ""
echo "========================================"
echo "📋 Next Steps:"
echo "1. Login at http://localhost:8080"
echo "2. Check dashboard for your project"
echo "3. Push code to GitHub: ruhil6789/Repo-Deployment"
echo "4. Watch server logs for deployment"
echo "5. Verify hostname is persistent (same URL)"
echo ""
echo "See TESTING_GUIDE.md for detailed instructions"

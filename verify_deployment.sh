#!/bin/bash

echo "🔍 Verifying Deployment System"
echo "=============================="
echo ""

# Check if webhook endpoint is accessible
echo "1. Testing webhook endpoint..."
if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/webhooks/github | grep -q "401\|400\|200"; then
    echo "   ✅ Webhook endpoint is accessible"
else
    echo "   ⚠️  Webhook endpoint might not be accessible"
fi

# Check project configuration
echo ""
echo "2. Checking project configuration..."
PROJECT=$(sqlite3 deployments.db "SELECT id, name, repo_owner, repo_name FROM projects WHERE user_id = 2 LIMIT 1;" 2>/dev/null)
if [ -n "$PROJECT" ]; then
    echo "   ✅ Project found: $PROJECT"
    echo ""
    echo "   📝 GitHub Webhook URL should be:"
    echo "      http://localhost:8080/webhooks/github"
    echo "      (or your ngrok URL if using ngrok)"
    echo ""
    echo "   📝 Configure in GitHub:"
    echo "      Settings → Webhooks → Add webhook"
    echo "      URL: http://localhost:8080/webhooks/github"
    echo "      Content type: application/json"
    echo "      Events: Just the push event"
    echo "      Secret: (from .env WEBHOOK_SECRET)"
else
    echo "   ⚠️  No project found for user"
fi

# Check latest deployment
echo ""
echo "3. Latest deployment status:"
LATEST=$(sqlite3 deployments.db "SELECT d.id, d.status, d.hostname, d.commit_sha, d.created_at FROM deployments d JOIN projects p ON d.project_id = p.id WHERE p.user_id = 2 ORDER BY d.id DESC LIMIT 1;" 2>/dev/null)
if [ -n "$LATEST" ]; then
    echo "$LATEST" | awk -F'|' '{print "   ID: " $1 "\n   Status: " $2 "\n   Hostname: " $3 "\n   Commit: " substr($4,1,7) "\n   Created: " $5}'
else
    echo "   No deployments yet"
fi

echo ""
echo "=============================="
echo "✅ System ready for testing!"

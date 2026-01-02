# Quick Test Guide - Step by Step

## 🚀 Test the Vercel-like Deployment System

### Step 1: Verify Server is Running
```bash
# Check if server is running
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

### Step 2: Check Dashboard
1. Open browser: http://localhost:8080/dashboard
2. Login if needed (GitHub OAuth or Email/Password)
3. You should see:
   - Your project: "Test Project"
   - Current deployments
   - Hostname (if any deployment is deployed)

### Step 3: Configure GitHub Webhook (if not done)
1. Go to: https://github.com/ruhil6789/Repo-Deployment/settings/hooks
2. Click "Add webhook"
3. Set:
   - **Payload URL**: `http://localhost:8080/webhooks/github` (or your ngrok URL)
   - **Content type**: `application/json`
   - **Secret**: (from your `.env` file, `WEBHOOK_SECRET`)
   - **Events**: Select "Just the push event"
4. Click "Add webhook"

### Step 4: Push Code to GitHub
```bash
# In your repository directory
cd /path/to/Repo-Deployment

# Make a small change
echo "# Test deployment $(date)" >> README.md

# Commit and push
git add .
git commit -m "Test automatic deployment - $(date +%H:%M:%S)"
git push origin main
```

### Step 5: Watch Server Logs
Watch your server terminal. You should see:
```
✅ Deployment X enqueued for build
Worker 1: Processing deployment X
✅ Build completed successfully for deployment X
✅ Successfully deployed to Kubernetes: test-project.localhost
```

### Step 6: Check Dashboard
1. Refresh dashboard: http://localhost:8080/dashboard
2. New deployment should appear in "Recent Deployments"
3. Status should change: pending → building → deployed
4. **Hostname should be persistent** (same as previous, or new if first time)

### Step 7: Visit Hostname
1. Click the "Visit" button or copy the hostname
2. Open: `http://test-project.localhost` (or your hostname)
3. Should show your deployed application

### Step 8: Push Again (Test Persistence)
```bash
# Make another change
echo "# Second test $(date)" >> README.md
git add .
git commit -m "Second test - $(date +%H:%M:%S)"
git push origin main
```

**Expected Result:**
- ✅ New deployment created
- ✅ **SAME hostname** (not a new one!)
- ✅ Hostname automatically shows new code
- ✅ Dashboard shows both deployments with same hostname

### Step 9: Verify in Database
```bash
# Check all deployments for your project
sqlite3 deployments.db "
SELECT 
    d.id, 
    d.status, 
    d.hostname, 
    substr(d.commit_sha, 1, 7) as commit,
    d.created_at
FROM deployments d 
JOIN projects p ON d.project_id = p.id 
WHERE p.user_id = 2 
ORDER BY d.id DESC 
LIMIT 5;
"

# All deployments should have THE SAME hostname!
```

## ✅ Success Criteria

Your system works like Vercel when:
1. ✅ Push code → Deployment happens automatically
2. ✅ Same hostname for all deployments of same project
3. ✅ Hostname updates to show new code automatically
4. ✅ No manual steps needed
5. ✅ Dashboard shows real-time updates

## 🐛 Troubleshooting

### Webhook not triggering?
- Check GitHub webhook delivery page
- Verify webhook URL is correct
- Check server logs for errors
- Verify webhook secret matches

### Deployment stuck in "pending"?
- Check if build queue is working
- Verify Docker is running
- Check worker pool logs

### Hostname not persistent?
- Check database: `SELECT hostname FROM deployments WHERE project_id = 1;`
- All should be the same (except old ones)
- New deployments should reuse hostname

### Can't access hostname?
- If using `localhost`, make sure you're accessing from same machine
- If using Kubernetes, check ingress configuration
- Verify DNS/hosts file if needed

## 📊 Monitoring

Watch these in real-time:
```bash
# Server logs
tail -f server.log  # or watch terminal

# Database changes
watch -n 2 'sqlite3 deployments.db "SELECT id, status, hostname FROM deployments ORDER BY id DESC LIMIT 5;"'

# Docker images
watch -n 2 'docker images | grep deploy'
```

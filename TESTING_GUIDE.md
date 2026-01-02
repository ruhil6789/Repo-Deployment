# Testing Guide - Vercel-like Deployment System

## Quick Test Steps

### 1. Start the Server
```bash
cd /data/Go/ride-booking
go run cmd/api/main.go
```

### 2. Verify Server is Running
- Open: http://localhost:8080
- You should see the login page

### 3. Login
- Use GitHub OAuth or Email/Password
- You'll be redirected to dashboard

### 4. Check Your Project
- Dashboard should show your project: "Test Project"
- It should have a persistent hostname (e.g., `test-project.localhost`)

### 5. Test Automatic Deployment

#### Option A: Push to GitHub (Real Test)
1. Make a change to your repository: `ruhil6789/Repo-Deployment`
2. Commit and push:
   ```bash
   git add .
   git commit -m "Test automatic deployment"
   git push origin main
   ```
3. Check server logs - you should see:
   - `✅ Deployment X enqueued for build`
   - `✅ Build completed successfully`
   - `✅ Successfully deployed to Kubernetes`
4. Check dashboard - new deployment should appear
5. Visit the hostname - should show updated code

#### Option B: Manual Webhook Test (Without Push)
```bash
# Get your ngrok URL (if using ngrok)
# Or use localhost if webhook is configured locally

# Send test webhook (replace with your values)
curl -X POST http://localhost:8080/webhooks/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{
    "ref": "refs/heads/main",
    "repository": {
      "owner": {"login": "ruhil6789"},
      "name": "Repo-Deployment"
    },
    "head_commit": {
      "id": "test123456789",
      "message": "Test deployment"
    }
  }'
```

### 6. Verify Persistent Hostname

1. **First Deployment:**
   - Check dashboard - note the hostname (e.g., `test-project.localhost`)
   - Visit the URL

2. **Second Deployment (after push):**
   - Push new code
   - Check dashboard - hostname should be THE SAME
   - Visit the URL - should show NEW code

3. **Verify in Database:**
   ```bash
   sqlite3 deployments.db "SELECT id, project_id, hostname, status FROM deployments ORDER BY id DESC LIMIT 5;"
   ```
   - All deployments for same project should have same hostname

### 7. Check Kubernetes (if configured)

```bash
# List deployments
kubectl get deployments

# Check specific project deployment
kubectl get deployment project-1 -o yaml

# Check ingress
kubectl get ingress

# Check pods
kubectl get pods
```

## Expected Behavior (Vercel-like)

✅ **One hostname per project** - Never changes
✅ **Automatic updates** - Push code, see changes immediately
✅ **No manual steps** - Everything happens automatically
✅ **Same URL forever** - Like Vercel's preview URLs

## Troubleshooting

### Projects not showing?
- Check browser console (F12)
- Verify you're logged in: `localStorage.getItem('token')`
- Check API: `fetch('/api/projects', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('token')}}).then(r => r.json()).then(console.log)`

### Deployments not triggering?
- Check webhook URL in GitHub settings
- Verify webhook secret matches `.env` file
- Check server logs for webhook errors

### Hostname not updating?
- Check database: `SELECT hostname FROM deployments WHERE project_id = 1 ORDER BY id DESC;`
- Verify hostname manager is assigning correctly
- Check Kubernetes ingress is updated

### Build failing?
- Check Docker is running: `docker ps`
- Check build logs in database
- Verify repository is accessible

## Test Checklist

- [ ] Server starts without errors
- [ ] Can login via GitHub/Email
- [ ] Projects appear on dashboard
- [ ] Hostname is shown for project
- [ ] Push to GitHub triggers webhook
- [ ] New deployment is created
- [ ] Build completes successfully
- [ ] Hostname remains the same
- [ ] New code is visible at hostname
- [ ] Dashboard shows updated deployment

## Success Criteria

🎉 **System works like Vercel when:**
1. You push code → Deployment happens automatically
2. Same hostname shows new code
3. No manual intervention needed
4. Dashboard updates in real-time

# Webhook Testing Guide

## ✅ Current Status

- **Webhook endpoint**: `/webhooks/github` - Working
- **Database project**: Updated to `ruhil6789/Repo-Deployment`
- **Hostname generation**: Fixed (needs server restart)

## 🧪 How to Test Webhook

### Option 1: Local Test (No ngrok needed)

```bash
# 1. Make sure server is running with latest code
cd /data/Go/ride-booking
go run cmd/api/main.go

# 2. In another terminal, run test script
./test_webhook_local.sh
```

### Option 2: Real GitHub Webhook (Requires ngrok)

#### Step 1: Set up ngrok

1. Sign up: https://dashboard.ngrok.com/signup
2. Get authtoken: https://dashboard.ngrok.com/get-started/your-authtoken
3. Install authtoken:
   ```bash
   ngrok config add-authtoken YOUR_AUTHTOKEN_HERE
   ```

#### Step 2: Start ngrok

```bash
# In one terminal
ngrok http 8080

# Copy the HTTPS URL (e.g., https://abc123.ngrok.io)
```

#### Step 3: Configure GitHub Webhook

1. Go to: https://github.com/ruhil6789/Repo-Deployment/settings/hooks
2. Click "Add webhook"
3. Fill in:
   - **Payload URL**: `https://YOUR-NGROK-URL.ngrok.io/webhooks/github`
   - **Content type**: `application/json`
   - **Secret**: `nncfebvjhebhjvrevjejrvhjelv`
   - **Events**: Just the push event
4. Click "Add webhook"

#### Step 4: Test with Real Push

**Option A: Use GitHub Web Interface**
1. Go to: https://github.com/ruhil6789/Repo-Deployment
2. Click "Add file" → "Create new file"
3. Name: `test.txt`
4. Content: `Test webhook`
5. Commit message: `Test webhook trigger`
6. Click "Commit new file"

**Option B: Use Git (if hooks are fixed)**
```bash
cd /tmp/Repo-Deployment
echo "Test" >> README.md
git add README.md
git commit -m "Test webhook"
git push origin main
```

## 🔍 Verify Webhook Worked

1. **Check server logs** - Should show webhook received
2. **Check database**:
   ```bash
   sqlite3 deployments.db "SELECT id, status, commit_sha, hostname, branch FROM deployments ORDER BY id DESC LIMIT 1;"
   ```
3. **Check GitHub**:
   - Go to webhook settings
   - Click your webhook
   - Check "Recent Deliveries" - should show 200 status

## 📝 Test Scripts Available

- `test_webhook_local.sh` - Test locally without ngrok
- `test_webhook_real.sh` - Test with real repo data
- `test_webhook_with_sig.sh` - Test with signature

## ⚠️ Important Notes

1. **Restart server** after code changes
2. **ngrok URL changes** each time you restart (unless paid plan)
3. **Webhook secret** must match in GitHub and code
4. **Project must exist** in database before webhook works

## 🐛 Troubleshooting

- **"Invalid signature"**: Check webhook secret matches
- **"Project not found"**: Create project in database first
- **"UNIQUE constraint failed"**: Clear old deployments or restart server
- **Webhook not received**: Check ngrok is running and URL is correct

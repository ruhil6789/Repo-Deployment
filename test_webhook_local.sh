#!/bin/bash

echo "🧪 Testing Webhook Locally (Simulating GitHub Push)"
echo "===================================================="
echo ""

# Make sure server is running first!
echo "⚠️  Make sure your server is running: go run cmd/api/main.go"
echo ""

SECRET="nncfebvjhebhjvrevjejrvhjelv"

PAYLOAD='{
  "ref": "refs/heads/main",
  "head_commit": {
    "id": "'$(openssl rand -hex 7)'",
    "message": "Test webhook - local test",
    "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
  },
  "repository": {
    "name": "Repo-Deployment",
    "owner": {
      "login": "ruhil6789"
    }
  }
}'

SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
SIGNATURE_HEADER="sha256=$SIGNATURE"

echo "Sending webhook to: http://localhost:8080/webhooks/github"
echo ""

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST http://localhost:8080/webhooks/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -H "X-Hub-Signature-256: $SIGNATURE_HEADER" \
  -d "$PAYLOAD")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE/d')

echo "Response:"
echo "$BODY" | python3 -m json.tool 2>/dev/null || echo "$BODY"
echo ""
echo "HTTP Status: $HTTP_CODE"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Webhook test successful!"
    echo ""
    echo "Check database:"
    sqlite3 deployments.db "SELECT id, status, commit_sha, commit_msg, branch, created_at FROM deployments ORDER BY id DESC LIMIT 1;"
else
    echo "❌ Webhook test failed"
fi

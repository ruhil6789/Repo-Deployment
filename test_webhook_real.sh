#!/bin/bash

# Test webhook with real repository data
SECRET="nncfebvjhebhjvrevjejrvhjelv"

PAYLOAD='{
  "ref": "refs/heads/main",
  "head_commit": {
    "id": "test123456789",
    "message": "Test webhook from real repo",
    "timestamp": "2024-01-02T12:00:00Z"
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

echo "🧪 Testing Webhook for Repo-Deployment"
echo "========================================"
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
    echo "✅ Webhook working! Check database:"
    echo "   sqlite3 deployments.db \"SELECT id, status, commit_sha, branch FROM deployments ORDER BY id DESC LIMIT 1;\""
else
    echo "❌ Webhook test failed"
fi

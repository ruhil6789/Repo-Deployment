#!/bin/bash

# Webhook secret (must match your code)
SECRET="nncfebvjhebhjvrevjejrvhjelv"

# Test payload
PAYLOAD='{
  "ref": "refs/heads/main",
  "head_commit": {
    "id": "abc123def456",
    "message": "Test commit message",
    "timestamp": "2024-01-02T12:00:00Z"
  },
  "repository": {
    "name": "test-repo",
    "owner": {
      "login": "ruhil6789"
    }
  }
}'

# Generate HMAC signature
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
SIGNATURE_HEADER="sha256=$SIGNATURE"

echo "🔍 Testing GitHub Webhook"
echo "=========================="
echo "Payload: $PAYLOAD"
echo "Signature: $SIGNATURE_HEADER"
echo ""

# Test the webhook
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
else
    echo "❌ Webhook test failed"
fi

#!/bin/bash

# Test webhook payload
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

echo "Testing webhook endpoint..."
echo "Payload: $PAYLOAD"
echo ""

# Test without signature (will fail signature check but we can see the flow)
curl -X POST http://localhost:8080/webhooks/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -d "$PAYLOAD" \
  -v

echo ""
echo "Test complete!"

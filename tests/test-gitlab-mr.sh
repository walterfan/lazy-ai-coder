#!/bin/bash
# Test the get_gitlab_merge_request MCP tool
# Usage: ./test-gitlab-mr.sh [project] [mr_id]

PROJECT="${1:-internal-common-sdk/python-sdk-3.0}"
MR_ID="${2:-42}"

echo "Testing get_gitlab_merge_request tool"
echo "Project: $PROJECT"
echo "MR ID: $MR_ID"
echo ""

curl -X POST http://localhost:8888/api/v1/mcp \
  -H "Content-Type: application/json" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 1,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"get_gitlab_merge_request\",
      \"arguments\": {
        \"project\": \"$PROJECT\",
        \"merge_request_id\": \"$MR_ID\"
      }
    }
  }" | jq

echo ""
echo "Usage: $0 <project-path> <mr-id>"
echo "Example: $0 myorg/myrepo 42"

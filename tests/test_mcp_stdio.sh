#!/bin/bash
# Test script for MCP server in stdio mode

set -e

echo "Testing MCP Server in stdio transport mode..."
echo "============================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to send JSON-RPC request and get response
test_mcp_request() {
    local name=$1
    local request=$2

    echo -e "${BLUE}Test: $name${NC}"
    echo "Request:"
    echo "$request" | jq '.'
    echo ""

    echo "Response:"
    echo "$request" | ./lazy-ai-coder mcp 2>/dev/null | head -1 | jq '.'
    echo ""
    echo "---"
    echo ""
}

# Check if binary exists
if [ ! -f "./lazy-ai-coder" ]; then
    echo -e "${YELLOW}Building lazy-ai-coder...${NC}"
    go build -o lazy-ai-coder
fi

echo -e "${GREEN}1. Testing Initialize${NC}"
test_mcp_request "Initialize" '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    }
  }
}'

echo -e "${GREEN}2. Testing List Tools${NC}"
test_mcp_request "List Tools" '{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}'

echo -e "${GREEN}3. Testing List Resources${NC}"
test_mcp_request "List Resources" '{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "resources/list",
  "params": {}
}'

echo -e "${GREEN}4. Testing List Prompts${NC}"
test_mcp_request "List Prompts" '{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "prompts/list",
  "params": {}
}'

echo ""
echo -e "${GREEN}✓ MCP Server stdio transport test completed!${NC}"
echo ""
echo "Summary:"
echo "--------"
echo "The MCP server provides:"
echo "  • Tools: 6 MCP tools (GitLab, code review, prompt building, etc.)"
echo "  • Resources: Config files and GitLab projects"
echo "  • Prompts: 117 prompts from database"
echo ""
echo "To use with Claude Desktop or Cursor, add to your MCP config:"
echo "  {\"command\": \"$(pwd)/lazy-ai-coder\", \"args\": [\"mcp\"], \"env\": {...}}"

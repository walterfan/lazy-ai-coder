#!/bin/bash

# Manual testing examples for MCP HTTP server
# These examples show how to test the MCP endpoints using curl

SERVER_URL="http://localhost:8888"

echo "=========================================="
echo "MCP HTTP Server - Manual Test Examples"
echo "=========================================="
echo ""

# Check if server is running
echo "1. Check server info"
echo "   GET /api/v1/mcp/info"
echo ""
curl -s "${SERVER_URL}/api/v1/mcp/info" | jq .
echo ""
echo ""

# List tools (simple)
echo "2. List tools (simple GET)"
echo "   GET /api/v1/mcp/tools"
echo ""
curl -s "${SERVER_URL}/api/v1/mcp/tools" | jq .
echo ""
echo ""

# Initialize
echo "3. Initialize (JSON-RPC)"
echo "   POST /api/v1/mcp"
echo ""
curl -s -X POST "${SERVER_URL}/api/v1/mcp" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "curl-client",
        "version": "1.0.0"
      }
    }
  }' | jq .
echo ""
echo ""

# List tools (JSON-RPC)
echo "4. List tools (JSON-RPC)"
echo "   POST /api/v1/mcp"
echo ""
curl -s -X POST "${SERVER_URL}/api/v1/mcp" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }' | jq '.result.tools[] | {name, description}'
echo ""
echo ""

# Call LLM chat tool
echo "5. Call LLM chat tool"
echo "   POST /api/v1/mcp (tools/call)"
echo ""
curl -s -X POST "${SERVER_URL}/api/v1/mcp" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "llm_chat",
      "arguments": {
        "user_prompt": "Say hello in one word",
        "system_prompt": "You are a helpful assistant."
      }
    }
  }' | jq .
echo ""
echo ""

# Test error handling
echo "6. Test error handling (invalid method)"
echo "   POST /api/v1/mcp"
echo ""
curl -s -X POST "${SERVER_URL}/api/v1/mcp" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "invalid_method",
    "params": {}
  }' | jq .
echo ""
echo ""

echo "=========================================="
echo "Manual tests complete!"
echo "=========================================="
echo ""
echo "For automated testing, use pytest:"
echo "  pytest tests/test_mcp_http.py -v"
echo ""
echo "Or use the test runner:"
echo "  ./run-tests.sh"
echo ""


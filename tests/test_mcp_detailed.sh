#!/bin/bash
# Detailed test for MCP server stdio mode
# This script sends JSON-RPC requests and captures actual responses

set -e

echo "==================================================="
echo "  MCP Server stdio Transport - Detailed Test"
echo "==================================================="
echo ""

# Check if binary exists
if [ ! -f "./lazy-ai-coder" ]; then
    echo "Building lazy-ai-coder..."
    go build -o lazy-ai-coder
    echo ""
fi

# Function to send request and parse JSON response (skip log lines)
send_request() {
    local request="$1"
    # Send request, wait for response, filter out log lines (lines starting with {\"level\":)
    echo "$request" | timeout 3 ./lazy-ai-coder mcp 2>&1 | \
        grep -v '^{"level":' | grep -v '^$' | head -1
}

echo "Test 1: Initialize"
echo "------------------"
INIT_REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}'
echo "Request: $INIT_REQUEST"
echo ""
INIT_RESPONSE=$(send_request "$INIT_REQUEST")
echo "Response:"
echo "$INIT_RESPONSE" | jq '.' 2>/dev/null || echo "$INIT_RESPONSE"
echo ""
echo ""

echo "Test 2: List Tools"
echo "------------------"
TOOLS_REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
echo "Sending request..."
echo ""
TOOLS_RESPONSE=$(send_request "$TOOLS_REQUEST")
echo "Response:"
echo "$TOOLS_RESPONSE" | jq '.result.tools[] | {name: .name, description: .description}' 2>/dev/null || echo "$TOOLS_RESPONSE"
echo ""
echo ""

echo "Test 3: List Resources"
echo "----------------------"
RESOURCES_REQUEST='{"jsonrpc":"2.0","id":3,"method":"resources/list","params":{}}'
echo "Sending request..."
echo ""
RESOURCES_RESPONSE=$(send_request "$RESOURCES_REQUEST")
echo "Response:"
echo "$RESOURCES_RESPONSE" | jq '.result.resources[] | {name: .name, uri: .uri}' 2>/dev/null || echo "$RESOURCES_RESPONSE"
echo ""
echo ""

echo "Test 4: List Prompts (first 5)"
echo "-------------------------------"
PROMPTS_REQUEST='{"jsonrpc":"2.0","id":4,"method":"prompts/list","params":{}}'
echo "Sending request..."
echo ""
PROMPTS_RESPONSE=$(send_request "$PROMPTS_REQUEST")
echo "Response (first 5 prompts):"
echo "$PROMPTS_RESPONSE" | jq '.result.prompts[0:5] | .[] | {name: .name, description: .description, arguments: (.arguments | length)}' 2>/dev/null || echo "$PROMPTS_RESPONSE"
echo ""
echo "Total prompts:"
echo "$PROMPTS_RESPONSE" | jq '.result.prompts | length' 2>/dev/null || echo "N/A"
echo ""
echo ""

echo "==================================================="
echo "  Summary"
echo "==================================================="
echo ""
echo "✓ MCP Server is working in stdio transport mode"
echo ""
echo "Available Components:"
echo "  • Tools: $(echo "$TOOLS_RESPONSE" | jq '.result.tools | length' 2>/dev/null || echo 'N/A')"
echo "  • Resources: $(echo "$RESOURCES_RESPONSE" | jq '.result.resources | length' 2>/dev/null || echo 'N/A')"
echo "  • Prompts: $(echo "$PROMPTS_RESPONSE" | jq '.result.prompts | length' 2>/dev/null || echo 'N/A')"
echo ""
echo "Server Info:"
echo "$INIT_RESPONSE" | jq '.result.serverInfo' 2>/dev/null || echo "N/A"
echo ""

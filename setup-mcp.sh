#!/bin/bash

# Setup script for lazy-ai-coder MCP server
# This script helps you configure the MCP server for use with Claude Desktop

set -e

echo "=========================================="
echo "Lazy AI Coder - MCP Server Setup"
echo "=========================================="
echo ""

# Check if running on macOS or Linux
if [[ "$OSTYPE" == "darwin"* ]]; then
    CONFIG_DIR="$HOME/Library/Application Support/Claude"
    CONFIG_FILE="$CONFIG_DIR/claude_desktop_config.json"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    CONFIG_DIR="$HOME/.config/Claude"
    CONFIG_FILE="$CONFIG_DIR/claude_desktop_config.json"
else
    echo "❌ Unsupported operating system: $OSTYPE"
    echo "This script supports macOS and Linux only."
    echo "For Windows, manually configure: %APPDATA%\Claude\claude_desktop_config.json"
    exit 1
fi

echo "Detected OS: $OSTYPE"
echo "Config directory: $CONFIG_DIR"
echo "Config file: $CONFIG_FILE"
echo ""

# Build the binary
echo "📦 Building lazy-ai-coder..."
go build -o lazy-ai-coder
BINARY_PATH="$(pwd)/lazy-ai-coder"
echo "✅ Built: $BINARY_PATH"
echo ""

# Get configuration from user
echo "Please provide the following information:"
echo ""

read -p "GitLab Base URL (e.g., https://gitlab.com): " GITLAB_URL
read -p "GitLab Token: " GITLAB_TOKEN
read -p "LLM Base URL (e.g., https://api.openai.com/v1): " LLM_URL
read -p "LLM API Key: " LLM_KEY
read -p "LLM Model (e.g., gpt-4): " LLM_MODEL
read -p "PlantUML URL (optional, press Enter to skip): " PLANTUML_URL

echo ""
echo "Creating MCP configuration..."

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Create the configuration JSON
cat > "$CONFIG_FILE" << EOF
{
  "mcpServers": {
    "lazy-ai-coder": {
      "command": "$BINARY_PATH",
      "args": ["mcp"],
      "env": {
        "GITLAB_BASE_URL": "$GITLAB_URL",
        "GITLAB_TOKEN": "$GITLAB_TOKEN",
        "LLM_BASE_URL": "$LLM_URL",
        "LLM_API_KEY": "$LLM_KEY",
        "LLM_MODEL": "$LLM_MODEL"
EOF

if [ -n "$PLANTUML_URL" ]; then
cat >> "$CONFIG_FILE" << EOF
,
        "PLANTUML_URL": "$PLANTUML_URL"
EOF
fi

cat >> "$CONFIG_FILE" << EOF

      }
    }
  }
}
EOF

echo "✅ Configuration saved to: $CONFIG_FILE"
echo ""
echo "=========================================="
echo "✨ Setup Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Restart Claude Desktop if it's running"
echo "2. Look for the 🔌 icon in Claude Desktop"
echo "3. You should see 'lazy-ai-coder' with 6 tools available"
echo ""
echo "Available MCP Tools:"
echo "  • get_gitlab_file_content"
echo "  • get_gitlab_merge_request"
echo "  • get_gitlab_project_id"
echo "  • llm_chat"
echo "  • llm_analyze_code"
echo "  • generate_plantuml"
echo ""
echo "Example prompts to try:"
echo "  • 'Show me the content of README.md from myorg/myrepo'"
echo "  • 'Review merge request #123 in myorg/myrepo'"
echo "  • 'Analyze this code for security issues: [paste code]'"
echo ""
echo "For more information, see MCP_README.md"
echo ""


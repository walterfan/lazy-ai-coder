# Lazy AI Coder v0.1

- [Lazy AI Coder v0.1](#lazy-ai-coder-v01)
  - [🆕 MCP Server Mode](#-mcp-server-mode)
    - [Two MCP Modes](#two-mcp-modes)
      - [Mode 1: Stdio MCP Server (for Cursor/Claude Desktop)](#mode-1-stdio-mcp-server-for-cursorclaude-desktop)
      - [Mode 2: HTTP MCP Server (Built into Web Server)](#mode-2-http-mcp-server-built-into-web-server)
    - [Available MCP Tools](#available-mcp-tools)
    - [Integration Examples](#integration-examples)
      - [Option 1: Cursor (AI Code Editor)](#option-1-cursor-ai-code-editor)
      - [Option 2: Claude Desktop](#option-2-claude-desktop)
    - [📖 Detailed Setup Guides](#-detailed-setup-guides)
    - [📚 API Documentation (Swagger/OpenAPI)](#-api-documentation-swaggeropenapi)
    - [🧪 Testing MCP Functionality](#-testing-mcp-functionality)
  - [Usage](#usage)
    - [Use case](#use-case)
      - [summarize code and draw uml and mindmap](#summarize-code-and-draw-uml-and-mindmap)
      - [help review merge request](#help-review-merge-request)
      - [more cases](#more-cases)
  - [Setup](#setup)
    - [Step 0: Import Prompts (First Time Setup)](#step-0-import-prompts-first-time-setup)
    - [Option 1: Web-based Settings (Recommended)](#option-1-web-based-settings-recommended)
    - [Option 2: Environment Variables](#option-2-environment-variables)


The Lazy AI Coder is a tool designed to help developers understand, write, and review code, etc.

It supports features such as code summarization, UML/mindmap generation, and merge request reviews using LLM APIs, GitLab APIs, and PlantUML.

**🎉 NEW: This tool now supports MCP (Model Context Protocol) server mode!** You can use it as an MCP server with Claude Desktop, VSCode, or any MCP-compatible client. See [MCP_README.md](MCP_README.md) for details.

![snapshot](doc/snapshot.png)

## 🆕 MCP Server Mode

This tool can now run as an **MCP (Model Context Protocol) server**, enabling AI assistants like Claude Desktop and Cursor to access your GitLab repositories and LLM capabilities directly!

### Two MCP Modes

#### Mode 1: Stdio MCP Server (for Cursor/Claude Desktop)

```bash
# Build the project
go build -o lazy-ai-coder

# Run as stdio MCP server
./lazy-ai-coder mcp
```

**Use Case**: Direct integration with Cursor IDE and Claude Desktop via stdio transport.

#### Mode 2: HTTP MCP Server (Built into Web Server)

```bash
# Run web server with MCP endpoints enabled
./lazy-ai-coder web -p 8888
```

**Use Case**: Web-based access + MCP functionality over HTTP. Perfect for:
- Testing MCP tools with pytest
- Custom integrations via HTTP API
- Accessing both web UI and MCP tools
- Remote MCP server access

**MCP Endpoints**:
- `POST /api/v1/mcp` - Main JSON-RPC endpoint
- `GET /api/v1/mcp/info` - Server information
- `GET /api/v1/mcp/tools` - List available tools

### Available MCP Tools

- **GitLab Integration**: Read files, review merge requests, search projects
- **LLM Chat**: Direct LLM access with customizable prompts
- **Code Analysis**: Automated code review, security analysis, bug detection
- **Diagram Generation**: Create PlantUML diagrams and mind maps

### Integration Examples

#### Option 1: Cursor (AI Code Editor)

Cursor has native MCP support built-in. To integrate:

1. **Open Cursor Settings**
   - Press `Cmd/Ctrl + Shift + P`
   - Type "Cursor Settings"
   - Or go to `Cursor` → `Settings` → `Cursor Settings`

2. **Navigate to MCP Settings**
   - In settings, search for "MCP" or "Model Context Protocol"
   - Or go to the "Features" section

3. **Add Server Configuration**
   - Click "Edit Config" or "Add MCP Server"
   - Add the following configuration:

**For Cursor Settings UI:**
- **Server Name**: `lazy-ai-coder`
- **Command**: `/full/path/to/lazy-ai-coder` (use absolute path!)
- **Arguments**: `mcp`
- **Environment Variables**:
  ```
  GITLAB_BASE_URL=https://gitlab.com
  GITLAB_TOKEN=glpat-your-token
  LLM_BASE_URL=https://api.openai.com/v1
  LLM_API_KEY=sk-your-api-key
  LLM_MODEL=gpt-4
  PLANTUML_URL=http://www.plantuml.com/plantuml
  ```

**Or Edit Config File Directly:**

Cursor stores MCP configuration in `~/.cursor/mcp_config.json` (macOS/Linux) or `%APPDATA%\Cursor\mcp_config.json` (Windows):

```json
{
  "mcpServers": {
    "lazy-ai-coder": {
      "command": "~/lazy-ai-coder/lazy-ai-coder",
      "args": ["mcp"],
      "env": {
        "GITLAB_BASE_URL": "https://gitlab.com",
        "GITLAB_TOKEN": "glpat-your-token-here",
        "LLM_BASE_URL": "https://api.openai.com/v1",
        "LLM_API_KEY": "sk-your-api-key-here",
        "LLM_MODEL": "gpt-4",
        "PLANTUML_URL": "http://www.plantuml.com/plantuml"
      }
    }
  }
}
```

4. **Restart Cursor**
   - Completely quit and restart Cursor
   - The MCP server will start automatically

5. **Use the Tools**
   - Open Cursor's AI chat (Cmd/Ctrl + L)
   - Ask questions like:
     ```
     "Show me the README.md from myorg/myrepo"
     "Review merge request #123 in myorg/myrepo"
     "Analyze this file for security issues"
     ```
   - Cursor will automatically use the MCP tools when relevant!

**Tips for Cursor:**
- 💡 The MCP tools work alongside Cursor's built-in codebase context
- 🔍 Use `@lazy-ai-coder` to explicitly reference MCP tools
- 📝 Tools will auto-activate when you mention GitLab, merge requests, or code analysis
- 🚀 You can access your entire GitLab workspace without leaving Cursor!

#### Option 2: Claude Desktop

Add to your `claude_desktop_config.json`:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "lazy-ai-coder": {
      "command": "/path/to/lazy-ai-coder",
      "args": ["mcp"],
      "env": {
        "GITLAB_BASE_URL": "https://gitlab.com",
        "GITLAB_TOKEN": "your-token",
        "LLM_BASE_URL": "https://api.openai.com/v1",
        "LLM_API_KEY": "your-api-key",
        "LLM_MODEL": "gpt-4"
      }
    }
  }
}
```

After adding the config:
1. Restart Claude Desktop
2. Look for the 🔌 icon in Claude
3. You should see 6 tools available from lazy-ai-coder

### 📖 Detailed Setup Guides

- **Cursor**: See integration guide above or `cursor-mcp-config.json` for config
- **Claude Desktop**: See config above or `mcp-config-example.json`
- **HTTP MCP Testing**: See `tests/README.md` for pytest-based testing

### 📚 API Documentation (Swagger/OpenAPI)

The web server includes **interactive Swagger UI documentation** for all APIs:

```bash
# Start the server
./lazy-ai-coder web -p 8888

# Access Swagger UI
open http://localhost:8888/swagger/index.html
```

**Features:**
- ✅ Interactive API documentation
- ✅ Try-it-out functionality for all endpoints
- ✅ Request/response examples
- ✅ Schema definitions
- ✅ Authentication testing

**Documented APIs:**
- LLM chat and processing (`/api/v1/process`)
- PlantUML diagram generation (`/api/v1/draw`)
- MCP JSON-RPC endpoints (`/api/v1/mcp`)
- Session management (`/api/v1/sessions`)
- Configuration APIs (`/api/v1/prompts`, `/api/v1/projects`)

See `API_DOCUMENTATION.md` and `SWAGGER_SETUP.md` for complete details.

### 🧪 Testing MCP Functionality

The web server mode includes HTTP MCP endpoints that can be tested with pytest:

```bash
# Start the web server
./lazy-ai-coder web -p 8888

# In another terminal, install and run tests

# Option A: Using Poetry (recommended)
poetry install --only test
poetry run pytest tests/test_mcp_http.py -v

# Option B: Using pip
pip install -r tests/requirements.txt
pytest tests/test_mcp_http.py -v

# Option C: Use the quick test script (auto-detects Poetry/pip)
./run-tests.sh --fast --html

# Option D: Use Make
make test-fast
```

**Test Features**:
- ✅ Protocol compliance testing
- ✅ All 6 MCP tools tested
- ✅ Error handling validation
- ✅ Concurrent request testing
- ✅ Integration workflow tests
- ✅ HTML test reports
- ✅ Coverage analysis

See `tests/README.md` for complete testing documentation.

---

## Usage

Open http://10.100.212.8:8888 or other host you use

* You can use this tool to help you to understand and write code, e.g.
  * Explain or summarize the code in natural language, and draw a UML or mindmap diagram
  * Review merge request by inputting the merge request Id
  * Write Code , refactor code or other coding work
  * ...
* You can change the related system or user prompts in config/config.yaml
* You can change the gitlab projects configuration in config/config.yaml

### Use case
#### summarize code and draw uml and mindmap
1. configure your API keys and GitLab settings in the Settings page first
2. select prompt "1.summarize""
3. select "computer language"(java by default) and "output language"(chinese by default)
4. select gitlab project and configure repo/branch, or input local code path or remote code url, or "Gitlab Code Path" 
5. click "submit"
6. wait for a while for answer
7. click "Draw Image" to generate UML and Mindmap
8. click "Save Image" to save UML and Mindmap
9. click "Save Answer" to save the answer
10. click "Copy Answer" to copy the answer into clipboard (it requests you run the tool on localhost or use a HTTPS proxy)

#### help review merge request
1. configure your API keys and GitLab settings in the Settings page first
2. select prompt "5.review_mr" or use the Review page for specialized review types
3. select "computer language"(java by default) and "output language"(chinese by default)
4. select gitlab project and configure repo/branch
5. input merge request Id
6. click "submit"

#### more cases
You can change config.yml for the predefined prompt, or just edit the system prompt or user prompt in the text area.

## Setup

### Step 0: Import Prompts (First Time Setup)

Import all prompt templates from config to database:

```bash
# Build the application
go build -o lazy-ai-coder

# Import prompts (73 templates)
./lazy-ai-coder import prompts

# Verify
sqlite3 db/lazy_ai_coder.db "SELECT COUNT(*) FROM prompts;"
# Should show: 73
```

**Options:**
- `--dry-run` - Preview without importing
- `--update` - Update existing prompts
- `--file <path>` - Use custom YAML file

See `IMPORT_COMMAND_GUIDE.md` for details.

### Option 1: Web-based Settings (Recommended)

1. Start the application (see build and run steps below)
2. Open the web interface
3. Click the "Settings" button in the top right corner
4. Configure your API keys and settings:
   - **LLM API Key**: Your API key for the LLM service
   - **LLM Model**: Select your preferred model (GPT-4, Claude, etc.)
   - **LLM Base URL**: Base URL for your LLM API service (e.g., https://api.openai.com/v1)
   - **GitLab Base URL**: Your GitLab instance URL
   - **GitLab Token**: Your GitLab personal access token
5. Click "Save Settings" - these will be stored in your browser's local storage

### Option 2: Environment Variables

1. create .env like below

```
PLANTUML_URL="http://10.100.212.8:8000"

ADMIN_USERNAME="{{your_admin_username}}"
ADMIN_PASSWORD="{{your_admin_password}}"
```

**Note**: Web-based settings will override environment variables when both are present.

2. build

```
make
```

3. run

```shell
# it is the folder to save the UML or mindmap images
./start.sh
```

4. test

* open http://$host:8888/
* select predefined prompt
* do some change, e.g. input your code path
* click submit

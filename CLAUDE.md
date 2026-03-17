# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Lazy AI Coder** is a full-stack application that integrates GitLab, LLM APIs, and PlantUML to help developers understand, write, and review code. It supports both a web-based UI and operates as an MCP (Model Context Protocol) server.

**Key Features:**
- Code summarization and analysis with LLMs
- GitLab integration (read files, review merge requests)
- MCP server mode for Claude Desktop, Cursor, and other MCP clients
- PlantUML diagram generation (UML, mindmaps)
- Web-based UI with real-time streaming

## Build and Run Commands

### Build
```bash
# Standard build
make build
# Or directly:
go build -o lazy-ai-coder

# Build with Swagger documentation
make build-with-swagger
# Or:
swag init -g main.go --output docs
go build -o lazy-ai-coder
```

### Run Modes

**1. Web Server Mode (HTTP + MCP endpoints)**
```bash
# Default port 8888
./lazy-ai-coder web -p 8888
# Or via Make:
make run-web

# Access:
# - Web UI: http://localhost:8888
# - Swagger UI: http://localhost:8888/swagger/index.html
# - MCP HTTP endpoint: http://localhost:8888/api/v1/mcp
```

**2. MCP Stdio Server Mode (for Cursor/Claude Desktop)**
```bash
./lazy-ai-coder mcp
# Or:
make run-mcp
```

**3. Import Prompts (First-time Setup)**
```bash
# Import 73 prompt templates from config/prompts.yaml to database
./lazy-ai-coder import prompts

# Options:
./lazy-ai-coder import prompts --dry-run  # Preview without importing
./lazy-ai-coder import prompts --update   # Update existing prompts
```

### Testing

**Go Tests:**
```bash
# Run Go tests
go test ./...
go test -v ./internal/mcp/...
```

**Python MCP HTTP Tests:**
```bash
# Must start web server first
./lazy-ai-coder web -p 8888 &

# Option 1: Using Poetry (recommended)
poetry install --only test
poetry run pytest tests/test_mcp_http.py -v

# Option 2: Using pip
pip install -r tests/requirements.txt
pytest tests/test_mcp_http.py -v

# Option 3: Quick test script (auto-detects Poetry/pip)
./run-tests.sh --fast     # Fast tests
./run-tests.sh --html     # With HTML report
./run-tests.sh --coverage # With coverage

# Option 4: Make targets
make test-fast
make test-html
make test-coverage

# Skip external dependencies (GitLab, LLM)
pytest tests/test_mcp_http.py -v -m "not gitlab and not llm"
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Development server (http://localhost:5173)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Type check
npm run type-check

# Or via Make from root:
make frontend
```

### Docker Deployment

```bash
# Using docker-compose
cd deploy
cp .env.example .env
# Edit .env with credentials
docker-compose up -d

# Or from root:
make docker-compose-up

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down
```

## Architecture

### High-Level Structure

```
lazy-ai-coder/
├── cmd/                    # CLI commands (cobra)
│   ├── root.go            # Root command
│   ├── web.go             # Web server command
│   ├── mcp.go             # MCP stdio server command
│   ├── import.go          # Import prompts/projects
│   └── config.go          # Config management
├── internal/              # Private application code
│   ├── auth/              # JWT authentication, user service
│   ├── chat/              # LLM chat service & routes
│   ├── diagram/           # PlantUML diagram generation
│   ├── handlers/          # HTTP handlers (auth, projects, prompts)
│   ├── mcp/               # MCP protocol implementation
│   │   ├── server.go      # MCP server (stdio & HTTP)
│   │   ├── tools.go       # Tool definitions
│   │   └── handlers.go    # Tool handlers
│   ├── models/            # API request/response types
│   ├── oauth/             # OAuth integration (GitLab)
│   ├── services/          # Business logic (project, prompt services)
│   ├── smartprompt/       # Smart prompt generator
│   └── util/              # Utilities (file reader, converter)
├── pkg/                   # Public library code
│   ├── database/          # Database initialization (SQLite, Postgres, MySQL)
│   ├── models/            # GORM database models
│   │   ├── user.go        # User model
│   │   ├── project.go     # Project model
│   │   └── prompt.go      # Prompt model
│   ├── auth/              # Auth utilities
│   ├── authz/             # Authorization (Casbin)
│   ├── handlers/          # Base handlers
│   └── metrics/           # Prometheus metrics
├── frontend/              # Vue 3 + TypeScript frontend
│   ├── src/
│   │   ├── views/         # Page components
│   │   ├── components/    # Reusable components
│   │   ├── stores/        # Pinia state management
│   │   ├── services/      # API service layer
│   │   ├── router/        # Vue Router
│   │   └── types/         # TypeScript types
│   ├── package.json
│   └── vite.config.ts
├── config/
│   ├── config.yaml        # Main configuration
│   └── prompts.yaml       # Prompt templates (73 templates)
├── tests/                 # Python pytest tests for MCP HTTP
├── deploy/                # Docker compose deployment
├── docs/                  # Swagger documentation
└── main.go               # Application entry point
```

### Key Design Patterns

**Backend (Go):**
- **Clean Architecture**: Repository → Service → Handler/Routes layers
- **Dependency Injection**: Services injected into handlers
- **Middleware Pattern**: Auth, logging, CORS middleware
- **MCP Tools Pattern**: Tool registration with handlers in `internal/mcp/`

**Frontend (Vue 3):**
- **Composition API**: Modern Vue 3 pattern
- **Pinia Stores**: Centralized state management (settings, prompts, projects)
- **Service Layer**: `apiService.ts` for all API calls
- **Component-based**: Reusable components in `components/`, pages in `views/`

**Database:**
- **GORM Models**: In `pkg/models/` with soft deletes
- **Multi-DB Support**: SQLite (default), PostgreSQL, MySQL
- **Migrations**: SQL files in `deploy/db/migrations/`

## MCP (Model Context Protocol) Implementation

### Two MCP Modes

**Mode 1: Stdio MCP Server** (for Cursor/Claude Desktop)
- Transport: stdin/stdout with JSON-RPC 2.0
- Command: `./lazy-ai-coder mcp`
- Use case: Direct integration with IDEs and desktop apps

**Mode 2: HTTP MCP Server** (built into web server)
- Transport: HTTP POST with JSON-RPC 2.0
- Endpoint: `http://localhost:8888/api/v1/mcp`
- Additional endpoints: `/api/v1/mcp/info`, `/api/v1/mcp/tools`
- Use case: Testing, custom integrations, web-based access

### Available MCP Tools (6 tools)

1. **get_gitlab_file_content** - Fetch file from GitLab
2. **get_gitlab_merge_request** - Fetch MR changes
3. **get_gitlab_project_id** - Get project ID from path
4. **llm_chat** - Chat with LLM
5. **llm_analyze_code** - Analyze code (review/explain/optimize/security/bugs)
6. **generate_plantuml** - Generate UML/mindmap diagrams

### MCP Integration Examples

**Cursor Integration:**
Edit `~/.cursor/mcp_config.json`:
```json
{
  "mcpServers": {
    "lazy-ai-coder": {
      "command": "/absolute/path/to/lazy-ai-coder",
      "args": ["mcp"],
      "env": {
        "GITLAB_BASE_URL": "https://gitlab.com",
        "GITLAB_TOKEN": "glpat-your-token",
        "LLM_BASE_URL": "https://api.openai.com/v1",
        "LLM_API_KEY": "sk-your-key",
        "LLM_MODEL": "gpt-4"
      }
    }
  }
}
```

**Claude Desktop Integration:**
Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) with same format.

## Configuration

### Environment Variables

Required for MCP/LLM functionality:
```bash
LLM_API_KEY=sk-your-api-key
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4
LLM_TEMPERATURE=1.0
LLM_MAX_TOKEN=4096

GITLAB_TOKEN=glpat-your-token
GITLAB_BASE_URL=https://gitlab.com

PLANTUML_URL=http://www.plantuml.com/plantuml
```

### Configuration Files

- `config/config.yaml` - Main config (database, projects)
- `config/prompts.yaml` - 73 prompt templates for various tasks
- `.env` - Environment variables (not committed)

### Database Configuration

Supports SQLite (default), PostgreSQL, MySQL. Configure in `config/config.yaml`:
```yaml
database:
  type: sqlite  # or postgres, mysql
  file_path: ./db/lazy_ai_coder.db  # for SQLite
```

Initialize with migrations:
```bash
# PostgreSQL
docker exec -i pgvector psql -U postgres -d lazy_ai_coder < deploy/db/migrations/*.sql
```

## API Documentation

### Swagger UI
Access interactive API docs at:
```
http://localhost:8888/swagger/index.html
```

Regenerate Swagger docs:
```bash
make swagger
# Or:
swag init -g main.go --output docs
```

### Key API Endpoints

**LLM Processing:**
- `POST /api/v1/process` - Process LLM requests
- `POST /api/v1/draw` - Generate PlantUML diagrams
- `POST /api/v1/chat` - LLM chat

**MCP:**
- `POST /api/v1/mcp` - JSON-RPC MCP endpoint
- `GET /api/v1/mcp/info` - Server info
- `GET /api/v1/mcp/tools` - List tools

**Authentication:**
- `POST /api/v1/auth/login` - User login (JWT)
- `GET /api/v1/auth/callback` - OAuth callback

**Configuration:**
- `GET /api/v1/prompts` - List prompts
- `POST /api/v1/prompts` - Create prompt
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project

## Common Development Tasks

### Adding a New MCP Tool

1. **Define tool schema** in `internal/mcp/tools.go`:
```go
func createMyNewTool() Tool {
    return Tool{
        Name: "my_new_tool",
        Description: "What this tool does",
        InputSchema: InputSchema{
            Type: "object",
            Properties: map[string]Property{
                "param1": {Type: "string", Description: "Parameter description"},
            },
            Required: []string{"param1"},
        },
    }
}
```

2. **Implement handler** in `internal/mcp/handlers.go`:
```go
func (s *Server) handleMyNewTool(args map[string]interface{}) (*CallToolResult, error) {
    param1 := getStringArg(args, "param1", "")
    // Tool logic here
    return &CallToolResult{
        Content: []ContentItem{{Type: "text", Text: "Result"}},
    }, nil
}
```

3. **Register tool** in `internal/mcp/server.go`:
```go
s.RegisterTool(createMyNewTool(), s.handleMyNewTool)
```

### Adding a New Frontend View

1. Create view in `frontend/src/views/MyNewView.vue`
2. Add route in `frontend/src/router/index.ts`
3. Add navigation link in `frontend/src/components/NavigationBar.vue`
4. Create Pinia store if needed in `frontend/src/stores/myNewStore.ts`
5. Define types in `frontend/src/types/index.ts`

### Adding a New Prompt Template

Edit `config/prompts.yaml`:
```yaml
- name: my_prompt
  description: "What this prompt does"
  tags: ["tag1", "tag2"]
  system_prompt: "System instructions"
  user_prompt: "User template with {{variables}}"
```

Then import:
```bash
./lazy-ai-coder import prompts --update
```

## Testing Strategy

### Unit Tests (Go)
- Test business logic in services
- Mock external dependencies
- Location: `*_test.go` files alongside code

### Integration Tests (Python)
- Test MCP HTTP endpoints
- Test protocol compliance
- Location: `tests/test_mcp_http.py`
- Run with: `make test-fast`

### Frontend Tests
- Not currently implemented
- Future: Use Vitest for unit tests, Playwright for E2E

## Security Considerations

- **JWT Authentication**: Token-based auth for web APIs
- **OAuth Integration**: GitLab OAuth for user login
- **CORS**: Configured in web server
- **Input Validation**: Validate all API inputs
- **SQL Injection**: Use GORM parameterized queries
- **Secrets Management**: Never commit API keys or tokens

## Performance Notes

- **WebSocket Streaming**: Real-time LLM responses via WebSocket
- **Database Indexes**: Ensure indexes on frequently queried fields
- **Frontend Bundle**: Use Vite for optimized builds
- **Caching**: Redis available in Docker deployment for caching

## Deployment

**Development:**
```bash
make dev  # Builds with Swagger, starts web server
```

**Production (Docker):**
```bash
cd deploy
cp .env.example .env
# Edit .env with production credentials
docker-compose up -d
```

**Services in Docker deployment:**
- lazy-ai-coder (port 8888) - Main app
- PostgreSQL with pgvector (port 5432)
- Redis (port 6379) - Password protected
- PlantUML server (port 8000)
- PgWeb (port 8081) - Database UI
- JupyterLab (port 8889) - Notebooks
- Nginx (ports 80/443) - Reverse proxy

## Troubleshooting

**MCP Server won't start:**
- Check all required env vars are set (LLM_API_KEY, GITLAB_TOKEN, etc.)
- Verify `config/config.yaml` and `config/prompts.yaml` exist

**Frontend build fails:**
- Run `npm run type-check` to find TypeScript errors
- Clear cache: `rm -rf node_modules/.vite && npm install`

**Tests failing:**
- Ensure web server is running: `./lazy-ai-coder web -p 8888`
- Skip external tests: `pytest -m "not gitlab and not llm"`

**Database issues:**
- Default SQLite DB location: `./db/lazy_ai_coder.db`
- Run migrations if schema missing
- Check database type in `config/config.yaml`

## Code Style and Conventions

**Go:**
- Use `gofmt` for formatting: `make fmt`
- camelCase for unexported, PascalCase for exported
- snake_case for JSON tags and database fields
- Handle all errors explicitly
- Use context for request-scoped data

**TypeScript/Vue:**
- Use Vue 3 Composition API (not Options API)
- TypeScript strict mode enabled
- camelCase for variables/functions
- PascalCase for components
- Type all function parameters and returns

**Commits:**
- Use conventional commits format
- Keep commits atomic and focused

## Resources

- Main README: `/README.md`
- MCP Guide: `/doc/MCP_README.md`
- API Documentation: `/doc/API_DOCUMENTATION.md`
- Frontend Guide: `/frontend/README.md`
- Testing Guide: `/tests/README.md`
- Deployment Guide: `/deploy/README.md`
- Swagger UI: `http://localhost:8888/swagger/index.html` (when server running)

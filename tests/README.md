# MCP HTTP Server Tests

Pytest-based test suite for the MCP HTTP server functionality.

## Setup

### 1. Install Dependencies

**Option A: Using Poetry (Recommended)**

```bash
# Install Poetry if you haven't already
curl -sSL https://install.python-poetry.org | python3 -

# Install test dependencies
poetry install --only test

# Or install all dependencies (including dev tools)
poetry install
```

**Option B: Using pip (Legacy)**

```bash
# Create virtual environment (optional but recommended)
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install test dependencies
pip install -r tests/requirements.txt
```

### 2. Start the Server

```bash
# Build the server
go build -o lazy-ai-coder

# Start in background (Option 1)
./lazy-ai-coder web -p 8888 &

# Or use the start script (Option 2)
./start.sh

# Or run in foreground (Option 3)
./lazy-ai-coder web -p 8888
```

### 3. Configure Environment Variables

```bash
# Optional: Set test configuration
export MCP_TEST_URL="http://localhost:8888"
export TEST_GITLAB_PROJECT="your-org/your-repo"
export TEST_GITLAB_FILE="README.md"
export TEST_MR_ID="123"
```

## Running Tests

### Run All Tests

**With Poetry:**
```bash
poetry run pytest tests/test_mcp_http.py -v
```

**With pip:**
```bash
pytest tests/test_mcp_http.py -v
```

**Using test script (auto-detects Poetry or pip):**
```bash
./run-tests.sh
```

### Run Specific Test Classes

```bash
# Test server info endpoints
poetry run pytest tests/test_mcp_http.py::TestMCPServerInfo -v

# Test JSON-RPC protocol
poetry run pytest tests/test_mcp_http.py::TestMCPProtocol -v

# Test GitLab tools
poetry run pytest tests/test_mcp_http.py::TestGitLabTools -v

# Test LLM tools
poetry run pytest tests/test_mcp_http.py::TestLLMTools -v
```

**Note:** If not using Poetry, replace `poetry run pytest` with `pytest`

### Run Specific Tests

```bash
# Test initialize handshake
pytest tests/test_mcp_http.py::TestMCPProtocol::test_initialize -v

# Test list tools
pytest tests/test_mcp_http.py::TestMCPProtocol::test_list_tools -v

# Test GitLab file fetching
pytest tests/test_mcp_http.py::TestGitLabTools::test_get_gitlab_file_content -v
```

### Run with Markers

```bash
# Skip GitLab tests (no GitLab access needed)
poetry run pytest tests/test_mcp_http.py -v -m "not gitlab"

# Skip LLM tests (no LLM API needed)
poetry run pytest tests/test_mcp_http.py -v -m "not llm"

# Run only integration tests
poetry run pytest tests/test_mcp_http.py -v -m integration

# Run only GitLab tests
poetry run pytest tests/test_mcp_http.py -v -m gitlab
```

### Generate HTML Report

```bash
poetry run pytest tests/test_mcp_http.py -v --html=report.html --self-contained-html
```

### Run with Coverage

```bash
# Run tests with coverage
poetry run pytest tests/test_mcp_http.py -v --cov=internal/mcp --cov-report=html

# View coverage report
open htmlcov/index.html
```

### Run in Parallel

```bash
# Run tests in parallel (4 workers)
poetry run pytest tests/test_mcp_http.py -v -n 4
```

### Using Quick Test Script

The `run-tests.sh` script automatically detects whether you're using Poetry or pip:

```bash
# Fast tests
./run-tests.sh --fast

# With HTML report
./run-tests.sh --html

# With coverage
./run-tests.sh --coverage
```

## Test Categories

### 1. Server Info Tests (`TestMCPServerInfo`)
- `test_server_info` - Test server information endpoint
- `test_list_tools_simple` - Test simple tools listing

### 2. Protocol Tests (`TestMCPProtocol`)
- `test_initialize` - Test MCP handshake
- `test_list_tools` - Test JSON-RPC tools/list
- `test_invalid_method` - Test error handling
- `test_invalid_json` - Test malformed requests

### 3. GitLab Tools (`TestGitLabTools`)
- `test_get_gitlab_file_content` - Fetch files from GitLab
- `test_get_gitlab_merge_request` - Fetch merge requests
- `test_get_gitlab_project_id` - Get project IDs
- `test_gitlab_tool_missing_params` - Parameter validation

### 4. LLM Tools (`TestLLMTools`)
- `test_llm_chat` - Test LLM chat functionality
- `test_llm_analyze_code` - Test code analysis
- `test_llm_analyze_invalid_type` - Test validation

### 5. PlantUML Tools (`TestPlantUMLTools`)
- `test_generate_plantuml` - Test diagram generation

### 6. Error Handling (`TestErrorHandling`)
- `test_malformed_jsonrpc` - Malformed requests
- `test_unknown_tool` - Unknown tool calls
- `test_concurrent_requests` - Concurrent request handling

### 7. Integration Tests (`TestIntegration`)
- `test_fetch_and_analyze_workflow` - Full workflow test

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TEST_URL` | `http://localhost:8888` | MCP server URL |
| `TEST_GITLAB_PROJECT` | `myorg/myrepo` | GitLab project for tests |
| `TEST_GITLAB_FILE` | `README.md` | GitLab file for tests |
| `TEST_MR_ID` | `123` | Merge request ID for tests |

## Example Output

```bash
$ pytest tests/test_mcp_http.py -v

tests/test_mcp_http.py::TestMCPServerInfo::test_server_info PASSED       [ 10%]
tests/test_mcp_http.py::TestMCPServerInfo::test_list_tools_simple PASSED [ 20%]
tests/test_mcp_http.py::TestMCPProtocol::test_initialize PASSED          [ 30%]
tests/test_mcp_http.py::TestMCPProtocol::test_list_tools PASSED          [ 40%]
tests/test_mcp_http.py::TestMCPProtocol::test_invalid_method PASSED      [ 50%]
tests/test_mcp_http.py::TestGitLabTools::test_get_gitlab_file_content PASSED [60%]
tests/test_mcp_http.py::TestLLMTools::test_llm_chat PASSED               [ 70%]
tests/test_mcp_http.py::TestLLMTools::test_llm_analyze_code PASSED       [ 80%]
tests/test_mcp_http.py::TestPlantUMLTools::test_generate_plantuml PASSED [ 90%]
tests/test_mcp_http.py::TestIntegration::test_fetch_and_analyze_workflow PASSED [100%]

===================== 10 passed in 5.23s =====================
```

## Troubleshooting

### Server Not Running

**Error**: `MCP server is not running`

**Solution**:
```bash
./lazy-ai-coder web -p 8888
```

### Connection Refused

**Error**: `requests.exceptions.ConnectionError`

**Solution**: Check server is running on port 8888:
```bash
curl http://localhost:8888/api/v1/mcp/info
```

### GitLab Tests Failing

**Solution**: Set proper GitLab credentials:
```bash
export GITLAB_BASE_URL="https://gitlab.com"
export GITLAB_TOKEN="glpat-your-token"
```

### LLM Tests Failing

**Solution**: Set proper LLM credentials:
```bash
export LLM_BASE_URL="https://api.openai.com/v1"
export LLM_API_KEY="sk-your-key"
export LLM_MODEL="gpt-4"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: MCP Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Build server
      run: go build -o lazy-ai-coder
    
    - name: Start server
      run: |
        ./lazy-ai-coder web -p 8888 &
        sleep 5
    
    - name: Set up Python
      uses: actions/setup-python@v4
      with:
        python-version: '3.10'
    
    - name: Install dependencies
      run: |
        pip install -r tests/requirements.txt
    
    - name: Run tests
      run: |
        pytest tests/test_mcp_http.py -v -m "not gitlab and not llm"
```

## Quick Test Script

For quick testing during development:

```bash
#!/bin/bash
# test-mcp-quick.sh

# Start server if not running
if ! curl -s http://localhost:8888/api/v1/mcp/info > /dev/null; then
    echo "Starting server..."
    ./lazy-ai-coder web -p 8888 &
    sleep 2
fi

# Run basic tests
pytest tests/test_mcp_http.py::TestMCPServerInfo -v
pytest tests/test_mcp_http.py::TestMCPProtocol -v
```

## Further Reading

- [Pytest Documentation](https://docs.pytest.org/)
- [MCP Protocol Spec](https://spec.modelcontextprotocol.io/)
- Main README: [../README.md](../README.md)


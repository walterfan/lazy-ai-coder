#!/bin/bash

# Quick test script for MCP HTTP server
# Usage: ./run-tests.sh [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "MCP HTTP Server Test Runner"
echo "=========================================="
echo ""

# Check if server is running
SERVER_URL="http://localhost:8888"
if ! curl -s "${SERVER_URL}/api/v1/mcp/info" > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Server not running at ${SERVER_URL}${NC}"
    echo ""
    echo "Please start the server first:"
    echo "  ./lazy-ai-coder web -p 8888"
    echo ""
    echo "Or run in background:"
    echo "  ./start.sh"
    echo ""
    exit 1
fi

echo -e "${GREEN}✅ Server is running at ${SERVER_URL}${NC}"
echo ""

# Check Python and Poetry
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}❌ Python 3 is not installed${NC}"
    exit 1
fi

# Check if Poetry is installed
if ! command -v poetry &> /dev/null; then
    echo -e "${YELLOW}📦 Poetry not found. Installing dependencies with pip...${NC}"
    if ! python3 -c "import pytest" 2> /dev/null; then
        pip3 install -r tests/requirements.txt
    fi
    PYTEST_CMD="python3 -m pytest"
else
    # Use Poetry
    if [ ! -d ".venv" ] && [ ! -d "$(poetry env info --path 2>/dev/null)" ]; then
        echo -e "${YELLOW}📦 Installing test dependencies with Poetry...${NC}"
        poetry install --only test
    fi
    PYTEST_CMD="poetry run pytest"
fi

# Parse command line arguments
TEST_ARGS="-v"
SKIP_EXTERNAL=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --fast)
            SKIP_EXTERNAL="-m 'not gitlab and not llm and not plantuml'"
            echo "Running fast tests (skipping external dependencies)..."
            shift
            ;;
        --gitlab-only)
            SKIP_EXTERNAL="-m gitlab"
            echo "Running only GitLab tests..."
            shift
            ;;
        --llm-only)
            SKIP_EXTERNAL="-m llm"
            echo "Running only LLM tests..."
            shift
            ;;
        --html)
            TEST_ARGS="$TEST_ARGS --html=report.html --self-contained-html"
            echo "Will generate HTML report..."
            shift
            ;;
        --coverage)
            TEST_ARGS="$TEST_ARGS --cov=internal/mcp --cov-report=html"
            echo "Will generate coverage report..."
            shift
            ;;
        -h|--help)
            echo "Usage: ./run-tests.sh [options]"
            echo ""
            echo "Options:"
            echo "  --fast         Skip tests requiring external services (GitLab, LLM, PlantUML)"
            echo "  --gitlab-only  Run only GitLab tests"
            echo "  --llm-only     Run only LLM tests"
            echo "  --html         Generate HTML test report"
            echo "  --coverage     Generate coverage report"
            echo "  -h, --help     Show this help message"
            echo ""
            echo "Examples:"
            echo "  ./run-tests.sh                  # Run all tests"
            echo "  ./run-tests.sh --fast           # Run tests without external dependencies"
            echo "  ./run-tests.sh --html           # Run all tests and generate HTML report"
            echo "  ./run-tests.sh --fast --html    # Fast tests with HTML report"
            echo ""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help to see available options"
            exit 1
            ;;
    esac
done

# Run tests
echo "Running tests..."
echo "Command: $PYTEST_CMD tests/test_mcp_http.py $TEST_ARGS $SKIP_EXTERNAL"
echo ""

if $PYTEST_CMD tests/test_mcp_http.py $TEST_ARGS $SKIP_EXTERNAL; then
    echo ""
    echo -e "${GREEN}=========================================="
    echo "✅ All tests passed!"
    echo -e "==========================================${NC}"
    
    if [[ $TEST_ARGS == *"--html"* ]]; then
        echo ""
        echo "📊 HTML report generated: report.html"
        echo "   Open with: open report.html"
    fi
    
    if [[ $TEST_ARGS == *"--cov"* ]]; then
        echo ""
        echo "📊 Coverage report generated: htmlcov/index.html"
        echo "   Open with: open htmlcov/index.html"
    fi
else
    echo ""
    echo -e "${RED}=========================================="
    echo "❌ Some tests failed"
    echo -e "==========================================${NC}"
    exit 1
fi


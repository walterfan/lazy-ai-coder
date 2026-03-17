"""
Pytest configuration and shared fixtures
"""

import pytest
import os


def pytest_configure(config):
    """Configure pytest"""
    config.addinivalue_line(
        "markers", "gitlab: mark test as requiring GitLab access"
    )
    config.addinivalue_line(
        "markers", "llm: mark test as requiring LLM API access"
    )
    config.addinivalue_line(
        "markers", "plantuml: mark test as requiring PlantUML server"
    )
    config.addinivalue_line(
        "markers", "integration: mark test as integration test"
    )


@pytest.fixture(scope="session")
def base_url():
    """Base URL for the MCP server"""
    return os.getenv("MCP_TEST_URL", "http://localhost:8888")


@pytest.fixture(scope="session")
def mcp_endpoint(base_url):
    """MCP JSON-RPC endpoint"""
    return f"{base_url}/api/v1/mcp"


@pytest.fixture
def gitlab_config():
    """GitLab configuration for tests"""
    return {
        "project": os.getenv("TEST_GITLAB_PROJECT", "myorg/myrepo"),
        "file": os.getenv("TEST_GITLAB_FILE", "README.md"),
        "mr_id": os.getenv("TEST_MR_ID", "123"),
        "branch": "main"
    }


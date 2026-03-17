"""
Test suite for MCP HTTP server endpoints

Run tests:
    pytest tests/test_mcp_http.py -v
    pytest tests/test_mcp_http.py -v -k test_initialize
    pytest tests/test_mcp_http.py -v --html=report.html
"""

import pytest
import requests
import json
import os
from typing import Dict, Any


# Configuration
BASE_URL = os.getenv("MCP_TEST_URL", "http://localhost:8888")
MCP_ENDPOINT = f"{BASE_URL}/api/v1/mcp"
INFO_ENDPOINT = f"{BASE_URL}/api/v1/mcp/info"
TOOLS_ENDPOINT = f"{BASE_URL}/api/v1/mcp/tools"

# Test data
TEST_GITLAB_PROJECT = os.getenv("TEST_GITLAB_PROJECT", "myorg/myrepo")
TEST_GITLAB_FILE = os.getenv("TEST_GITLAB_FILE", "README.md")
TEST_MR_ID = os.getenv("TEST_MR_ID", "123")


class TestMCPServerInfo:
    """Test MCP server information endpoints"""
    
    def test_server_info(self):
        """Test GET /api/v1/mcp/info endpoint"""
        response = requests.get(INFO_ENDPOINT)
        assert response.status_code == 200
        
        data = response.json()
        assert data["name"] == "lazy-ai-coder"
        assert data["version"] == "1.0.0"
        assert data["protocolVersion"] == "2024-11-05"
        assert data["transport"] == "http"
        assert data["capabilities"]["tools"] is True
    
    def test_list_tools_simple(self):
        """Test GET /api/v1/mcp/tools endpoint"""
        response = requests.get(TOOLS_ENDPOINT)
        assert response.status_code == 200
        
        data = response.json()
        assert "tools" in data
        assert "count" in data
        assert data["count"] == 6
        
        tool_names = [tool["name"] for tool in data["tools"]]
        assert "get_gitlab_file_content" in tool_names
        assert "get_gitlab_merge_request" in tool_names
        assert "llm_chat" in tool_names


class TestMCPProtocol:
    """Test MCP JSON-RPC protocol"""
    
    def send_jsonrpc_request(self, method: str, params: Dict = None) -> Dict:
        """Helper to send JSON-RPC requests"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": method,
            "params": params or {}
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        assert response.status_code == 200
        return response.json()
    
    def test_initialize(self):
        """Test MCP initialize handshake"""
        params = {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "pytest-client",
                "version": "1.0.0"
            }
        }
        
        response = self.send_jsonrpc_request("initialize", params)
        
        assert response["jsonrpc"] == "2.0"
        assert response["id"] == 1
        assert "result" in response
        
        result = response["result"]
        assert result["protocolVersion"] == "2024-11-05"
        assert result["serverInfo"]["name"] == "lazy-ai-coder"
        assert result["capabilities"]["tools"] is not None
    
    def test_list_tools(self):
        """Test tools/list method"""
        response = self.send_jsonrpc_request("tools/list", {})
        
        assert response["jsonrpc"] == "2.0"
        assert "result" in response
        
        tools = response["result"]["tools"]
        assert len(tools) == 6
        
        # Verify all expected tools are present
        tool_names = [tool["name"] for tool in tools]
        expected_tools = [
            "get_gitlab_file_content",
            "get_gitlab_merge_request",
            "get_gitlab_project_id",
            "llm_chat",
            "llm_analyze_code",
            "generate_plantuml"
        ]
        
        for tool_name in expected_tools:
            assert tool_name in tool_names
    
    def test_invalid_method(self):
        """Test error handling for invalid method"""
        response = self.send_jsonrpc_request("invalid_method", {})
        
        assert "error" in response
        assert response["error"]["code"] == -32601
        assert "not found" in response["error"]["message"].lower()
    
    def test_invalid_json(self):
        """Test error handling for invalid JSON"""
        response = requests.post(MCP_ENDPOINT, data="invalid json")
        
        assert response.status_code == 400


class TestGitLabTools:
    """Test GitLab-related MCP tools"""
    
    def send_tool_call(self, tool_name: str, arguments: Dict) -> Dict:
        """Helper to call a tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        assert response.status_code == 200
        return response.json()
    
    @pytest.mark.gitlab
    def test_get_gitlab_file_content(self):
        """Test get_gitlab_file_content tool"""
        response = self.send_tool_call(
            "get_gitlab_file_content",
            {
                "project": TEST_GITLAB_PROJECT,
                "file_path": TEST_GITLAB_FILE,
                "branch": "main"
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert len(result["content"]) > 0
        assert result["content"][0]["type"] == "text"
    
    @pytest.mark.gitlab
    def test_get_gitlab_merge_request(self):
        """Test get_gitlab_merge_request tool"""
        response = self.send_tool_call(
            "get_gitlab_merge_request",
            {
                "project": TEST_GITLAB_PROJECT,
                "merge_request_id": TEST_MR_ID
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert len(result["content"]) > 0
    
    @pytest.mark.gitlab
    def test_get_gitlab_project_id(self):
        """Test get_gitlab_project_id tool"""
        response = self.send_tool_call(
            "get_gitlab_project_id",
            {
                "project_name": TEST_GITLAB_PROJECT
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert "ID:" in result["content"][0]["text"]
    
    def test_gitlab_tool_missing_params(self):
        """Test GitLab tool with missing parameters"""
        response = self.send_tool_call(
            "get_gitlab_file_content",
            {
                "project": TEST_GITLAB_PROJECT
                # Missing file_path
            }
        )
        
        # Should return an error result
        assert "result" in response
        result = response["result"]
        assert result.get("isError") is True


class TestLLMTools:
    """Test LLM-related MCP tools"""
    
    def send_tool_call(self, tool_name: str, arguments: Dict) -> Dict:
        """Helper to call a tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        assert response.status_code == 200
        return response.json()
    
    @pytest.mark.llm
    def test_llm_chat(self):
        """Test llm_chat tool"""
        response = self.send_tool_call(
            "llm_chat",
            {
                "user_prompt": "Say hello",
                "system_prompt": "You are a helpful assistant."
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert len(result["content"]) > 0
        assert result["content"][0]["type"] == "text"
    
    @pytest.mark.llm
    def test_llm_analyze_code(self):
        """Test llm_analyze_code tool"""
        code = """
def hello():
    print("Hello, World!")
"""
        
        response = self.send_tool_call(
            "llm_analyze_code",
            {
                "code": code,
                "analysis_type": "review",
                "language": "python"
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert len(result["content"]) > 0
    
    def test_llm_analyze_invalid_type(self):
        """Test llm_analyze_code with invalid analysis type"""
        response = self.send_tool_call(
            "llm_analyze_code",
            {
                "code": "print('test')",
                "analysis_type": "invalid_type"
            }
        )
        
        # Should still work, just use default analysis
        assert "result" in response


class TestPlantUMLTools:
    """Test PlantUML diagram generation tools"""
    
    def send_tool_call(self, tool_name: str, arguments: Dict) -> Dict:
        """Helper to call a tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        assert response.status_code == 200
        return response.json()
    
    @pytest.mark.plantuml
    def test_generate_plantuml(self):
        """Test generate_plantuml tool"""
        script = """
@startuml
Alice -> Bob: Hello
Bob -> Alice: Hi
@enduml
"""
        
        response = self.send_tool_call(
            "generate_plantuml",
            {
                "script": script,
                "type": "uml"
            }
        )
        
        assert "result" in response
        result = response["result"]
        assert "content" in result
        assert "URL:" in result["content"][0]["text"]


class TestErrorHandling:
    """Test error handling and edge cases"""
    
    def send_jsonrpc_request(self, method: str, params: Dict = None, request_id=1) -> Dict:
        """Helper to send JSON-RPC requests"""
        request = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": method,
            "params": params or {}
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        return response
    
    def test_malformed_jsonrpc(self):
        """Test malformed JSON-RPC request"""
        response = requests.post(MCP_ENDPOINT, json={
            "method": "initialize"
            # Missing jsonrpc and id fields
        })
        
        # Server should handle gracefully
        assert response.status_code in [200, 400]
    
    def test_unknown_tool(self):
        """Test calling unknown tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": "unknown_tool",
                "arguments": {}
            }
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        assert response.status_code == 200
        
        data = response.json()
        assert "error" in data
    
    def test_concurrent_requests(self):
        """Test handling multiple concurrent requests"""
        import concurrent.futures
        
        def make_request(i):
            request = {
                "jsonrpc": "2.0",
                "id": i,
                "method": "tools/list",
                "params": {}
            }
            return requests.post(MCP_ENDPOINT, json=request)
        
        with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(make_request, i) for i in range(10)]
            results = [f.result() for f in concurrent.futures.as_completed(futures)]
        
        # All requests should succeed
        for result in results:
            assert result.status_code == 200
            data = result.json()
            assert "result" in data


class TestIntegration:
    """Integration tests for complete workflows"""
    
    def send_tool_call(self, tool_name: str, arguments: Dict) -> Dict:
        """Helper to call a tool"""
        request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        response = requests.post(MCP_ENDPOINT, json=request)
        return response.json()
    
    @pytest.mark.integration
    @pytest.mark.gitlab
    @pytest.mark.llm
    def test_fetch_and_analyze_workflow(self):
        """Test workflow: fetch GitLab file -> analyze with LLM"""
        # Step 1: Fetch file from GitLab
        fetch_response = self.send_tool_call(
            "get_gitlab_file_content",
            {
                "project": TEST_GITLAB_PROJECT,
                "file_path": TEST_GITLAB_FILE,
                "branch": "main"
            }
        )
        
        assert "result" in fetch_response
        
        # Step 2: Extract code from response
        content = fetch_response["result"]["content"][0]["text"]
        
        # Step 3: Analyze the code
        analyze_response = self.send_tool_call(
            "llm_analyze_code",
            {
                "code": content,
                "analysis_type": "explain"
            }
        )
        
        assert "result" in analyze_response
        assert len(analyze_response["result"]["content"]) > 0


# Fixtures
@pytest.fixture(scope="session", autouse=True)
def check_server_running():
    """Check if the MCP server is running before tests"""
    try:
        response = requests.get(INFO_ENDPOINT, timeout=2)
        if response.status_code != 200:
            pytest.exit("MCP server is not responding. Please start it with: ./lazy-ai-coder web -p 8888")
    except requests.exceptions.RequestException:
        pytest.exit("MCP server is not running. Please start it with: ./lazy-ai-coder web -p 8888")


@pytest.fixture
def sample_code():
    """Sample code for testing"""
    return """
def calculate_sum(a, b):
    return a + b

def main():
    result = calculate_sum(5, 3)
    print(f"Result: {result}")
"""


# Run with: pytest tests/test_mcp_http.py -v
# Run specific test: pytest tests/test_mcp_http.py::TestMCPProtocol::test_initialize -v
# Run with markers: pytest tests/test_mcp_http.py -v -m "not gitlab"


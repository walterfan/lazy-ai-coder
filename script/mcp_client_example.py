#!/usr/bin/env python3
"""
Example MCP Client demonstrating how to use prompts and resources
"""

import requests
import json

class MCPClient:
    def __init__(self, base_url="http://localhost:8888"):
        self.base_url = base_url
        self.mcp_endpoint = f"{base_url}/api/v1/mcp"
        self.request_id = 0

    def _call_jsonrpc(self, method, params=None):
        """Make a JSON-RPC 2.0 call to the MCP server"""
        self.request_id += 1
        payload = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params or {}
        }

        response = requests.post(
            self.mcp_endpoint,
            json=payload,
            headers={"Content-Type": "application/json"}
        )
        response.raise_for_status()
        return response.json()

    # Prompt Methods

    def list_prompts(self):
        """List all available prompts"""
        result = self._call_jsonrpc("prompts/list")
        return result.get("result", {}).get("prompts", [])

    def get_prompt(self, name, arguments=None):
        """
        Get a specific prompt with variable substitution

        Args:
            name: Prompt name (e.g., "correct_syntax")
            arguments: Dict of variables to substitute (e.g., {"language": "Python", "function": "add"})
        """
        params = {"name": name}
        if arguments:
            params["arguments"] = arguments

        result = self._call_jsonrpc("prompts/get", params)
        return result.get("result", {})

    # Resource Methods

    def list_resources(self):
        """List all available resources"""
        result = self._call_jsonrpc("resources/list")
        return result.get("result", {}).get("resources", [])

    def read_resource(self, uri):
        """
        Read a resource by URI

        Args:
            uri: Resource URI (e.g., "file://config/config.yaml" or "gitlab://projects")
        """
        result = self._call_jsonrpc("resources/read", {"uri": uri})
        contents = result.get("result", {}).get("contents", [])
        return contents[0] if contents else None

    # Convenience Methods

    def search_prompts(self, keyword):
        """Search prompts by keyword in name or description"""
        prompts = self.list_prompts()
        keyword_lower = keyword.lower()
        return [
            p for p in prompts
            if keyword_lower in p.get("name", "").lower()
            or keyword_lower in p.get("description", "").lower()
        ]


def main():
    """Example usage"""
    client = MCPClient()

    print("=" * 60)
    print("MCP Client Example - Prompts and Resources")
    print("=" * 60)

    # 1. List all prompts
    print("\n1. Listing all prompts...")
    prompts = client.list_prompts()
    print(f"   Found {len(prompts)} prompts")
    print(f"   First 3 prompts:")
    for prompt in prompts[:3]:
        print(f"   - {prompt['name']}: {prompt['description']}")
        if prompt.get('arguments'):
            args = [arg['name'] for arg in prompt['arguments']]
            print(f"     Arguments: {', '.join(args)}")

    # 2. Search for specific prompts
    print("\n2. Searching for code review prompts...")
    code_prompts = client.search_prompts("review")
    print(f"   Found {len(code_prompts)} prompts matching 'review':")
    for prompt in code_prompts[:5]:
        print(f"   - {prompt['name']}: {prompt['description']}")

    # 3. Get a specific prompt with variables
    print("\n3. Getting 'correct_syntax' prompt with variables...")
    prompt_result = client.get_prompt(
        "correct_syntax",
        arguments={
            "function": "calculateTotal",
            "language": "Python"
        }
    )
    print(f"   Description: {prompt_result.get('description')}")
    print(f"   Messages:")
    for msg in prompt_result.get('messages', []):
        role = msg['role']
        text = msg['content']['text'][:100] + "..." if len(msg['content']['text']) > 100 else msg['content']['text']
        print(f"   - [{role}] {text}")

    # 4. List all resources
    print("\n4. Listing all resources...")
    resources = client.list_resources()
    print(f"   Found {len(resources)} resources:")
    for resource in resources:
        print(f"   - {resource['name']} ({resource['uri']})")
        print(f"     Type: {resource['mimeType']}, Description: {resource['description']}")

    # 5. Read a resource
    print("\n5. Reading GitLab projects resource...")
    gitlab_resource = client.read_resource("gitlab://projects")
    if gitlab_resource:
        print(f"   URI: {gitlab_resource['uri']}")
        print(f"   Content (first 200 chars):")
        text = gitlab_resource['text'][:200] + "..." if len(gitlab_resource['text']) > 200 else gitlab_resource['text']
        print(f"   {text}")

    # 6. Read a config file resource
    print("\n6. Reading config file resource...")
    config_resource = client.read_resource("file://config/config.yaml")
    if config_resource:
        print(f"   URI: {config_resource['uri']}")
        lines = config_resource['text'].split('\n')
        print(f"   First 10 lines:")
        for line in lines[:10]:
            print(f"   {line}")

    # 7. Example: Using a prompt for code generation
    print("\n7. Example workflow - Using prompt for code generation...")
    code_gen_prompts = client.search_prompts("write")
    if code_gen_prompts:
        prompt_name = code_gen_prompts[0]['name']
        print(f"   Found prompt: {prompt_name}")

        # Get the prompt with specific variables
        prompt = client.get_prompt(
            prompt_name,
            arguments=code_gen_prompts[0].get('arguments', {})
        )
        print(f"   You can now use this prompt with an LLM to generate code!")

    print("\n" + "=" * 60)
    print("Example completed!")
    print("=" * 60)


if __name__ == "__main__":
    main()

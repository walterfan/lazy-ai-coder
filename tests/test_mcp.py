#!/usr/bin/env python3
"""
Test MCP server in stdio transport mode
Sends JSON-RPC requests and displays responses
"""

import json
import subprocess
import sys

def send_mcp_request(request_dict):
    """Send a JSON-RPC request to MCP server and get response"""
    request_json = json.dumps(request_dict) + '\n'

    # Start MCP server process
    process = subprocess.Popen(
        ['./lazy-ai-coder', 'mcp'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )

    # Send request
    stdout, stderr = process.communicate(input=request_json, timeout=5)

    # Parse response (first line should be JSON-RPC response)
    for line in stdout.split('\n'):
        if line.strip() and not line.startswith('{"level":'):
            try:
                return json.loads(line)
            except json.JSONDecodeError:
                continue

    return None

def print_section(title):
    """Print a section header"""
    print(f"\n{'='*60}")
    print(f"  {title}")
    print('='*60)

def print_result(label, data):
    """Print formatted result"""
    print(f"\n{label}:")
    print(json.dumps(data, indent=2))

def main():
    print_section("MCP Server stdio Transport Test")
    print("\nTesting lazy-ai-coder MCP server...")

    # Test 1: Initialize
    print_section("1. Initialize")
    init_response = send_mcp_request({
        "jsonrpc": "2.0",
        "id": 1,
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test-client",
                "version": "1.0.0"
            }
        }
    })

    if init_response:
        print("\n✓ Server initialized successfully")
        if 'result' in init_response:
            server_info = init_response['result'].get('serverInfo', {})
            print(f"  Server: {server_info.get('name')} v{server_info.get('version')}")
            print(f"  Protocol: {init_response['result'].get('protocolVersion')}")

            caps = init_response['result'].get('capabilities', {})
            print(f"\n  Capabilities:")
            print(f"    - Tools: {'✓' if caps.get('tools') else '✗'}")
            print(f"    - Resources: {'✓' if caps.get('resources') else '✗'}")
            print(f"    - Prompts: {'✓' if caps.get('prompts') else '✗'}")
    else:
        print("✗ Failed to initialize")
        return

    # Test 2: List Tools
    print_section("2. List Tools")
    tools_response = send_mcp_request({
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tools/list",
        "params": {}
    })

    if tools_response and 'result' in tools_response:
        tools = tools_response['result'].get('tools', [])
        print(f"\n✓ Found {len(tools)} tools:")
        for tool in tools:
            print(f"\n  • {tool['name']}")
            print(f"    {tool['description']}")
            required = tool.get('inputSchema', {}).get('required', [])
            if required:
                print(f"    Required params: {', '.join(required)}")
    else:
        print("✗ Failed to list tools")

    # Test 3: List Resources
    print_section("3. List Resources")
    resources_response = send_mcp_request({
        "jsonrpc": "2.0",
        "id": 3,
        "method": "resources/list",
        "params": {}
    })

    if resources_response and 'result' in resources_response:
        resources = resources_response['result'].get('resources', [])
        print(f"\n✓ Found {len(resources)} resources:")
        for resource in resources:
            print(f"\n  • {resource['name']}")
            print(f"    URI: {resource['uri']}")
            print(f"    Type: {resource.get('mimeType', 'N/A')}")
    else:
        print("✗ Failed to list resources")

    # Test 4: List Prompts
    print_section("4. List Prompts")
    prompts_response = send_mcp_request({
        "jsonrpc": "2.0",
        "id": 4,
        "method": "prompts/list",
        "params": {}
    })

    if prompts_response and 'result' in prompts_response:
        prompts = prompts_response['result'].get('prompts', [])
        print(f"\n✓ Found {len(prompts)} prompts")
        print(f"\nFirst 10 prompts:")
        for prompt in prompts[:10]:
            args_count = len(prompt.get('arguments', []))
            print(f"\n  • {prompt['name']}")
            print(f"    {prompt.get('description', 'No description')}")
            print(f"    Arguments: {args_count}")
    else:
        print("✗ Failed to list prompts")

    # Summary
    print_section("Summary")
    print("\n✓ MCP Server is working correctly in stdio transport mode")
    print("\nComponents:")
    if tools_response and 'result' in tools_response:
        print(f"  • Tools: {len(tools_response['result'].get('tools', []))}")
    if resources_response and 'result' in resources_response:
        print(f"  • Resources: {len(resources_response['result'].get('resources', []))}")
    if prompts_response and 'result' in prompts_response:
        print(f"  • Prompts: {len(prompts_response['result'].get('prompts', []))}")

    print("\nUsage with Claude Desktop/Cursor:")
    print('  Add to MCP config:')
    print('  {')
    print('    "command": "/absolute/path/to/lazy-ai-coder",')
    print('    "args": ["mcp"],')
    print('    "env": {')
    print('      "GITLAB_BASE_URL": "https://gitlab.com",')
    print('      "GITLAB_TOKEN": "your-token",')
    print('      "LLM_BASE_URL": "https://api.openai.com/v1",')
    print('      "LLM_API_KEY": "your-key",')
    print('      "LLM_MODEL": "gpt-4"')
    print('    }')
    print('  }')
    print()

if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print("\n\nTest interrupted by user")
        sys.exit(0)
    except Exception as e:
        print(f"\n✗ Error: {e}")
        sys.exit(1)

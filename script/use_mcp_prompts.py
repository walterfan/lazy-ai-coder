#!/usr/bin/env python3
"""
Example: Using MCP Prompts Programmatically

This script demonstrates how to:
1. List available prompts from the MCP server
2. Get a specific prompt with arguments
3. Use the prompt in your application
"""

import json
import subprocess
import sys

def send_mcp_request(method, params=None):
    """Send a JSON-RPC request to the MCP server"""
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": method,
        "params": params or {}
    }

    # Determine path to lazy-ai-coder binary
    import os
    script_dir = os.path.dirname(os.path.abspath(__file__))
    binary_path = os.path.join(script_dir, '..', 'lazy-ai-coder')

    process = subprocess.Popen(
        [binary_path, 'mcp'],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )

    request_json = json.dumps(request) + '\n'
    stdout, stderr = process.communicate(input=request_json, timeout=5)

    # Parse response (skip log lines)
    for line in stdout.split('\n'):
        if line.strip() and not line.startswith('{"level":'):
            try:
                return json.loads(line)
            except json.JSONDecodeError:
                continue

    return None

def list_prompts():
    """List all available prompts"""
    print("=" * 60)
    print("Available MCP Prompts")
    print("=" * 60)

    response = send_mcp_request("prompts/list")
    if not response or 'result' not in response:
        print("Failed to list prompts")
        return []

    prompts = response['result'].get('prompts', [])
    print(f"\nFound {len(prompts)} prompts:\n")

    for i, prompt in enumerate(prompts[:20], 1):  # Show first 20
        args_count = len(prompt.get('arguments', []))
        print(f"{i:2d}. {prompt['name']}")
        print(f"    {prompt.get('description', 'No description')}")
        print(f"    Arguments: {args_count}")
        print()

    if len(prompts) > 20:
        print(f"... and {len(prompts) - 20} more prompts")
        print()

    return prompts

def get_prompt_details(prompt_name):
    """Get details for a specific prompt"""
    print("=" * 60)
    print(f"Prompt Details: {prompt_name}")
    print("=" * 60)

    # First list prompts to find it
    response = send_mcp_request("prompts/list")
    if not response or 'result' not in response:
        print("Failed to get prompts")
        return None

    prompts = response['result'].get('prompts', [])
    prompt = next((p for p in prompts if p['name'] == prompt_name), None)

    if not prompt:
        print(f"Prompt '{prompt_name}' not found")
        return None

    print(f"\nName: {prompt['name']}")
    print(f"Description: {prompt.get('description', 'N/A')}")
    print(f"\nArguments:")

    arguments = prompt.get('arguments', [])
    if arguments:
        for arg in arguments:
            required = "✓ Required" if arg.get('required', False) else "Optional"
            print(f"  • {arg['name']} ({required})")
            print(f"    {arg.get('description', 'No description')}")
    else:
        print("  No arguments")

    return prompt

def use_prompt(prompt_name, arguments):
    """Use a prompt with specific arguments"""
    print("=" * 60)
    print(f"Using Prompt: {prompt_name}")
    print("=" * 60)

    response = send_mcp_request("prompts/get", {
        "name": prompt_name,
        "arguments": arguments
    })

    if not response or 'result' not in response:
        print("Failed to get prompt")
        return None

    result = response['result']
    messages = result.get('messages', [])

    print("\nGenerated Messages:")
    print("-" * 60)

    for msg in messages:
        role = msg.get('role', 'unknown')
        content = msg.get('content', {})
        text = content.get('text', '')

        print(f"\n[{role.upper()}]")
        print(text)

    return result

# ============================================================================
# Example Usage
# ============================================================================

def example_1_list_prompts():
    """Example 1: List all available prompts"""
    print("\n" + "=" * 60)
    print("EXAMPLE 1: List All Prompts")
    print("=" * 60)

    prompts = list_prompts()
    return len(prompts)

def example_2_get_prompt_details():
    """Example 2: Get details for a specific prompt"""
    print("\n" + "=" * 60)
    print("EXAMPLE 2: Get Prompt Details")
    print("=" * 60)

    get_prompt_details("review_code")

def example_3_use_code_review():
    """Example 3: Use code review prompt"""
    print("\n" + "=" * 60)
    print("EXAMPLE 3: Use Code Review Prompt")
    print("=" * 60)

    code = """
def calculate_total(items):
    total = 0
    for item in items:
        total = total + item['price']
    return total
"""

    use_prompt("review_code", {
        "code": code,
        "language": "python"
    })

def example_4_use_convert_code():
    """Example 4: Convert JavaScript to Python"""
    print("\n" + "=" * 60)
    print("EXAMPLE 4: Convert Code")
    print("=" * 60)

    js_code = """
function getUserNames(users) {
    return users
        .filter(u => u.active)
        .map(u => u.name);
}
"""

    use_prompt("convert_code", {
        "code": js_code,
        "source_language": "javascript",
        "target_language": "python"
    })

def main():
    """Run all examples"""
    print("\n" + "=" * 60)
    print("MCP Prompts - Usage Examples")
    print("=" * 60)

    try:
        # Example 1: List prompts
        prompt_count = example_1_list_prompts()

        # Example 2: Get prompt details
        example_2_get_prompt_details()

        # Example 3: Use code review
        example_3_use_code_review()

        # Example 4: Convert code
        example_4_use_convert_code()

        print("\n" + "=" * 60)
        print("Summary")
        print("=" * 60)
        print(f"\n✓ Successfully demonstrated MCP prompt usage")
        print(f"✓ Total prompts available: {prompt_count}")
        print("\nYou can now:")
        print("  • Use these prompts in Cursor with /prompt <name>")
        print("  • Use in Claude Code with @mcp prompt:<name>")
        print("  • Integrate in your own Python scripts (see this file)")

    except KeyboardInterrupt:
        print("\n\nInterrupted by user")
        sys.exit(0)
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    main()

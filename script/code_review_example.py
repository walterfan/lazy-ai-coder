#!/usr/bin/env python3
"""
Quick example: Using MCP code_review tool

Usage:
    python examples/code_review_example.py
"""

import requests
import json
import sys

def code_review(code, language="", focus="all", base_url="http://localhost:8888"):
    """Perform code review using MCP code_review tool"""
    mcp_endpoint = f"{base_url}/api/v1/mcp"
    
    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "code_review",
            "arguments": {
                "code": code,
                "language": language,
                "focus": focus
            }
        }
    }
    
    try:
        response = requests.post(
            mcp_endpoint,
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=60  # LLM calls can take time
        )
        response.raise_for_status()
        result = response.json()
        
        if "error" in result:
            print(f"❌ Error: {result['error']}")
            return None
        
        content = result.get("result", {}).get("content", [])
        if content and len(content) > 0:
            return content[0].get("text", "")
        
        return None
    except requests.exceptions.RequestException as e:
        print(f"❌ Request failed: {e}")
        print(f"   Make sure the MCP server is running at {base_url}")
        return None

def main():
    print("=" * 70)
    print("MCP Code Review Tool - Example")
    print("=" * 70)
    print()
    
    # Example 1: Security review
    print("Example 1: Security Review")
    print("-" * 70)
    sql_code = "SELECT * FROM users WHERE username = '" + username + "'"
    review = code_review(sql_code, language="sql", focus="security")
    if review:
        print(review)
    print()
    
    # Example 2: Performance review
    print("Example 2: Performance Review")
    print("-" * 70)
    python_code = """
def process_items(items):
    result = []
    for i in range(len(items)):
        for j in range(len(items)):
            result.append(items[i] + items[j])
    return result
"""
    review = code_review(python_code, language="python", focus="performance")
    if review:
        print(review)
    print()
    
    # Example 3: Comprehensive review
    print("Example 3: Comprehensive Review (All)")
    print("-" * 70)
    go_code = """
package main

import (
    "database/sql"
    "fmt"
)

func GetUser(db *sql.DB, id string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", id)
    rows, err := db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var user User
    if rows.Next() {
        rows.Scan(&user.ID, &user.Name)
    }
    return &user, nil
}
"""
    review = code_review(go_code, language="go", focus="all")
    if review:
        print(review)
    print()
    
    # Example 4: Review from file (if provided)
    if len(sys.argv) > 1:
        file_path = sys.argv[1]
        print(f"Example 4: Review File - {file_path}")
        print("-" * 70)
        try:
            with open(file_path, 'r') as f:
                file_code = f.read()
            
            # Determine language from extension
            lang_map = {
                '.go': 'go',
                '.py': 'python',
                '.js': 'javascript',
                '.ts': 'typescript',
                '.java': 'java',
                '.rs': 'rust',
                '.cpp': 'cpp',
                '.c': 'c',
            }
            ext = file_path[file_path.rfind('.'):]
            language = lang_map.get(ext, '')
            
            review = code_review(file_code, language=language, focus="all")
            if review:
                print(review)
        except FileNotFoundError:
            print(f"❌ File not found: {file_path}")
        except Exception as e:
            print(f"❌ Error reading file: {e}")
    
    print()
    print("=" * 70)
    print("Done!")
    print("=" * 70)
    print()
    print("Usage tips:")
    print("  - Review a file: python examples/code_review_example.py path/to/file.go")
    print("  - Focus options: all, security, performance, quality, style")
    print("  - Make sure MCP server is running: ./lazy-ai-coder web -p 8888")

if __name__ == "__main__":
    # Fix the undefined variable
    username = "user123"  # This would come from user input in real code
    
    main()


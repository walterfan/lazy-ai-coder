#!/usr/bin/env python3
"""
SSE Client Test for MCP Server
Tests Server-Sent Events functionality
"""

import requests
import json
import sys
from datetime import datetime

def listen_to_sse(url="http://localhost:9999/api/v1/mcp/events", duration=30):
    """
    Listen to SSE events from the MCP server

    Args:
        url: SSE endpoint URL
        duration: How long to listen (seconds)
    """
    print(f"🚀 Connecting to SSE endpoint: {url}")
    print(f"⏱️  Will listen for {duration} seconds...")
    print("=" * 60)

    try:
        response = requests.get(
            url,
            stream=True,
            headers={
                'Accept': 'text/event-stream',
                'Cache-Control': 'no-cache'
            },
            timeout=duration
        )

        response.raise_for_status()
        print("✅ Connected successfully!")
        print("=" * 60)

        event_type = None
        event_id = None
        data_lines = []
        event_count = 0

        for line in response.iter_lines():
            if not line:
                # Empty line indicates end of event
                if data_lines:
                    event_count += 1
                    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]

                    # Parse data
                    data_str = '\n'.join(data_lines)
                    try:
                        data = json.loads(data_str)
                        data_formatted = json.dumps(data, indent=2)
                    except json.JSONDecodeError:
                        data_formatted = data_str

                    # Print event
                    print(f"\n[{timestamp}] Event #{event_count}")
                    if event_type:
                        print(f"📌 Type: {event_type}")
                    if event_id:
                        print(f"🆔 ID: {event_id}")
                    print(f"📦 Data: {data_formatted}")
                    print("-" * 60)

                    # Reset for next event
                    event_type = None
                    event_id = None
                    data_lines = []
                continue

            line = line.decode('utf-8')

            if line.startswith('event:'):
                event_type = line[6:].strip()
            elif line.startswith('id:'):
                event_id = line[3:].strip()
            elif line.startswith('data:'):
                data_lines.append(line[5:].strip())
            elif line.startswith(':'):
                # Comment line, ignore
                pass

    except requests.exceptions.Timeout:
        print(f"\n⏱️  {duration} seconds elapsed. Disconnecting...")
    except requests.exceptions.RequestException as e:
        print(f"\n❌ Connection error: {e}")
        return 1
    except KeyboardInterrupt:
        print("\n\n⚠️  Interrupted by user")
    finally:
        print("\n" + "=" * 60)
        print(f"📊 Total events received: {event_count}")
        print("👋 Disconnected")

    return 0


def test_sse_connection():
    """Quick connection test"""
    url = "http://localhost:9999/api/v1/mcp/events"

    print("🧪 Testing SSE Connection...")
    print(f"Endpoint: {url}\n")

    try:
        response = requests.get(
            url,
            stream=True,
            headers={'Accept': 'text/event-stream'},
            timeout=5
        )

        response.raise_for_status()

        # Read first few events
        lines_read = 0
        for line in response.iter_lines():
            if lines_read > 10:
                break
            if line:
                print(line.decode('utf-8'))
            lines_read += 1

        print("\n✅ SSE connection test passed!")
        return 0

    except Exception as e:
        print(f"\n❌ SSE connection test failed: {e}")
        return 1


def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(
        description='Test MCP Server SSE functionality'
    )
    parser.add_argument(
        '--url',
        default='http://localhost:9999/api/v1/mcp/events',
        help='SSE endpoint URL (default: http://localhost:9999/api/v1/mcp/events)'
    )
    parser.add_argument(
        '--duration',
        type=int,
        default=30,
        help='How long to listen in seconds (default: 30)'
    )
    parser.add_argument(
        '--test',
        action='store_true',
        help='Run quick connection test only'
    )

    args = parser.parse_args()

    if args.test:
        return test_sse_connection()
    else:
        return listen_to_sse(args.url, args.duration)


if __name__ == '__main__':
    sys.exit(main())

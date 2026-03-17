#!/bin/bash
# Stop the lazy-ai-coder web app (started by start.sh).

BINARY_NAME="${BINARY_NAME:-lazy-ai-coder}"
PID_FILE=".lazy-ai-coder.pid"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

stopped=0

if [[ -f "$PID_FILE" ]]; then
  PID=$(cat "$PID_FILE")
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null || true
    sleep 1
    if kill -0 "$PID" 2>/dev/null; then
      kill -9 "$PID" 2>/dev/null || true
    fi
    stopped=1
  fi
  rm -f "$PID_FILE"
fi

if [[ $stopped -eq 0 ]]; then
  if pkill -f "$BINARY_NAME web" 2>/dev/null; then
    stopped=1
  fi
fi

if [[ $stopped -eq 1 ]]; then
  echo "Stopped lazy-ai-coder web."
else
  echo "No lazy-ai-coder web process found."
fi

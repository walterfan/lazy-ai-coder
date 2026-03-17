#!/bin/bash
# Start the lazy-ai-coder web app in the background.
# Stop with: ./stop.sh

set -e

BINARY_NAME="${BINARY_NAME:-lazy-ai-coder}"
PORT="${PORT:-8888}"
PID_FILE=".lazy-ai-coder.pid"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [[ ! -x ./$BINARY_NAME ]]; then
  echo "Error: ./$BINARY_NAME not found or not executable. Run: make build"
  exit 1
fi

if [[ -f "$PID_FILE" ]]; then
  OLD_PID=$(cat "$PID_FILE")
  if kill -0 "$OLD_PID" 2>/dev/null; then
    echo "Already running (PID $OLD_PID). Stop with: ./stop.sh"
    exit 1
  fi
  rm -f "$PID_FILE"
fi

mkdir -p web/images
nohup ./$BINARY_NAME web -p "$PORT" > /dev/null 2>&1 &
echo $! > "$PID_FILE"
echo "Started lazy-ai-coder web on port $PORT (PID $(cat "$PID_FILE")). Stop with: ./stop.sh"
echo "  Web UI:    http://localhost:$PORT"
echo "  Swagger:   http://localhost:$PORT/swagger/index.html"

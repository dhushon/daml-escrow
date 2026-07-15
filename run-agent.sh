#!/bin/bash
# Check if Gemini key is exported
if [ -z "$GEMINI_API_KEY" ]; then
  echo "Error: Please export your GEMINI_API_KEY before running."
  exit 1
fi

# 1. Start the LiteLLM gateway silently in the background
echo "Starting local routing proxy..."
litellm --config ./litellm-config.yaml --port 4000 > /dev/null 2>&1 &
PROXY_PID=$!

# Ensure the proxy processes terminate gracefully when you exit Claude Code
trap "kill $PROXY_PID" EXIT

# Wait a brief moment for the proxy port to bind
sleep 1.5

# 2. Fire up Claude Code targeting your custom project settings
echo "Launching local/Gemini routing environment..."
claude


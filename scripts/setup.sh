#!/bin/bash
set -euo pipefail
PLUGIN_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$PLUGIN_ROOT/bin"
(cd "$PLUGIN_ROOT/tools" && go build -o "$PLUGIN_ROOT/bin/claude-code-hooks" ./cmd/claude-code-hooks)
echo "Setup complete: $PLUGIN_ROOT/bin/claude-code-hooks"

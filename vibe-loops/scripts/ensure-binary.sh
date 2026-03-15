#!/bin/bash
set -euo pipefail
PLUGIN_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BINARY_NAME="${1:-$(basename "$PLUGIN_ROOT")}"
BIN_PATH="$PLUGIN_ROOT/bin/$BINARY_NAME"
if [ -f "$BIN_PATH" ] && [ "$BIN_PATH" -nt "$PLUGIN_ROOT/tools/go.mod" ]; then
  exit 0
fi
command -v go &>/dev/null || { echo "[ensure-binary] Go not found" >&2; exit 1; }
mkdir -p "$PLUGIN_ROOT/bin"
(cd "$PLUGIN_ROOT/tools" && go build -o "$BIN_PATH" "./cmd/$BINARY_NAME")

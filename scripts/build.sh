#!/usr/bin/env bash
set -euo pipefail

ENV="${1:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Validate ENV
if [[ "$ENV" != "dev" && "$ENV" != "prod" ]]; then
    echo "Usage: $0 <dev|prod>" >&2
    exit 1
fi

# Validate branch
EXPECTED_BRANCH="$([[ "$ENV" == "prod" ]] && echo "main" || echo "dev")"
CURRENT_BRANCH="$(git -C "$ROOT_DIR" rev-parse --abbrev-ref HEAD)"
if [[ "$CURRENT_BRANCH" != "$EXPECTED_BRANCH" ]]; then
    echo "Error: $ENV build requires branch '$EXPECTED_BRANCH', current branch is '$CURRENT_BRANCH'" >&2
    exit 1
fi

# Read base_path from backend config
CONFIG_FILE="$ROOT_DIR/backend/config.yaml"
if [[ -f "$CONFIG_FILE" ]]; then
    BASE_PATH=$(grep '^\s*base_path:' "$CONFIG_FILE" | sed 's/.*base_path:[[:space:]]*//' | tr -d '"' | tr -d "'" | tr -d ' ')
else
    BASE_PATH=""
fi
echo "base_path: '${BASE_PATH}'"

# Build frontend
cd "$ROOT_DIR/frontend"
npm ci
VITE_BASE_PATH="$BASE_PATH" npm run build

# Copy to backend embed dir
rm -rf "$ROOT_DIR/backend/web/dist"
cp -r "$ROOT_DIR/frontend/dist" "$ROOT_DIR/backend/web/dist"

# Build backend
mkdir -p "$ROOT_DIR/bin"
cd "$ROOT_DIR/backend"
go build -o "../bin/pm-work-tracker" ./cmd/server

echo "Build complete: bin/pm-work-tracker"

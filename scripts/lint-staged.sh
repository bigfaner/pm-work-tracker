#!/usr/bin/env bash
# scripts/lint-staged.sh — Run linters on staged files
set -euo pipefail

echo "=== Checking staged Go files ==="
staged_go=$(git diff --cached --name-only --diff-filter=ACM -- 'backend/**/*.go')
if [ -n "$staged_go" ]; then
  echo "Running gofmt..."
  for f in $staged_go; do
    if [ -f "$f" ]; then
      gofmt -l "$f" | while read -r unformatted; do
        echo "  NOT FORMATTED: $unformatted"
        gofmt -w "$unformatted"
        git add "$unformatted"
      done
    fi
  done

  echo "Running golangci-lint..."
  cd backend && golangci-lint run ./... 2>&1 || { echo "FAIL: golangci-lint found issues"; exit 1; }
  cd ..
else
  echo "No staged Go files."
fi

echo "=== Checking staged TS/TSX files ==="
staged_ts=$(git diff --cached --name-only --diff-filter=ACM -- 'frontend/**/*.{ts,tsx}')
if [ -n "$staged_ts" ]; then
  if command -v npx &>/dev/null; then
    echo "Running prettier..."
    for f in $staged_ts; do
      if [ -f "$f" ]; then
        npx --prefix frontend prettier --write "$f" 2>/dev/null && git add "$f" || true
      fi
    done
  else
    echo "npx not found, skipping prettier."
  fi
else
  echo "No staged TS/TSX files."
fi

echo "=== lint-staged done ==="

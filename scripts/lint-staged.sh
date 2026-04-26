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

check_sqlite_keywords() {
    local pattern='SUBSTR\(|CAST.*AS INTEGER|datetime\(.*now.*\)|pragma_table_info'
    local files=$(git diff --cached --name-only --diff-filter=ACM -- 'backend/internal/repository/gorm/*.go' | grep -v '_test.go')
    if [ -n "$files" ]; then
        local matches
        matches=$(grep -En "$pattern" $files | grep -v 'nosqlite') || true
        if [ -n "$matches" ]; then
            echo "ERROR: Hardcoded SQLite syntax found in repo layer:"
            echo "$matches"
            echo ""
            echo "Use pkg/dbutil.Dialect instead. See docs/features/db-dialect-compat/."
            return 1
        fi
    fi
    return 0
}

echo "=== Checking SQLite keywords in repo layer ==="
check_sqlite_keywords || { echo "FAIL: SQLite keyword check failed"; exit 1; }

echo "=== lint-staged done ==="

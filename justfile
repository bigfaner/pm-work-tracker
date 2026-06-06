claude:
    claude --dangerously-skip-permissions

claude-c:
    claude --dangerously-skip-permissions -c

setup:
    git config core.hooksPath .githooks
    @echo "Git hooks configured."

release env:
    ./scripts/build.sh {{env}}

start:
    ./bin/darwin-amd64/pm-work-tracker -config ./bin/config.yaml

start-linux:
    ./bin/linux-amd64/pm-work-tracker -config ./bin/config.yaml

start-windows:
    ./bin/windows-amd64/pm-work-tracker.exe -config ./bin/config.yaml

# --- forge standard recipes ---

frontend_dir := "./frontend"
backend_dir  := "./backend"

project-type:
    @echo "mixed"

# Language-level targets (fast feedback, per-task submit gate)

[group("language")]
compile scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npx tsc --noEmit ;;
      backend)  cd "{{backend_dir}}" && go vet ./... ;;
      "")       (cd "{{frontend_dir}}" && npx tsc --noEmit) && (cd "{{backend_dir}}" && go vet ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language")]
build scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run build ;;
      backend)  cd "{{backend_dir}}" && go build ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run build) && (cd "{{backend_dir}}" && go build ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language")]
install scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm install ;;
      backend)  cd "{{backend_dir}}" && go mod download ;;
      "")       (cd "{{frontend_dir}}" && npm install) && (cd "{{backend_dir}}" && go mod download) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language")]
clean scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && rm -rf dist ;;
      backend)  cd "{{backend_dir}}" && go clean ./... ;;
      "")       (cd "{{frontend_dir}}" && rm -rf dist) && (cd "{{backend_dir}}" && go clean ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language")]
ci: install compile build unit-test lint

[group("language-test")]
unit-test:
    #!/usr/bin/env bash
    set -euo pipefail
    (cd "{{frontend_dir}}" && npm test) && (cd "{{backend_dir}}" && go test ./...)

[group("language-test")]
lint scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run lint ;;
      backend)  cd "{{backend_dir}}" && golangci-lint run ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run lint) && (cd "{{backend_dir}}" && golangci-lint run ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language-test")]
fmt scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run fmt ;;
      backend)  cd "{{backend_dir}}" && gofmt -w . ;;
      "")       (cd "{{frontend_dir}}" && npm run fmt) && (cd "{{backend_dir}}" && gofmt -w .) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[group("language-test")]
check scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run lint && npx tsc --noEmit ;;
      backend)  cd "{{backend_dir}}" && golangci-lint run ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run lint && npx tsc --noEmit) && (cd "{{backend_dir}}" && golangci-lint run ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

# Convenience targets (not invoked by forge skills)

run scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run preview ;;
      backend)  cd "{{backend_dir}}" && go run cmd/server/main.go ;;
      "")       cd "{{backend_dir}}" && go run cmd/server/main.go & backend_pid=$!; trap "kill $backend_pid 2>/dev/null" EXIT; cd "{{frontend_dir}}" && npm run preview ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

dev scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run dev ;;
      backend)  cd "{{backend_dir}}" && go run cmd/server/main.go -dev ;;
      "")       cd "{{backend_dir}}" && go run cmd/server/main.go -dev & backend_pid=$!; trap "kill $backend_pid 2>/dev/null" EXIT; cd "{{frontend_dir}}" && npm run dev ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

# --- Surface: backend (api) ---
# API tests are self-contained: Vitest global setup starts its own backend server.
# backend-dev/probe/teardown are for manual development workflows.

# user-customized
[group("backend")]
[unix]
backend-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{backend_dir}}" && go run cmd/server/main.go -dev &
    _pid=$!
    mkdir -p tests/results
    printf '%s\n' "$_pid" > tests/results/.pid-backend
    echo "Backend started (PID $_pid)"

# user-customized
[group("backend")]
[windows]
backend-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{backend_dir}}" && go run cmd/server/main.go -dev &
    _pid=$!
    mkdir -p tests/results
    printf '%s\n' "$_pid" > tests/results/.pid-backend
    echo "Backend started (PID $_pid)"

# user-customized
[group("backend")]
[unix]
backend-probe:
    #!/usr/bin/env bash
    set -euo pipefail
    for _i in {1..6}; do
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            echo "OK: backend (http://localhost:8080)"
            exit 0
        fi
        sleep 5
    done
    echo "FAIL: backend health check timed out" >&2
    exit 1

# user-customized
[group("backend")]
[windows]
backend-probe:
    #!/usr/bin/env bash
    set -euo pipefail
    for _i in {1..6}; do
        if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
            echo "OK: backend (http://localhost:8080)"
            exit 0
        fi
        sleep 5
    done
    echo "FAIL: backend health check timed out" >&2
    exit 1

# user-customized
[group("backend")]
[unix]
backend-test journey='':
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d tests/backend/node_modules ]; then npm install --prefix tests/backend; fi
    if [ "{{journey}}" != "" ]; then
        cd tests/backend && npx vitest run "{{journey}}"
    else
        cd tests/backend && npx vitest run
    fi

# user-customized
[group("backend")]
[windows]
backend-test journey='':
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d tests/backend/node_modules ]; then npm install --prefix tests/backend; fi
    if [ "{{journey}}" != "" ]; then
        cd tests/backend && npx vitest run "{{journey}}"
    else
        cd tests/backend && npx vitest run
    fi

# user-customized
[group("backend")]
[unix]
backend-teardown:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -f tests/results/.pid-backend ]; then
        kill "$(tr -d '\r' < tests/results/.pid-backend)" 2>/dev/null || true
        rm -f tests/results/.pid-backend
    fi

# user-customized
[group("backend")]
[windows]
backend-teardown:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -f tests/results/.pid-backend ]; then
        kill "$(tr -d '\r' < tests/results/.pid-backend)" 2>/dev/null || true
        rm -f tests/results/.pid-backend
    fi

[group("backend")]
backend:
    #!/usr/bin/env bash
    set -euo pipefail
    just backend-test

# --- Surface: frontend (web) ---
# Web E2E tests need full stack: backend + frontend dev server running.

# user-customized
[group("frontend")]
[unix]
frontend-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    just backend-dev
    cd "{{frontend_dir}}" && npm run dev &
    _pid=$!
    mkdir -p tests/results
    printf '%s\n' "$_pid" > tests/results/.pid-frontend
    echo "Frontend started (PID $_pid)"

# user-customized
[group("frontend")]
[windows]
frontend-dev:
    #!/usr/bin/env bash
    set -euo pipefail
    just backend-dev
    cd "{{frontend_dir}}" && npm run dev &
    _pid=$!
    mkdir -p tests/results
    printf '%s\n' "$_pid" > tests/results/.pid-frontend
    echo "Frontend started (PID $_pid)"

# user-customized
[group("frontend")]
[unix]
frontend-probe:
    #!/usr/bin/env bash
    set -euo pipefail
    for _i in {1..6}; do
        if curl -sf http://localhost:5173 > /dev/null 2>&1; then
            echo "OK: frontend (http://localhost:5173)"
            exit 0
        fi
        sleep 5
    done
    echo "FAIL: frontend health check timed out" >&2
    exit 1

# user-customized
[group("frontend")]
[windows]
frontend-probe:
    #!/usr/bin/env bash
    set -euo pipefail
    for _i in {1..6}; do
        if curl -sf http://localhost:5173 > /dev/null 2>&1; then
            echo "OK: frontend (http://localhost:5173)"
            exit 0
        fi
        sleep 5
    done
    echo "FAIL: frontend health check timed out" >&2
    exit 1

# user-customized
[group("frontend")]
[unix]
frontend-test journey='':
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d tests/frontend/node_modules ]; then npm install --prefix tests/frontend; fi
    if [ "{{journey}}" != "" ]; then
        cd tests/frontend && npx playwright test "{{journey}}"
    else
        cd tests/frontend && npx playwright test
    fi

# user-customized
[group("frontend")]
[windows]
frontend-test journey='':
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d tests/frontend/node_modules ]; then npm install --prefix tests/frontend; fi
    if [ "{{journey}}" != "" ]; then
        cd tests/frontend && npx playwright test "{{journey}}"
    else
        cd tests/frontend && npx playwright test
    fi

# user-customized
[group("frontend")]
[unix]
frontend-teardown:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -f tests/results/.pid-frontend ]; then
        kill "$(tr -d '\r' < tests/results/.pid-frontend)" 2>/dev/null || true
        rm -f tests/results/.pid-frontend
    fi
    just backend-teardown

# user-customized
[group("frontend")]
[windows]
frontend-teardown:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -f tests/results/.pid-frontend ]; then
        kill "$(tr -d '\r' < tests/results/.pid-frontend)" 2>/dev/null || true
        rm -f tests/results/.pid-frontend
    fi
    just backend-teardown

[group("frontend")]
frontend:
    #!/usr/bin/env bash
    set -euo pipefail
    just frontend-dev && just frontend-probe && just frontend-test; rc=$?; just frontend-teardown; exit $rc

# Infra tests (build/lint checks, no surface configured)

[group("infra")]
infra-test:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -d tests/infra/node_modules ]; then npm install --prefix tests/infra; fi
    cd tests/infra && npx vitest run

# --- end forge standard recipes ---

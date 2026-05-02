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

compile scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npx tsc --noEmit ;;
      backend)  cd "{{backend_dir}}" && go vet ./... ;;
      "")       (cd "{{frontend_dir}}" && npx tsc --noEmit) && (cd "{{backend_dir}}" && go vet ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

build scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run build ;;
      backend)  cd "{{backend_dir}}" && go build ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run build) && (cd "{{backend_dir}}" && go build ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

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

test scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm test ;;
      backend)  cd "{{backend_dir}}" && go test -race ./... ;;
      "")       (cd "{{frontend_dir}}" && npm test) && (cd "{{backend_dir}}" && go test -race ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

[arg("feature", long)]
test-e2e feature="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ "{{feature}}" != "" ]; then
        cd tests/e2e && npx playwright test {{feature}}/
    else
        [ ! -d tests/e2e/node_modules ] && npm install --prefix tests/e2e
        cd tests/e2e && npx playwright test
    fi

lint scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run lint ;;
      backend)  cd "{{backend_dir}}" && golangci-lint run ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run lint) && (cd "{{backend_dir}}" && golangci-lint run ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

fmt scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npx prettier --write . ;;
      backend)  cd "{{backend_dir}}" && gofmt -w . ;;
      "")       (cd "{{frontend_dir}}" && npx prettier --write .) && (cd "{{backend_dir}}" && gofmt -w .) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

check scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm run lint && npx tsc --noEmit ;;
      backend)  cd "{{backend_dir}}" && golangci-lint run ./... ;;
      "")       (cd "{{frontend_dir}}" && npm run lint && npx tsc --noEmit) && (cd "{{backend_dir}}" && golangci-lint run ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

clean scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && rm -rf dist ;;
      backend)  cd "{{backend_dir}}" && go clean ./... ;;
      "")       (cd "{{frontend_dir}}" && rm -rf dist) && (cd "{{backend_dir}}" && go clean ./...) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

install scope="":
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{scope}}" in
      frontend) cd "{{frontend_dir}}" && npm install ;;
      backend)  cd "{{backend_dir}}" && go mod download ;;
      "")       (cd "{{frontend_dir}}" && npm install) && (cd "{{backend_dir}}" && go mod download) ;;
      *)        echo "[forge] invalid scope '{{scope}}'; expected frontend/backend" >&2; exit 1 ;;
    esac

ci: install compile build test lint

e2e-setup:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ ! -f tests/e2e/package.json ]; then
        echo "Error: tests/e2e/package.json not found" >&2
        exit 1
    fi
    if [ ! -d tests/e2e/node_modules ]; then
        npm install --prefix tests/e2e
    fi
    npx --prefix tests/e2e playwright install chromium
    echo "OK: e2e dependencies ready"

[arg("feature", long)]
e2e-verify feature="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "{{feature}}" ]; then
        echo "Usage: just e2e-verify --feature <slug>" >&2
        exit 1
    fi
    if [ ! -d "tests/e2e/{{feature}}" ]; then
        echo "Error: tests/e2e/{{feature}}/ not found" >&2
        exit 1
    fi
    matches=$(grep -rn '// VERIFY:' "tests/e2e/{{feature}}/" --include='*.spec.ts' || true)
    if [ -n "$matches" ]; then
        count=$(echo "$matches" | wc -l | tr -d ' ')
        echo "Error: $count unresolved // VERIFY: marker(s) in tests/e2e/{{feature}}/" >&2
        echo "" >&2
        echo "$matches" >&2
        echo "" >&2
        echo "Replace each // VERIFY: comment with a real assertion before running tests." >&2
        exit 1
    fi
    echo "OK: no unresolved // VERIFY: markers in tests/e2e/{{feature}}/"

# --- end forge standard recipes ---

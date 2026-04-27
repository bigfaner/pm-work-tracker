claude:
    claude --dangerously-skip-permissions

claude-c:
    claude --dangerously-skip-permissions -c

setup:
    git config core.hooksPath .githooks
    @echo "Git hooks configured."

dev:
    cd backend && go run ./cmd/server/main.go -dev &
    cd frontend && npm run dev

build:
    cd backend && go build ./...
    cd frontend && npm run build

test:
    cd backend && go test -race ./...
    cd frontend && npm test

lint:
    cd backend && golangci-lint run ./...

# Run e2e tests: "just test-e2e" (regression) or "just test-e2e --feature <slug>" (feature tests)
[arg("feature", long)]
test-e2e feature="":
    #!/usr/bin/env bash
    if [ "{{feature}}" != "" ]; then
        scripts_dir="docs/features/{{feature}}/testing/scripts"
        fail=0
        for spec in "$scripts_dir"/*.spec.ts; do
            [ -f "$spec" ] && npx tsx "$spec" || fail=$((fail+1))
        done
        [ "$fail" -eq 0 ]
    else
        [ ! -d tests/e2e/node_modules ] && npm install --prefix tests/e2e
        fail=0
        for spec in $(find tests/e2e -mindepth 2 -name '*.spec.ts'); do
            npx tsx "$spec" || fail=$((fail+1))
        done
        [ "$fail" -eq 0 ]
    fi

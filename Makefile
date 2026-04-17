.PHONY: dev build test setup lint

setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured."

dev:
	@echo "Starting backend..."
	cd backend && go run ./cmd/server & \
	echo "Starting frontend..." && \
	cd frontend && npm run dev

build:
	cd backend && go build ./...
	cd frontend && npm run build

test:
	cd backend && go test ./...
	cd frontend && npm test

lint:
	cd backend && golangci-lint run ./...

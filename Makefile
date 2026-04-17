.PHONY: dev build test

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

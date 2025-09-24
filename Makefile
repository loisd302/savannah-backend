# Makefile for Savannah Backend API

.PHONY: help build run dev migrate migrate-up migrate-down migrate-status clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  dev           - Run in development mode (if you have air installed)"
	@echo "  migrate-up    - Run all pending migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-status- Show migration status"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"

# Build the application
build:
	go build -o backend.exe .

# Run the application
run: build
	./backend.exe

# Run in development mode with auto-reload (requires air)
dev:
	air

# Run migrations
migrate-up:
	go run cmd/migrate.go -action=up

migrate-down:
	go run cmd/migrate.go -action=down

migrate-status:
	go run cmd/migrate.go -action=status

# Alias for migrate-up
migrate: migrate-up

# Clean build artifacts
clean:
	rm -f backend.exe
	go clean

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Build for production
build-prod:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o backend.exe .
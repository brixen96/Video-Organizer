.PHONY: help build run test clean install dev lint format check-env migrate

# Default target
help:
	@echo "Video Organizer - Available Commands"
	@echo "===================================="
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make dev         - Run in development mode with auto-reload"
	@echo "  make test        - Run all tests"
	@echo "  make lint        - Run linters"
	@echo "  make format      - Format code with gofmt"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make install     - Install dependencies"
	@echo "  make check-env   - Check if .env file exists"
	@echo "  make migrate     - Run database migrations"

# Build the application
build: check-env
	@echo "Building video-organizer..."
	@go build -o video-organizer cmd/video-organizer/main.go
	@echo "Build complete! Binary: ./video-organizer"

# Run the application
run: check-env
	@echo "Starting video-organizer..."
	@go run cmd/video-organizer/main.go

# Development mode (requires air: go install github.com/cosmtrek/air@latest)
dev: check-env
	@echo "Starting in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to 'go run'..."; \
		go run cmd/video-organizer/main.go; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "Test coverage:"
	@go tool cover -func=coverage.txt | tail -1

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint code
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"; \
		echo "Falling back to basic checks..."; \
		go vet ./...; \
	fi

# Format code
format:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Code formatted"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f video-organizer video-organizer.exe
	@rm -f coverage.txt coverage.html
	@rm -rf dist/ build/
	@echo "Clean complete"

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Setup development environment
setup: check-env
	@echo "Setting up development environment..."
	@echo "Installing dependencies..."
	@go mod download
	@echo "Creating necessary directories..."
	@mkdir -p frontend/.thumbnails
	@mkdir -p frontend/.performers
	@mkdir -p old_logs
	@echo "Setup complete"

# Database migrations (placeholder for future implementation)
migrate:
	@echo "Running database migrations..."
	@echo "Note: Migrations not yet implemented"

# Build for production (with optimizations)
build-prod:
	@echo "Building for production..."
	@CGO_ENABLED=1 go build -ldflags="-w -s" -o video-organizer cmd/video-organizer/main.go
	@echo "Production build complete"

# Cross-compile for Windows (from Linux/Mac)
build-windows:
	@echo "Cross-compiling for Windows..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o video-organizer.exe cmd/video-organizer/main.go
	@echo "Windows build complete"

# Cross-compile for Linux (from Windows/Mac)
build-linux:
	@echo "Cross-compiling for Linux..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o video-organizer-linux cmd/video-organizer/main.go
	@echo "Linux build complete"

# Run security checks
security:
	@echo "Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc > /dev/null; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Show project statistics
stats:
	@echo "Project Statistics"
	@echo "=================="
	@echo "Go files:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo "Lines of Go code:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "Test files:"
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l

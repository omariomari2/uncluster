# HTML Formatter & JSX Converter Makefile

.PHONY: build build-cli build-server run test clean help

# Default target
all: build

# Build both CLI and server
build: build-cli build-server

# Build CLI tool
build-cli:
	@echo "Building CLI tool..."
	@go build -o htmlfmt cmd/htmlfmt/main.go
	@echo "CLI tool built: htmlfmt"

# Build web server
build-server:
	@echo "Building web server..."
	@go build -o htmlfmt-server api/server.go api/handlers.go
	@echo "Web server built: htmlfmt-server"

# Run the web server
run:
	@echo "Starting web server..."
	@go run api/server.go api/handlers.go

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f htmlfmt htmlfmt-server
	@echo "Clean complete"

# Cross-platform builds
build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o htmlfmt.exe cmd/htmlfmt/main.go
	@GOOS=windows GOARCH=amd64 go build -o htmlfmt-server.exe api/server.go api/handlers.go

build-macos:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o htmlfmt-macos cmd/htmlfmt/main.go
	@GOOS=darwin GOARCH=amd64 go build -o htmlfmt-server-macos api/server.go api/handlers.go

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o htmlfmt-linux cmd/htmlfmt/main.go
	@GOOS=linux GOARCH=amd64 go build -o htmlfmt-server-linux api/server.go api/handlers.go

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Development server with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@air

# Help
help:
	@echo "HTML Formatter & JSX Converter"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build both CLI and server"
	@echo "  build-cli      - Build CLI tool only"
	@echo "  build-server   - Build web server only"
	@echo "  run            - Run the web server"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  build-windows  - Build for Windows"
	@echo "  build-macos    - Build for macOS"
	@echo "  build-linux    - Build for Linux"
	@echo "  deps           - Install dependencies"
	@echo "  dev            - Start development server (requires air)"
	@echo "  help           - Show this help message"

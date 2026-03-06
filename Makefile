.PHONY: build build-cli build-server run test clean help

all: build

build: build-cli build-server

build-cli:
	@echo "Building CLI tool..."
	@go build -o uncluster-split ./cmd/uncluster-split
	@echo "CLI tool built: uncluster-split"

build-server:
	@echo "Building web server..."
	@go build -o htmlfmt-server main.go
	@echo "Web server built: htmlfmt-server"

run:
	@echo "Starting web server..."
	@go run main.go

test:
	@echo "Running tests..."
	@go test ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -f htmlfmt htmlfmt-server uncluster-split uncluster-split.exe uncluster-split-macos uncluster-split-linux htmlfmt-server.exe htmlfmt-server-macos htmlfmt-server-linux
	@echo "Clean complete"

build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o uncluster-split.exe ./cmd/uncluster-split
	@GOOS=windows GOARCH=amd64 go build -o htmlfmt-server.exe main.go

build-macos:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o uncluster-split-macos ./cmd/uncluster-split
	@GOOS=darwin GOARCH=amd64 go build -o htmlfmt-server-macos main.go

build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o uncluster-split-linux ./cmd/uncluster-split
	@GOOS=linux GOARCH=amd64 go build -o htmlfmt-server-linux main.go

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

dev:
	@echo "Starting development server with hot reload..."
	@air

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

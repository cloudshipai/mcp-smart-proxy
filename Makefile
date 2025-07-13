# MCP Smart Proxy Makefile

.PHONY: build run test clean deps test-local test-full test-ai

# Build the application
build:
	go build -o mcp-smart-proxy ./cmd/mcp-smart-proxy

# Run the application
run:
	go run ./cmd/mcp-smart-proxy -config ./mcp.json -addr :8080

# Run with local test config
run-local:
	go run ./cmd/mcp-smart-proxy -config ./temp_scripts/mcp-local.json -addr :8080

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run local tests (without external MCP servers)
test-local:
	go run cmd/test/main.go

# Run full integration test with AI simulation
test-full:
	./temp_scripts/run_full_test.sh

# Run AI client simulator (requires server running)
test-ai:
	cd temp_scripts && go run ai_client_simulator.go

# Clean build artifacts
clean:
	rm -f mcp-smart-proxy

# Test with curl commands
test-curl:
	@echo "Testing health endpoint..."
	curl -s http://localhost:8080/api/v1/health || echo "Server not running"
	@echo "\nTesting tools list..."
	curl -s http://localhost:8080/api/v1/tools | jq . || echo "Server not running or jq not installed"
	@echo "\nTesting tool discovery..."
	curl -s -X POST http://localhost:8080/api/v1/discover \
		-H 'Content-Type: application/json' \
		-d '{"query":"I need to search files"}' | jq . || echo "Server not running or jq not installed"

# Setup environment for testing
setup-env:
	@echo "Setting up test environment..."
	@echo "1. Install Node.js MCP servers for testing:"
	@echo "   npm install -g @modelcontextprotocol/server-filesystem"
	@echo "   npm install -g @modelcontextprotocol/server-brave-search"
	@echo "2. Set environment variables:"
	@echo "   export OPENAI_API_KEY=your_openai_key"
	@echo "   export GEMINI_API_KEY=your_gemini_key"
	@echo "   export BRAVE_API_KEY=your_brave_key"
	@echo "3. Update mcp.json with your API keys"

# Docker build (optional)
docker-build:
	docker build -t mcp-smart-proxy .

# Help
help:
	@echo "Available commands:"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application"
	@echo "  run-local  - Run with local test MCP server"
	@echo "  deps       - Install dependencies"
	@echo "  test-local - Run basic local tests"
	@echo "  test-full  - Run full integration test with AI simulation"
	@echo "  test-ai    - Run AI client simulator (requires server running)"
	@echo "  test-curl  - Test with curl commands"
	@echo "  setup-env  - Show setup instructions"
	@echo "  clean      - Clean build artifacts"
	@echo "  help       - Show this help"
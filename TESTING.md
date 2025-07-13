# MCP Smart Proxy Testing Guide

## Quick Start Testing

### 1. Basic Setup Test
```bash
# Build the application
make build

# Run local test utility
go run cmd/test/main.go
```

### 2. Start the Server
```bash
# Using make
make run

# Or directly
go run . -config ./mcp.json -addr :8080
```

### 3. Test API Endpoints

#### Health Check
```bash
curl http://localhost:8080/api/v1/health
```

#### List All Tools
```bash
curl http://localhost:8080/api/v1/tools
```

#### Discover Tools (requires LLM provider)
```bash
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"I need to search files"}'
```

#### Use a Tool
```bash
curl -X POST http://localhost:8080/api/v1/use/read_file \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"path":"./README.md"}}'
```

## Environment Setup

### LLM Provider Configuration
Choose one or more LLM providers:

#### OpenAI
```bash
export OPENAI_API_KEY=your_openai_api_key_here
```

#### Google Gemini
```bash
export GEMINI_API_KEY=your_gemini_api_key_here
```

### MCP Server Setup
Install test MCP servers:

```bash
# Filesystem server
npm install -g @modelcontextprotocol/server-filesystem

# Brave search server  
npm install -g @modelcontextprotocol/server-brave-search

# GitHub server
npm install -g @modelcontextprotocol/server-github
```

### Configure API Keys
Update `mcp.json` with your API keys:

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"],
      "env": {}
    },
    "brave-search": {
      "command": "npx", 
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "your_actual_brave_api_key"
      }
    }
  }
}
```

## Testing Without External Dependencies

You can test the proxy even without real MCP servers by:

1. **Mock Mode**: The proxy will handle missing servers gracefully
2. **Local Test**: Run `go run cmd/test/main.go` for basic functionality tests
3. **Health Check**: Test HTTP endpoints without MCP servers

## Test Scenarios

### 1. Tool Discovery
Test the LLM-powered tool selection:

```bash
# Search for file operations
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"I want to read a file"}'

# Search for web operations  
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"I need to search the internet"}'
```

### 2. Tool Execution
Test actual tool usage:

```bash
# Read a file (if filesystem server is configured)
curl -X POST http://localhost:8080/api/v1/use/read_file \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"path":"./README.md"}}'

# List directory
curl -X POST http://localhost:8080/api/v1/use/list_directory \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"path":"./"}}'
```

### 3. Tool Cache Refresh
Test dynamic tool discovery:

```bash
# Refresh tool cache
curl -X POST http://localhost:8080/api/v1/refresh
```

## Troubleshooting

### Common Issues

1. **No LLM Provider**: Tool discovery will fail, but other endpoints work
2. **MCP Server Connection Failed**: Check server installation and paths
3. **API Key Issues**: Verify environment variables and mcp.json configuration

### Debug Mode
Add logging to see detailed operations:

```bash
# Run with verbose output
go run . -config ./mcp.json -addr :8080 2>&1 | tee debug.log
```

### Test with Different Configurations

Create alternative config files:

```bash
# Test with minimal config
cp mcp.json mcp-minimal.json
# Edit to have only filesystem server

# Run with minimal config
go run . -config ./mcp-minimal.json -addr :8080
```

## Performance Testing

### Load Testing
Use `wrk` or `ab` for load testing:

```bash
# Install wrk
sudo apt install wrk

# Test health endpoint
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/health

# Test tool discovery  
wrk -t12 -c400 -d30s -s post.lua http://localhost:8080/api/v1/discover
```

### Memory Testing
Monitor memory usage during operation:

```bash
# Monitor with htop
htop

# Or use ps
watch 'ps aux | grep mcp-smart-proxy'
```

## Integration Testing

Test with actual AI clients by implementing the MCP proxy as a tool provider:

```python
# Example Python client
import requests

def test_proxy():
    # Discover tools
    response = requests.post('http://localhost:8080/api/v1/discover', 
                           json={'query': 'read files'})
    tools = response.json()['recommendedTools']
    
    # Use a tool
    if tools:
        tool_name = tools[0]['name']
        response = requests.post(f'http://localhost:8080/api/v1/use/{tool_name}',
                               json={'arguments': {'path': './test.txt'}})
        result = response.json()
        print(result)

test_proxy()
```
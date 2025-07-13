# MCP Smart Proxy

An intelligent proxy server for Model Context Protocol (MCP) that provides smart tool discovery and routing using LLM-powered selection. Solves "tool degradation" by intelligently selecting the most relevant tools from multiple MCP servers.

## 🎯 Problem Solved

When AI systems connect to multiple MCP servers with 100+ tools total, they suffer from "tool degradation" - too many options confuse the AI and reduce performance. MCP Smart Proxy solves this by:

- **Intelligent Tool Selection**: Uses LLM (OpenAI/Gemini) to select at most 5 most relevant tools
- **Unified Interface**: Single API endpoint for multiple MCP servers  
- **Dynamic Discovery**: Automatically caches and updates tools from all servers
- **Seamless Integration**: Works with existing MCP clients and servers

## 🏗️ Architecture

```
AI Client → MCP Smart Proxy → Multiple MCP Servers
              ↓
    [LLM-Powered Tool Selection]
```

Instead of seeing 100+ tools, AI clients see just a few smart endpoints that dynamically route to the best tools.

## 🚀 Quick Start

### 1. Install and Build

```bash
git clone <repository>
cd mcp-smart-proxy
make build
```

### 2. Configure MCP Servers

Create `mcp.json` configuration:

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
        "BRAVE_API_KEY": "your-api-key"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token"
      }
    }
  }
}
```

### 3. Set LLM Provider

```bash
# Choose one:
export OPENAI_API_KEY=your_openai_key
export GEMINI_API_KEY=your_gemini_key
```

### 4. Start the Proxy

```bash
# Start with your configuration
./mcp-smart-proxy -config mcp.json -addr :8080

# The proxy will:
# 1. Connect to all MCP servers
# 2. Discover and cache all tools
# 3. Start HTTP API on port 8080
```

### 5. Use the API

```bash
# Discover relevant tools for a task
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"I need to search for files containing specific text"}'

# Response: Top 5 most relevant tools ranked by LLM
{
  "recommendedTools": [
    {"name": "search_files", "description": "Search for files containing specific text", ...},
    {"name": "read_file", "description": "Read the contents of a file", ...},
    {"name": "list_directory", "description": "List files and directories", ...}
  ]
}

# Execute a specific tool
curl -X POST http://localhost:8080/api/v1/use/search_files \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"query":"TODO", "path":"/project"}}'
```

## 📖 Complete Usage Guide

### Command Line Options

```bash
./mcp-smart-proxy [options]

Options:
  -config string    Path to MCP configuration file (default "./mcp.json")
  -addr string      Address to listen on (default ":8080")

Examples:
  ./mcp-smart-proxy -config ./my-servers.json -addr :9000
  ./mcp-smart-proxy -config /etc/mcp/production.json
```

### Configuration File Format

The `-config` flag points to a JSON file that defines your MCP servers:

```json
{
  "mcpServers": {
    "server-name": {
      "command": "executable",
      "args": ["arg1", "arg2"],
      "env": {
        "ENV_VAR": "value"
      }
    }
  }
}
```

**Real Examples:**

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"],
      "env": {}
    },
    "postgres": {
      "command": "uvx",
      "args": ["mcp-server-postgres", "--connection-string", "postgresql://..."],
      "env": {
        "POSTGRES_PASSWORD": "secret"
      }
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "your-brave-key"
      }
    }
  }
}
```

### API Endpoints

#### `GET /api/v1/health`
Health check endpoint.

**Response:** `200 OK` with `"OK"`

#### `GET /api/v1/tools`
List all discovered tools from all MCP servers.

**Response:**
```json
{
  "recommendedTools": [
    {
      "name": "read_file",
      "description": "Read the contents of a file from the filesystem", 
      "inputSchema": {...},
      "serverName": "filesystem"
    }
  ]
}
```

#### `POST /api/v1/discover`
Get LLM-recommended tools for a specific query (max 5 tools).

**Request:**
```json
{
  "query": "I need to analyze database performance and find slow queries"
}
```

**Response:**
```json
{
  "recommendedTools": [
    {"name": "query_database", "description": "Execute SQL queries", "serverName": "postgres"},
    {"name": "analyze_performance", "description": "Analyze query performance", "serverName": "postgres"},
    {"name": "list_tables", "description": "List database tables", "serverName": "postgres"},
    {"name": "read_file", "description": "Read log files", "serverName": "filesystem"},
    {"name": "search_files", "description": "Search for error patterns", "serverName": "filesystem"}
  ]
}
```

#### `POST /api/v1/use/{tool}`
Execute a specific tool with arguments.

**Request:**
```json
{
  "arguments": {
    "path": "/var/log/app.log",
    "query": "ERROR"
  }
}
```

**Response:**
```json
{
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 5 errors in /var/log/app.log:\nLine 102: ERROR: Database connection failed\n..."
      }
    ]
  }
}
```

#### `POST /api/v1/refresh`
Refresh tool cache by reconnecting to all MCP servers.

**Response:** `200 OK` with `"Tools refreshed successfully"`

### LLM Provider Configuration

The proxy uses LLM providers to intelligently select tools. Configure one:

**OpenAI:**
```bash
export OPENAI_API_KEY=sk-...
```

**Google Gemini:**
```bash
export GEMINI_API_KEY=AIza...
```

**Selection Logic:**
- Returns **at most 5 tools** ranked by relevance
- Prioritizes tools that directly solve the query
- Includes supporting tools that provide context
- Maintains ranking order (most relevant first)

## 🧪 Testing

### Local Testing (No External Dependencies)

```bash
# Test with local mock server
make run-local

# Test basic functionality
make test-local

# Test with curl commands
make test-curl
```

### Full Integration Testing

```bash
# Complete AI simulation test
make test-full

# Test MCP client connectivity  
make test-ai
```

### Manual Testing

```bash
# Start server
./mcp-smart-proxy -config mcp.json -addr :8080

# Test discovery
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"send an email notification"}'

# Test execution  
curl -X POST http://localhost:8080/api/v1/use/send_email \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"to":"user@example.com","subject":"Test","body":"Hello"}}'
```

## 🔧 Development

### Project Structure

```
pkg/types/              # Public interfaces and types
internal/
├── llm/               # LLM provider implementations (OpenAI, Gemini)
├── mcp/               # MCP client protocol implementation  
├── proxy/             # Core proxy logic and tool caching
└── server/            # HTTP server and API endpoints
cmd/mcp-smart-proxy/   # Main application entry point
temp_scripts/          # Test utilities and examples
```

### Adding New LLM Providers

1. Implement the `LLMProvider` interface in `internal/llm/`
2. Add factory function to `NewProvider()`
3. Follow the 5-tool limit pattern

### Adding New Features

- **Tool Caching**: Modify `internal/proxy/proxy.go`
- **API Endpoints**: Add to `internal/server/server.go`  
- **MCP Protocol**: Extend `internal/mcp/client.go`

## 📋 Troubleshooting

### Common Issues

**"No LLM provider configured"**
```bash
# Set one of these:
export OPENAI_API_KEY=your_key
export GEMINI_API_KEY=your_key
```

**"Failed to connect to server X"**
- Check MCP server installation: `npm install -g @modelcontextprotocol/server-filesystem`
- Verify paths and arguments in `mcp.json`
- Check environment variables and API keys

**"Address already in use"**
```bash
# Kill existing processes
lsof -ti:8080 | xargs kill -9
```

**"Tool not found"**
- Use `/api/v1/tools` to see available tools
- Use `/api/v1/refresh` to reload tool cache
- Check MCP server logs for connection issues

### Debug Mode

```bash
# Run with verbose logging
./mcp-smart-proxy -config mcp.json -addr :8080 2>&1 | tee debug.log
```

### Performance Tuning

- **Tool Discovery**: 1-2 seconds with LLM selection
- **Tool Execution**: <500ms per call
- **Memory Usage**: ~50MB base + tools cache
- **Concurrent Requests**: Thread-safe operations

## 🚀 Production Deployment

### Requirements

- Go 1.21+
- Node.js (for npm-based MCP servers)
- LLM API access (OpenAI or Gemini)
- Network access to MCP servers

### Environment Setup

```bash
# Production environment
export GEMINI_API_KEY=production_key
export LOG_LEVEL=info

# Start with production config
./mcp-smart-proxy -config /etc/mcp/production.json -addr :8080
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN make build

FROM node:18-alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/mcp-smart-proxy /usr/local/bin/
COPY mcp.json /etc/mcp/
EXPOSE 8080
CMD ["mcp-smart-proxy", "-config", "/etc/mcp/mcp.json", "-addr", ":8080"]
```

### Health Monitoring

```bash
# Health check endpoint
curl http://localhost:8080/api/v1/health

# Tool availability check
curl http://localhost:8080/api/v1/tools | jq '.recommendedTools | length'
```

## 📚 Examples

### AI Assistant Integration

```python
import requests

class MCPSmartProxyClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
    
    def discover_tools(self, user_query):
        """Get relevant tools for user query"""
        response = requests.post(f"{self.base_url}/api/v1/discover",
                               json={"query": user_query})
        return response.json()["recommendedTools"]
    
    def use_tool(self, tool_name, **kwargs):
        """Execute a specific tool"""
        response = requests.post(f"{self.base_url}/api/v1/use/{tool_name}",
                               json={"arguments": kwargs})
        return response.json()["result"]

# Usage
client = MCPSmartProxyClient()

# User: "Find files containing 'TODO' in my project"
tools = client.discover_tools("Find files containing 'TODO' in my project")
# Returns: [search_files, read_file, list_directory]

result = client.use_tool("search_files", query="TODO", path="/project")
print(result)
```

### Claude Desktop Integration

Add to Claude Desktop config:

```json
{
  "mcpServers": {
    "smart-proxy": {
      "command": "node",
      "args": ["./mcp_proxy_server.js"],
      "env": {}
    }
  }
}
```

## 🤝 Contributing

1. Follow Go best practices and idioms
2. Add tests for new features
3. Update documentation
4. Ensure thread safety for concurrent operations

## 📄 License

See LICENSE file for details.
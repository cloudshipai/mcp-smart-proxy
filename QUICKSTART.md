# MCP Smart Proxy - Quick Start Guide

Get up and running with MCP Smart Proxy in 5 minutes!

## 🎯 What This Does

**Problem**: AI systems get confused when they have access to 100+ tools from multiple MCP servers  
**Solution**: MCP Smart Proxy uses LLM intelligence to show AI clients only the 5 most relevant tools for each query

## ⚡ 5-Minute Setup

### Step 1: Build
```bash
git clone <repository>
cd mcp-smart-proxy
make build
```

### Step 2: Configure LLM Provider
```bash
# Choose one:
export OPENAI_API_KEY=sk-your-key-here
export GEMINI_API_KEY=AIza-your-key-here
```

### Step 3: Create MCP Configuration

Create `mcp.json` with your MCP servers:

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
        "BRAVE_API_KEY": "your-brave-api-key"
      }
    }
  }
}
```

### Step 4: Start the Proxy
```bash
./mcp-smart-proxy -config mcp.json -addr :8080
```

**Output:**
```
2025/07/13 18:29:22 Initializing Smart Proxy...
2025/07/13 18:29:22 Connecting to server: filesystem
2025/07/13 18:29:22 Server filesystem provided 15 tools
2025/07/13 18:29:22 Connecting to server: brave-search  
2025/07/13 18:29:22 Server brave-search provided 8 tools
2025/07/13 18:29:22 Discovered 23 tools from 2 servers
2025/07/13 18:29:22 Starting server on :8080
```

### Step 5: Test It!

```bash
# Get smart tool recommendations (max 5 tools)
curl -X POST http://localhost:8080/api/v1/discover \
  -H 'Content-Type: application/json' \
  -d '{"query":"I need to search for files containing TODO items"}'

# Execute a specific tool
curl -X POST http://localhost:8080/api/v1/use/search_files \
  -H 'Content-Type: application/json' \
  -d '{"arguments":{"query":"TODO","path":"/project"}}'
```

## 🧪 Testing Without Real MCP Servers

Want to test immediately? Use our local test server:

```bash
# Test with mock server (no dependencies needed)
make run-local

# Run AI simulation
make test-full
```

This starts a local MCP server with mock tools for testing.

## 🔧 Configuration Examples

### Basic Setup (Filesystem Only)
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "env": {}
    }
  }
}
```

### Production Setup (Multiple Servers)
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx", 
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/data"],
      "env": {}
    },
    "postgres": {
      "command": "uvx",
      "args": ["mcp-server-postgres", "--connection-string", "postgresql://user:pass@localhost/db"],
      "env": {
        "POSTGRES_PASSWORD": "secret"
      }
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_your_token"
      }
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "your_brave_key"
      }
    }
  }
}
```

## 🚀 Usage Patterns

### 1. AI Assistant Integration

```python
import requests

class SmartProxy:
    def __init__(self, url="http://localhost:8080"):
        self.url = url
    
    def ask(self, user_request):
        # Get relevant tools
        tools = requests.post(f"{self.url}/api/v1/discover", 
                             json={"query": user_request}).json()
        
        # Use the most relevant tool
        if tools["recommendedTools"]:
            tool = tools["recommendedTools"][0]
            print(f"Using tool: {tool['name']}")
            # ... execute tool
```

### 2. Command Line Usage

```bash
# Start proxy
./mcp-smart-proxy -config production.json -addr :8080

# Query for relevant tools
curl -X POST localhost:8080/api/v1/discover \
  -d '{"query":"analyze database performance"}' | jq

# Execute chosen tool
curl -X POST localhost:8080/api/v1/use/query_database \
  -d '{"arguments":{"sql":"SELECT * FROM slow_queries"}}' | jq
```

### 3. Claude Desktop Integration

Add to your Claude Desktop MCP config:

```json
{
  "mcpServers": {
    "smart-proxy": {
      "command": "node",
      "args": ["./temp_scripts/mcp_proxy_server.js"],
      "env": {}
    }
  }
}
```

## 📊 What You Get

**Before MCP Smart Proxy:**
- AI sees 50+ tools from multiple servers
- Tool degradation reduces performance
- Complex setup for each AI client

**After MCP Smart Proxy:**
- AI sees max 5 relevant tools per query
- LLM-powered intelligent selection
- Single proxy endpoint for all servers

## 🔍 Key Features in Action

### Intelligent Tool Selection
```bash
# Query: "I need to fix a bug in my Python code"
# Returns: [search_files, read_file, write_file, run_tests, check_syntax]

# Query: "Send an email about the weather in Paris"  
# Returns: [send_email, get_weather, translate_text]
```

### Automatic Caching
- Tools discovered on startup
- Cached for fast response
- Refreshable via `/api/v1/refresh`

### Production Ready
- Thread-safe operations
- Graceful error handling
- Health monitoring endpoints
- Docker support

## 🆘 Troubleshooting

**"No tools discovered"**
→ Check MCP server installation and config paths

**"LLM provider not configured"** 
→ Set `OPENAI_API_KEY` or `GEMINI_API_KEY`

**"Connection refused"**
→ Verify MCP servers are accessible and API keys are valid

**"Tool not found"**
→ Use `/api/v1/tools` to see all available tools

## 🎯 Next Steps

1. **Production Setup**: Configure with your real MCP servers
2. **Integration**: Connect your AI assistant or agent framework  
3. **Monitoring**: Use health endpoints for uptime monitoring
4. **Scaling**: Deploy with Docker for production workloads

## 📚 Full Documentation

For complete documentation, API reference, and advanced configuration:
→ See [README.md](README.md)
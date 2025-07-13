// Package types provides common data structures for the MCP Smart Proxy
package types

import (
	"context"
	"time"
)

// MCPServer represents a configured MCP server
type MCPServer struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// MCPConfig represents the mcp.json configuration
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// Tool represents a tool from an MCP server
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
	ServerName  string      `json:"serverName"`
}

// ToolCache manages cached tools from all servers
type ToolCache struct {
	Tools     map[string]Tool   `json:"tools"`
	LastSync  time.Time         `json:"lastSync"`
	ServerMap map[string]string `json:"serverMap"` // tool name -> server name
}

// ProxyRequest represents a request to discover tools
type ProxyRequest struct {
	Query string `json:"query"`
}

// ToolRequest represents a request to use a tool
type ToolRequest struct {
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// ProxyResponse represents the response from the proxy
type ProxyResponse struct {
	RecommendedTools []Tool                 `json:"recommendedTools,omitempty"`
	Result           map[string]interface{} `json:"result,omitempty"`
	Error            string                 `json:"error,omitempty"`
}

// LLMProvider interface for different LLM providers
type LLMProvider interface {
	SelectBestTools(ctx context.Context, query string, availableTools []Tool) ([]Tool, error)
}

// MCPClient interface for interacting with MCP servers
type MCPClient interface {
	ListTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error)
	Close() error
}
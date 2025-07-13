// Package proxy provides the core smart proxy functionality
package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"mcp-smart-proxy/internal/llm"
	"mcp-smart-proxy/internal/mcp"
	"mcp-smart-proxy/pkg/types"
)

// SmartProxy is the main proxy server that manages MCP servers and tool selection
type SmartProxy struct {
	config      types.MCPConfig
	toolCache   *types.ToolCache
	llmProvider types.LLMProvider
	clients     map[string]types.MCPClient
	mu          sync.RWMutex
}

// New creates a new SmartProxy instance
func New(configPath string) (*SmartProxy, error) {
	// Load configuration
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config types.MCPConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize LLM provider
	llmProvider, err := llm.NewProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LLM provider: %w", err)
	}

	proxy := &SmartProxy{
		config:      config,
		toolCache:   &types.ToolCache{Tools: make(map[string]types.Tool), ServerMap: make(map[string]string)},
		llmProvider: llmProvider,
		clients:     make(map[string]types.MCPClient),
	}

	return proxy, nil
}

// Initialize discovers all tools from configured MCP servers
func (p *SmartProxy) Initialize(ctx context.Context) error {
	log.Println("Initializing Smart Proxy...")

	// Discover all tools from configured servers
	if err := p.discoverAllTools(ctx); err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	log.Printf("Discovered %d tools from %d servers", len(p.toolCache.Tools), len(p.config.MCPServers))
	return nil
}

// discoverAllTools connects to all configured MCP servers and caches their tools
func (p *SmartProxy) discoverAllTools(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for serverName, serverConfig := range p.config.MCPServers {
		log.Printf("Connecting to server: %s", serverName)

		client, err := mcp.NewStdioClient(serverConfig.Command, serverConfig.Args, serverConfig.Env)
		if err != nil {
			log.Printf("Failed to connect to server %s: %v", serverName, err)
			continue
		}

		p.clients[serverName] = client

		tools, err := client.ListTools(ctx)
		if err != nil {
			log.Printf("Failed to list tools from server %s: %v", serverName, err)
			client.Close()
			delete(p.clients, serverName)
			continue
		}

		// Cache tools
		for _, tool := range tools {
			tool.ServerName = serverName
			p.toolCache.Tools[tool.Name] = tool
			p.toolCache.ServerMap[tool.Name] = serverName
		}

		log.Printf("Server %s provided %d tools", serverName, len(tools))
	}

	p.toolCache.LastSync = time.Now()
	return nil
}

// ListTools returns all cached tools
func (p *SmartProxy) ListTools(ctx context.Context) ([]types.Tool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var tools []types.Tool
	for _, tool := range p.toolCache.Tools {
		tools = append(tools, tool)
	}

	return tools, nil
}

// DiscoverTools uses LLM to select the most relevant tools for a query
func (p *SmartProxy) DiscoverTools(ctx context.Context, query string) ([]types.Tool, error) {
	p.mu.RLock()
	allTools := make([]types.Tool, 0, len(p.toolCache.Tools))
	for _, tool := range p.toolCache.Tools {
		allTools = append(allTools, tool)
	}
	p.mu.RUnlock()

	// Use LLM to select best tools
	selectedTools, err := p.llmProvider.SelectBestTools(ctx, query, allTools)
	if err != nil {
		return nil, fmt.Errorf("failed to select tools: %w", err)
	}

	return selectedTools, nil
}

// UseTool executes a specific tool with the given arguments
func (p *SmartProxy) UseTool(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	p.mu.RLock()
	serverName, exists := p.toolCache.ServerMap[toolName]
	if !exists {
		p.mu.RUnlock()
		return nil, fmt.Errorf("tool %s not found", toolName)
	}

	client, exists := p.clients[serverName]
	if !exists {
		p.mu.RUnlock()
		return nil, fmt.Errorf("client for server %s not available", serverName)
	}
	p.mu.RUnlock()

	// Execute tool
	result, err := client.CallTool(ctx, toolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool %s: %w", toolName, err)
	}

	return result, nil
}

// RefreshTools rediscovers all tools from configured servers
func (p *SmartProxy) RefreshTools(ctx context.Context) error {
	log.Println("Refreshing tool cache...")

	// Close existing clients
	p.mu.Lock()
	for _, client := range p.clients {
		client.Close()
	}
	p.clients = make(map[string]types.MCPClient)
	p.toolCache.Tools = make(map[string]types.Tool)
	p.toolCache.ServerMap = make(map[string]string)
	p.mu.Unlock()

	// Rediscover tools
	return p.discoverAllTools(ctx)
}

// Close shuts down the proxy and all MCP clients
func (p *SmartProxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}

	return nil
}
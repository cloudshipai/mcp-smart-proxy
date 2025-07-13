package main

import (
	"fmt"
	"os"
	"time"
)

// Tool represents a tool from an MCP server (duplicate for test package)
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
	ServerName  string      `json:"serverName"`
}

// ToolCache manages cached tools from all servers (duplicate for test package)
type ToolCache struct {
	Tools     map[string]Tool   `json:"tools"`
	LastSync  time.Time         `json:"lastSync"`
	ServerMap map[string]string `json:"serverMap"`
}

func main() {
	fmt.Println("=== MCP Smart Proxy Local Test ===")
	
	// Test environment setup
	fmt.Println("1. Testing environment setup...")
	
	// Check for LLM provider
	if os.Getenv("OPENAI_API_KEY") == "" && os.Getenv("GEMINI_API_KEY") == "" {
		fmt.Println("⚠️  No LLM provider configured. Set OPENAI_API_KEY or GEMINI_API_KEY for full testing")
		fmt.Println("   You can still test basic functionality without LLM provider")
	} else {
		fmt.Println("✅ LLM provider found")
	}
	
	// Test tool cache functionality
	fmt.Println("\n2. Testing tool cache...")
	cache := &ToolCache{
		Tools:     make(map[string]Tool),
		ServerMap: make(map[string]string),
		LastSync:  time.Now(),
	}
	
	// Add mock tools
	mockTool := Tool{
		Name:        "test_tool",
		Description: "A test tool for local testing",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "string",
				},
			},
		},
		ServerName: "mock-server",
	}
	
	cache.Tools["test_tool"] = mockTool
	cache.ServerMap["test_tool"] = "mock-server"
	
	fmt.Printf("✅ Tool cache populated with %d tools\n", len(cache.Tools))
	
	fmt.Println("\n3. Test commands to run manually:")
	fmt.Println("   Start the server: go run .")
	fmt.Println("   Test health: curl http://localhost:8080/api/v1/health")
	fmt.Println("   List tools: curl http://localhost:8080/api/v1/tools")
	fmt.Println("   Discover tools: curl -X POST http://localhost:8080/api/v1/discover -H 'Content-Type: application/json' -d '{\"query\":\"I need to search files\"}'")
	
	fmt.Println("\n=== Local Test Complete ===")
}
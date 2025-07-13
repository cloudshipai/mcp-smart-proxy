// Package mcp provides MCP client implementations
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"mcp-smart-proxy/pkg/types"
)

// StdioClient implements MCPClient using stdio protocol
type StdioClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Scanner
}

// NewStdioClient creates a new MCP client using stdio protocol
func NewStdioClient(command string, args []string, env map[string]string) (*StdioClient, error) {
	cmd := exec.Command(command, args...)

	// Set environment variables
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	client := &StdioClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewScanner(stdout),
	}

	// Initialize MCP connection
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// initialize sends the MCP initialize request
func (c *StdioClient) initialize() error {
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "mcp-smart-proxy",
				"version": "1.0.0",
			},
		},
	}

	if err := c.sendRequest(initReq); err != nil {
		return err
	}

	// Read and discard the initialize response
	_, err := c.readResponse()
	return err
}

// sendRequest sends a JSON-RPC request to the MCP server
func (c *StdioClient) sendRequest(req map[string]interface{}) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = c.stdin.Write(append(data, '\n'))
	return err
}

// readResponse reads a JSON-RPC response from the MCP server
func (c *StdioClient) readResponse() (map[string]interface{}, error) {
	if !c.reader.Scan() {
		return nil, fmt.Errorf("failed to read response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal(c.reader.Bytes(), &response); err != nil {
		return nil, err
	}

	return response, nil
}

// ListTools retrieves all available tools from the MCP server
func (c *StdioClient) ListTools(ctx context.Context) ([]types.Tool, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	if err := c.sendRequest(req); err != nil {
		return nil, err
	}

	response, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: %v", response)
	}

	toolsData, ok := result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no tools in response: %v", result)
	}

	var tools []types.Tool
	for _, toolData := range toolsData {
		toolMap, ok := toolData.(map[string]interface{})
		if !ok {
			continue
		}

		tool := types.Tool{
			Name:        getString(toolMap, "name"),
			Description: getString(toolMap, "description"),
			InputSchema: toolMap["inputSchema"],
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool executes a tool on the MCP server
func (c *StdioClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	if err := c.sendRequest(req); err != nil {
		return nil, err
	}

	response, err := c.readResponse()
	if err != nil {
		return nil, err
	}

	if errorData, exists := response["error"]; exists {
		return nil, fmt.Errorf("tool error: %v", errorData)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	return result, nil
}

// Close closes the MCP client and terminates the server process
func (c *StdioClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// getString safely extracts a string value from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
// Package llm provides LLM provider implementations for tool selection
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"mcp-smart-proxy/pkg/types"

	"github.com/sashabaranov/go-openai"
	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// OpenAIProvider implements LLMProvider using OpenAI's API
type OpenAIProvider struct {
	client *openai.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	client := openai.NewClient(apiKey)
	return &OpenAIProvider{client: client}
}

// SelectBestTools selects the most relevant tools using OpenAI
func (p *OpenAIProvider) SelectBestTools(ctx context.Context, query string, availableTools []types.Tool) ([]types.Tool, error) {
	toolsJSON, _ := json.Marshal(availableTools)

	prompt := fmt.Sprintf(`You are a tool selection expert. Given the user query and available tools, select the most relevant tools that would help answer the query.

RULES:
- Select AT MOST 5 tools 
- Rank them by relevance (most relevant first)
- Include tools that could directly solve the query
- Include tools that could provide supporting information
- Always prioritize quality over quantity

User Query: %s

Available Tools:
%s

Return a JSON array of tool names only, ranked by relevance. Example: ["most_relevant", "second_choice", "supporting_tool"]`,
		query, string(toolsJSON))

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 200,
	})

	if err != nil {
		return nil, err
	}

	var selectedNames []string
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &selectedNames); err != nil {
		return nil, err
	}

	return filterToolsByNames(selectedNames, availableTools), nil
}

// GeminiProvider implements LLMProvider using Google's Gemini API
type GeminiProvider struct {
	client *genai.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &GeminiProvider{client: client}, nil
}

// SelectBestTools selects the most relevant tools using Gemini
func (p *GeminiProvider) SelectBestTools(ctx context.Context, query string, availableTools []types.Tool) ([]types.Tool, error) {
	model := p.client.GenerativeModel("gemini-pro")

	toolsJSON, _ := json.Marshal(availableTools)
	prompt := fmt.Sprintf(`You are a tool selection expert. Given the user query and available tools, select the most relevant tools that would help answer the query.

RULES:
- Select AT MOST 5 tools
- Rank them by relevance (most relevant first) 
- Include tools that could directly solve the query
- Include tools that could provide supporting information
- Always prioritize quality over quantity

User Query: %s

Available Tools:
%s

Return only a JSON array of tool names, ranked by relevance. Example: ["most_relevant", "second_choice", "supporting_tool"]`,
		query, string(toolsJSON))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	content := resp.Candidates[0].Content.Parts[0]
	var selectedNames []string
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", content)), &selectedNames); err != nil {
		return nil, err
	}

	return filterToolsByNames(selectedNames, availableTools), nil
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	return p.client.Close()
}

// NewProvider creates an LLM provider based on environment variables
func NewProvider() (types.LLMProvider, error) {
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return NewOpenAIProvider(apiKey), nil
	}

	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		return NewGeminiProvider(apiKey)
	}

	return nil, fmt.Errorf("no LLM provider configured. Set OPENAI_API_KEY or GEMINI_API_KEY")
}

// filterToolsByNames filters tools by their names and limits to max 5 tools
func filterToolsByNames(selectedNames []string, availableTools []types.Tool) []types.Tool {
	var selectedTools []types.Tool
	toolMap := make(map[string]types.Tool)
	for _, tool := range availableTools {
		toolMap[tool.Name] = tool
	}

	// Limit to at most 5 tools
	maxTools := 5
	if len(selectedNames) > maxTools {
		selectedNames = selectedNames[:maxTools]
	}

	for _, name := range selectedNames {
		if tool, exists := toolMap[name]; exists {
			selectedTools = append(selectedTools, tool)
		}
	}

	return selectedTools
}
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	ollamaapi "github.com/ollama/ollama/api"
)

// Client encapsula o transporte stdio do mcp-go.
// O servidor MCP é spawned como processo filho; Close() o finaliza corretamente.
type Client struct {
	inner *mcpclient.Client // <-- CORREÇÃO APLICADA AQUI
}

// NewClient spawna o servidor MCP como processo filho e executa o handshake do protocolo.
func NewClient(ctx context.Context, cmd string, args []string) (*Client, error) {
	c, err := mcpclient.NewStdioMCPClient(cmd, nil, args...)
	if err != nil {
		return nil, fmt.Errorf("create stdio mcp client: %w", err)
	}

	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: "host-api", Version: "1.0.0"}

	if _, err = c.Initialize(ctx, req); err != nil {
		c.Close()
		return nil, fmt.Errorf("mcp handshake: %w", err)
	}

	return &Client{inner: c}, nil
}

// Close termina o processo filho do servidor MCP. Evita processos zumbi no Docker.
func (c *Client) Close() {
	c.inner.Close()
}

// CallTool invoca uma ferramenta no servidor MCP com os argumentos fornecidos.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (string, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = arguments

	result, err := c.inner.CallTool(ctx, req)
	if err != nil {
		return "", fmt.Errorf("mcp call tool %q: %w", name, err)
	}
	if result.IsError {
		return "", fmt.Errorf("mcp tool %q returned an error state", name)
	}

	return extractText(result), nil
}

// ListTools consulta o servidor MCP e converte as ferramentas para o formato ollamaapi.Tool.
func (c *Client) ListTools(ctx context.Context) ([]ollamaapi.Tool, error) {
	result, err := c.inner.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("mcp list tools: %w", err)
	}
	return convertTools(result.Tools)
}

// extractText extrai conteúdo de texto de um CallToolResult via round-trip JSON.
func extractText(result *mcp.CallToolResult) string {
	data, err := json.Marshal(result.Content)
	if err != nil {
		return ""
	}
	var items []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, item := range items {
		if item.Type == "text" {
			sb.WriteString(item.Text)
		}
	}
	return sb.String()
}

// convertTools traduz []mcp.Tool para []ollamaapi.Tool via round-trip JSON.
// Evita dependência direta nos tipos anônimos de struct do Ollama SDK.
func convertTools(mcpTools []mcp.Tool) ([]ollamaapi.Tool, error) {
	type intermediate struct {
		Type     string `json:"type"`
		Function struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Parameters  struct {
				Type       string                 `json:"type"`
				Required   []string               `json:"required,omitempty"`
				Properties map[string]interface{} `json:"properties"`
			} `json:"parameters"`
		} `json:"function"`
	}

	tools := make([]ollamaapi.Tool, 0, len(mcpTools))
	for _, t := range mcpTools {
		var im intermediate
		im.Type = "function"
		im.Function.Name = t.Name
		im.Function.Description = t.Description
		im.Function.Parameters.Type = t.InputSchema.Type
		im.Function.Parameters.Required = t.InputSchema.Required
		im.Function.Parameters.Properties = t.InputSchema.Properties

		data, err := json.Marshal(im)
		if err != nil {
			return nil, fmt.Errorf("marshal tool %q: %w", t.Name, err)
		}
		var tool ollamaapi.Tool
		if err := json.Unmarshal(data, &tool); err != nil {
			return nil, fmt.Errorf("unmarshal tool %q: %w", t.Name, err)
		}
		tools = append(tools, tool)
	}
	return tools, nil
}
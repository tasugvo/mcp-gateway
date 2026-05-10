package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	ollamaapi "github.com/ollama/ollama/api"
)

type Client struct {
	api   *ollamaapi.Client
	model string
}

func NewClient(baseURL, model string) *Client {
	u, _ := url.Parse(baseURL)
	return &Client{
		api:   ollamaapi.NewClient(u, http.DefaultClient),
		model: model,
	}
}

// Chat envia o histórico completo + ferramentas ao modelo.
// Stream desabilitado: bloqueia até a resposta completa.
func (c *Client) Chat(
	ctx context.Context,
	messages []ollamaapi.Message,
	tools []ollamaapi.Tool,
) (*ollamaapi.Message, error) {
	streamOff := false
	req := &ollamaapi.ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   &streamOff,
	}

	var response ollamaapi.Message
	err := c.api.Chat(ctx, req, func(resp ollamaapi.ChatResponse) error {
		if resp.Done {
			response = resp.Message
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ollama chat: %w", err)
	}
	return &response, nil
}
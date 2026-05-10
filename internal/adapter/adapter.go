// Padrão: Adapter.
// Traduz []ollamaapi.ToolCall (formato Ollama) em chamadas JSON-RPC do mcp-go.
// Múltiplos tool_calls são executados em paralelo via goroutines + channel.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"loud-host/internal/mcp"
	ollamaapi "github.com/ollama/ollama/api"
)

type Adapter struct {
	mcp    *mcp.Client
	logger *slog.Logger
}

func New(mcpClient *mcp.Client, logger *slog.Logger) *Adapter {
	return &Adapter{mcp: mcpClient, logger: logger}
}

type toolResult struct {
	index   int
	content string
	err     error
}

// Execute despacha todos os ToolCalls em goroutines paralelas.
// Resultados são coletados via channel e remontados na ordem original.
func (a *Adapter) Execute(ctx context.Context, calls []ollamaapi.ToolCall) ([]ollamaapi.Message, error) {
	resultCh := make(chan toolResult, len(calls))
	var wg sync.WaitGroup

	for i, call := range calls {
		wg.Add(1)
		go func(idx int, tc ollamaapi.ToolCall) {
			defer wg.Done()
			a.logger.Info("dispatching mcp tool",
				"index", idx,
				"tool", tc.Function.Name,
				"args", tc.Function.Arguments,
			)

			// CORREÇÃO DEFINITIVA: JSON Round-Trip
			// Transforma o struct rígido do Ollama em JSON e converte para um map genérico
			var args map[string]any
			argBytes, _ := json.Marshal(tc.Function.Arguments)
			json.Unmarshal(argBytes, &args)

			content, err := a.mcp.CallTool(ctx, tc.Function.Name, args)

			resultCh <- toolResult{index: idx, content: content, err: err}
		}(i, call)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	ordered := make([]toolResult, len(calls))
	for r := range resultCh {
		ordered[r.index] = r
	}

	messages := make([]ollamaapi.Message, 0, len(calls))
	for _, r := range ordered {
		if r.err != nil {
			return nil, fmt.Errorf("tool[%d] execution failed: %w", r.index, r.err)
		}
		messages = append(messages, ollamaapi.Message{
			Role:    "tool",
			Content: r.content,
		})
	}

	return messages, nil
}
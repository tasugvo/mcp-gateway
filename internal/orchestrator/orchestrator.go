// Padrão: Mediator.
// Ollama e MCP não se conhecem. Todo roteamento e gerenciamento de estado
// flui exclusivamente por este struct. Implementa o ciclo de 6 passos do documento.
package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"loud-host/internal/adapter"
	"loud-host/internal/conversation"
	"loud-host/internal/mcp"
	"loud-host/internal/ollama"
	ollamaapi "github.com/ollama/ollama/api"
)

type Orchestrator struct {
	ollama  *ollama.Client
	adapter *adapter.Adapter
	state   *conversation.State
	tools   []ollamaapi.Tool
	logger  *slog.Logger
}

// New inicializa o Orchestrator e carrega o catálogo de ferramentas do servidor MCP.
func New(ctx context.Context, ollamaClient *ollama.Client, mcpClient *mcp.Client, logger *slog.Logger) (*Orchestrator, error) {
	tools, err := mcpClient.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch mcp tool list: %w", err)
	}
	logger.Info("mcp tools loaded", "count", len(tools))

	return &Orchestrator{
		ollama:  ollamaClient,
		adapter: adapter.New(mcpClient, logger),
		state:   conversation.New(),
		tools:   tools,
		logger:  logger,
	}, nil
}

// Run inicia o loop REPL, lendo input do usuário via stdin.
func (o *Orchestrator) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprint(os.Stdout, "> ")

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			fmt.Fprint(os.Stdout, "> ")
			continue
		}

		if err := o.handleTurn(ctx, input); err != nil {
			o.logger.Error("turn failed", "error", err)
		}

		fmt.Fprint(os.Stdout, "> ")
	}

	return scanner.Err()
}

// handleTurn executa um ciclo completo usuário→assistente.
// Loop interno: repete inferência até o modelo retornar texto sem tool_calls.
func (o *Orchestrator) handleTurn(ctx context.Context, userInput string) error {
	// Passo 2: State Update (User)
	o.state.Append(ollamaapi.Message{Role: "user", Content: userInput})

	for {
		// Passo 3: Inference Request
		o.logger.Info("inference request",
			"message_count", o.state.Len(),
			"tools_available", len(o.tools),
		)

		msg, err := o.ollama.Chat(ctx, o.state.Snapshot(), o.tools)
		if err != nil {
			return fmt.Errorf("ollama inference: %w", err)
		}

		o.logger.Info("inference response",
			"has_tool_calls", len(msg.ToolCalls) > 0,
			"tool_call_count", len(msg.ToolCalls),
		)

		// Passo 4: Interceptação
		if len(msg.ToolCalls) == 0 {
			// Estado terminal: modelo retornou texto puro ao usuário.
			o.state.Append(*msg)
			fmt.Fprintln(os.Stdout, msg.Content)
			return nil
		}

		// Tool call detectado: pausa interação com usuário.
		o.state.Append(*msg)

		// Passo 5: MCP Execution (paralelo via goroutines)
		toolMessages, err := o.adapter.Execute(ctx, msg.ToolCalls)
		if err != nil {
			return fmt.Errorf("adapter execute: %w", err)
		}

		// Passo 6: Context Injection
		for _, tm := range toolMessages {
			o.state.Append(tm)
		}
		o.logger.Info("context injected", "tool_result_count", len(toolMessages))

		// Reinicia o loop: nova inferência com o contexto enriquecido.
	}
}
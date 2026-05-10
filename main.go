package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"loud-host/config"
	"loud-host/internal/mcp"
	"loud-host/internal/ollama"
	"loud-host/internal/orchestrator"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mcpClient, err := mcp.NewClient(ctx, cfg.MCPServerCmd, cfg.MCPServerArgs)
	if err != nil {
		slog.Error("mcp client init failed", "error", err)
		os.Exit(1)
	}
	defer mcpClient.Close()

	ollamaClient := ollama.NewClient(cfg.OllamaBaseURL, cfg.OllamaModel)

	orch, err := orchestrator.New(ctx, ollamaClient, mcpClient, logger)
	if err != nil {
		slog.Error("orchestrator init failed", "error", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("shutdown signal received", "signal", sig.String())
		cancel()
		mcpClient.Close()
		os.Exit(0)
	}()

	if err := orch.Run(ctx); err != nil {
		slog.Error("orchestrator terminated with error", "error", err)
		os.Exit(1)
	}
}
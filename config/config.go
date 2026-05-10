package config

import (
	"os"
	"strings"
)

type Config struct {
	OllamaBaseURL string
	OllamaModel   string
	MCPServerCmd  string
	MCPServerArgs []string
}

func Load() *Config {
	rawArgs := getEnv("MCP_SERVER_ARGS", "-y @modelcontextprotocol/server-filesystem /data/knowledge")
	return &Config{
		OllamaBaseURL: getEnv("OLLAMA_BASE_URL", "http://host.docker.internal:11434"),
		OllamaModel:   getEnv("OLLAMA_MODEL", "llama3.2"),
		MCPServerCmd:  getEnv("MCP_SERVER_CMD", "npx"),
		MCPServerArgs: strings.Fields(rawArgs),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
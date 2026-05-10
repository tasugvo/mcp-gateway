FROM golang:1.26-bookworm

# Instala Node.js e NPM para o servidor MCP
RUN apt-get update && apt-get install -y curl \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Instala o servidor de arquivos do MCP
RUN npm install -g @modelcontextprotocol/server-filesystem

WORKDIR /app

# Mantém o container aberto para aceitar comandos externos
CMD ["tail", "-f", "/dev/null"]

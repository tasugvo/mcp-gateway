# Host MCP

Host para servidores MCP (Model Context Protocol) — camada intermediária entre clientes de IA e múltiplos MCP Servers, centralizando conexão, descoberta de ferramentas e orquestração de chamadas.

---

## O que é MCP?

O Model Context Protocol (MCP) é um protocolo que permite que modelos de IA se conectem a ferramentas externas de forma padronizada.

Um MCP Server pode expor funcionalidades como:

- acesso a banco de dados;
- leitura de arquivos;
- integração com APIs;
- automações;
- execução de ferramentas customizadas.

---

## Objetivo do projeto

Este projeto implementa um Host MCP responsável por:

- conectar múltiplos MCP Servers;
- centralizar comunicação;
- gerenciar chamadas;
- servir como ponto único de integração para clientes LLM.

A proposta é manter uma arquitetura leve, simples e extensível.

---

## Arquitetura

O projeto atua nas camadas de:

- Integração
- Orquestração
- Infraestrutura para IA

O padrão de orquestração adotado é o **Mediator**: o modelo de linguagem e os MCP Servers não se comunicam diretamente. Todo roteamento, gerenciamento de estado e execução de ferramentas fluem exclusivamente pelo Orchestrator, que implementa o ciclo de inferência → interceptação → execução MCP → injeção de contexto.

Fluxo simplificado:

```text
Cliente / LLM
       │
       ▼
   Host MCP
       │
 ┌─────┼─────┐
 ▼     ▼     ▼
MCP   MCP   MCP
Server Server Server
```

### Stack

| Componente | Tecnologia |
|---|---|
| Runtime | Go 1.26 |
| Integração MCP | `mcp-go` |
| Modelo local | Ollama (`llama3.2` por padrão) |
| Infraestrutura | Docker + Docker Compose |

### Configuração via variáveis de ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://host.docker.internal:11434` | Endereço da instância Ollama |
| `OLLAMA_MODEL` | `llama3.2` | Modelo a ser utilizado |
| `MCP_SERVER_CMD` | `npx` | Comando de inicialização do MCP Server |
| `MCP_SERVER_ARGS` | `-y @modelcontextprotocol/server-filesystem /data/knowledge` | Argumentos do servidor MCP |

---

## Roadmap

### Correções e melhorias imediatas

- [ ] **Leitura de arquivos via MCP** — corrigir limitação atual em que o modelo apenas lista diretórios, sem acesso à leitura individual ou múltipla de arquivos para composição de respostas
- [ ] **Refatoração geral do código** — revisar estrutura de pacotes, remover acoplamentos desnecessários e aplicar boas práticas idiomáticas em Go

### Interface e usabilidade

- [ ] **Camada de View para CLI** — interface interativa via linha de comando para operação direta do projeto sem necessidade de integração externa

### Configuração e governança

- [ ] **Configuração parametrizável do serviço:**
  - Governança de uso dos modelos (limites, políticas de acesso)
  - Configuração de integração com modelos locais (ex: endereço do Ollama)
  - Whitelist e blacklist de IPs
  - Governança de tools por modelo ou por camada de negócio

### Observabilidade

- [ ] **Modo desenvolvedor:**
  - Verbose de prompt e requisição
  - Inspeção do ciclo de inferência em tempo real
- [ ] **Sistema de logs otimizado** — registro estruturado para controle de tráfego, uso de tools e acesso a arquivos MCP

### Integrações

- [ ] Multi-integração a outros serviços MCP externos
- [ ] Integração com internet (consultas via mecanismos de busca)
- [ ] Integração a bancos de dados relacionais e não-relacionais
- [ ] Integração a APIs públicas
- [ ] Orquestração de modelos locais (Ollama) e cloud (OpenAI, Anthropic, etc.)

### Modos de operação

- [ ] Modos configuráveis de operação (ex: modo econômico, modo de alta performance)

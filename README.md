# Host MCP

Projeto simples de Host para servidores MCP (Model Context Protocol).

O objetivo deste projeto é atuar como uma camada intermediária entre clientes de IA e múltiplos MCP Servers, centralizando conexão, descoberta de ferramentas e orquestração de chamadas.

Repositório oficial:
https://github.com/tasugvo/host-mcp

---

# O que é MCP?

O Model Context Protocol (MCP) é um protocolo que permite que modelos de IA se conectem a ferramentas externas de forma padronizada.

Um MCP Server pode expor funcionalidades como:

- acesso a banco de dados;
- leitura de arquivos;
- integração com APIs;
- automações;
- execução de ferramentas customizadas.

---

# Objetivo do projeto

Este projeto implementa um Host MCP simples responsável por:

- conectar múltiplos MCP Servers;
- centralizar comunicação;
- gerenciar chamadas;
- servir como ponto único de integração para clientes LLM.

A proposta é manter uma arquitetura leve, simples e extensível.

---

# Arquitetura

O projeto atua principalmente nas camadas de:

- Integração
- Orquestração
- Infraestrutura para IA

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

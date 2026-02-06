# LLMProxy - High-Performance LLM Reverse Proxy

[English](#english) | [中文](#中文)

---

## English

### Overview

LLMProxy is a simple reverse proxy for Large Language Model (LLM) APIs. It provides load balancing, failover, and unified API endpoints for multiple LLM providers.

**Note**: When using this proxy, you can fill the API key with any value. The proxy will forward your actual provider API keys internally.

### Features

- Multi-backend support (DeepSeek, Poe, VolcEngine, etc.)
- Load balancing: `random` (weighted) or `round-robin` modes
- Automatic failover & retry
- Streaming support
- CORS enabled
- Simple YAML configuration

### Quick Start

1. **Clone & Build**
   ```bash
   git clone https://github.com/piaoranyc/llmproxy.git
   cd llmproxy
   go build -o llmproxy.exe
   ```

2. **Configure** - Edit `config.yaml` with your API keys:
   ```yaml
   server:
     port: 8080

   timeout: 180s
   retry: 3
   mode: random  # random or round-robin

   backends:
     - name: deepseek
       url: https://api.deepseek.com
       api_key: your-api-key-here
       weight: 5
       default_model: deepseek-chat
       models:
         - deepseek-chat
         - deepseek-reasoner

     - name: poe
       url: https://api.poe.com/v1
       api_key: your-poe-key
       weight: 5
       default_model: glm-4.7
       models:
         - glm-4.7
         - gpt-5.2-instant
         - claude-opus-4.5
   ```

3. **Run**
   ```bash
   ./llmproxy
   ```

4. **Test**
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "deepseek-chat",
       "messages": [{"role": "user", "content": "Hello!"}]
     }'
   ```

### Configuration

| Field | Description | Default |
|-------|-------------|---------|
| `server.port` | Proxy port | `8080` |
| `timeout` | Request timeout | `180s` |
| `retry` | Retry count on failure | `3` |
| `mode` | Load balancing: `random` or `round-robin` | `random` |
| `backends[].api_key` | **Provider API key** (required) | – |
| `backends[].weight` | Weight for random selection | `1` |
| `backends[].default_model` | Default model for this backend | – |
| `backends[].models` | Available models | – |

### API Endpoints

- `POST /v1/chat/completions` - Chat completion (OpenAI-compatible)
- `GET /v1/models` - List all available models
- `GET /health` - Health check

### License

MIT License. See [LICENSE](LICENSE).

---

## 中文

### 概述

LLMProxy 是一个简单的大型语言模型（LLM）API 反向代理，提供负载均衡、故障转移和统一的 API 端点。

**注意**：使用此代理时，API 密钥可以随便填写。代理会在内部使用你配置的提供商 API 密钥。

### 功能特性

- 支持多个后端（DeepSeek、Poe、火山引擎等）
- 负载均衡：`random`（加权随机）或 `round-robin`（轮询）
- 自动故障转移与重试
- 流式响应支持
- CORS 支持
- 简单的 YAML 配置

### 快速开始

1. **克隆与编译**
   ```bash
   git clone https://github.com/piaoranyc/llmproxy.git
   cd llmproxy
   go build -o llmproxy.exe
   ```

2. **配置** - 编辑 `config.yaml` 填入你的 API 密钥：
   ```yaml
   server:
     port: 8080

   timeout: 180s
   retry: 3
   mode: random  # random 或 round-robin

   backends:
     - name: deepseek
       url: https://api.deepseek.com
       api_key: 你的-api-key
       weight: 5
       default_model: deepseek-chat
       models:
         - deepseek-chat
         - deepseek-reasoner

     - name: poe
       url: https://api.poe.com/v1
       api_key: 你的-poe-key
       weight: 5
       default_model: glm-4.7
       models:
         - glm-4.7
         - gpt-5.2-instant
         - claude-opus-4.5
   ```

3. **运行**
   ```bash
   ./llmproxy
   ```

4. **测试**
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "deepseek-chat",
       "messages": [{"role": "user", "content": "你好！"}]
     }'
   ```

### 配置说明

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `server.port` | 代理端口 | `8080` |
| `timeout` | 请求超时时间 | `180s` |
| `retry` | 失败重试次数 | `3` |
| `mode` | 负载均衡模式：`random` 或 `round-robin` | `random` |
| `backends[].api_key` | **提供商的 API 密钥**（必填） | – |
| `backends[].weight` | 随机选择的权重 | `1` |
| `backends[].default_model` | 该后端的默认模型 | – |
| `backends[].models` | 可用模型列表 | – |

### API 端点

- `POST /v1/chat/completions` - 聊天补全（兼容 OpenAI）
- `GET /v1/models` - 获取所有可用模型
- `GET /health` - 健康检查

### 许可证

MIT 许可证。详见 [LICENSE](LICENSE)。

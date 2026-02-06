# LLMProxy - High-Performance LLM Reverse Proxy

[English](#english) | [中文](#中文)

---

## English

### Overview

LLMProxy is a simple, high-performance reverse proxy for Large Language Model (LLM) APIs. It provides load balancing, failover, and unified API endpoints for multiple LLM providers, allowing you to seamlessly switch between different AI models and services.

### Features

- **Multi‑Backend Support**: Connect to multiple LLM providers simultaneously (DeepSeek, Poe, VolcEngine, etc.)
- **Load Balancing**: Choose between `random` (weighted) and `round‑robin` distribution modes
- **Automatic Failover & Retry**: If a backend fails, the proxy automatically retries with another backend
- **Model Mapping**: Map generic model names to provider‑specific models; fallback to default models when needed
- **Streaming Support**: Full compatibility with streaming responses (`text/event‑stream`)
- **CORS Enabled**: Ready for browser‑based applications
- **Health Check Endpoint**: Monitor proxy status and backend count
- **Simple Configuration**: YAML‑based configuration with clear defaults

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/piaoranyc/llmproxy.git
   cd llmproxy
   ```

2. **Install Go** (if not already installed)
   - Download from [golang.org](https://golang.org/dl/)

3. **Build the proxy**
   ```bash
   go build -o llmproxy.exe   # Windows
   # or
   go build -o llmproxy       # Linux/macOS
   ```

4. **Configure your backends**
   Edit `config.yaml` with your API keys and endpoints (see example below).

5. **Run the proxy**
   ```bash
   ./llmproxy
   # or specify a custom config file
   ./llmproxy -c /path/to/config.yaml
   ```

6. **Test the proxy**
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "deepseek-chat",
       "messages": [{"role": "user", "content": "Hello!"}],
       "stream": false
     }'
   ```

### Configuration

Example `config.yaml`:

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

#### Configuration Fields

| Field | Description | Default |
|-------|-------------|---------|
| `server.port` | Port the proxy listens on | `8080` |
| `timeout` | HTTP request timeout | `180s` |
| `retry` | Number of retries on failure | `3` |
| `mode` | Load‑balancing mode: `random` (weighted) or `round‑robin` | `random` |
| `backends[].name` | Backend identifier (for logging) | – |
| `backends[].url` | Base URL of the LLM API | – |
| `backends[].api_key` | API key for authentication | – |
| `backends[].weight` | Weight for random selection (higher = more traffic) | `1` |
| `backends[].default_model` | Model to use when none is specified | First model in `models` |
| `backends[].models` | List of models supported by this backend | – |

### API Endpoints

#### `POST /v1/chat/completions`
Main chat completion endpoint compatible with OpenAI‑style requests.

**Request Body** (JSON):
```json
{
  "model": "deepseek-chat",  // or "auto" to let proxy choose
  "messages": [...],
  "stream": false
}
```

**Response**: Same as the underlying LLM provider's response.

#### `GET /v1/models`
Returns a list of all available models across all backends.

#### `GET /health`
Health check endpoint returning proxy status and backend count.

### Load‑Balancing Modes

- **`random` (default)**: Weighted random selection. Backends with higher `weight` receive more traffic.
- **`round‑robin`**: Strict round‑robin rotation, ignoring weights.

### Model Resolution

1. If `model` is empty or `"auto"`, the proxy picks a backend according to the load‑balancing mode and uses its `default_model`.
2. If a specific model is requested, the proxy searches all backends for that model name.
3. If the model is not found, a random backend is selected and its `default_model` is used (with a warning log).

### Development

#### Prerequisites
- Go 1.20+ (global random seeding is automatically handled)

#### Building from Source
```bash
go mod download
go build -o llmproxy
```

#### Running Tests
```bash
go test ./...
```

### License

MIT License. See [LICENSE](LICENSE) for details.

---

## 中文

### 概述

LLMProxy 是一个简单、高性能的大型语言模型（LLM）API 反向代理。它为多个 LLM 提供商提供负载均衡、故障转移和统一的 API 端点，让你可以在不同的 AI 模型和服务之间无缝切换。

### 功能特性

- **多后端支持**：同时连接多个 LLM 提供商（DeepSeek、Poe、火山引擎等）
- **负载均衡**：支持 `random`（加权随机）和 `round‑robin`（轮询）两种分发模式
- **自动故障转移与重试**：若某个后端失败，代理会自动尝试其他后端
- **模型映射**：将通用模型名称映射到提供商特定的模型；必要时回退到默认模型
- **流式响应支持**：完全兼容流式响应（`text/event‑stream`）
- **CORS 支持**：可直接用于浏览器应用
- **健康检查端点**：监控代理状态和后端数量
- **简单配置**：基于 YAML 的配置文件，提供清晰的默认值

### 快速开始

1. **克隆仓库**
   ```bash
   git clone https://github.com/piaoranyc/llmproxy.git
   cd llmproxy
   ```

2. **安装 Go**（如果尚未安装）
   - 从 [golang.org](https://golang.org/dl/) 下载

3. **编译代理**
   ```bash
   go build -o llmproxy.exe   # Windows
   # 或
   go build -o llmproxy       # Linux/macOS
   ```

4. **配置后端**
   编辑 `config.yaml`，填入你的 API 密钥和端点（参见下方示例）。

5. **运行代理**
   ```bash
   ./llmproxy
   # 或指定自定义配置文件
   ./llmproxy -c /path/to/config.yaml
   ```

6. **测试代理**
   ```bash
   curl http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "deepseek-chat",
       "messages": [{"role": "user", "content": "你好！"}],
       "stream": false
     }'
   ```

### 配置

示例 `config.yaml`：

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

#### 配置字段说明

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `server.port` | 代理监听的端口 | `8080` |
| `timeout` | HTTP 请求超时时间 | `180s` |
| `retry` | 失败时的重试次数 | `3` |
| `mode` | 负载均衡模式：`random`（加权随机）或 `round‑robin`（轮询） | `random` |
| `backends[].name` | 后端标识（用于日志） | – |
| `backends[].url` | LLM API 的基础 URL | – |
| `backends[].api_key` | 用于认证的 API 密钥 | – |
| `backends[].weight` | 随机选择的权重（值越高，流量越多） | `1` |
| `backends[].default_model` | 未指定模型时使用的模型 | `models` 列表中的第一个 |
| `backends[].models` | 该后端支持的模型列表 | – |

### API 端点

#### `POST /v1/chat/completions`
主聊天补全端点，兼容 OpenAI 风格的请求。

**请求体**（JSON）：
```json
{
  "model": "deepseek-chat",  // 或 "auto" 让代理自动选择
  "messages": [...],
  "stream": false
}
```

**响应**：与底层 LLM 提供商的响应相同。

#### `GET /v1/models`
返回所有后端中所有可用模型的列表。

#### `GET /health`
健康检查端点，返回代理状态和后端数量。

### 负载均衡模式

- **`random`（默认）**：加权随机选择。`weight` 越高的后端接收的流量越多。
- **`round‑robin`**：严格的轮询选择，忽略权重。

### 模型解析逻辑

1. 如果 `model` 为空或为 `"auto"`，代理根据负载均衡模式选择一个后端，并使用该后端的 `default_model`。
2. 如果指定了具体的模型名称，代理在所有后端中搜索该模型。
3. 如果未找到该模型，则随机选择一个后端并使用其 `default_model`（同时记录警告日志）。

### 开发

#### 环境要求
- Go 1.20+（全局随机种子已自动处理）

#### 从源码编译
```bash
go mod download
go build -o llmproxy
```

#### 运行测试
```bash
go test ./...
```

### 许可证

MIT 许可证。详见 [LICENSE](LICENSE) 文件。

---

### 贡献

欢迎提交 Issue 和 Pull Request！

### 联系方式

- GitHub: [piaoranyc/llmproxy](https://github.com/piaoranyc/llmproxy)
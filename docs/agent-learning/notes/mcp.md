# Step 9 · MCP（Model Context Protocol）详解

> 📚 对应学习计划：[README §Step 9](../README.md)
> 🗓️ 时间：今天
> 💻 关键代码：
> - [`code/06-agent/mcp-sdk-demo/`](../../../code/06-agent/mcp-sdk-demo/) — 独立官方 SDK Demo（177 行）
> - [`code/06-agent/go-agent-demo/mcp/`](../../../code/06-agent/go-agent-demo/mcp/) — 主项目内的 MCP 模块（423 行）

---

## 🎯 核心问题

> 如果 Tool 不在 Agent 这个 Go 程序里，而是在另一个独立服务里，Agent 怎么调用？

之前我们写的 Tool Registry、Tool Schema、Tool Handler 全部**耦合在 Agent 进程内**。

**MCP（Model Context Protocol）= 把 Tool 能力抽象成可跨进程、跨语言、跨网络的标准协议。**

---

## 🧠 MCP 是什么（一句话）

> **MCP = Tool 能力的标准化提供方 + 标准化连接协议。**

| 我们的 Tool Registry | MCP |
|---------------------|-----|
| Tool 在本进程 | Tool 在任意进程/服务/网络 |
| Go 函数调用 | JSON-RPC over Stdio/HTTP |
| 无协议 | 标准 MCP 协议 |
| Agent 自己实现 | Agent 直接对接 MCP Client |

---

## 📐 MCP 的 4 个核心概念

```
┌──────────────────────────────────────────────────┐
│  MCP Client (在 Agent 进程内)                      │
│       ↓                                            │
│  Transport (Stdio / Streamable HTTP)              │
│       ↓                                            │
│  MCP Server (独立进程 / 独立服务)                   │
│       ↓                                            │
│  Tools / Resources / Prompts                      │
└──────────────────────────────────────────────────┘
```

| 概念 | 职责 | 类比 |
|------|------|------|
| **MCP Server** | 提供 Tool/Resource/Prompt | 餐厅厨房 |
| **MCP Client** | 连接 Server，调用能力 | 餐厅服务员 |
| **Transport** | Client 和 Server 之间的通信通道 | 传送带 |
| **Protocol** | 通信的格式（JSON-RPC）| 点菜单的语言 |

---

## 🧪 项目里两套 MCP 实现（学习用）

### 实现 1：`mcp-sdk-demo`（独立官方 SDK Demo）

**两个文件，177 行，纯官方 SDK**：

#### `server/main.go`（约 90 行）

```go
import "github.com/modelcontextprotocol/go-sdk/mcp"

// 1. 定义 Tool 的输入输出结构
type GreetInput struct {
    Name string `json:"name" jsonschema:"the name of the person"`
}

type GreetOutput struct {
    Message string `json:"message"`
}

// 2. 实现 Tool Handler
func Greet(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input GreetInput,
) (*mcp.CallToolResult, GreetOutput, error) {
    return nil, GreetOutput{Message: "你好，" + input.Name}, nil
}

// 3. 创建 MCP Server
server := mcp.NewServer(&mcp.Implementation{Name: "mcp-sdk-demo-server"}, nil)

// 4. 注册 Tool
mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "..."}, Greet)

// 5. 用 Streamable HTTP Handler 暴露
handler := mcp.NewStreamableHTTPHandler(...)
http.Handle("/mcp", handler)
http.ListenAndServe(":8080", nil)
```

**3 个 Tool**：`greet`、`get_user`、`query_order`。

#### `client/main.go`（约 85 行）

```go
// 1. 创建 Client
client := mcp.NewClient(&mcp.Implementation{...}, nil)

// 2. 创建 Transport（指向 Server）
transport := &mcp.StreamableClientTransport{Endpoint: "http://localhost:8080/mcp"}

// 3. 连接
session, _ := client.Connect(ctx, transport, nil)

// 4. 列出 Server 的 Tool
toolsResult, _ := session.ListTools(ctx, &mcp.ListToolsParams{})

// 5. 调用 Tool
result, _ := session.CallTool(ctx, &mcp.CallToolParams{
    Name: "get_user",
    Arguments: map[string]any{"user_id": "1001"},
})
```

---

### 实现 2：`go-agent-demo/mcp/`（主项目内的混合实现）

**5 个文件，423 行**：

| 文件 | 行数 | 职责 |
|------|------|------|
| `mcp/server/server.go` | 210 | **自定义** MCP Server（模拟协议，复用现有 Tool Registry）|
| `mcp/client/client.go` | 85 | **官方 SDK** MCP Client |
| `mcp/client/tools.go` | 63 | MCP Tool → LLM Tool 转换 |
| `mcp/protocol/jsonrpc.go` | 42 | JSON-RPC 数据结构 |
| `mcp/transport/memory.go` | 23 | 内存 Transport（教学用）|

#### 为什么混合？

- **Client 用真 SDK**：MCP 客户端的协议交互复杂，直接用 SDK 省心
- **Server 自定义**：先理解核心职责（ListTools / CallTool），不直接陷入完整 MCP 协议

**好处**：等熟悉后，Server 可以无缝切换成 `mcp.AddTool` 真实现，Client 代码**0 改动**。

---

## 🔧 Server 端实现细节

### `mcp/server/server.go`

**核心职责只有 2 个**：

```go
// 1. 列出所有 Tool
func (s *Server) ListTools() []ToolInfo {
    allTools := s.registry.All()
    result := make([]ToolInfo, 0, len(allTools))
    for _, tool := range allTools {
        result = append(result, ToolInfo{
            Name:        tool.Name,
            Description: tool.Description,
        })
    }
    return result
}

// 2. 调用指定 Tool
func (s *Server) CallTool(
    ctx context.Context,
    name string,
    args map[string]any,
) (string, error) {
    tool, ok := s.registry.Get(name)
    if !ok {
        return "", fmt.Errorf("tool not found: %s", name)
    }
    
    argsJSON, _ := json.Marshal(args)
    return tool.Handler(string(argsJSON))
}
```

**关键设计**：**复用现有 Tool Registry**。

```go
type Server struct {
    registry *tools.Registry  // ← 直接复用，不重新发明 Tool
}
```

这样不破坏之前 `tools/order.go` 等已有代码。

### JSON-RPC 处理

```go
func (s *Server) HandleJSONRPC(requestJSON string) (string, error) {
    var request protocol.Request
    json.Unmarshal([]byte(requestJSON), &request)
    
    switch request.Method {
    case "tools/list":
        return s.handleToolsList(request)
    case "tools/call":
        return s.handleToolsCall(request)
    default:
        return s.errorResponse(request.ID, -32601, "method not found")
    }
}
```

**这是真实 MCP 协议的简化版**：
- 真实 MCP 用 `initialize` / `tools/list` / `tools/call` 等 method
- 我们的 Demo 只实现了 `tools/list` 和 `tools/call`

---

## 🔧 Client 端实现细节

### `mcp/client/client.go`

**官方 SDK Client**，核心 3 个方法：

```go
func New(ctx context.Context, endpoint string) (*Client, error) {
    mcpClient := mcp.NewClient(...)
    transport := &mcp.StreamableClientTransport{Endpoint: endpoint}
    session, _ := mcpClient.Connect(ctx, transport, nil)
    return &Client{client: mcpClient, session: session}, nil
}

func (c *Client) ListTools(ctx) ([]*mcp.Tool, error) {
    result, _ := c.session.ListTools(ctx, &mcp.ListToolsParams{})
    return result.Tools, nil
}

func (c *Client) CallTool(ctx, name, arguments) (*mcp.CallToolResult, error) {
    return c.session.CallTool(ctx, &mcp.CallToolParams{
        Name:      name,
        Arguments: arguments,
    })
}
```

---

## 🔄 关键转换：MCP Tool → LLM Tool

### 为什么要转换？

**MCP Server 返回的 Tool 定义 ≠ LLM Function Calling 需要的格式**。

```
MCP Tool:            { Name, Description, InputSchema (JSON Schema 对象) }
LLM Tool (OpenAI):   { Function: { Name, Description, Parameters (map) } }
```

需要把 MCP Tool 列表转换成 LLM 能识别的 Tool Schema，才能让 LLM 决定调用哪个 Tool。

### 实现（`mcp/client/tools.go`）

```go
func ConvertTools(
    tools []*mcp.Tool,
) []openai.ChatCompletionToolParam {
    result := make([]openai.ChatCompletionToolParam, 0, len(tools))

    for _, tool := range tools {
        // 1. InputSchema 是 JSON Schema 对象，转成 map
        inputSchema, _ := json.Marshal(tool.InputSchema)
        var parameters map[string]any
        json.Unmarshal(inputSchema, &parameters)

        // 2. 构造 OpenAI 格式
        result = append(result, openai.ChatCompletionToolParam{
            Function: openai.FunctionDefinitionParam{
                Name:        tool.Name,
                Description: openai.String(tool.Description),
                Parameters:  parameters,
            },
        })
    }

    return result
}
```

**这是 MCP 接入 Agent Loop 的关键一步**。

---

## 🔗 MCP 在 Agent Loop 中的位置

### 改造前

```go
// Agent 只调本地 Tool Registry
toolDefinitions := BuildToolDefinitions()  // ← 写死
```

### 改造后

```go
// Agent 同时有本地 Tool 和 MCP Tool
toolDefinitions := []openai.ChatCompletionToolParam{}

if a.useTool {
    // 1. 本地 Tool
    localTools := BuildToolDefinitions()
    toolDefinitions = append(toolDefinitions, localTools...)
    
    // 2. MCP Tool
    if a.mcpClient != nil {
        mcpTools, _ := a.mcpClient.ListTools(ctx)
        mcpDefinitions := mcpclient.ConvertTools(mcpTools)
        toolDefinitions = append(toolDefinitions, mcpDefinitions...)
    }
}
```

**Agent 不需要知道 Tool 来自本地还是 MCP**，统一调用：

```go
isMCP, _ := a.isMCPTool(toolName)
if isMCP {
    result, _ := a.mcpClient.CallTool(ctx, toolName, args)
} else {
    result, _ := a.toolRegistry.Get(toolName).Handler(argsJSON)
}
```

---

## 🚌 Transport：MCP 的传输层

### 三种主流 Transport

| Transport | 适用 | 现状 |
|-----------|------|------|
| **Stdio** | 本地进程（Client 启动 Server 子进程）| 经典 |
| **HTTP + SSE** | 远程服务 | **已过时** |
| **Streamable HTTP** ✅ | 远程服务、HTTP/2、流式 | **官方推荐** |

> "官方文档明确建议新部署使用 Streamable HTTP，而不是老的 SSE。"

### Streamable HTTP 示例（`mcp-sdk-demo/server/main.go`）

```go
handler := mcp.NewStreamableHTTPHandler(
    func(r *http.Request) *mcp.Server {
        return server
    },
    &mcp.StreamableHTTPOptions{
        JSONResponse: true,
    },
)
http.Handle("/mcp", handler)
http.ListenAndServe(":8080", nil)
```

Client 端：

```go
transport := &mcp.StreamableClientTransport{
    Endpoint: "http://localhost:8080/mcp",
}
```

---

## 🧠 为什么 MCP 重要？

### 问题场景

公司内部有 5 个系统：
- 订单系统（Python 写的）
- 用户系统（Java 写的）
- CRM 系统（Ruby 写的）
- 数据分析系统（Go 写的）
- 文档系统（Node 写的）

每个系统都有自己的 API。

**没有 MCP**：
- Agent 要为每个系统写一个 Tool 适配
- N 个系统 × M 个 Agent = N×M 个集成
- 一个系统改了 API，所有 Agent 都要改

**有 MCP**：
- 每个系统暴露 MCP Server
- Agent 只需要接 MCP Client
- N 个系统 + M 个 Agent = N + M 个集成
- 一个系统改了 API，只需要改对应的 MCP Server

### MCP 解决的本质问题

> **数据源之间的标准化连接问题**。

跟 LangChain/LlamaIndex 解决的是同一类问题，但 MCP 更轻、更聚焦。

---

## 📊 MCP vs Tool Registry 对比

| 维度 | Tool Registry | MCP |
|------|---------------|-----|
| **Tool 位置** | 同进程 | 跨进程/网络 |
| **调用方式** | Go 函数 | JSON-RPC |
| **语言** | 强类型 Go | 任何语言 |
| **进程模型** | 单进程 | 多进程 |
| **标准化** | 无 | MCP 协议 |
| **解耦程度** | 紧耦合 | 松耦合 |
| **适用规模** | Demo / 小型 | 真实生产 |

**关键洞察**：你之前的 Tool Registry 不是"白写了"，**MCP 是 Tool Registry 的超集**。

---

## ⚠️ 当前实现的局限

### 局限 1：Server 不是真 MCP Server

`mcp/server/server.go` 是**模拟的 MCP Server**，只实现了 `tools/list` 和 `tools/call`。

**生产应该**：用 `mcp.AddTool` + `mcp.NewServer` 真实现。

### 局限 2：没有 Resources 和 Prompts

完整 MCP 还包括：
- **Resources**：结构化数据（数据库行、文件内容）
- **Prompts**：可复用的 Prompt 模板

当前 Demo 只有 Tools。

### 局限 3：没有认证 / 权限

真实生产 MCP Server 需要：
- OAuth / API Key 认证
- Tool 黑名单 / 白名单
- 调用频率限制
- 审计日志

### 局限 4：Memory Transport 不实用

`mcp/transport/memory.go` 是教学用的，Client 和 Server 通过内存函数回调通信。

**真实应该是**：Stdio 或 Streamable HTTP。

---

## 🔮 下一步

### 9.1 把自定义 Server 换成官方 SDK

```go
// 当前（自定义）
server := mcp.NewServer(...)
handler := server.HandleJSONRPC

// 改造后（官方 SDK）
sdkServer := mcp.NewServer(...)
mcp.AddTool(sdkServer, &mcp.Tool{Name: "..."}, handler)
handler := mcp.NewStreamableHTTPHandler(...)
```

### 9.2 接入真实业务 MCP Server

比如接入 GitHub MCP Server：
```bash
npx -y @modelcontextprotocol/server-github
```

### 9.3 在 Agent 中实现 Tool 来源透明

```go
// Agent 不关心 Tool 来自哪里
result := a.executeTool(ctx, toolName, args)
//   ↑ 内部判断本地 / MCP
```

---

## ✅ 自检

- [ ] 1. MCP 解决了什么问题？Tool Registry 不够用吗？
- [ ] 2. MCP 的 4 个核心概念（Server / Client / Transport / Protocol）分别是什么？
- [ ] 3. 为什么需要把 MCP Tool 转换成 LLM Tool？转换的关键是什么？
- [ ] 4. Streamable HTTP 比 SSE 强在哪？为什么官方推荐？
- [ ] 5. 项目里两套 MCP 实现（`mcp-sdk-demo` vs `mcp/`）的区别是什么？

---

## 📂 关键代码

### 独立 Demo
- `code/06-agent/mcp-sdk-demo/server/main.go` (~90 行) - 官方 SDK MCP Server
- `code/06-agent/mcp-sdk-demo/client/main.go` (~85 行) - 官方 SDK MCP Client

### 主项目集成
- `code/06-agent/go-agent-demo/mcp/server/server.go` (210 行) - 自定义 MCP Server
- `code/06-agent/go-agent-demo/mcp/client/client.go` (85 行) - 官方 SDK Client
- `code/06-agent/go-agent-demo/mcp/client/tools.go` (63 行) - MCP Tool → LLM Tool
- `code/06-agent/go-agent-demo/mcp/protocol/jsonrpc.go` (42 行) - JSON-RPC
- `code/06-agent/go-agent-demo/mcp/transport/memory.go` (23 行) - 内存 Transport

---

## 🔗 相关笔记

- [Router 详解](./router.md) - Agent Loop 前的 Capability 路由
- [Agent 架构总览](./agent-architecture.md) - MCP 在整体架构中的位置

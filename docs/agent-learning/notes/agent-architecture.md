# Agent 架构总览：3 大支柱 + Harness

> 📚 对应学习计划：[README §Step 6-12](../README.md) 整体回顾
> 🗓️ 时间：今天
> 💻 对应代码：[`code/06-agent/go-agent-demo/`](../../../code/06-agent/go-agent-demo/)

---

## 🎯 核心问题

> 一个完整的 Agent 系统，到底由哪些部分组成？

经过 Step 5-8 的学习，你的 Agent 现在已经**不只是 ReAct 循环**，而是一个**完整的 Agent Runtime**。

本文档帮你把所有学过的概念**串成一张完整的架构图**。

---

## 🏛️ Agent 的 3 大上下文支柱 + Router

```
                    User Query
                         │
                         ▼
                 ┌───────────────┐
                 │    Router     │ ← 新增：能力路由
                 │  (LLM 决策)   │
                 └───────┬───────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
   User Memory     Knowledge         External
   (Memory)        (RAG)             (Tools)
        │                │                │
        ▼                ▼                ▼
    用户是谁         知识是什么       外部能做什么
```

**四大组件**：

| 组件 | 职责 | 典型实现 |
|------|------|---------|
| **Router** | 「这次需要哪些能力」| LLM + Capability 配置 |
| **Memory** | 「用户是谁」| key-value + 向量检索 |
| **RAG** | 「知识是什么」| 文档 Embedding + 向量库 |
| **Tools** | 「外部能做什么」| Function Calling / **MCP** |

**Router 的价值**：按需启用能力，省 Token、省延迟、减少幻觉。详见 [router.md](./router.md)。

---

## 📊 完整 Agent 运行时架构

```
┌─────────────────────────────────────────────────────────────┐
│                       User Input                              │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                         Router                                │
│  (LLM: 决定 use_memory / use_rag / use_tool)                │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Query Rewrite                              │
│  (LLM: 自然语言 → 检索关键词)                                │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Memory Retrieval                             │
│  ┌────────────────┐    ┌────────────────┐                  │
│  │ Keyword Search │ +  │ Vector Search  │ → Hybrid Score   │
│  └────────────────┘    └────────────────┘                  │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     RAG Retrieval                             │
│  Knowledge Base → Vector Search → Top K Chunks              │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Build LLM Context                            │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ [System] Agent 角色 + 行为约束                        │  │
│  │ [Memory] 检索到的用户长期记忆                          │  │
│  │ [RAG]    检索到的相关知识                              │  │
│  │ [Tools]  当前可用的 Tool Schema                       │  │
│  │ [History] 最近 N 条消息（含 Summary）                  │  │
│  │ [Current] 当前用户 Query                              │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                       LLM Call                                │
└──────────────────────────┬──────────────────────────────────┘
                           ▼
                ┌──────────┴──────────┐
                ▼                     ▼
        ┌──────────────┐      ┌──────────────┐
        │  Tool Calls  │      │ Final Answer │
        └──────┬───────┘      └──────┬───────┘
               │                     │
               ▼                     ▼
        ┌──────────────┐      ┌──────────────────┐
        │ Tool Execute │      │Memory Extraction │
        │ + Result     │      │ (LLM 提取事实)    │
        │ 回到 LLM     │      │ + Save to Memory │
        └──────────────┘      └──────────────────┘
```

---

## 🧩 每个模块的职责清单

| 模块 | 职责 | 关键代码 |
|------|------|---------|
| **ContextManager** | 管理当前对话上下文，摘要压缩 | `memory/context.go` |
| **Memory** | 跨对话长期记忆 + Hybrid Search | `memory/memory.go` |
| **Memory Extractor** | 从用户消息提取长期事实 | `memory/extractor.go` |
| **Query Rewriter** | 自然语言 → 检索关键词 | `agent/query_rewrite.go` |
| **Embedding** | 文本 → 向量 | `embedding/embedding.go` |
| **KnowledgeBase** | 文档切片 + 向量化 | `rag/knowledge.go` |
| **Retriever** | Top K 检索 | `rag/retriever.go` |
| **RAG** | 检索 + LLM 生成答案 | `rag/rag.go` |
| **Tool Registry** | 工具注册与查找 | `tools/registry.go` |
| **Tool Schema** | 工具说明（给 LLM）| `agent/schema.go` |
| **Router** | Capability 路由（按需启用）| `agent/router.go` |
| **MCP Server** | 自定义 MCP Server | `mcp/server/server.go` |
| **MCP Client** | 官方 SDK MCP Client | `mcp/client/client.go` |
| **MCP Tool Converter** | MCP Tool → LLM Tool | `mcp/client/tools.go` |
| **Agent Loop** | 整合所有模块的循环 | `agent/agent.go` |
| **Output Parser** | 清理 LLM 输出（<think> / Markdown）| `llm/output.go` |

---

## 🧠 几个关键设计决策

### 决策 1：Memory 和 RAG 是两个独立模块

虽然两者底层都用「向量检索 + 关键词检索」，但**职责完全不同**：

| | Memory | RAG |
|--|--|--|
| 数据来源 | 用户对话 | 离线文档 |
| 更新时机 | 每次对话 | 文档更新时 |
| 内容性质 | 用户事实 | 通用知识 |
| 检索 query | 用户当前问题 | 任务相关问题 |

**如果混在一起**：难以维护、难以优化、概念混乱。

### 决策 2：Query Rewrite 在 Memory Retrieval 之前

不直接用原 query，而是先 LLM 改写：

```
原 query："你还记得我的职业吗？"   ← 难搜
    ↓ Rewrite
queries: ["user_profession", "职业", "工作"]   ← 易搜
```

详见 [query-rewrite.md](./query-rewrite.md)。

### 决策 3：Memory Extraction 只看 User Message

```go
// ❌ 错误：把 user + assistant 都给 LLM
userPrompt := userMsg + assistantAnswer

// ✅ 正确：只把 user 给 LLM
userPrompt := userMsg
```

避免 Assistant 的「幻觉回答」被错误提取成 Memory 事实。

### 决策 4：Hybrid Search 而不是单一检索

```
Hybrid Score = 0.5 × Keyword Score + 0.5 × Vector Similarity
```

- Keyword 擅长精确 key 匹配（`user_profession` 完全匹配）
- Vector 擅长语义相似（"工作" ≈ "职业"）
- 两者互补

### 决策 5：Router 按需启用能力（不是固定执行）

**不是**每次都跑 Memory + RAG + Tools，而是 Router 先判断需要哪些：

```go
decision := router.Decide(ctx, userQuery, capabilities)
// decision: { use_memory, use_rag, use_tool }

if decision.UseMemory { /* 查 Memory */ }
if decision.UseRAG    { /* 检索 RAG */ }
if decision.UseTool   { /* 给 LLM Tool Schema */ }
```

详见 [router.md](./router.md)。

### 决策 6：本地 Tool 和 MCP Tool 统一抽象

Agent 不区分 Tool 来源，统一通过 Tool Name 调用：

```go
isMCP := a.isMCPTool(toolName)
if isMCP {
    return a.mcpClient.CallTool(ctx, toolName, args)
}
return a.toolRegistry.Get(toolName).Handler(argsJSON)
```

详见 [mcp.md](./mcp.md)。

---

## 🔗 Context Engineering 视角

把所有学过的概念放到 **Context Engineering** 这个更高层的视角：

> **Context Engineering = 怎么把历史、Memory、RAG、Tool Result 等各种信息组织成 LLM 能正确理解的上下文。**

| 我们学过的 | 在 Context Engineering 里是 |
|-----------|------------------------|
| ContextManager | 控制 history 长度 |
| Memory + Hybrid Search | 注入 user facts |
| RAG | 注入 relevant knowledge |
| Tool Schema + Result | 注入可用工具和结果 |
| System Prompt | 全局行为约束 |

**Context Engineering 是 2025+ Agent 工程的核心理念**。

---

## 🛠️ Harness Engineering 视角

这是更新的概念（2026 年由 Mitchell Hashimoto / HashiCorp 创始人提出）：

> **Harness（驾驭）= 包裹在 LLM 外面的运行环境，让 Agent 能稳定干活。**

### Harness 的核心职责

```
Harness
├── 任务定义（Task Definition）
├── 上下文选择（Context Selection）
├── 工具访问（Tool Access）
├── 项目记忆（Project Memory）
├── 任务状态（Task State）
├── 可观测性（Observability）
├── 验证（Validation）
├── 权限（Permissions）
└── 扩展（Extensions，含 Skills / Subagents / MCP）
```

### 我们的项目跟 Harness 的对应

| Harness 职责 | 我们项目里的实现 |
|------------|--------------|
| 任务定义 | User Prompt + Agent Loop |
| 上下文选择 | ContextManager + Memory Retrieval + RAG |
| 工具访问 | Tool Registry + Tool Schema |
| 项目记忆 | longTermMemory |
| 任务状态 | ContextManager.messages |
| 可观测性 | 打印 Log（生产应该接 LangSmith）|
| 验证 | 当前没有 |
| 权限 | 当前没有 |
| 扩展 | 当前没有 |

> 我们实际上已经在写一个 **Mini Agent Harness**。

---

## 📐 当前 Agent Loop 的 7 步详解

`agent/agent.go` 的 `Run` 方法：

```go
func (a *Agent) Run(ctx context.Context) (string, error) {
    toolDefinitions := BuildToolDefinitions()

    for {
        // 1. Context 摘要检查
        a.contextManager.MaybeSummarize(ctx, a.client, a.model)

        // 2. Memory Retrieval（用 Query Rewrite + Hybrid Search）
        memoryContext := a.buildMemoryContext(ctx, userQuery)

        // 3. RAG Retrieval（如启用）
        // ragContext := a.rag.Retrieve(ctx, userQuery)

        // 4. 构造 Messages（含 Memory + RAG Context）
        messages := buildMessages(system, memoryContext, summary, recent)

        // 5. 调 LLM
        resp := a.client.Chat.Completions.New(ctx, ...)

        // 6. 处理 Tool Calls 或 Final Answer
        ...

        // 7. Memory Extraction（保存新事实）
        a.extractAndSaveMemory(ctx, userQuery, answer)
    }
}
```

**每一步**都有专门的笔记：

| Step | 对应笔记 |
|------|---------|
| Router | [router.md](./router.md) |
| 1. Context 摘要 | [context-management.md](./context-management.md) |
| 2. Memory Retrieval | [memory.md](./memory.md) |
| 2. Query Rewrite | [query-rewrite.md](./query-rewrite.md) |
| 3. RAG | [rag.md](./rag.md) |
| 3. Embedding | [embedding-vector-search.md](./embedding-vector-search.md) |
| 6. Workflow | [agent-orchestration.md](./agent-orchestration.md) |
| Tool 来源透明 | [mcp.md](./mcp.md) |

---

## 🎓 面试角度：怎么讲清楚你做的 Agent

如果你在面试中被问「你怎么用 Go 写 Agent」，按这个顺序讲：

### 30 秒电梯演讲

> 我用 Go 写了一个完整的 Agent Runtime，包含三大支柱：
> 1. **Tools**（订单/用户/支付/物流查询）
> 2. **Memory**（用户长期记忆 + Hybrid Search）
> 3. **RAG**（知识库检索 + 生成）
> 加上 **ContextManager** 控制对话长度，**Query Rewriter** 优化检索，
> **Memory Extractor** 自动从对话提取事实。

### 5 分钟深入讲

按这个顺序：

1. **架构图**：3 大支柱 → Context 拼装 → Agent Loop
2. **Agent Loop**：ReAct 循环，处理 Tool Calls
3. **ContextManager**：消息历史 + 摘要
4. **Memory**：Hybrid Search（关键词 + 向量），通过 Query Rewrite 提升召回
5. **Memory Extraction**：用 LLM 自动提取，避免污染
6. **RAG**：KnowledgeBase + Retriever + Ask

### 加分项

- 提到 **Context Engineering** 视角
- 提到 **Harness Engineering** 概念
- 提到 **Hybrid Search Score Fusion** 的权重设计
- 提到 **Memory Extraction** 的反污染设计

---

## 📊 演进路径：当前 vs 生产

```
当前 Demo                    生产级 Agent Harness
─────────────────────────────────────────────────
进程内 Memory           →    Redis + 向量 DB
Fake Embedding          →    OpenAI / BGE Embedding
朴素 Chunk 切分         →    RecursiveCharacterTextSplitter
无评估                  →    LangSmith / LangFuse 观测
单 LLM Provider         →    多 Provider + 路由
无权限控制              →    Tool 黑名单 + 二次确认
无 Skills               →    MCP Server + Skills 注册
无 Subagents            →    Multi-Agent 协作
```

---

## 🔗 相关笔记

- [Memory 完整体系](./memory.md)
- [RAG 完整链路](./rag.md)
- [Query Rewrite 详解](./query-rewrite.md)
- [Embedding & Vector Search](./embedding-vector-search.md)
- [Context Management 详解](./context-management.md)
- [Agent 编排模式详解](./agent-orchestration.md)

---

## ✅ 自检

- [ ] 1. Agent 的 3 大上下文支柱是什么？各自回答什么问题？
- [ ] 2. Memory 和 RAG 的本质区别是什么？
- [ ] 3. Query Rewrite 在 Agent Loop 的哪一步？
- [ ] 4. Hybrid Search Score 公式是什么？为什么要归一化？
- [ ] 5. 什么是 Harness Engineering？跟 Context Engineering 有什么关系？

---

## 📂 项目结构（最新）

```
go-agent-demo/
├── main.go                       # 入口：装配所有模块
├── config/config.go              # Provider 配置
├── llm/
│   ├── client.go                 # OpenAI Client
│   └── output.go                 # Output Parser（CleanThinking）
├── embedding/
│   └── embedding.go              # Embedder 接口 + FakeEmbedder + CosineSimilarity
├── memory/
│   ├── context.go                # ContextManager（当前对话）
│   ├── memory.go                 # Memory（跨对话 + Hybrid Search）
│   └── extractor.go              # Memory Extraction
├── rag/
│   ├── knowledge.go              # KnowledgeBase + Document + Chunk
│   ├── retriever.go              # Vector Search
│   └── rag.go                    # RAG Ask 完整链路
├── tools/
│   ├── tool.go                   # Tool 接口
│   ├── registry.go               # Tool Registry
│   └── order.go                  # 业务 Tool 实现
├── mcp/                          # MCP 模块（混合实现）
│   ├── server/server.go          # 自定义 MCP Server
│   ├── client/client.go          # 官方 SDK MCP Client
│   ├── client/tools.go           # MCP Tool → LLM Tool 转换
│   ├── protocol/jsonrpc.go       # JSON-RPC 数据结构
│   └── transport/memory.go       # 内存 Transport（教学用）
├── agent/
│   ├── agent.go                  # Agent Loop（核心）
│   ├── schema.go                 # Tool Schema（给 LLM 看）
│   ├── memory.go                 # Agent 中的 Memory 集成
│   ├── query_rewrite.go          # Query Rewriter
│   └── router.go                 # Capability Router
└── workflow/
    └── order.go                  # 三种 Workflow 实现

mcp-sdk-demo/                      # 独立 MCP SDK Demo
├── server/main.go                # 官方 SDK MCP Server
└── client/main.go                # 官方 SDK MCP Client
```

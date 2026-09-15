# Router · Capability 路由详解

> 📚 对应学习计划：[README §Step 8 扩展](../README.md)（属于 Step 8 之后引入的关键架构思想）
> 🗓️ 时间：今天
> 💻 关键代码：[`agent/router.go`](../../../code/06-agent/go-agent-demo/agent/router.go)（~170 行）

---

## 🎯 核心问题

> 每次用户问问题，都要查 Memory + 检索 RAG + 调用 Tool 吗？太浪费了，能不能按需启用？

真实 Agent 不可能所有事情都让 LLM 自动执行。**Router** 就是解决「**按需启用能力**」的问题。

---

## 💡 核心思想

```
                    User Query
                         │
                         ▼
                 ┌───────────────┐
                 │    Router     │
                 │  (LLM 决策)   │
                 └───────┬───────┘
                         │
        ┌────────────────┼────────────────┐
        ▼                ▼                ▼
   ┌─────────┐      ┌─────────┐      ┌─────────┐
   │ Memory? │      │  RAG?   │      │ Tools?  │
   └─────────┘      └─────────┘      └─────────┘
        │                │                │
        ▼                ▼                ▼
   (use_memory)    (use_rag)         (use_tool)
      bool            bool             bool
```

**Router 不回答问题，只决定用哪些能力。**

---

## 🧠 为什么需要 Router？

### 问题场景

每次用户问问题，Agent 当前做的事：

1. 调 LLM 做 Query Rewrite
2. 查 Memory（Hybrid Search）
3. 检索 RAG（Vector Search）
4. 把 Memory Context + RAG Context 注入
5. 调用 LLM
6. 处理 Tool Calls
7. 做 Memory Extraction

**全跑一遍**：每个 query 都要花 5-7 次 LLM 调用 + 多次检索。

### Router 优化

| 用户问题 | Router 决策 |
|---------|----------|
| "你好" | `use_memory=false, use_rag=false, use_tool=false`（直接 LLM 回答） |
| "你还记得我的职业吗？" | `use_memory=true, use_rag=false, use_tool=false` |
| "订单 10001 怎么退款？" | `use_memory=false, use_rag=true, use_tool=false` |
| "帮我查用户 1001 的订单 10001" | `use_memory=false, use_rag=false, use_tool=true` |
| "你还记得我吗？根据规则告诉我能不能退款" | `use_memory=true, use_rag=true, use_tool=true` |

**收益**：省 50%+ Token，省 50%+ 延迟。

---

## 🔧 实现（`agent/router.go`）

### 3 个核心数据结构

```go
type Capability struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type RouteDecision struct {
    UseMemory bool `json:"use_memory"`
    UseRAG    bool `json:"use_rag"`
    UseTool   bool `json:"use_tool"`
}

type Router struct {
    client *openai.Client
    model  string
}
```

### Router Prompt 设计

```
你是一个 Agent Router。

你的任务不是回答用户问题，
而是判断当前用户问题需要使用 Agent 的哪些能力。

当前 Agent 可用能力：

[
  {"name": "memory", "description": "查询用户过去明确保存的长期记忆..."},
  {"name": "rag", "description": "查询当前 Agent 的业务知识库..."},
  {"name": "tools", "description": "查询实时业务数据或执行外部操作..."}
]

判断规则：
1. Memory：用于查询用户过去明确保存的长期记忆
2. RAG：用于查询当前 Agent 的业务知识库
3. Tools：用于查询实时业务数据或执行外部操作
4. 如果用户问题不需要以上任何能力，则全部设置为 false
5. 可以同时选择多个能力

只返回 JSON，不要输出其他内容。
```

**关键设计**：
- Router 的输出是 **结构化 boolean**（不是自由文本）
- Prompt 动态化（capabilities 从配置注入）
- 严格的 JSON 输出格式

### 关键实现

```go
func (r *Router) Decide(
    ctx context.Context,
    query string,
    capabilities []Capability,
) (RouteDecision, error) {
    // 1. 把 capabilities 序列化成 JSON
    capabilityJSON, _ := json.MarshalIndent(capabilities, "", "  ")
    
    // 2. 拼 Prompt
    prompt := fmt.Sprintf(`...`, capabilityJSON, query)
    
    // 3. 调 LLM
    resp, _ := r.client.Chat.Completions.New(ctx, ...)
    
    // 4. 清理 <think> 标签 + Markdown
    content := cleanRouterThinking(resp.Choices[0].Message.Content)
    content = cleanRouterJSON(content)
    
    // 5. 解析 RouteDecision
    var decision RouteDecision
    json.Unmarshal([]byte(content), &decision)
    
    return decision, nil
}
```

---

## 🔗 Router 在 Agent Loop 中的位置

### Router 只判断一次

```
User Query
    ↓
[1. Router.Decide]  ← 只调一次 LLM
    ↓
RouteDecision { use_memory, use_rag, use_tool }
    ↓
Agent Loop for {
    [2. 根据 decision 准备能力]
    [3. 调 LLM（带 Memory/RAG/Tool Context）]
    [4. 处理 Tool Calls / Final Answer]
    [5. Memory Extraction]
    [6. 回到 step 3]
}
```

**关键**：Router 在 Agent Loop **外面**调用一次，决策结果贯穿整个 Loop。

---

## 📋 Router 与各能力的配合

### Memory 决策（`use_memory=true`）

```go
if decision.UseMemory {
    memoryContext := a.buildMemoryContext(ctx, userQuery)
    messages = append(messages, memoryContext)
} else {
    // 不查 Memory，跳过这一步
}
```

### RAG 决策（`use_rag=true`）

```go
if decision.UseRAG {
    ragResults, _ := a.rag.Retrieve(ctx, userQuery)
    // 拼到 Context
}
```

### Tools 决策（`use_tool=true`）

```go
if decision.UseTool {
    // 1. 本地 Tool
    localTools := BuildToolDefinitions()
    toolDefinitions = append(toolDefinitions, localTools...)
    
    // 2. MCP Tool（如启用）
    if a.mcpClient != nil {
        mcpTools, _ := a.mcpClient.ListTools(ctx)
        toolDefinitions = append(toolDefinitions, mcpclient.ConvertTools(mcpTools)...)
    }
} else {
    // LLM 看不到任何 Tool Schema，自然不会调
    toolDefinitions = nil
}
```

**关键**：如果不提供 Tool Schema 给 LLM，LLM 就不会调 Tool → **从源头切断**。

---

## 🧪 Router 测试场景

项目里 `main.go` 的 `runRouterTest` 覆盖了 5 个典型场景：

```go
testCases := []RouterTestCase{
    {
        name:        "1. Memory 路由",
        query:       "你还记得我的职业吗？",
        expectedContains: []string{"use_memory"},
    },
    {
        name:        "2. RAG 路由",
        query:       "订单 10001 怎么退款？",
        expectedContains: []string{"use_rag"},
    },
    {
        name:        "3. Tool 路由",
        query:       "帮我查用户 1001 的订单 10001",
        expectedContains: []string{"use_tool"},
    },
    {
        name:        "4. 闲聊路由",
        query:       "你好",
        expectedContains: []string{"use_memory\":false"},  // 期望 false
    },
    {
        name:        "5. 多能力路由",
        query:       "你还记得我吗？根据规则告诉我能不能退款",
        expectedContains: []string{"use_memory\":true", "use_rag\":true"},
    },
}
```

---

## ⚠️ Router 设计的常见误区

### 误区 1：把 Router 当成另一个 Agent

❌ Router 是 Agent 的**前置决策器**，不是另一个 Agent
✅ Router 只决定"用哪些能力"，不参与实际执行

### 误区 2：Router 输出自由文本

❌ "我觉得应该用 Memory"
✅ **结构化 boolean**（可机器解析、可条件分支）

### 误区 3：把 Router 放在 Agent Loop 内

❌ 每轮 Loop 都重新调 Router（浪费）
✅ **Router 只调一次**，决策贯穿整个对话

### 误区 4：Capabilities 写死在 Prompt 里

❌ Prompt 里固定写死三种能力
✅ **Capabilities 动态注入**，方便扩展（如增加 "Skills"）

### 误区 5：Router 不带容错

❌ LLM 返回非法 JSON，整个 Agent 挂掉
✅ 加 fallback（默认全部 true 或 false）

---

## 📊 Router 的成本 vs 收益

### 成本

- 每次对话 **多 1 次 LLM 调用**（Router）
- 约 200~500 token 的 Prompt
- 约 100~300 token 的输出

### 收益（按平均对话估算）

| 场景 | 无 Router | 有 Router | 节省 |
|------|---------|---------|------|
| 闲聊 | 1 次 LLM + 0 检索 | 1 次 Router + 1 次 LLM | **持平** |
| Memory 类 | 1 次 LLM + 3 次检索 | 1 次 Router + 1 次 LLM + 1 次检索 | **省 67%** |
| RAG 类 | 1 次 LLM + Memory + RAG + Tool Schema | 1 次 Router + 1 次 LLM + RAG | **省 60%** |
| Tool 类 | 全部能力 + 全部 Tool Schema | 1 次 Router + 1 次 LLM + Tool Schema | **省 70%** |

**结论**：Router 几乎总是赚的，特别是 Memory/RAG/Tool 不需要的场景。

---

## 🔮 Router 的演进方向

### 演进 1：基于历史决策的 Router

```
用户连续 3 轮都在问"订单 XXXX"
→ Router 推断：当前主要是 Tool 类
→ 优先 use_tool，跳过 Memory/RAG
```

### 演进 2：Router 自学习

记录「Router 决策 → 用户反馈」，训练更准的 Router。

### 演进 3：Router 决策解释

```
RouteDecision: { use_memory: true, use_rag: false, use_tool: false }
Reason: "用户问的是个人偏好，不是业务知识"
```

便于调试 + 后续可追溯。

---

## ✅ 自检

- [ ] 1. 为什么需要 Router？不加 Router 会怎样？
- [ ] 2. Router 决策的输出是什么格式？为什么？
- [ ] 3. Router 应该在 Agent Loop 内还是外？为什么？
- [ ] 4. Router Prompt 里的 Capabilities 为什么动态注入而不是写死？
- [ ] 5. 5 个典型测试场景分别测什么？

---

## 📂 关键代码

- `agent/router.go` (~170 行) - Router 完整实现
- `agent/agent.go` - Agent Loop 中调用 Router.Decide
- `main.go` `runRouterTest` - 5 个测试场景

---

## 🔗 相关笔记

- [MCP 详解](./mcp.md) - Router 决定是否启用 Tool（本地 + MCP）
- [Memory 完整体系](./memory.md) - Router 决定是否启用 Memory
- [RAG 完整链路](./rag.md) - Router 决定是否启用 RAG
- [Agent 架构总览](./agent-architecture.md) - Router 在整体架构中的位置

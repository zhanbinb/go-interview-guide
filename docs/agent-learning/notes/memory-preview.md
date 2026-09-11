# Step 8 · Memory 预备（Preview）

> 📚 对应学习计划：[README §Step 8](../README.md)
> 🗓️ 状态：⚠️ **预览**（下一阶段将完整学习）
> 💻 关键代码：[`memory/memory.go`](../../../code/06-agent/go-agent-demo/memory/memory.go)（38 行）

---

## 🎯 核心问题

> 这次对话结束以后，Agent 以后还能不能记住某些信息？

例如：
- 用户说「我叫张三」→ 下次对话 Agent 应该知道
- 用户说「回答简洁点」→ 所有未来对话都遵守
- 用户上周问过 XX → 这次还能引用

**Memory = Agent 跨任务 / 跨会话持久化的信息**

---

## 🧠 Memory vs ContextManager（再次强调）

这是 Agent 开发里**最容易混淆的两个概念**。

| 维度 | Memory | ContextManager |
|------|--------|----------------|
| **核心问题** | "下次对话 Agent 还记不记得？" | "这次请求 LLM 看哪些历史？" |
| **范围** | 跨对话 | 当前对话 |
| **存储位置** | 持久化存储（生产用 Redis/DB）| 进程内（`ContextManager.messages`）|
| **生命周期** | 长期（重启后还在）| 短暂（Agent Run 结束就没了）|
| **当前实现复杂度** | 极简（38 行 map）| 复杂（175 行 + LLM 摘要）|
| **典型内容** | 结构化事实（"user_name=张三"）| 完整对话消息 |

**一个比喻**：

| | ContextManager | Memory |
|--|--|--|
| 类比 | 短期记忆（今天的事） | 长期记忆（人生经验） |
| 类比 | 工作记忆 | 知识库 |

---

## 💻 当前 Memory 实现（最简版）

```go
type Memory struct {
    data map[string]string
}

func NewMemory() *Memory {
    return &Memory{
        data: make(map[string]string),
    }
}

func (m *Memory) Set(key, value string) {
    m.data[key] = value
}

func (m *Memory) Get(key string) string {
    return m.data[key]
}
```

**38 行代码，实现了一个 key-value 内存存储**。

### Demo 中的使用

```go
// main.go testMemory()
func testMemory() {
    fmt.Println("===== Memory Demo =====")
    m := memory.NewMemory()

    // 第一轮对话：用户告诉 Agent 自己叫什么
    m.Set("user_name", "张三")

    // 第二轮对话：Agent 从 Memory 中获取之前保存的信息
    name := m.Get("user_name")
    fmt.Println("Memory user_name:", name)
}
```

**这已经足够理解 Memory 的本质**：
> Agent 可以跨越单次执行、保存并重新读取的信息。

---

## ⚠️ 当前实现的局限性

### 局限 1：不持久化

进程退出 Memory 就没了。

**生产解决方案**：

| 存储方案 | 适用场景 |
|---------|---------|
| **文件**（JSON/SQLite） | 单机部署 |
| **Redis** | 分布式、需要高性能 |
| **PostgreSQL** | 需要复杂查询、长期归档 |
| **向量数据库**（pgvector / Pinecone）| 语义检索（RAG）|

### 局限 2：不区分用户/会话

所有用户共享同一个 Memory。

```go
// 当前：所有用户共享
m := memory.NewMemory()
m.Set("user_name", "张三")

// 应该：按用户隔离
m := memory.NewMemory(userID="user_1001")
```

**生产解决方案**：Memory 按 `user_id` 区分（多租户隔离）。

### 局限 3：手动 Set（无自动事实抽取）

当前要程序手动 `Set("user_name", "张三")`。

**生产应该让 LLM 自动提取**：

```go
// 用户说："我叫张三"
// ↓ LLM 自动提取
m.Set("user_name", "张三")
m.Set("language", "zh-CN")
m.Set("preference", "concise")
```

### 局限 4：没自动注入到 Context

当前 Memory 不会自动拼接到 messages 里。

**生产应该**：

```
BuildMessages():
  [System Prompt: 你有用户的长期记忆: ...]
  [Memory 摘要: user_name=张三, language=zh-CN]
  [最近 messages...]
```

### 局限 5：Map[string]string 太简单

只能存字符串，无法存复杂结构。

**生产应该用结构化 Memory**：

```go
type Memory struct {
    Facts    []Fact           // 事实列表
    Episodes []Episode        // 关键事件
    Skills   map[string]Skill // 学到的技能
}
```

---

## 🧩 Memory 的核心问题（生产环境要解决的）

### 问题 1：存什么？

不存什么？存什么？这是 Memory 设计的核心。

| 应该存 | 不应该存 |
|--------|---------|
| 用户偏好（"回答简洁"） | 完整对话历史（那是 ContextManager 的事）|
| 关键事实（"user_name=张三"） | 临时上下文（"订单 10001"）|
| 用户身份信息（VIP、年龄） | 一次性查询结果 |
| 学到的技能 | 调试日志 |

### 问题 2：什么时候更新？

| 时机 | 动作 |
|------|------|
| 用户明确告知偏好 | 立即更新（高优先级） |
| 对话结束时 | LLM 异步提取事实 |
| Memory 读取时 | 可能的冲突解决 |
| 定期 | 压缩 / 清理过期 Memory |

### 问题 3：什么时候给 LLM 看？

不是所有 Memory 都要每次都塞给 LLM：

- **System Prompt**：核心稳定的事实（如用户名、语言）
- **每次请求**：相关的事实（按当前任务相关性筛选）
- **完全不注入**：过期或低价值的事实

### 问题 4：怎么避免污染？

如果 Memory 里错误信息没清理，会持续误导 LLM。

**解决方案**：
- 信任度分数
- 冲突解决（最新覆盖旧）
- 定期 Review

---

## 🔮 完整 Memory 系统的可能架构

```
┌─────────────────────────────────────────┐
│            User Input                    │
└─────────────┬───────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│         Agent Loop                       │
│  ┌───────────────────────────────────┐  │
│  │ 1. Read Memory (按 user_id)        │  │
│  │ 2. Inject to Context               │  │
│  │ 3. LLM Call                        │  │
│  │ 4. Tool Execution                  │  │
│  │ 5. Fact Extraction (async)         │  │
│  │ 6. Update Memory                   │  │
│  └───────────────────────────────────┘  │
└─────────────┬───────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│         Memory Store                     │
│  - Redis (短期事实)                      │
│  - PostgreSQL (长期档案)                 │
│  - Vector DB (语义检索)                  │
└─────────────────────────────────────────┘
```

---

## 🎯 Step 8 完整学习目标

下一阶段将系统学习：

### 8.1 Redis 持久化（1h）

把 `Memory` 升级为支持 Redis 持久化：

```go
type RedisMemory struct {
    client *redis.Client
    userID string
}

func (m *RedisMemory) Set(key, value string) error {
    return m.client.HSet(ctx, "memory:"+m.userID, key, value).Err()
}
```

### 8.2 用户隔离（30 min）

按 `user_id` 隔离 Memory：

```go
m := memory.NewMemory("user_1001")
m2 := memory.NewMemory("user_1002")
m.Set("user_name", "张三")  // 只在 user_1001 下面
```

### 8.3 事实自动抽取（1h）

让 LLM 从对话中自动提取 key-value：

```go
extractPrompt := `
请从下面的对话中提取用户的关键信息。

输出 JSON：
{
  "facts": [
    {"key": "user_name", "value": "张三"},
    {"key": "preference", "value": "concise"}
  ]
}

对话：
{conversation}
`
```

### 8.4 Memory 注入策略（30 min）

设计「什么时候把 Memory 注入到 Context」：

```go
func (cm *ContextManager) BuildMessages() []Message {
    messages := []Message{}
    
    // 1. System prompt + 关键 Memory（每次都注入）
    if cm.memory != nil {
        sysMsg := "用户信息：" + cm.memory.GetCore()
        messages = append(messages, openai.SystemMessage(sysMsg))
    }
    
    // 2. 摘要 + 最近消息
    ...
}
```

---

## 🤔 Memory vs RAG 的关系

**RAG（Retrieval-Augmented Generation）** 本质是 Memory 的一种实现：

| | Memory (Map) | RAG (Vector) |
|--|--|--|
| 检索方式 | 精确匹配 key | 语义相似度 |
| 存储内容 | 结构化事实 | 文档片段 |
| 适用场景 | 用户偏好、关键属性 | 知识库、长文档 |
| 复杂度 | 低 | 高（需要 embedding + 向量 DB）|

**按计划**：RAG 不在主线，作为 Memory 的扩展话题。如有需要可单独学习。

---

## 🔗 跟其他模块的关系

| 模块 | 关系 |
|------|------|
| `memory/context.go` | **完全独立**：ContextManager 管当前对话，Memory 管跨对话 |
| `agent/agent.go` | 未来会在 Agent Loop 里调用 `memory.Get()` 注入到 Context |
| `main.go` | 当前演示用 `testMemory()` 简单测试 |

---

## 📊 进度

| 子任务 | 状态 |
|--------|------|
| 理解 Memory 概念 | ✅ |
| 跑通当前最简 Memory | ✅ |
| Redis 持久化 | ⬜ Step 8.1 |
| 用户隔离 | ⬜ Step 8.2 |
| 事实自动抽取 | ⬜ Step 8.3 |
| 注入策略 | ⬜ Step 8.4 |

---

## ✅ 自检（基础级，Step 8 完成后会加进阶版）

- [ ] 1. Memory 和 ContextManager 的本质区别是什么？
- [ ] 2. 当前 Memory 实现的 5 个局限分别是什么？
- [ ] 3. 为什么 Memory 要按 `user_id` 隔离？
- [ ] 4. Memory 和 RAG 是什么关系？
- [ ] 5. 什么信息应该存 Memory，什么不应该？

---

## 📂 关键代码

- `memory/memory.go` (38 行) - 当前最简版实现
- `main.go` `testMemory()` - 演示用法

---

## 📚 推荐阅读

- LangGraph Memory 文档：https://langchain-ai.github.io/langgraph/concepts/memory/
- MemGPT 论文：长上下文 LLM 的虚拟 Memory 管理
- Zep：生产级 Memory 系统参考实现 https://www.getzep.com/

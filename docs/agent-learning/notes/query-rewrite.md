# Query Rewrite 详解

> 📚 对应学习计划：[README §Step 8](../README.md)（Memory 子主题）
> 🗓️ 时间：今天
> 💻 关键代码：[`agent/query_rewrite.go`](../../../code/06-agent/go-agent-demo/agent/query_rewrite.go)（199 行）

---

## 🎯 核心问题

> 用户问「你还记得我的职业吗？」→ 怎么转换成 Memory 系统能搜到的关键词？

直接拿自然语言去搜 Memory，效果往往很差：
- 用户问句：包含助词、语气词（「你还记得」「吗」）
- Memory key：通常是英文 snake_case（如 `user_profession`）

**关键词匹配失败**。

---

## 💡 Query Rewrite 的核心思想

> **不要用用户的原话去检索。先让 LLM 把原话改写成检索关键词。**

```
用户原话：「你还记得我的职业吗？」
    ↓ Query Rewriter (LLM)
关键词列表：["user_profession", "职业", "工作"]
    ↓ Memory Retrieval
命中：user_profession = Go 后端开发工程师 ✅
```

---

## 🔧 实现（`agent/query_rewrite.go`）

### 数据结构

```go
type QueryRewriter struct {
    client *openai.Client
    model  string
}

type QueryRewriteResult struct {
    Queries []string `json:"queries"`
}

func (q *QueryRewriter) Rewrite(
    ctx context.Context,
    userQuery string,
) ([]string, error)
```

### System Prompt（核心设计）

```
你是一个 Memory Query Rewriter。

你的任务是：
把用户的自然语言问题转换成适合长期记忆检索的关键词。

例如：

用户："你还记得我的职业吗？"
返回：
{
  "queries": [
    "user_profession",
    "职业",
    "工作"
  ]
}

用户："我之前说过我想找什么工作？"
返回：
{
  "queries": [
    "career_goal",
    "工作",
    "求职"
  ]
}

重要规则：
1. 不要回答用户的问题
2. 只负责生成检索 Query
3. 不要创建 Memory
4. 不要猜测 Memory 中不存在的信息
5. 优先生成可能对应 Memory key 的英文关键词
6. 可以同时生成中文语义关键词
7. 最多生成 5 个 Query
8. 只输出 JSON
9. 不要输出 Markdown
10. 不要输出解释
```

**为什么这些规则重要？**

| 规则 | 防止的问题 |
|------|----------|
| 不要回答用户问题 | Rewriter 跑去回答问题，忘了改写 |
| 不要创建 Memory | Rewriter 错误地添加新 Memory |
| 优先生成英文 key | 匹配 Memory 的 `user_profession` 这种 snake_case |
| 同时生成中文 | 用语义检索兜底 |
| 最多 5 个 | 避免过多 Query 拖慢检索 |

---

## 📐 Query Rewrite 在 Memory Retrieval 中的位置

```
User Query: "你还记得我的职业吗？"
    ↓
[1. Query Rewrite]  ← agent/query_rewrite.go
    ↓
queries: ["user_profession", "职业", "工作"]
    ↓
[2. 对每个 query 分别检索 Memory]
    ↓
candidate_set: {user_profession, career_goal, ...}
    ↓
[3. Candidate Merge]（同一 Memory 被多个 Query 命中 → 保留最高分）
    ↓
[4. Global Ranking]
    ↓
[5. Top K = 5]
    ↓
[6. 注入 LLM Context]
```

详见 [`agent/memory.go` `buildMemoryContext`](../../../code/06-agent/go-agent-demo/agent/memory.go)。

---

## 🛠️ 关键细节

### 1. 输出清洗（避免解析失败）

```go
// 1. 清理 <think> 标签
content = llm.CleanThinking(content)

// 2. 清理 Markdown Code Fence
content = cleanJSONContent(content)

// 3. JSON Parse
var result QueryRewriteResult
json.Unmarshal([]byte(content), &result)
```

**为什么要两层清洗？**
- 推理类模型（DeepSeek-R1、MiniMax thinking mode）会输出 `<think>...</think>`
- 一些 LLM 会用 ` ```json ... ``` ` 包 JSON
- 不清洗直接解析 → 失败

### 2. 去重 + 限制数量

```go
seen := make(map[string]bool)
var queries []string

for _, query := range result.Queries {
    query = strings.TrimSpace(query)
    if query == "" { continue }
    query = strings.ToLower(query)              // 统一小写
    if seen[query] { continue }                 // 去重
    seen[query] = true
    queries = append(queries, query)
    if len(queries) >= 5 { break }              // 最多 5 个
}
```

**为什么这些处理？**
- LLM 可能输出重复 query
- LLM 可能输出大小写不一致
- LLM 可能输出空字符串
- LLM 可能输出超过 5 个

### 3. 失败兜底

```go
queries, err := a.queryRewriter.Rewrite(ctx, query)
if err != nil {
    fmt.Println("Query Rewrite failed:", err)
    queries = []string{query}   // ← 兜底：用原 query
}
```

**为什么不直接返回 error？**
- Query Rewrite 是「锦上添花」，失败不应阻塞 Agent
- 兜底用原 query，至少 Keyword Search 还能跑

---

## ⚠️ 当前实现的局限

### 局限 1：依赖 LLM 准确改写

Query Rewrite 的质量取决于 LLM 的指令遵循能力。

**生产改进**：
- 用小模型专门训练做改写（成本低）
- 加 few-shot 例子提升稳定性

### 局限 2：英文 key 假设

Prompt 里说「优先生成英文 key」，但实际 Memory key 可能不是英文。

**生产改进**：Prompt 动态化，把现有 Memory 的 key 列表注入。

### 局限 3：没考虑历史 Query

如果用户连续问「我叫什么」→「我做什么工作」，第二问可以结合上下文改写。

**生产改进**：带历史的 Query Rewriter。

### 局限 4：5 个 Query 是硬编码

不同场景最优 Query 数量不同：
- 简单问题：1~2 个就够
- 模糊问题：可能需要 10+ 个

**生产改进**：动态决定数量，或召回后过滤低分结果。

---

## 🤔 为什么用 LLM 改写而不是规则匹配？

| 方案 | 优点 | 缺点 |
|------|------|------|
| **规则匹配** | 简单、快、稳定 | 只能覆盖少量固定模式 |
| **LLM 改写** ✅ | 灵活、能处理任意表达 | 慢、成本高、有不确定性 |

**经验**：小规模固定场景用规则，大规模开放场景用 LLM。

**混合方案**：先用规则匹配常见模式，匹配不上再调 LLM。

---

## 🔮 Query Rewrite 在 RAG 中的应用

Query Rewrite **不只用于 Memory**，在 RAG 里也常用：

```
用户问："Go 怎么排查内存泄漏？"
    ↓ Query Rewrite
queries: ["Go memory leak", "Go 内存泄漏排查", "Go pprof memory"]
    ↓
对每个 query 分别做 Vector Search
    ↓
合并 Top K 结果
```

**本质相同**：把自然语言问题转换为多种检索 query，提高召回率。

---

## 📊 Query Rewrite 的演进路径

```
当前 Demo                  生产级
──────────────────────────────────────
LLM 改写               →    规则 + LLM 混合
英文 key 假设          →    动态注入现有 key
5 个硬编码             →    动态 Top K
无历史                 →    带上下文的改写
无评估                 →    Recall@K / MRR 评估
```

---

## ✅ 自检

- [ ] 1. 为什么需要 Query Rewrite？直接用原 query 检索不行吗？
- [ ] 2. Query Rewrite Prompt 里「优先生成英文 key」的意义是什么？
- [ ] 3. 为什么要两层清洗（`CleanThinking` + `cleanJSONContent`）？
- [ ] 4. Query Rewrite 失败时为什么用原 query 兜底，而不是返回 error？
- [ ] 5. Query Rewrite 在 RAG 里也能用吗？为什么？

---

## 📂 关键代码

- `agent/query_rewrite.go` (199 行) - QueryRewriter 完整实现
- `agent/memory.go` (483 行) - 在 buildMemoryContext 中调用 Rewrite

---

## 🔗 相关笔记

- [Memory 完整体系](./memory.md) - Query Rewrite 的下游应用
- [RAG 完整链路](./rag.md) - Query Rewrite 的另一个应用场景
- [Agent 架构总览](./agent-architecture.md) - Query Rewrite 在整体架构中的位置

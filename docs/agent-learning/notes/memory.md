# Step 8 · Memory 完整体系

> 📚 对应学习计划：[README §Step 8](../README.md)
> 🗓️ 时间：今天（学习 + 实现）
> 💻 关键代码：[`memory/`](../../../code/06-agent/go-agent-demo/memory/)（memory.go 610 行 + extractor.go 120 行 + context.go 212 行）

---

## 🎯 核心问题

> Agent 怎么 **自动** 记住用户事实？怎么 **智能** 检索需要的记忆？

Memory 不是简单的 `map[string]string`。真正的 Agent Memory 是一个 **完整链路**：

```
用户对话 → Memory Extraction → 长期 Memory → Memory Retrieval → 注入 Context
   ↑                                                            ↓
   └──────────────── 下一轮对话继续累积 ←──────────────────────┘
```

---

## 🏗️ Memory 系统完整架构

```
┌────────────────────────────────────────────────────────┐
│                      Agent Loop                          │
│                                                          │
│  User Message                                            │
│     ↓                                                    │
│  [1. Query Rewrite]  ← agent/query_rewrite.go            │
│     ↓                                                    │
│  [2. Hybrid Search]  ← memory/memory.go                  │
│     - Keyword Score                                     │
│     - Vector Score                                      │
│     - Score Fusion (0.5 + 0.5)                          │
│     ↓                                                    │
│  [3. Memory Context]  ← 拼成 system message              │
│     ↓                                                    │
│  [4. LLM Call]                                           │
│     ↓                                                    │
│  [5. Final Answer]                                       │
│     ↓                                                    │
│  [6. Memory Extraction] ← memory/extractor.go            │
│     - 只从 user message 提取                             │
│     - 避免 Assistant 答案污染                            │
│     ↓                                                    │
│  [7. Save Memory]                                        │
└────────────────────────────────────────────────────────┘
```

**7 个步骤**，每一步都有明确职责。

---

## 🧠 Memory 的 3 个核心组件

### 组件 1：Memory 存储（`memory/memory.go`，610 行）

数据结构：

```go
type Memory struct {
    data     map[string]string      // key -> value
    vectors  map[string][]float64   // key -> embedding
    embedder embedding.Embedder     // 用于生成 vector
}
```

**特点**：
- 同时维护**原始数据**和**向量表示**
- 每次 `Save(key, value)` 自动生成 vector
- 支持多种检索方式

### 组件 2：Memory Extractor（`memory/extractor.go`，120 行）

**职责**：从用户消息中提取值得长期保存的事实。

```go
type ExtractedMemory struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type ExtractMemoryResult struct {
    ShouldSave bool              `json:"should_save"`
    Memories   []ExtractedMemory `json:"memories"`
}
```

**核心 Prompt**（节选）：

```
适合保存的信息：
- 用户姓名、职业、技术方向、技能
- 长期目标、用户偏好、长期项目
- 稳定的个人背景信息

不要保存：
- 普通闲聊、一次性临时问题
- 当前订单查询结果、当前时间
- 无长期价值的信息
```

### 组件 3：Memory Retrieval（`memory/memory.go` 内）

**多种检索策略**：

| 方法 | 实现 | 适用 |
|------|------|------|
| `Search()` | 简单字符串包含 | 精确 key 已知 |
| `SearchRelevantTopK()` | 关键词打分（10/5/2）| 关键词匹配 |
| `SearchVectorTopK()` | Cosine Similarity | 语义相似 |
| `SearchHybridTopK()` | 0.5 × Keyword + 0.5 × Vector | 综合最优 |

---

## 🔍 Query Rewrite：检索的前置优化

### 核心问题

用户问：「你还记得我的职业吗？」

直接拿这句话去搜 Memory，可能**搜不到**（如果 Memory 里存的是 `user_profession: Go 后端开发工程师`）。

**Query Rewrite** = 用 LLM 把自然语言问题转换为适合 Memory 检索的关键词。

### 实现（`agent/query_rewrite.go`，199 行）

```go
type QueryRewriteResult struct {
    Queries []string `json:"queries"`
}

func (q *QueryRewriter) Rewrite(
    ctx context.Context,
    userQuery string,
) ([]string, error)
```

### Query Rewrite Prompt 关键设计

```
你是一个 Memory Query Rewriter。

把用户的自然语言问题转换成适合长期记忆检索的关键词。

例如：
  "你还记得我的职业吗？"
  → ["user_profession", "职业", "工作"]

  "我之前说过我想找什么工作？"
  → ["career_goal", "工作", "求职"]

重要规则：
1. 不要回答用户的问题
2. 只负责生成检索 Query
3. 不要创建 Memory
4. 不要猜测 Memory 中不存在的信息
5. 优先生成可能对应 Memory key 的英文关键词
6. 最多生成 5 个 Query
```

**核心思路**：用英文 key + 中文语义词双管齐下，兼容 Keyword 和 Vector 两种检索。

---

## 🎯 Memory Extraction：避免污染的关键

### 核心问题（来自实际调试）

测试场景：
1. Session 1：用户说「我叫张三，我是 Go 后端开发工程师」
2. Session 2：用户问「你还记得我的年龄吗？」
3. ❌ **Bug**：Memory 里多出 `user_age: 28`（LLM 自己猜的）

### 根因

Memory Extraction 时，**把 Assistant 的回答也作为输入**给 LLM，让 LLM 提取 Memory。LLM 看到 Assistant 说"我猜您 28 岁"，就把这个当成事实保存。

### 解决方案（`agent/memory.go` `extractAndSaveMemory`）

```go
// ★ 只把 userQuery 交给 Memory Extraction ★
userPrompt := fmt.Sprintf(
    "用户消息：\n%s",
    userQuery,    // ← 注意：没有 assistantAnswer
)

resp, err := a.client.Chat.Completions.New(ctx, ...)
```

### 强化 Prompt

```
1. 不要根据用户的问题推测答案
2. 不要根据助手回答推测用户事实
3. 不要把助手自己的描述保存为用户事实
4. 如果用户只是询问某个 Memory 是否存在，不要创建新的 Memory
5. 如果用户没有明确提供某项信息，不要保存该信息
6. 不要根据上下文猜测用户属性
```

**这是 Agent Memory 设计的核心难点**：**只保存用户明确表达的事实**。

---

## 🔬 Hybrid Search：Keyword + Vector 双引擎

### 为什么需要 Hybrid？

| 检索方式 | 优势 | 劣势 |
|---------|------|------|
| Keyword（关键词）| 精确 key 匹配强（`user_profession` 完全匹配得 10 分）| 中文语义、错别字、同义词差 |
| Vector（向量）| 语义相似、跨语言 | 精确 key 弱 |

**结论**：两种互补，结合最好。

### Hybrid Score 公式

```go
// memory/memory.go SearchHybridTopK
hybridScore = 0.5 * keywordScore + 0.5 * vectorScore
```

其中：
- `keywordScore` = 关键词得分 / 最大关键词得分（**归一化到 [0, 1]**）
- `vectorScore` = Cosine Similarity（天然在 [0, 1]）

### 关键词打分规则

```go
// 完全匹配 key（如 query="user_profession", key="user_profession"）
score += 10

// key 部分包含（如 query="profession", key="user_profession"）
score += 5

// value 部分包含（如 value 含 "Go"）
score += 2
```

### 为什么 Keyword Score 要归一化？

如果不归一化：
- Keyword 最高可能得 100 分
- Vector 最高 1 分

直接相加 → Keyword 完全压制 Vector。

归一化后两者权重才真正平衡。

---

## 🧪 Agent Loop 中的完整 Memory 链路

### `agent/agent.go` Agent Loop（核心 7 步）

```go
for {
    // 1. Context 摘要检查
    a.contextManager.MaybeSummarize(...)

    // 2. Memory Retrieval：检索相关长期记忆
    memoryContext := a.buildMemoryContext(ctx, userQuery)

    // 3. 构造消息（含 Memory Context）
    messages := buildMessages(system, memoryContext, summary, recent)

    // 4. 调用 LLM
    resp := a.client.Chat.Completions.New(ctx, ...)

    // 5. 处理 Tool Calls / 得到 Final Answer
    ...

    // 6. Memory Extraction：保存新事实
    a.extractAndSaveMemory(ctx, userQuery, assistantAnswer)

    // 7. 进入下一轮（或结束）
}
```

### Memory 注入策略

**当前实现**：把检索到的 Memory 拼成 system message。

```
[长期记忆检索结果]
以下是从用户历史长期记忆中检索到的相关信息。
只能使用这些已有信息，不要自行猜测。

- user_name: 张三
- user_profession: Go 后端开发工程师
- career_goal: 正在寻找 Go 后端开发相关的工作
```

**关键 prompt 措辞**：
- 「只能使用这些已有信息」→ 防幻觉
- 「不要自行猜测」→ 防污染
- 「如果无法从当前上下文得到答案，应明确说明没有相关记录」→ 兜底

---

## 🛡️ Memory 系统设计的关键原则

### 原则 1：Memory Extraction 只看 User Message

❌ 把 user + assistant 都给 LLM → 容易污染
✅ 只把 user 给 LLM → 安全

### 原则 2：检索失败时明确告知 LLM

❌ 不告诉 LLM「没找到」→ LLM 会瞎猜
✅ 明确告诉「没找到」+「不要猜测」→ LLM 老实回答

### 原则 3：User 主动询问的 Memory 不能创建

❌ 用户问「你记得我的 user_role 吗？」→ 自动创建 `user_role: unknown`
✅ 必须用户明确提供信息才能创建

### 原则 4：Memory Save 失败不应影响主流程

```go
// 错误处理：Memory 失败不能让整个 Agent 挂掉
if err := a.extractAndSaveMemory(ctx, userQuery, answer); err != nil {
    fmt.Println("Memory save failed:", err)
    // 不返回 error，继续往下走
}
```

---

## ⚠️ 当前实现的局限

### 局限 1：Fake Embedding 不是真语义

当前 `FakeEmbedder` 用字符 Hash 生成向量，**只是演示用**。

```go
// embedding/embedding.go
func (e *FakeEmbedder) Embed(text string) ([]float64, error) {
    vector := make([]float64, e.Dimension)
    for _, r := range text {
        index := int(r) % e.Dimension
        vector[index] += 1   // ← 只是字符计数
    }
    normalize(vector)
    return vector, nil
}
```

**生产应该**：接 OpenAI Embedding API 或开源模型（BGE / M3E 等）。

### 局限 2：Memory 不持久化

进程退出 Memory 就没了。

**生产应该**：Redis / PostgreSQL / 向量数据库。

### 局限 3：Hybrid Search 权重固定

`0.5 × Keyword + 0.5 × Vector` 是硬编码。

**生产应该**：根据场景调权重，或加 Rerank 阶段。

### 局限 4：Output Parser 脆弱

LLM 返回的 JSON 偶尔带 `<think>` 标签、Markdown code fence。

```go
// llm/output.go CleanThinking 处理 <think> 标签
// agent/memory.go cleanJSONContent 处理 ```json``` 代码块
```

**生产应该**：用 OpenAI Structured Outputs 或 instructor 库。

### 局限 5：没有 Memory 冲突解决

如果两次保存同一个 key，新值覆盖旧值，没有版本管理。

**生产应该**：保留历史 + 冲突标记。

---

## 📊 演进路径：从 Demo 到生产

```
当前 Demo                 生产级
─────────────────────────────────────────────
FakeEmbedder         →    OpenAI/Cohere/BGE Embedding
内存 map              →    Redis + 向量 DB（Qdrant/Milvus）
简单 Hybrid Score     →    BM25 + Vector + Rerank
JSON 解析             →    Structured Outputs / instructor
进程内 Memory         →    持久化 + 多租户隔离
无冲突解决            →    版本化 + 冲突标记
```

---

## 🔗 跟其他模块的关系

| 模块 | 关系 |
|------|------|
| `memory/context.go` | **不同**：ContextManager 管当前对话，Memory 管跨对话 |
| `agent/query_rewrite.go` | Memory Retrieval 的前置优化 |
| `agent/agent.go` | Agent Loop 中调用 Memory Retrieval + Extraction |
| `embedding/` | Memory 用来生成 vector |
| `llm/output.go` | Memory Extraction 解析 LLM JSON 时用 |

---

## ✅ 自检

- [ ] 1. Agent Memory 完整链路的 7 个步骤分别是什么？
- [ ] 2. 为什么 Memory Extraction 只看 User Message，不看 Assistant Answer？
- [ ] 3. Hybrid Score 公式是什么？为什么要归一化 Keyword Score？
- [ ] 4. Query Rewrite 解决的是什么问题？
- [ ] 5. 当前 Memory 实现的 5 个局限分别是什么？

---

## 📂 关键代码

- `memory/memory.go` (610 行) - Memory 存储 + 多种检索
- `memory/extractor.go` (120 行) - Memory Extraction
- `agent/query_rewrite.go` (199 行) - Query Rewriter
- `agent/memory.go` (483 行) - Agent Loop 中的 Memory 集成
- `embedding/embedding.go` (142 行) - FakeEmbedder + Cosine Similarity
- `llm/output.go` (32 行) - Output Parser（CleanThinking）

---

## 🔗 相关笔记

- [Query Rewrite 详解](./query-rewrite.md) - Memory 的前置优化
- [Embedding & Vector Search](./embedding-vector-search.md) - 检索的数学基础
- [RAG 完整链路](./rag.md) - Memory 的孪生兄弟（知识库检索）
- [Agent 架构总览](./agent-architecture.md) - 3 大支柱（Memory / RAG / Tools）

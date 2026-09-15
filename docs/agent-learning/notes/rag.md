# RAG 完整链路详解

> 📚 对应学习计划：[README §Step 8](../README.md)（RAG 作为 Memory 的扩展）
> 🗓️ 时间：今天
> 💻 关键代码：[`rag/`](../../../code/06-agent/go-agent-demo/rag/)（knowledge.go 92 行 + retriever.go 76 行 + rag.go 97 行 = 265 行）

---

## 🎯 核心问题

> 不能每次都把整个 PDF 塞给 LLM。怎么让 Agent 从大量文档中找到相关内容回答用户？

**RAG（Retrieval-Augmented Generation）= 先检索相关知识，再让 LLM 基于检索结果生成答案。**

---

## 📊 RAG vs Memory vs Tools（核心区分）

这是 Agent 学习的**最重要的概念区分之一**：

| 维度 | Memory | RAG | Tools |
|------|--------|-----|-------|
| **存什么** | 用户事实、偏好 | 知识库文档 | 外部 API |
| **回答什么** | "你是谁" | "XX 怎么用" | "查订单" |
| **例子** | `user_profession=Go` | "Go GMP 调度详解" | `query_order()` |
| **典型来源** | 用户对话提取 | 文档/PDF/网页 | 业务系统 |

> 一句话：
> - **Memory**：记住"用户是谁"
> - **RAG**：记住"知识是什么"
> - **Tools**：操作"外部世界"

---

## 🏗️ RAG 完整架构

```
                    离线阶段（Indexing）
┌─────────────────────────────────────────────┐
│ Document (PDF/Markdown/网页)                 │
│      ↓ 切分                                   │
│ Chunks (文本片段)                            │
│      ↓ Embedding                              │
│ Vectors (向量)                                │
│      ↓ 存入                                   │
│ Vector DB / 内存                              │
└─────────────────────────────────────────────┘

                    在线阶段（Retrieval + Generation）
┌─────────────────────────────────────────────┐
│ User Query                                    │
│      ↓ Embedding                              │
│ Query Vector                                  │
│      ↓ Vector Search                          │
│ Top K Chunks (相关知识片段)                   │
│      ↓ 拼成 Context                           │
│ LLM (基于 Context 生成答案)                  │
│      ↓                                        │
│ Final Answer                                  │
└─────────────────────────────────────────────┘
```

**两阶段**：离线建索引 + 在线检索生成。

---

## 🔧 当前 Demo 实现

### 1. Document & Chunk（`rag/knowledge.go`）

```go
type Document struct {
    ID      string
    Content string
}

type Chunk struct {
    ID         string
    DocumentID string
    Content    string
    Vector     []float64
}
```

**文档被切成多个 Chunk**，每个 Chunk 单独 Embedding，单独检索。

### 2. 切分逻辑（最简单的版本）

```go
func splitDocument(content string) []string {
    parts := strings.Split(content, "\n\n")  // 按双换行切
    result := make([]string, 0)

    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part != "" {
            result = append(result, part)
        }
    }

    return result
}
```

**最朴素的按段落切分**。生产应该用 LangChain 的 RecursiveCharacterTextSplitter 或类似工具。

### 3. KnowledgeBase（`rag/knowledge.go`）

```go
type KnowledgeBase struct {
    chunks   []Chunk
    embedder embedding.Embedder
}

func (kb *KnowledgeBase) AddDocument(doc Document) error {
    parts := splitDocument(doc.Content)

    for index, content := range parts {
        vector, err := kb.embedder.Embed(content)
        if err != nil {
            return err
        }

        kb.chunks = append(kb.chunks, Chunk{
            ID: fmt.Sprintf("%s-chunk-%d", doc.ID, index+1),
            DocumentID: doc.ID,
            Content: content,
            Vector: vector,
        })
    }

    return nil
}
```

**核心**：每个 Chunk 同步生成 vector，存入 KnowledgeBase。

### 4. Retriever（`rag/retriever.go`）

```go
type Retriever struct {
    knowledgeBase *KnowledgeBase
    embedder      embedding.Embedder
}

func (r *Retriever) Retrieve(query string, topK int) ([]RetrievalResult, error) {
    queryVector, err := r.embedder.Embed(query)

    results := make([]RetrievalResult, 0, len(r.knowledgeBase.Chunks()))

    for _, chunk := range r.knowledgeBase.Chunks() {
        similarity := embedding.CosineSimilarity(queryVector, chunk.Vector)

        results = append(results, RetrievalResult{
            Chunk: chunk,
            Similarity: similarity,
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Similarity > results[j].Similarity
    })

    if topK > len(results) {
        topK = len(results)
    }

    return results[:topK], nil
}
```

**核心**：遍历所有 Chunk → 计算 Cosine Similarity → 排序 → Top K。

### 5. RAG Ask（`rag/rag.go`）

完整 RAG 链路：

```go
func (r *RAG) Ask(ctx context.Context, query string) (string, error) {
    // 1. 检索
    results, err := r.Retrieve(ctx, query)

    // 2. 拼 Context
    var contextBuilder strings.Builder
    for _, result := range results {
        contextBuilder.WriteString(fmt.Sprintf("- %s\n", result.Chunk.Content))
    }

    // 3. 拼 Prompt
    prompt := fmt.Sprintf(`请根据下面提供的知识回答用户问题。

要求：
1. 优先使用提供的知识回答
2. 不要编造知识中不存在的信息
3. 如果知识中没有答案，请明确说明无法从知识库中找到答案

知识：
%s

用户问题：
%s`, contextBuilder.String(), query)

    // 4. 调 LLM
    resp, err := r.client.Chat.Completions.New(ctx, ...)
    return resp.Choices[0].Message.Content, nil
}
```

---

## 🛡️ RAG Prompt 的关键设计

### 三条铁律（缺一不可）

```
1. 优先使用提供的知识回答
2. 不要编造知识中不存在的信息
3. 如果知识中没有答案，请明确说明无法从知识库中找到答案
```

**为什么这三条都重要？**

| 规则 | 没有会怎样 |
|------|----------|
| 优先使用知识 | LLM 可能忽略检索结果，凭自己"幻觉"回答 |
| 不要编造 | LLM 可能基于半真半假的信息瞎编 |
| 没答案就说没 | LLM 不会承认不知道，会硬凑答案 |

---

## 📊 RAG 演进路径

```
Demo 版                   生产级
─────────────────────────────────────
朴素段落切分        →    RecursiveCharacterTextSplitter
全部 Embedding      →    增量 Embedding
内存存储            →    Qdrant / Milvus / pgvector
无 Metadata         →    文档来源、时间、标签
无 Rerank           →    BGE Reranker / Cohere Rerank
无混合检索          →    Keyword + Vector + Rerank
无评估              →    Recall@K / MRR / NDCG
```

---

## ⚠️ 常见误区

### 误区 1：以为上了 Embedding 就是 RAG 了

❌ 错。
- 只有 **KnowledgeBase + Retriever + Ask** 三个组件齐全才是完整 RAG
- 只 Embedding 不检索不生成 → 只是个"向量存储"

### 误区 2：Chunk 越大越好

❌ 错。
- Chunk 太大 → 检索精确但 LLM 上下文被噪声淹没
- Chunk 太小 → 检索召回率高但缺上下文

**经验值**：256~512 token，视文档类型调。

### 误区 3：相似度高就是好结果

❌ 不一定。
- Embedding 模型本身的语义理解能力是上限
- 高相似度但 Chunk 内容不相关 → 是 Embedding 模型的局限

**解决**：加 Rerank 阶段精排。

### 误区 4：所有文档都该用 RAG

❌ 错。
- 实时性要求高 → 不用 RAG，直接调 API
- 数据量小 → 直接塞 Context 就行
- 数据量大 + 知识型查询 → ✅ RAG 最合适

---

## 🤔 实际项目中的决策树

```
Agent 需要查外部知识？
├── 知识经常变化（订单状态、库存）
│   └── 不用 RAG，用 Tools 调实时 API
├── 知识稳定但量大（产品手册、文档）
│   └── 用 RAG
├── 知识稳定且量小（FAQ）
    └── 直接塞 System Prompt，不用 RAG
└── 多种混合
    └── Tools + RAG 共存，按场景路由
```

---

## 🔗 RAG 与其他模块的关系

| 模块 | 关系 |
|------|------|
| `embedding/` | RAG 依赖 Embedder 生成向量 |
| `agent/` | Agent 调用 RAG.Ask() 获取答案 |
| `memory/` | **RAG 和 Memory 是兄弟**，都做检索 |

### RAG vs Memory 的本质区别

| | Memory | RAG |
|--|--|--|
| **存什么** | 用户事实（key-value）| 知识片段（chunk）|
| **检索目标** | 用户的稳定属性 | 任务相关的知识 |
| **写入方式** | 用户对话提取 | 离线索引 |
| **更新频率** | 每次对话 | 文档更新时 |

> **可以这么理解**：Memory 是"用户档案"，RAG 是"知识库"。

---

## 🔮 下一步可深挖方向

| 方向 | 内容 |
|------|------|
| **真实 Embedding** | 接 OpenAI / BGE |
| **真实向量库** | Qdrant / Milvus / pgvector |
| **Hybrid Search** | Keyword + Vector（已在 Memory 实现，RAG 可复用）|
| **Rerank** | Cohere Rerank / BGE Reranker |
| **Streaming Ask** | 流式输出 RAG 答案 |
| **多模态 RAG** | 图像 + 表格的检索 |

---

## ✅ 自检

- [ ] 1. RAG 完整链路的两个阶段（离线 / 在线）分别做什么？
- [ ] 2. RAG 和 Memory 的本质区别是什么？
- [ ] 3. Chunk 切分太大会怎样？太小会怎样？
- [ ] 4. RAG Prompt 的三条铁律分别防什么？
- [ ] 5. 什么时候该用 RAG，什么时候该直接调 API？

---

## 📂 关键代码

- `rag/knowledge.go` (92 行) - Document / Chunk / KnowledgeBase
- `rag/retriever.go` (76 行) - Vector Search 检索
- `rag/rag.go` (97 行) - 完整 RAG Ask 链路

---

## 🔗 相关笔记

- [Embedding & Vector Search](./embedding-vector-search.md) - RAG 的数学基础
- [Memory 完整体系](./memory.md) - RAG 的孪生兄弟
- [Agent 架构总览](./agent-architecture.md) - 3 大支柱全景图

# Embedding & Vector Search 详解

> 📚 对应学习计划：[README §Step 8](../README.md)（Memory 子主题）
> 🗓️ 时间：今天
> 💻 关键代码：[`embedding/`](../../../code/06-agent/go-agent-demo/embedding/)（142 行）

---

## 🎯 核心问题

> 计算机怎么理解「Go 后端开发」和「找 Go 工作」在语义上很相近？

传统关键词匹配：**字符串包含** → 「Go 后端开发」包含「Go」吗？是的。包含「工作」吗？是的。得 2 分。

**问题**：用户问「我想找什么工作？」→ 关键词「工作」匹配 Memory `user_profession: Go 后端开发工程师」value 部分 → 得 2 分。但语义上「想找工作」和「user_profession」其实是**强相关**的。

**Embedding + Vector Search** 解决的就是这个问题。

---

## 🧠 Embedding 是什么

### 一句话定义

**Embedding = 把文本转换成数学向量，语义相近的文本向量也相近。**

```
"Go 后端开发工程师"
    ↓ Embedding
[0.12, -0.34, 0.56, ..., 0.78]   ← 1536 维向量（OpenAI text-embedding-3-small）

"找 Go 后端开发的工作"
    ↓ Embedding
[0.11, -0.32, 0.55, ..., 0.77]   ← 跟上面很接近！

"今天天气不错"
    ↓ Embedding
[-0.89, 0.23, -0.45, ..., 0.12]  ← 完全不同的方向
```

### 为什么有效？

- 训练数据中「Go 后端开发」和「Go 工作」经常出现在相似上下文
- 模型学到了它们语义相近
- 把这种语义关系编码到向量空间的几何关系中

---

## 🔧 当前 Demo 的 Embedding 实现

### `FakeEmbedder`（142 行）

```go
type FakeEmbedder struct {
    Dimension int
}

func (e *FakeEmbedder) Embed(text string) ([]float64, error) {
    vector := make([]float64, e.Dimension)
    text = strings.ToLower(strings.TrimSpace(text))

    if text == "" {
        return vector, nil
    }

    for _, r := range text {
        index := int(r) % e.Dimension
        vector[index] += 1
    }

    normalize(vector)
    return vector, nil
}
```

### ⚠️ 这是假的 Embedding

**它只是字符计数 + 归一化**：

| 文本 | Fake Vector |
|------|------------|
| "Go" | [0, 1, 0, 0, 0, 1, 0, 0] |
| "Go 后端" | [0, 1.4, 0, 0.7, 0, 1.4, 0, 0] |
| "找工作" | [0, 0, 1, 0, 0, 0, 0, 1.4] |

「Go 后端」和「找工作」的 Fake Vector 完全不同 → FakeEmbedder **不具备语义检索能力**。

### 为什么用 Fake？

1. **避免依赖外部 API**：学习阶段不想让你纠结 API Key 和成本
2. **强制理解机制**：你能清晰看到"向量"和"相似度"是怎么算出来的
3. **接口对齐**：生产换真 Embedding 时，**Memory 层代码不用改**

### 接口设计

```go
type Embedder interface {
    Embed(text string) ([]float64, error)
}
```

只要实现这个接口，就能换：
- `OpenAIEmbedder`（接 text-embedding-3-small）
- `BGEEmbedder`（开源 BGE 模型）
- `FakeEmbedder`（学习用）

**Memory 层只依赖 Embedder 接口，不知道具体用了哪个**。

---

## 📐 Cosine Similarity（余弦相似度）

### 为什么用余弦而不是欧氏距离？

| 度量 | 含义 | 特点 |
|------|------|------|
| 欧氏距离 | 两点直线距离 | 受向量长度影响 |
| **余弦相似度** | 两向量夹角 | **只看方向，不看长度** |

**对文本 Embedding 来说，长度通常不重要**（长文本自然向量大），重要的是「方向」（语义方向）。

### 公式

```
                A · B          (点积)
cos(A, B) = ───────────── = ──────────────
             ||A|| × ||B||    (A 的模 × B 的模)
```

返回值：[-1, 1]
- **1** = 完全相同方向（语义相同）
- **0** = 正交（不相关）
- **-1** = 完全相反（语义相反）

### 实现（`embedding/embedding.go`）

```go
func CosineSimilarity(a []float64, b []float64) float64 {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    if len(a) != len(b) {
        return 0
    }

    var dotProduct, normA, normB float64

    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }

    if normA == 0 || normB == 0 {
        return 0
    }

    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

**3 个关键点**：
1. 长度不一致 → 返回 0（防御性）
2. 任一向量为零向量 → 返回 0
3. 否则按公式计算

---

## 🔍 Vector Search（向量检索）

### 流程

```
Query "你还记得我的职业吗？"
    ↓ Embedder
queryVector = [0.1, -0.2, 0.3, ...]
    ↓
遍历 Memory 中所有 vectors
    ↓ Cosine Similarity
[user_name: 0.05, user_profession: 0.78, career_goal: 0.45, ...]
    ↓ Sort + Top K
返回 Top 3
```

### 实现（`memory/memory.go` `SearchVectorTopK`）

```go
func (m *Memory) SearchVectorTopK(query string, topK int) []VectorSearchResult {
    if m.embedder == nil {
        return nil
    }

    queryVector, err := m.embedder.Embed(query)
    if err != nil {
        return nil
    }

    results := make([]VectorSearchResult, 0, len(m.data))

    for key, value := range m.data {
        vector, exists := m.vectors[key]
        if !exists {
            continue
        }

        similarity := embedding.CosineSimilarity(queryVector, vector)

        results = append(results, VectorSearchResult{
            Item: MemoryItem{Key: key, Value: value},
            Similarity: similarity,
        })
    }

    sort.Slice(results, func(i, j int) bool {
        return results[i].Similarity > results[j].Similarity
    })

    if len(results) > topK {
        results = results[:topK]
    }

    return results
}
```

**核心步骤**：
1. Embedder 把 query 转成向量
2. 跟每个 Memory 的向量算 Cosine Similarity
3. 按相似度降序排序
4. 取 Top K

---

## 🧪 Embedding 向量什么时候生成？

### 当前实现：保存 Memory 时同步生成

```go
// memory/memory.go Save
func (m *Memory) Save(key string, value string) {
    m.data[key] = value

    // ★ 保存 Memory 的同时生成 Vector ★
    if m.embedder != nil {
        vector, err := m.embedder.Embed(key + " " + value)
        if err == nil {
            m.vectors[key] = vector
        }
    }
}
```

**为什么 `key + " " + value`？**  
让 key 和 value 都参与 Embedding，这样检索 query 时（可能只含 key 或只含 value）都能命中。

### 生产环境考虑

| 方案 | 优缺点 |
|------|--------|
| 同步生成 | 简单，但 Embedding 调用阻塞 Save |
| 异步生成 | Save 快，但检索时可能 vector 还没生成 |
| 单独索引 | 生产级，解耦存储和检索 |

---

## 📊 真实 Embedding 模型对比

| 模型 | 维度 | 特点 |
|------|------|------|
| OpenAI `text-embedding-3-small` | 1536 | 便宜，通用 |
| OpenAI `text-embedding-3-large` | 3072 | 更准，更贵 |
| BGE-large-zh-v1.5 | 1024 | 中文开源首选 |
| M3E | 1024 | 中文，效果略弱于 BGE |
| FakeEmbedder（Demo 用）| 8 | **完全无语义能力** |

---

## ⚠️ Vector Search 的常见误区

### 误区 1：以为换了 Embedding 就解决了语义检索

❌ 错。Embedding 只是把语义编码成向量，**能不能检索到取决于 Embedding 模型本身**。
- FakeEmbedder → 没语义能力
- OpenAI text-embedding-3 → 强语义能力

### 误区 2：Vector Search 一定优于 Keyword Search

❌ 错。
- 精确 key 匹配（`user_profession` 完全匹配）：Keyword 远超 Vector
- 中文近义词（"工作"/"职业"/"求职"）：Vector 远超 Keyword

**结论**：Hybrid Search 是最优解（详见 [memory.md](./memory.md)）。

### 误区 3：维度越高越好

❌ 不一定。
- 维度高 → 表达能力强，但存储和计算成本也高
- 维度低 → 便宜，但可能不够区分细微语义

**生产建议**：根据场景选维度，中文一般 1024 已经够用。

---

## 🔮 下一步：从 Fake Embedding 到真实 API

### 改造示例

```go
// embedding/openai.go（待实现）
type OpenAIEmbedder struct {
    client *openai.Client
    model  string
}

func (e *OpenAIEmbedder) Embed(text string) ([]float64, error) {
    resp, err := e.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
        Input: openai.EmbeddingNewParamsInputUnion{
            OfString: openai.String(text),
        },
        Model: openai.EmbeddingModel(e.model),
    })
    if err != nil {
        return nil, err
    }

    vector := make([]float64, len(resp.Data[0].Embedding))
    for i, v := range resp.Data[0].Embedding {
        vector[i] = float64(v)
    }

    return vector, nil
}
```

**关键点**：实现 `Embedder` 接口，**Memory 层代码 0 改动**。

---

## ✅ 自检

- [ ] 1. Embedding 是什么？为什么需要它？
- [ ] 2. 为什么 Vector Search 用 Cosine Similarity 而不是欧氏距离？
- [ ] 3. 当前 `FakeEmbedder` 为什么不具备真正的语义检索能力？
- [ ] 4. `Memory.Save` 时为什么要 `key + " " + value` 一起 Embed？
- [ ] 5. 真实 Embedding 模型跟 FakeEmbedder 的核心区别是什么？

---

## 📂 关键代码

- `embedding/embedding.go` (142 行) - Embedder 接口 + FakeEmbedder + CosineSimilarity

---

## 🔗 相关笔记

- [Memory 完整体系](./memory.md) - 用 Embedding 实现 Vector Search 的上层应用
- [RAG 完整链路](./rag.md) - Embedding + Vector Search 的另一个主要场景

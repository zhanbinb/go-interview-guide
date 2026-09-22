# Agentic RAG 完整链路 + 求职定位

> 📚 配套笔记：[rag.md](./rag.md)（RAG 基础）/ [langchain-langgraph.md](./langchain-langgraph.md)（LangChain 基础）
> 🗓️ 时间：今天（基于 ChatGPT 分享 [6aacda0a](https://chatgpt.com/share/6aacda0a-7680-83e8-9f76-32e96c598d13) + `langchain-agent-demo/17-30`）
> 🎯 主题：Agentic RAG 完整工程化 + 求职定位

> 💡 **本笔记覆盖**：之前 `rag.md` 没深入的 4 个生产话题：
> 1. 真实 Embedding（BGE vs Fake）
> 2. Metadata Filter（结构化过滤）
> 3. Reranker（精排，Cross-Encoder 概念）
> 4. VectorStore 持久化（offline / online 分离）
> + 求职定位建议

---

## 🎯 核心问题

> RAG Demo 跑通了，怎么升级到生产级？还有，作为 Go 后端 + Agent 学习者，**怎么定位自己的职业方向**？

之前 `rag.md` 写的 RAG 还是 Demo 版（Fake Embedding、内存存储、纯关键词匹配）。生产级 RAG 需要：
- **真实 Embedding 模型**
- **专业 Chunking 策略**
- **向量数据库**（FAISS / Milvus / pgvector）
- **多阶段检索**（粗召回 + 精排 + 过滤）
- **离线 / 在线分离**

---

## 1️⃣ 完整 RAG Pipeline（Demos 17-30）

### 演进路线

```
17_agentic_rag             → Agent + Tool(关键词搜索)
18_agentic_rag_retriever   → 抽出 Retriever 概念
19_agentic_rag_embedding   → 接入真实 Embedding
20_agentic_rag_vectorstore → FAISS VectorStore
21_rag_chunking            → RecursiveCharacterTextSplitter
22_rag_chunking_vectorstore→ Chunking + VectorStore 组合
23_agentic_rag_full        → 完整 Agentic RAG 链
24_rag_indexing            → Offline 索引（VectorStore 单独构建）
25_rag_retrieval           → Online 检索（只加载，不重建）
26_agentic_rag_persistent  → FAISS 序列化到磁盘
27_rag_metadata            → Metadata Filter
28_agentic_rag_metadata    → Agent + Metadata Filter
29_rag_reranker            → Vector Search + Reranker
30_agentic_rag_reranker    → Agent + Reranker
```

### 完整 Pipeline（Demo 23 + 24 + 25 + 26 + 27 + 29 综合）

```
【离线阶段 Indexing】一次构建，保存到磁盘
═══════════════════════════════════════════════════════════════

Markdown 文档
    ↓
Document Loader（17-23）
    ↓
Chunking（RecursiveCharacterTextSplitter）
    ↓ chunk_size=80, chunk_overlap=20
Chunks
    ↓
Embedding（BAAI/bge-small-zh-v1.5）
    ↓ 真实语义向量
VectorStore（FAISS.from_documents）
    ↓
保存到 faiss_index/ 目录
    ↓
pickle 序列化


【在线阶段 Retrieval】每次请求，加载即可
═══════════════════════════════════════════════════════════════

用户 Query
    ↓
Query Rewrite / Multi-Query（Step 6 学过）
    ↓
Vector Search（FAISS.similarity_search k=5）
    ↓ 粗召回 Top-5
Metadata Filter（filter={"category": "golang"}）
    ↓ 缩小范围
Reranker（Cross-Encoder）
    ↓ 精排 Top-N
Top-3 Chunks + Metadata
    ↓
LLM 生成最终答案
```

---

## 2️⃣ 真实 Embedding（BGE）

### vs Fake Embedder

| 维度 | Fake Embedder（之前） | BAAI/bge-small-zh-v1.5（现在） |
|------|---------------------|-------------------------------|
| 维度 | 8（手动字符 Hash）| 512（真实语义向量）|
| 语义 | ❌ 完全无语义 | ✅ 中文语义最强 |
| 速度 | 极快 | 中等（首次需下载模型）|
| 模型大小 | 0 KB | ~93 MB |
| 接口 | `embedding.Embed(text)` | `sentence_transformers.SentenceTransformer` |

### 集成代码（Demo 27 + 29）

```python
from sentence_transformers import SentenceTransformer

class LocalEmbedding(Embeddings):
    def __init__(self):
        self.model = SentenceTransformer("BAAI/bge-small-zh-v1.5")
    
    def embed_documents(self, texts):
        return self.model.encode(texts).tolist()
    
    def embed_query(self, text):
        return self.model.encode(text).tolist()

embedding = LocalEmbedding()
```

**关键认知**：
- LangChain 的 `Embeddings` 接口 = 任何 Embedding 实现的抽象
- 之前 FakeEmbedder 是为学习，现在 BGE 是真实使用
- **Demo 代码无需大改**，只换实现

### 真实 Embedding 选型

| 模型 | 维度 | 特点 |
|------|------|------|
| `BAAI/bge-small-zh-v1.5` | 512 | 中文首选，本地可跑 |
| `BAAI/bge-large-zh-v1.5` | 1024 | 中文更强，更慢 |
| `text-embedding-3-small` (OpenAI) | 1536 | 通用，便宜 |
| `text-embedding-3-large` (OpenAI) | 3072 | 通用，更准 |
| `BAAI/bge-m3` | 1024 | 多语言 |

---

## 3️⃣ Metadata Filter（结构化过滤）

### 核心问题

Vector Search 是**语义相似**，但企业场景常常需要**结构化过滤**：

```
"Go 的 channel 是什么"  → 语义搜索：返回含 "channel" 的所有内容
                        → 但 MySQL 文档里也提到 "channel"
                        → 不准确！

加上 Metadata Filter：
"category=golang" → 只在 Go 文档里搜 → 精准
```

### 实现（Demo 27）

```python
from langchain_core.documents import Document

# 1. 加载时加 Metadata
documents.append(Document(
    page_content=content,
    metadata={
        "source": file.name,
        "category": "golang",   # ← 关键
        "title": "Go 并发编程",
    },
))

# 2. 检索时 Filter
results = vectorstore.similarity_search(
    query,
    k=3,
    filter={"category": "golang"},   # ← Metadata Filter
)
```

### 为什么 Metadata Filter 重要？

| 场景 | 没用 Filter | 用了 Filter |
|------|------------|------------|
| "Go channel" | 可能返回 MySQL 也提到 channel 的内容 | 只返回 Go 文档 ✅ |
| "退款流程" | 可能返回无关文档 | 只返回售后/退款文档 ✅ |
| "VIP 用户折扣" | 可能返回所有用户文档 | 只返回 VIP 相关文档 ✅ |

**生产 RAG 几乎 100% 需要 Metadata Filter**，否则召回会有噪声。

### Metadata vs Vector Search 的本质区别

| 维度 | Vector Search | Metadata Filter |
|------|--------------|-----------------|
| **匹配方式** | 语义相似度（Embedding 距离）| 精确匹配（结构化字段）|
| **检索范围** | 整个知识库 | 子集（按 Metadata 过滤后）|
| **典型用途** | 找相关 Chunk | 缩小候选范围 |
| **可解释性** | 弱（黑盒距离）| 强（明确字段匹配）|

**两者结合** = 先 Vector 粗召回 → 再 Metadata 过滤 → 再 Reranker 精排。

---

## 4️⃣ Reranker（精排）

### 为什么需要 Reranker？

Vector Search 的局限：
- Bi-Encoder（独立编码 query 和 document），速度快但精度有限
- Top-5 里可能还有"看似相关但不准确"的结果

Reranker 的解决：
- Cross-Encoder（同时编码 query + document），精度高但慢
- 对 Top-5 再精排，挑出最相关的 Top-3

### 实现（Demo 29，简单版）

```python
def rerank(query, documents):
    query_keywords = ["channel", "goroutine", "通信", "并发", "缓冲", "发送", "接收"]
    
    scored = []
    for doc in documents:
        content = doc.page_content.lower()
        score = sum(1 for kw in query_keywords 
                    if kw in query.lower() and kw in content)
        scored.append((score, doc))
    
    scored.sort(key=lambda x: x[0], reverse=True)
    return [doc for _, doc in scored]

# 先 Vector Search 召回 5 个
candidates = vectorstore.similarity_search(query, k=5)

# 再 Reranker 精排
final = rerank(query, candidates)
```

### Bi-Encoder vs Cross-Encoder

| 维度 | Bi-Encoder (Vector Search) | Cross-Encoder (Reranker) |
|------|---------------------------|-------------------------|
| **编码方式** | Query / Doc 独立编码 | Query + Doc 一起编码 |
| **速度** | 快（可离线预先算 Doc 向量）| 慢（每次都要重新算）|
| **精度** | 中 | 高 |
| **适用规模** | 百万级 | 几十~几百个候选 |
| **典型用途** | 粗召回 | 精排 |

**面试回答模板**：
> Vector Search 用 Bi-Encoder 快速粗召回，Reranker 用 Cross-Encoder 对 Top-K 精排。
> Bi-Encoder 快但精度有限，Cross-Encoder 慢但精度高。
> 所以 Vector Search 负责覆盖度，Reranker 负责准确度。

### 生产 Reranker 选项

| 选项 | 特点 |
|------|------|
| Demo 关键词 Reranker | 教学用，不真实 |
| **BGE Reranker** | 中文首选，本地可跑 |
| Cohere Rerank | 商业 API，效果好 |
| 自训 Cross-Encoder | 大厂做法，需要标注数据 |

---

## 5️⃣ VectorStore 持久化（Offline / Online 分离）⭐⭐⭐

### 核心架构思想（来自 ChatGPT）

```
【启动阶段】一次性
─────────────────────────
1. 读取 Markdown
2. Chunking
3. Embedding
4. 建 FAISS 索引
5. 保存到磁盘

【在线阶段】每次请求
─────────────────────────
1. 加载 FAISS 索引（已有）
2. 对用户 Query 做 Embedding
3. Vector Search
4. （可选）Metadata Filter / Reranker
5. Top-K → LLM
```

### 为什么要分离？

| 不分离 | 分离 |
|--------|------|
| 每次启动 Agent 都重新建索引 | 启动只加载，不重建 |
| 索引时间 = 几分钟 | 加载时间 = 几秒 |
| 索引和查询逻辑耦合 | 索引和查询分离（offline / online）|
| 不利于大规模知识库 | 可以用专门工具管理索引 |

### 实现（Demo 26）

```python
from langchain_community.vectorstores import FAISS

# Offline Indexing
vectorstore = FAISS.from_documents(chunks, embedding)
vectorstore.save_local("./faiss_index")

# Online Retrieval
vectorstore = FAISS.load_local(
    "./faiss_index",
    embedding,
    allow_dangerous_deserialization=True,  # ← 必须显式确认
)
```

### ⚠️ 关键安全提示

LangChain 加载 FAISS 用了 `pickle`，**反序列化不可信文件有安全风险**。LangChain 强制要求：

```python
allow_dangerous_deserialization=True  # 必须显式确认
```

**生产原则**：
- ✅ 自己建的索引可以信任
- ❌ 别人给的索引文件不要随便加载

---

## 6️⃣ 完整生产 RAG Pipeline

```
┌─────────────────────────────────────────────────────────┐
│              Offline Indexing（每日/每周）                 │
│                                                           │
│   文档 → Chunking → Embedding → VectorStore → 保存磁盘  │
│                                                           │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│              Online Retrieval（每次请求）                  │
│                                                           │
│   Query → Query Rewrite                                   │
│         → Vector Search (粗召回 k=10)                    │
│         → Metadata Filter (缩范围)                       │
│         → Reranker (精排到 Top-3)                        │
│         → LLM 生成答案                                    │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### 关键指标

| 指标 | 含义 | 优化手段 |
|------|------|---------|
| **Recall@K** | Top-K 里有几个相关 | 增加 K / Hybrid Search / Query Rewrite |
| **Precision@K** | Top-K 里几个是真相关 | Reranker / Metadata Filter |
| **MRR** | 第一个相关结果排多前 | 优化排序（Reranker）|
| **NDCG** | 综合排序质量 | Reranker |

---

## 7️⃣ 完整 Agentic RAG 代码骨架（综合 Demo 23-30）

```python
# 1. Model + Tools
model = ChatOpenAI(...)

@tool
def search_knowledge(query: str, category: str = "") -> str:
    """搜索知识库（带 Metadata Filter）"""
    filter_dict = {"category": category} if category else None
    candidates = vectorstore.similarity_search(query, k=5, filter=filter_dict)
    final = rerank(query, candidates)  # 精排
    return "\n\n---\n\n".join(doc.page_content for doc in final[:3])

# 2. Agent
agent = create_agent(
    model=model,
    tools=[search_knowledge],
    system_prompt="""
你是一个技术知识助手。

优先调用 search_knowledge 工具查询知识库。
如果知识库没有相关资料，明确告诉用户。
不要编造知识库中不存在的内容。
""",
)

# 3. Run
result = agent.invoke({"messages": [{"role": "user", "content": "Go channel 有什么作用？"}]})
```

**Agent 自动决定**：
- 是否调用 search_knowledge
- 调用时传什么 category 参数
- 检索到 Top-3 后怎么组织答案

---

## 8️⃣ 求职定位建议 ⭐⭐⭐（来自 ChatGPT）

### 你现在的位置

**已有**（Go 后端 + Agent）：
- ✅ LLM API / Tool Calling / Function Calling
- ✅ Agent Loop / Memory / RAG / MCP
- ✅ LangChain / LangGraph（基础 + 生产扩展）
- ✅ Router / Workflow / Hybrid Search
- ✅ Go 后端开发经验

**企业级 Agent 完整地图**（还缺的部分）：

```
基础 Agent         RAG            Tool
LLM / Memory     检索           外部能力
   ✅             ⚠️             ✅
                  ↑
           刚补完 Agentic RAG + Metadata + Reranker

状态机            长期记忆        Multi-Agent
LangGraph        Memory Store   Supervisor / Swarm
   ✅              ⚠️             ❌

可观测性          Evaluation      Security
LangSmith        评估方法       Prompt Injection
   ❌              ❌             ❌
```

### 建议定位

> **企业级 Agent Backend 工程师**
> 
> 多年企业级后端开发经验（Go），掌握 Agent、RAG、LangGraph、MCP 以及企业级 Agent 服务开发能力。

**理由**（ChatGPT 原话）：
- 不要把自己培养成纯 Python AI 算法工程师
- 把 Go 后端经验 + Agent 能力**组合**，是差异化竞争力
- 企业 Agent 后端岗位 ≈ 后端能力 + Agent 概念 + LangGraph 实战

### 接下来优先级

**短期（2-4 周）补缺口**：
1. Multi-Agent（Supervisor / Swarm）⭐
2. 可观测性（LangSmith / LangFuse）⭐
3. Evaluation（Trajectory Eval / Outcome Eval）⭐

**中期**：
4. Security（Prompt Injection / 数据隔离）
5. 真实生产组件（Milvus / pgvector / BGE Reranker）

**不急**：
- 不需要成为 RAG 算法专家
- 不需要追 LangChain 源码细节
- 不需要把所有 Vector DB 都学一遍

---

## 9️⃣ 关键洞察

### 1. Agentic RAG = Agent 决策 + RAG 工具

之前学的 RAG 是「文档 → 检索 → LLM」的被动链路。
**Agentic RAG** 是「Agent 决定要不要检索、用什么参数、检索完怎么用」。

### 2. Metadata Filter 是企业 RAG 的入场券

没有 Metadata 的 RAG 是 Demo，有 Metadata 的 RAG 才是产品。
**Demo → 产品的分水岭 = Metadata Filter**。

### 3. Reranker 的本质是「精度换速度」

粗召回 + 精排的两阶段策略，是工程上**用空间换时间**的经典例子。

### 4. Offline / Online 分离是工程化基础

建索引慢，查询快 → 必须分离。
否则每次启动 Agent 都重建索引 → 几小时无法响应。

### 5. 你的定位决定了你该学什么

> **不要成为 RAG 算法专家**。
> **要成为「能用 Go 搭生产 Agent 后端」的工程师**。
> 
> 这决定了：
> - RAG 学到 Metadata + Reranker 够用，不必深入算法
> - Multi-Agent + 可观测性 + Evaluation 更值得投入
> - Go 后端经验是你的护城河

---

## 🔟 自检

- [ ] 1. Bi-Encoder 和 Cross-Encoder 的本质区别是什么？
- [ ] 2. 为什么需要 Metadata Filter？它跟 Vector Search 是替代还是互补？
- [ ] 3. Reranker 在 RAG Pipeline 中处于什么位置？为什么不直接用 Vector Search 的 Top-3？
- [ ] 4. 为什么要把 RAG 索引和查询分离？有什么好处？
- [ ] 5. 真实 Embedding 模型（BGE）和 Fake Embedder 的关键区别是什么？
- [ ] 6. 作为 Go 后端 + Agent 学习者，求职时该怎么定位？

---

## 📂 相关代码

- `langchain-agent-demo/17_agentic_rag/` - 第一个 Agentic RAG
- `langchain-agent-demo/19_agentic_rag_embedding/` - 真实 Embedding
- `langchain-agent-demo/20_agentic_rag_vectorstore/` - FAISS
- `langchain-agent-demo/23_agentic_rag_full/` - 完整 Pipeline
- `langchain-agent-demo/24_rag_indexing/` - Offline 索引
- `langchain-agent-demo/26_agentic_rag_persistent/` - FAISS 持久化
- `langchain-agent-demo/27_rag_metadata/` - Metadata Filter
- `langchain-agent-demo/29_rag_reranker/` - Reranker（精排）

**总计 14 个新增 demo · ~2500 行 Python**

---

## 🔗 相关笔记

- [RAG 完整链路](./rag.md) - 之前学的 RAG 基础（需要更新）
- [LangChain + LangGraph](./langchain-langgraph.md) - 框架视角
- [LangGraph 生产扩展](./langgraph-production-extensions.md) - HITL / Checkpointer
- [Memory 完整体系](./memory.md) - Memory 跟 RAG 的区别

---

## 📅 下一步主线（按 ChatGPT 建议）

短期优先（2-4 周）：
1. **Multi-Agent**（Supervisor / Swarm 模式）⭐⭐⭐
2. **可观测性**（LangSmith / LangFuse）⭐⭐
3. **Evaluation**（Trajectory / Outcome）⭐⭐

中期：
4. **Security**（Prompt Injection / 数据隔离）
5. 真实生产组件（Milvus / pgvector / BGE Reranker 替换教学版）

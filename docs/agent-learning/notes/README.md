# 学习笔记目录

> 按主题组织的深度笔记，配合代码一起看。
> 📅 最后更新：今天（新增 Step 9：MCP + Router）

## 📚 笔记清单

### 🏗️ 架构总览（建议先看）

| # | 笔记 | 说明 |
|---|------|------|
| 0 | [agent-architecture.md](./agent-architecture.md) | **Router + 3 大支柱 + Harness 视角**（建议从这开始） |

### ✅ Step 5-7 笔记

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 1 | [refactoring.md](./refactoring.md) | Step 5-7 重构 | 整个项目 | ✅ |
| 2 | [context-management.md](./context-management.md) | Step 6 · Context Management | `memory/context.go` | ✅ |
| 3 | [agent-orchestration.md](./agent-orchestration.md) | Step 7 · Agent 编排模式 | `workflow/order.go` | ✅ |

### ✅ Step 8 笔记

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 4 | [memory.md](./memory.md) | Step 8 · Memory 完整体系 | `memory/` (3 文件) | ✅ |
| 5 | [query-rewrite.md](./query-rewrite.md) | Step 8 · Query Rewrite | `agent/query_rewrite.go` | ✅ |
| 6 | [embedding-vector-search.md](./embedding-vector-search.md) | Step 8 · Embedding & 向量检索 | `embedding/` | ✅ |
| 7 | [rag.md](./rag.md) | Step 8 · RAG 完整链路 | `rag/` (3 文件) | ✅ |

### ✅ Step 9 笔记

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 8 | [router.md](./router.md) | Step 9 · Capability Router | `agent/router.go` | ✅ |
| 9 | [mcp.md](./mcp.md) | Step 9 · MCP 详解 | `mcp/` + `mcp-sdk-demo/` | ✅ |

### ✅ Step 10 笔记（LangChain/LangGraph）

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 10 | [langchain-langgraph.md](./langchain-langgraph.md) | Step 10 · LangChain + LangGraph（15 个 demo）| `langchain-agent-demo/` | ✅ |
| 11 | [langgraph-production-extensions.md](./langgraph-production-extensions.md) | Step 10 扩展 · 生产落地（HITL/Checkpointer/Memory vs Checkpoint/FastAPI）| — | ✅ |

### ✅ Step 11 笔记（Agentic RAG + 求职定位）

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 12 | [agentic-rag-full-pipeline.md](./agentic-rag-full-pipeline.md) | Step 11 · Agentic RAG 完整链路 + 求职定位 | `langchain-agent-demo/17-30` | ✅ |

### ✅ Step 12 笔记（Tool 安全 + 可靠性）

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 13 | [agent-tool-security-and-reliability.md](./agent-tool-security-and-reliability.md) | Step 12 · Tool 权限 / HITL / Idempotency / Multi-Agent | `langchain-agent-demo/31-38` | ✅ |

### ✅ Step 13 笔记（Multi-Agent 架构实战）

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 14 | [multi-agent-architecture.md](./multi-agent-architecture.md) | Step 13 · Multi-Agent 完整架构（Sub-Agent as Tool / 分布式 HTTP / Trace ID）| `langchain-agent-demo/39_multi_agent` | ✅ |

## 🗺️ 推荐阅读顺序

### 第一次系统学（按顺序）

1. **[agent-architecture.md](./agent-architecture.md)** — 先看整体架构图
2. [refactoring.md](./refactoring.md) — 理解模块切分
3. [context-management.md](./context-management.md) — Step 6
4. [agent-orchestration.md](./agent-orchestration.md) — Step 7
5. [memory.md](./memory.md) — Step 8（核心）
6. [embedding-vector-search.md](./embedding-vector-search.md) — 数学基础
7. [rag.md](./rag.md) — RAG 完整链路
8. [query-rewrite.md](./query-rewrite.md) — Memory 的前置优化

### 面试前复习（按主题跳）

- 面试被问 **Context Management** → 看 [context-management.md](./context-management.md)
- 面试被问 **Workflow vs Agent** → 看 [agent-orchestration.md](./agent-orchestration.md)
- 面试被问 **Memory 实现** → 看 [memory.md](./memory.md)
- 面试被问 **RAG** → 看 [rag.md](./rag.md)
- 面试被问 **Agent 架构** → 看 [agent-architecture.md](./agent-architecture.md)

## 📝 每篇笔记的结构

每篇笔记都包含：
- 🎯 核心问题（一句话讲清学这个干嘛）
- 🧠 关键概念（区别 / 对比 / 类比）
- 💻 代码解读（对应到实际文件 + 行数）
- ⚠️ 已知问题 / 局限
- 🔮 下一步 / 进阶方向
- ✅ 自检问题

## 📊 笔记覆盖度

| 学习路线 | 笔记 | 状态 |
|---------|------|------|
| Step 1-5 | 文档未单独整理，参考 [code/06-agent/go-agent-demo/README.md](../../../code/06-agent/go-agent-demo/README.md) | ✅ |
| Step 6 Context Management | [context-management.md](./context-management.md) | ✅ |
| Step 7 Agent 编排 | [agent-orchestration.md](./agent-orchestration.md) | ✅ |
| Step 8 Memory | [memory.md](./memory.md) + [query-rewrite.md](./query-rewrite.md) + [embedding-vector-search.md](./embedding-vector-search.md) | ✅ |
| Step 8 RAG 扩展 | [rag.md](./rag.md) | ✅ |
| Step 9 Router | [router.md](./router.md) | ✅ |
| Step 9 MCP | [mcp.md](./mcp.md) | ✅ |
| Step 10 LangChain/LangGraph | [langchain-langgraph.md](./langchain-langgraph.md) + [langgraph-production-extensions.md](./langgraph-production-extensions.md) | ✅ |
| Step 11 Agentic RAG | [agentic-rag-full-pipeline.md](./agentic-rag-full-pipeline.md) | ✅ |
| Step 12 Tool 安全 + 可靠性 | [agent-tool-security-and-reliability.md](./agent-tool-security-and-reliability.md) | ✅ |
| Step 13 Multi-Agent 架构 | [multi-agent-architecture.md](./multi-agent-architecture.md) | ✅ |
| Step 8 架构视角 | [agent-architecture.md](./agent-architecture.md) | ✅ |
| Step 9 MCP | 待学 | ⬜ |
| Step 10 Multi-Agent | 待学 | ⬜ |
| Step 11 Evaluation | 待学 | ⬜ |
| Step 12 工程化 | Harness 视角已提及，详细待学 | ⬜ |

## 🔗 相关资源

- [学习路线总览](../README.md)
- [Demo 代码 + README](../../../code/06-agent/go-agent-demo/README.md)

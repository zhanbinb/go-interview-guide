# Agent 开发极速上手学习计划

> 🎯 目标：系统性地概览 Agent 开发，建立全景认知
> 📚 假设：已有 Go 后端基础，正在准备面试
> ⏱️ 时间预算：12 步，约 13.5 小时（不含已完成部分）
> 📅 路线版本：v8（新增 Tool 安全 + 可靠性 + Multi-Agent 思路；12 步全部完成）

---

## 0. 5 分钟心智模型

**Agent ≠ ChatGPT**，核心区别：

| 维度 | 普通 LLM | Agent |
|------|---------|-------|
| 交互 | 一问一答 | 自主多步 |
| 工具 | 不能调用 | 能调函数 / API |
| 记忆 | 单次对话 | 短期 + 长期 |
| 规划 | 无 | ReAct / Plan-and-Execute |

> **Agent 本质 = LLM + 工具 + 循环（思考 → 行动 → 观察 → 再思考）**

---

## 1. 技术栈选择（择一即可）

| 方案 | 适合 | 推荐度 |
|------|------|--------|
| **Go + openai-go** | 你正在用 ✅ | ⭐⭐⭐⭐⭐ |
| Python + LangGraph | 严肃 Agent 框架 | ⭐⭐⭐⭐⭐ |
| TypeScript + Vercel AI SDK | 前端 / 全栈背景 | ⭐⭐⭐⭐ |
| Python + 裸 OpenAI API | 想理解原理 | ⭐⭐⭐⭐ |

> 💡 **当前选择**：Go + `openai-go`（Day 1 Demo 已用）。OpenAI 兼容协议一家通用（DeepSeek / MiniMax / OpenAI 都支持）。

---

## 2. 12 步学习路线（主线）

按 **「由内向外、由浅入深」** 组织：

### ✅ 已完成（Step 1-5）

| Step | 主题 | 状态 | 对应 Demo |
|------|------|------|----------|
| 1 | LLM API | ✅ | [code/06-agent/go-agent-demo/main.go](../../code/06-agent/go-agent-demo/main.go) |
| 2 | Tool Calling | ✅ | 同上 |
| 3 | Tool Result | ✅ | 同上 |
| 4 | Tool Registry | ✅ | 同上 |
| 5 | Agent Loop（ReAct） | ✅ | 同上 |

> 📖 Demo 文档：[code/06-agent/go-agent-demo/README.md](../../code/06-agent/go-agent-demo/README.md)

### 🎯 待学习（Step 6-12）

---

#### **Step 6 · Context Management**（2h）⭐ 下一步

**核心问题**：Agent 跑久了消息历史会爆，怎么管？

**学什么**：
- Token 计数 + 限制（不同模型上限：gpt-4o=128k, claude-sonnet-4-5=200k）
- 消息截断策略：保留最近 N 条 / 摘要压缩 / 滑动窗口
- Go 特有的 `context.Context`（超时、取消、跨 goroutine 传值）
- System Prompt 的组织方式

**怎么学**（最小动手）：
1. 给现有 Demo 加 **max iterations**（防止死循环）
2. 加 **context.WithTimeout**（API 调用超时）
3. 实现一个**简单消息截断器**（超过 N 条就压缩最早的）

---

#### **Step 7 · Agent 编排模式**（1.5h）

**核心问题**：除了 ReAct，主流还有哪些模式？

**必读**：[Anthropic《Building Effective Agents》](https://www.anthropic.com/research/building-effective-agents)（30 分钟讲清所有模式）

**5 个核心模式**：

| 模式 | 一句话 | 适用场景 |
|------|------|---------|
| Prompt Chaining | 流水线，固定步骤 | 翻译→审校→格式化 |
| Routing | 按类型路由到不同子流程 | 客服分类后派发 |
| Parallelization | 并行执行子任务 | 多角度评估同一个东西 |
| Orchestrator-Workers | 一个总指挥调度多个 worker | 复杂任务拆解 |
| Autonomous Agent | 完全自主决策 | 你已经做的 ReAct ✅ |

**怎么学**：通读 + 用 LangGraph 跑一个 Chaining 示例

---

#### **Step 8 · Memory**（2h）✅ 已完成

**核心问题**：怎么让 Agent **跨会话**"记住"东西？

**学了什么**：
- 长期记忆：key-value + 向量双存储
- 事实抽取：让 LLM 自动从对话中提取关键信息
- Hybrid Search：Keyword (10/5/2 打分) + Vector (Cosine Similarity)
- Query Rewrite：用 LLM 把自然语言转换为检索关键词
- RAG 完整链路：KnowledgeBase + Retriever + Ask

**笔记**：[notes/memory.md](notes/memory.md) + [notes/query-rewrite.md](notes/query-rewrite.md) + [notes/embedding-vector-search.md](notes/embedding-vector-search.md) + [notes/rag.md](notes/rag.md)

---

#### **Step 9 · MCP**（1h，概念为主）

**核心问题**：Tool 调用的**标准接口**是什么？

**学什么**：
- [Model Context Protocol 官方文档](https://modelcontextprotocol.io/)
- 三个核心概念：MCP Server / MCP Client / Tool Resource
- 看一个现成 MCP Server 实现（如 `@modelcontextprotocol/server-filesystem`）
- **不写代码**，理解 MCP 跟你现在的 `toolRegistry` 是同一层抽象的标准化

---

#### **Step 10 · Multi-Agent**（2h）

**核心问题**：多个 Agent 怎么协作？

**学什么**：
- 主流框架：[AutoGen](https://github.com/microsoft/autogen) / [CrewAI](https://github.com/crewAIInc/crewAI)
- 两种架构：
  - **Supervisor 模式**：一个总 Agent 调度其他 Agent
  - **Swarm 模式**：Agent 之间对等通信
- 选一个框架跑一个 2-Agent 协作 Demo（建议 CrewAI，上手快）

---

#### **Step 11 · Evaluation**（2h）⭐ 评估方法论

**核心问题**：怎么知道 Agent **变好了还是变烂了**？

**学什么**：
- **Trajectory Eval**：工具调用对不对、步骤对不对
- **Outcome Eval**：最终结果对不对
- 数据集怎么造（人工 / LLM 生成）
- 工具：[LangSmith](https://smith.langchain.com/) / [LangFuse](https://langfuse.com/) / DeepEval
- 跑一个最小评估（10 个 case 起步）

---

#### **Step 12 · Agent 工程化**（2h）

**核心问题**：怎么把 Demo 变成能上线的东西？

**学什么**：
- **部署**：FastAPI 包装 / Serverless
- **监控**：结构化日志、Token 消耗、错误率、延迟
- **成本控制**：模型选择（小模型预处理 + 大模型兜底）、缓存、限流
- **安全**：危险工具的二次确认、Prompt 注入防护、敏感信息脱敏

---

## 3. 必看资料清单

### 📖 官方文档（首选）

| 资料 | 链接 | 评级 |
|------|------|------|
| **Anthropic《Building Effective Agents》** | https://www.anthropic.com/research/building-effective-agents | ⭐⭐⭐⭐⭐ |
| Anthropic Prompt Engineering | https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview | ⭐⭐⭐⭐⭐ |
| LangGraph 官方文档 | https://langchain-ai.github.io/langgraph/ | ⭐⭐⭐⭐⭐ |
| OpenAI Function Calling | https://platform.openai.com/docs/guides/function-calling | ⭐⭐⭐⭐ |
| OpenAI Cookbook | https://cookbook.openai.com/ | ⭐⭐⭐⭐ |
| MCP 官方文档 | https://modelcontextprotocol.io/ | ⭐⭐⭐⭐ |

### 🎥 视频

- **Andrej Karpathy**：[Intro to Large Language Models](https://www.youtube.com/watch?v=zjkBMFhNj_g)（1h，建立直觉）
- **Harrison Chase**（LangChain CEO）YouTube 频道
- B 站搜「LLM Agent 入门」看几个高播放量搬运

### 📚 中文资料

- 宝玉翻译的 Anthropic 博文（博客 / 公众号「宝玉的分享」）
- 知乎搜「LLM Agent 入门」看 3~5 篇高赞
- LangChain 中文社区：https://langchain.com.cn

### 🛠️ 工具

- **大模型 API**：OpenAI / Anthropic / DeepSeek（国产便宜）/ 通义千问 / MiniMax
- **Agent 框架**：LangGraph（推荐）/ AutoGen / CrewAI
- **调试 / 观测**：LangSmith / LangFuse
- **协议**：MCP（Model Context Protocol）

---

## 4. 推荐 Demo 项目（任选一）

### 🥇 Demo A：个人助理 Agent（✅ 已完成 Go 版）

**功能**：能查天气 + 能算数学 + 能查日历  
**当前实现**：`query_order` + `get_user` 两个 mock 工具

### 🥈 Demo B：研究助手 Agent

**功能**：输入主题 → 自动搜索 → 总结 → 写报告  
**涉及**：搜索 API + Plan-and-Execute 模式  
**亮点**：理解多步推理与子任务拆分

### 🥉 Demo C：自动化运维 Agent（贴合 web3 背景）

**功能**：自然语言操作数据库 / 调用合约（dry-run）  
**涉及**：ReAct + 安全校验层  
**亮点**：跟你的实际工作直接相关

---

## 5. 推荐阅读的开源项目

| 项目 | 说明 | 看什么 |
|------|------|--------|
| [langgraph](https://github.com/langchain-ai/langgraph) | LangGraph 官方 | 看 `examples/` 文件夹（Step 7/8/10 都会用到） |
| [AutoGen](https://github.com/microsoft/autogen) | 微软多 Agent 框架 | 看 `test/` 里的用例（Step 10） |
| [crewAI](https://github.com/crewAIInc/crewAI) | 多 Agent 协作 | 看 `examples/`（Step 10） |
| [gpt-engineer](https://github.com/gpt-engineer-org/gpt-engineer) | 写代码的 Agent | 体会 Agent 完成复杂任务 |
| [dify](https://github.com/langgenius/dify) | 可视化 Agent 平台 | 跑起来玩，理解产品层 |
| [modelcontextprotocol](https://github.com/modelcontextprotocol) | MCP 官方 | 看 server 实现（Step 9） |

---

## 6. 环境准备

### Python 环境（用于跑 LangGraph / CrewAI 示例）

```bash
python -m venv .venv-agent
source .venv-agent/bin/activate
pip install langgraph langchain-openai crewai tavily-python python-dotenv
```

### API Key（至少准备一个）

```bash
# OpenAI（最通用）
export OPENAI_API_KEY=sk-...

# 或 DeepSeek（国产，便宜，OpenAI 兼容协议）
export OPENAI_API_KEY=sk-...
export OPENAI_BASE_URL=https://api.deepseek.com

# 或 Anthropic Claude
export ANTHROPIC_API_KEY=sk-ant-...
```

> 💡 **省钱建议**：调试阶段用 `gpt-4o-mini` / `deepseek-chat`，跑通后再换 `gpt-4o` / `claude-sonnet-4-5`。

---

## 7. 学完后自检清单

完成后能回答下面问题，说明对 Agent 领域有系统认知：

**Step 1-5（基础）**：
- [ ] 1. Function Calling 和普通 Prompt 调用的本质区别是什么？
- [ ] 2. ReAct 模式的「Thought / Action / Observation」是哪三步？
- [ ] 3. Agent 为什么会陷入死循环？怎么防止？
- [ ] 4. 为什么 `tool_call.id` 必须回传？

**Step 6-9（中级）**：
- [ ] 5. Token 成本和延迟在 Agent 里为什么会被放大？
- [ ] 6. context.Context 和 Agent 的「上下文」是什么关系？
- [ ] 7. Anthropic 提到的 5 种编排模式各自适用什么场景？
- [ ] 8. 短期记忆（thread）和长期记忆（store）的实现方式有什么不同？
- [ ] 9. MCP 解决的核心问题是什么？跟你写的 toolRegistry 是同一层抽象吗？

**Step 10-12（高级）**：
- [ ] 10. 单 Agent 和多 Agent 协作（supervisor / swarm）各适合什么场景？
- [ ] 11. Trajectory Eval 和 Outcome Eval 的区别？
- [ ] 12. Agent 上线前最关键的 3 个工程化考虑是什么？

---

## 8. 学习原则

1. **不要先看理论再看代码** —— 直接复制官方 Quick Start 跑起来，回头再补理论
2. **每步先跑通最小版本**，再加功能
3. **Debug 时打开 LangSmith / 日志**，看 Agent 每一步在干啥
4. **不建议先碰 AutoGen / CrewAI**，等 LangGraph 熟了再看（Step 10 才用）
5. **不要追新框架**，LangGraph + 主流大模型 API 已能覆盖 90% 场景
6. **每个 Step 跑通后再进下一步**，不要跳过

---

## 9. 进度追踪

| Step | 主题 | 状态 | 备注 |
|------|------|------|------|
| 1 | LLM API | ✅ | Day 1 Demo |
| 2 | Tool Calling | ✅ | Day 1 Demo |
| 3 | Tool Result | ✅ | Day 1 Demo |
| 4 | Tool Registry | ✅ | Day 1 Demo |
| 5 | Agent Loop（ReAct） | ✅ | Day 1 Demo |
| **6** | **Context Management** | ✅ | [笔记](notes/context-management.md) |
| 7 | Agent 编排模式 | ✅ | [笔记](notes/agent-orchestration.md) |
| **8** | **Memory** | ✅ | [笔记](notes/memory.md) |
| 9 | MCP + Router | ✅ | [笔记](notes/mcp.md) + [笔记](notes/router.md) |
| **10** | **LangChain / LangGraph** | ✅ | [笔记](notes/langchain-langgraph.md) |
| **11** | **Agentic RAG** | ✅ | [笔记](notes/agentic-rag-full-pipeline.md) |
| **12** | **Tool 安全 + 可靠性** | ✅ | [笔记](notes/agent-tool-security-and-reliability.md) |

---

## 10. 学习笔记

按主题整理的深度笔记，配合代码一起看：

### 架构总览
- [notes/agent-architecture.md](notes/agent-architecture.md) — **Router + 3 大支柱 + Harness 视角**（建议先看）

### Step 6-7 笔记
- [notes/refactoring.md](notes/refactoring.md) — 项目重构记录
- [notes/context-management.md](notes/context-management.md) — Step 6 Context Management
- [notes/agent-orchestration.md](notes/agent-orchestration.md) — Step 7 Agent 编排

### Step 8 笔记
- [notes/memory.md](notes/memory.md) — Memory 完整体系
- [notes/query-rewrite.md](notes/query-rewrite.md) — Query Rewrite
- [notes/embedding-vector-search.md](notes/embedding-vector-search.md) — Embedding & 向量检索
- [notes/rag.md](notes/rag.md) — RAG 完整链路

### Step 9 笔记（今天新完成）
- [notes/router.md](notes/router.md) — **Router 能力路由**（新增架构思想）
- [notes/mcp.md](notes/mcp.md) — **MCP 详解**（含独立 SDK Demo + 主项目集成）

**完整索引**：[notes/README.md](notes/README.md)

## 11. 下一步

🎉 **12 步学习路线全部完成！**

剩余 3 个补充方向（按需选择）：
1. **Multi-Agent 实战**（Supervisor / Swarm + Subgraph）⭐⭐⭐
2. **可观测性**（LangSmith / LangFuse 接入）⭐⭐⭐
3. **Evaluation**（Trajectory Eval / Outcome Eval）⭐⭐⭐

**建议下一步**：整理成最终学习报告 / 制作项目 Demo / 准备面试。

**路线完整度**：
```
✅ Step 1-5  基础（LLM / Tool / Loop）
✅ Step 6    Context Management
✅ Step 7    Agent 编排
✅ Step 8    Memory + RAG + Embedding + Query Rewrite
✅ Step 9    MCP + Router
✅ Step 10   LangChain + LangGraph
⬜ Step 11   Evaluation
⬜ Step 12   Agent 工程化
```

# Agent 开发极速上手学习计划

> 🎯 目标：周末两天跑通一个能用的 Agent Demo
> 📚 假设：已有 Go 后端基础，正在准备面试
> ⏱️ 时间预算：8 ~ 10 小时

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

| 方案 | 适合 | 上手时间 | 推荐度 |
|------|------|---------|--------|
| **Python + LangGraph** | 想做严肃 Agent 框架 | 2h | ⭐⭐⭐⭐⭐ |
| TypeScript + Vercel AI SDK | 前端 / 全栈背景 | 1h | ⭐⭐⭐⭐ |
| Python + 裸 OpenAI API | 想理解原理 | 1h | ⭐⭐⭐⭐ |
| Go + Eino（字节） | 想用 Go | 3h | ⭐⭐⭐ |

### 我的建议：选 Python + LangGraph

理由：
- Agent 领域 Python 是事实标准，生态最全
- 不熟 Python 也无所谓，Agent 这块 API 很薄
- LangGraph 把 Agent 抽象成「图」，概念清晰，官方文档质量高

---

## 2. 两天学习路线

### 📅 Day 1（4h）：跑通最小 Demo

**Step 1 · 30 min · Function Calling 原理**
- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
- [Anthropic Tool Use](https://docs.anthropic.com/en/docs/tool-use)
- 目标：理解「怎么让 LLM 返回结构化函数调用」

**Step 2 · 1h · 跑通 LangGraph Hello World**
- [LangGraph Quick Start](https://langchain-ai.github.io/langgraph/)
- 直接复制官方 `create_react_agent` 的 5 行代码跑通

**Step 3 · 2.5h · 做第一个 Agent Demo**
- 选下面 §4 的 Demo A/B/C 之一动手

### 📅 Day 2（4h）：进阶 + 读源码

**Step 4 · 1h · 理解 Agent 模式**
- [LangGraph 概念文档](https://langchain-ai.github.io/langgraph/concepts/)
- 重点：ReAct / Plan-and-Execute / Reflexion 区别

**Step 5 · 1h · Memory / 多 Agent / Human-in-the-loop**
- [LangGraph How-to 集合](https://langchain-ai.github.io/langgraph/how-tos/)
- 挑 3 个感兴趣的 how-to 跑一遍

**Step 6 · 2h · 读一个真实开源项目**
- 见 §5 推荐清单

---

## 3. 必看资料清单

### 📖 官方文档（首选）

| 资料 | 链接 | 评级 |
|------|------|------|
| LangGraph 官方文档 | https://langchain-ai.github.io/langgraph/ | ⭐⭐⭐⭐⭐ |
| Anthropic Prompt Engineering | https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview | ⭐⭐⭐⭐⭐ |
| OpenAI Cookbook | https://cookbook.openai.com/ | ⭐⭐⭐⭐ |
| LangChain 文档 | https://python.langchain.com/ | ⭐⭐⭐⭐ |

### 🎥 视频

- **Andrej Karpathy**：[Intro to Large Language Models](https://www.youtube.com/watch?v=zjkBMFhNj_g)（1h，建立直觉）
- **Harrison Chase**（LangChain CEO）YouTube 频道
- B 站搜「LLM Agent 入门」看几个高播放量搬运

### 📚 中文资料

- 宝玉翻译的 Anthropic 博文（博客 / 公众号「宝玉的分享」）
- 知乎搜「LLM Agent 入门」看 3 ~ 5 篇高赞
- LangChain 中文社区：https://langchain.com.cn

### 🛠️ 工具

- **大模型 API**：OpenAI / Anthropic / DeepSeek（国产便宜）/ 通义千问
- **Agent 框架**：LangGraph（推荐）/ AutoGen / CrewAI
- **调试工具**：[LangSmith](https://smith.langchain.com/)（看 Agent 每一步决策过程）

---

## 4. 推荐 Demo 项目（任选一）

### 🥇 Demo A：个人助理 Agent（首选）

**功能**：能查天气 + 能算数学 + 能查日历  
**工具**：3 个简单 function（`get_weather` / `calculate` / `search_notes`）  
**亮点**：真正理解「Agent 自己决定调哪个工具」

### 🥈 Demo B：研究助手 Agent

**功能**：输入主题 → 自动搜索 → 总结 → 写报告  
**涉及**：搜索 API + Plan-and-Execute 模式  
**亮点**：理解多步推理与子任务拆分

### 🥈 Demo C：自动化运维 Agent（贴合 web3 背景）

**功能**：自然语言操作数据库 / 调用合约（dry-run）  
**涉及**：ReAct + 安全校验层  
**亮点**：跟你的实际工作直接相关

---

## 5. 推荐阅读的开源项目

| 项目 | 说明 | 看什么 |
|------|------|--------|
| [langgraph](https://github.com/langchain-ai/langgraph) | LangGraph 官方 | 看 `examples/` 文件夹 |
| [AutoGen](https://github.com/microsoft/autogen) | 微软多 Agent 框架 | 看 `test/` 里的用例 |
| [crewAI](https://github.com/crewAIInc/crewAI) | 多 Agent 协作 | 看 `examples/` |
| [gpt-engineer](https://github.com/gpt-engineer-org/gpt-engineer) | 写代码的 Agent | 体会 Agent 完成复杂任务的感觉 |
| [dify](https://github.com/langgenius/dify) | 可视化 Agent 平台 | 跑起来玩，理解产品层 |

---

## 6. 环境准备（Day 1 之前搞定）

### Python 环境

```bash
python -m venv .venv-agent
source .venv-agent/bin/activate
pip install langgraph langchain-openai tavily-python python-dotenv
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

> 💡 **省钱建议**：调试阶段用 `gpt-4o-mini` 或 `deepseek-chat`，跑通后再换 `gpt-4o` / `claude-sonnet-4-5`。

---

## 7. 学完后自检清单

完成后能回答下面问题，说明超过 80% 的 Agent 初学者：

- [ ] 1. Function Calling 和普通 Prompt 调用的本质区别是什么？
- [ ] 2. ReAct 模式的「Thought / Action / Observation」是哪三步？
- [ ] 3. Agent 为什么会陷入死循环？怎么防止？
- [ ] 4. 短期记忆（thread）和长期记忆（store）的实现方式有什么不同？
- [ ] 5. 单 Agent 和多 Agent 协作（supervisor / swarm）各适合什么场景？
- [ ] 6. Token 成本和延迟在 Agent 里为什么会被放大？
- [ ] 7. 怎么评估一个 Agent 的好坏？需要看哪些指标？

---

## 8. 学习原则（血的教训）

1. **不要先看理论再看代码** —— 直接复制官方 Quick Start 跑起来，回头再补理论
2. **第一天就动手**，哪怕只是改一个参数观察行为变化
3. **Debug 时打开 LangSmith**，看 Agent 每一步在干啥，这是核心技能
4. **不建议先碰 AutoGen / CrewAI**，等 LangGraph 熟了再看，它们是更高层抽象
5. **不要追新框架**，LangGraph + 主流大模型 API 已经能覆盖 90% 场景

---

## 9. 进度追踪

| 阶段 | 状态 | 备注 |
|------|------|------|
| Day 1 Step 1：Function Calling 原理 | ⬜ | |
| Day 1 Step 2：LangGraph Hello World | ⬜ | |
| Day 1 Step 3：第一个 Demo 跑通 | ⬜ | |
| Day 2 Step 4：理解 Agent 模式 | ⬜ | |
| Day 2 Step 5：Memory / 多 Agent / HITL | ⬜ | |
| Day 2 Step 6：读一个开源项目源码 | ⬜ | |
| 自检清单 7 题全部能答 | ⬜ | |

---

## 10. 下一步可以做什么？

学完之后，可以告诉我，我会帮你：

- **A. 把 Demo A 的完整代码搭起来**（含 3 个工具 + LangGraph + 单元测试）
- **B. 把 Demo B 的搜索 + 总结链路写出来**
- **C. 把 Demo C 跟你之前的 scanner 项目结合，做一个「AI + 区块链事件分析」Agent**
- **D. 用你熟悉的 Go + Eino 重写一遍，对比两个生态差异**

挑一个，我就开干。

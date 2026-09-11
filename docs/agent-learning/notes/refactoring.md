# 项目重构：从单文件 main.go 到模块化架构

> 📚 对应学习计划：[Step 5-7 之间的过渡](../README.md)
> 🗓️ 时间：今天
> 🔗 学习来源：ChatGPT 分享 [6aa27a6d](https://chatgpt.com/share/6aa27a6d-0670-83e8-8e74-c0465ea4ecdc) / [6aa27a47](https://chatgpt.com/share/6aa27a47-78f8-83ee-826e-6c8aba41cb8c)

---

## 🎯 为什么重构

重构前的 `main.go` 是 **871 行单文件**，混合了：

| 内容 | 行数范围（旧） |
|------|---------|
| Provider 配置 | 1 ~ 50 |
| Tool 实现 + Schema | 192 ~ 400 |
| Agent Loop | 496 ~ 600 |
| 三个 Workflow Demo | 628 ~ 869 |
| Context / Memory 散落各处 | 各处 |

**问题**：
- 找一个知识点要在 800 行里翻
- 知识点边界模糊（Config、Agent、Tool、Workflow 混在一起）
- 不利于复习时按概念定位

**目标**：每个 Agent 知识点 → 一个独立模块

---

## 📁 重构后的模块结构

```
go-agent-demo/
├── main.go                      # 入口：装配 + 调用（业务代码几乎清空）
│
├── config/config.go             # Provider 配置（DeepSeek / MiniMax / OpenAI）
├── llm/client.go                # OpenAI 客户端创建
│
├── tools/
│   ├── tool.go                  # Tool 接口定义
│   ├── registry.go              # Tool 注册表
│   └── order.go                 # 业务 Tool 实现（用户/订单/支付/物流）
│
├── agent/
│   ├── agent.go                 # Agent Loop（核心循环）
│   └── schema.go                # Tool Schema（告诉 LLM 有哪些工具）
│
├── memory/
│   ├── context.go               # ContextManager（当前对话上下文）
│   └── memory.go                # Memory（跨对话长期记忆）
│
└── workflow/
    └── order.go                 # 三种 Workflow 实现
```

**总行数**：从 871 行单文件 → 1380 行分模块（多出来的主要是注释和接口边界）

---

## 🔨 重构顺序

> 一次拆一个，**确保还能跑再继续下一步**。

| 顺序 | 拆出 | 验证 | 学到的概念 |
|------|------|------|----------|
| 1 | `config` | go build + go run 通过 | Provider 抽象 |
| 2 | `llm` | go build + go run 通过 | 客户端创建边界 |
| 3 | `tools` | Tool 仍可调用 | Tool / Registry 分离 |
| 4 | `agent` | Agent Loop 跑通 | Agent 核心循环独立 |
| 5 | `memory/context` | ContextManager 集成 | 上下文管理独立 |
| 6 | `workflow` | 三个 Workflow 都能跑 | Workflow 与 Agent 分离 |
| 7 | 删除旧 main.go 的注释代码 | 旧代码完成教学使命 | 减法也是设计 |

---

## 🧠 核心设计原则

### 原则 1：每个模块对应一个 Agent 概念

| 模块 | 对应概念 | 在学习路线里 |
|------|---------|------------|
| `config` | Provider 抽象 | Step 1（LLM API）|
| `tools` | Tool Calling + Registry | Step 2-4 |
| `agent` | Agent Loop（ReAct）| Step 5 |
| `memory/context` | Context Management | Step 6 |
| `workflow` | Agent 编排模式 | Step 7 |
| `memory/memory` | 长期记忆 | Step 8 预备 |

**好处**：以后复习某个概念，直接定位到对应文件。

### 原则 2：依赖单向流动

```
main
  ↓
agent ──→ memory ──→ (llm client 注入)
  ↓           ↑
tools ────────┘
  ↓
workflow ──→ tools + config
```

**没有循环依赖**，每个模块只依赖自己真正需要的东西。

### 原则 3：删除旧代码

> 旧代码已经完成教学使命，再留着只会让项目越来越乱。

重构完成后，原 `main.go` 里**大量注释掉的旧代码被删除**。这是有意的减法。

---

## 📊 重构前后对比

| 维度 | 重构前 | 重构后 |
|------|--------|--------|
| 文件数 | 1 | 11 |
| main.go 行数 | 871 | 182（只剩装配 + 调用）|
| 找一个 Tool 实现 | 全文搜索 | 直接看 `tools/order.go` |
| 复习 Agent Loop | 翻 496 行附近 | 直接看 `agent/agent.go` |
| 知识点边界 | 模糊 | 清晰（一个文件 = 一个概念）|
| 编译错误定位 | 难 | 容易（按模块独立 build）|

---

## 💡 重构过程的关键决策

### 决策 1：先拆 `config` 还是 `llm`？

**先拆 `config`**。理由：
- `config` 是纯配置，零依赖
- 拆 `config` 时强制你想清楚「我要连哪个模型、Key 从哪里来」
- 不影响其他代码，改动最小

### 决策 2：`tools` 拆成几个文件？

**3 个文件**：
- `tool.go` - Tool 接口定义
- `registry.go` - Tool 注册表
- `order.go` - 业务 Tool 实现

理由：Tool 接口很薄、Registry 中等、业务实现最厚，三者复杂度不同。

### 决策 3：`agent/schema.go` 单独？

**是的**。理由：
- `schema.go`（Tool Schema）描述「LLM 看到的工具说明书」
- `tools/registry.go` 描述「程序怎么执行这个工具」
- **两者不是一回事**，必须分开

> 这是一个易踩的坑：很多初学者把 Tool Schema 写在 Tool struct 里，导致新增工具要改两个地方。

### 决策 4：`memory` 拆成 `context.go` 和 `memory.go`？

**是的**。这是 **Step 6 的核心论点**：
- `context.go` - 当前对话上下文（ContextManager）
- `memory.go` - 跨对话长期记忆（Memory）

**两个完全不同的概念**，必须分文件。

---

## ⚠️ 重构中遇到的问题

### 问题 1：OpenAI SDK 消息序列化丢失 Tool Call

`ContextManager.Summarize` 用 `json.Marshal` 序列化 messages 时，**Assistant Tool Call 结构丢失**。

**当前 Demo 的妥协**：摘要后清空 messages，下次请求 LLM 看到的是「摘要 + 新对话」。

**生产环境解决**：用 SDK 原生的 message 转换函数，保留 Tool Call 链。

### 问题 2：`workflow.KeysOf` 函数位置

`KeysOf` 是辅助函数，原来在 `main.go`，重构后放在 `workflow` 包。

**反思**：这个函数其实不属于 `workflow`，应该删掉或放到 `config` 包。**Demo 中保留是临时方案**。

---

## 🔗 相关笔记

- [Context Management 详解](./context-management.md) - Step 6
- [Agent 编排模式详解](./agent-orchestration.md) - Step 7
- [Memory 预备](./memory-preview.md) - Step 8 准备

---

## ✅ 自检

- [ ] 1. 为什么 `config` 先拆而不是 `agent`？
- [ ] 2. `tool.Schema` 和 `tool.Registry` 为什么不能合在一起？
- [ ] 3. `memory/context.go` 和 `memory/memory.go` 为什么必须分文件？
- [ ] 4. 重构后 main.go 应该剩下什么？（答：装配 + 调用，不含业务逻辑）

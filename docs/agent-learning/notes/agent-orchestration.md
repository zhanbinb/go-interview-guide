# Step 7 · Agent 编排（Orchestration）详解

> 📚 对应学习计划：[README §Step 7](../README.md)
> 🗓️ 时间：今天
> 💻 关键代码：[`workflow/order.go`](../../../code/06-agent/go-agent-demo/workflow/order.go)（239 行）

---

## 🎯 核心问题

> 哪些事情应该由程序控制？哪些事情应该由 LLM 决策？

这是 Agent 开发工程师每天都要回答的问题。

**编排（Orchestration）= 控制 Agent 的执行流程**

---

## 📊 三种编排模式（由简到繁）

### Level 1 · Fixed Workflow（固定顺序）

**程序决定所有步骤**，LLM 都不参与。

```go
func RunOrderWorkflow() string {
    user := tools.GetUser("1001")
    order := tools.QueryOrder("10001")
    payment := tools.QueryPayment("10001")
    logistics := tools.QueryLogistics("10001")
    return fmt.Sprintf("用户信息：%s\n订单信息：%s\n支付信息：%s\n物流信息：%s",
        user, order, payment, logistics)
}
```

**特点**：

| 维度 | 表现 |
|------|------|
| 决策者 | 程序（硬编码）|
| 灵活性 | 最低 |
| 确定性 | 最高 |
| 适用场景 | 流程固定的业务（用户开户、订单创建等）|

**对应 Anthropic 模式**：**Prompt Chaining**（流水线）

---

### Level 2 · Conditional Workflow（条件编排）

**程序根据上一步结果决定下一步**。

```go
func RunOrderWorkflow2() string {
    user := tools.GetUser("1001")
    order := tools.QueryOrder("10001")

    var nextResult string
    if strings.Contains(order, "已发货") {
        nextResult = tools.QueryLogistics("10001")
    } else {
        nextResult = tools.QueryPayment("10001")
    }
    return ...
}
```

更复杂的版本（`RunOrderWorkflow3` 用 switch）：

```go
switch {
case strings.Contains(order, "已发货"):
    logistics := tools.QueryLogistics("10001")
    result = fmt.Sprintf("订单已经发货。\n%s", logistics)
case strings.Contains(order, "待支付"):
    payment := tools.QueryPayment("10001")
    result = fmt.Sprintf("订单尚未完成支付。\n%s", payment)
default:
    result = "订单状态暂时无法判断。"
}
```

**特点**：

| 维度 | 表现 |
|------|------|
| 决策者 | 程序（基于前置结果）|
| 灵活性 | 中等（支持分支）|
| 确定性 | 高 |
| 适用场景 | 状态机驱动的业务流程 |

**对应 Anthropic 模式**：**Routing**（路由）

---

### Level 3 · Agentic Workflow（LLM + Workflow 组合）

**LLM 决策任务类型 + 程序执行 Workflow + LLM 总结**。

```go
func RunAgenticWorkflow(ctx, client, model) (string, error) {
    // 1️⃣ LLM 判断任务类型
    classifyResp, _ := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
        Model: model,
        Messages: []Message{
            openai.UserMessage(classifyPrompt),
        },
    })
    classification := classifyResp.Choices[0].Message.Content

    // 2️⃣ 根据类型进入对应 Workflow
    workflowResult := RunOrderWorkflow3()

    // 3️⃣ LLM 总结最终结果
    summaryPrompt := fmt.Sprintf(`请根据下面的业务查询结果，为用户生成最终回答。
        要求：使用中文、简洁清晰、不要提及 Workflow/Tool/LLM 等内部实现
        业务查询结果：%s`, workflowResult)
    finalResp, _ := client.Chat.Completions.New(ctx, ...)
    return finalResp.Choices[0].Message.Content, nil
}
```

**三阶段流程**：

```
用户请求
   ↓
┌─────────────────────────────────┐
│ 阶段 1: LLM 判断任务类型          │
│ 输出: order_analysis / unknown    │
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 阶段 2: 程序执行 Workflow         │
│ (确定性业务逻辑)                  │
└─────────────────────────────────┘
   ↓
┌─────────────────────────────────┐
│ 阶段 3: LLM 总结最终结果          │
│ 输出: 友好的中文回答              │
└─────────────────────────────────┘
```

**特点**：

| 维度 | 表现 |
|------|------|
| 决策者 | LLM 决策 + 程序执行 + LLM 总结 |
| 灵活性 | 高（开放性任务）|
| 确定性 | 中（LLM 决策有不确定性）|
| 适用场景 | 任务类型多样的开放性场景 |

**对应 Anthropic 模式**：**Orchestrator-Workers** 或 **Autonomous Agent**

---

## 🧠 核心洞察

> **让 LLM 做它擅长的不确定性决策，让代码做确定性的业务执行。**

这句话几乎概括了 Agent 编排的全部精髓。

### 为什么这样分工？

| | LLM | 代码 |
|--|-----|------|
| **擅长** | 理解自然语言、分类、生成 | 确定性逻辑、数据处理、状态管理 |
| **不擅长** | 精确数值、严格状态机 | 模糊语义、上下文理解 |

**强行让 LLM 做精确逻辑**（如「判断订单状态决定下一步」）→ 不稳定、成本高
**强行让代码做语义理解**（如「分类用户请求类型」）→ 维护噩梦

---

## 💡 实际项目中最常见的模式

```
用户请求（自然语言）
   ↓
LLM 分类/路由（任务类型）
   ↓
进入对应 Workflow（确定性业务逻辑）
   ↓
Workflow 内部由 Go 代码控制执行
   ↓
LLM 总结最终结果（友好输出）
```

**这就是 Agentic Workflow**。它是绝大多数生产级 Agent 系统的基本形态。

---

## 🔍 RunAgenticWorkflow 详解

### 阶段 1：分类 Prompt

```go
classifyPrompt := `
请判断下面的用户请求属于哪一种类型。

只允许返回以下两个值之一：
order_analysis
unknown

用户请求：
请帮我分析用户1001的订单10001，包括订单、支付和物流情况。
`
```

**设计要点**：
- 明确告诉 LLM「只允许返回 X 或 Y」
- 这是 OpenAI Structured Output 的朴素版（下一步会学习）

### 分类输出处理

```go
classification := resp.Choices[0].Message.Content

// 某些模型可能会输出：
// <think>
// ...
// </think>
// order_analysis
//
// Demo 中简单去掉 think 内容。
if start := strings.Index(classification, "<think>"); start >= 0 {
    if end := strings.Index(classification, "</think>"); end > start {
        classification = classification[:start] + classification[end+len("</think>"):]
    }
}
classification = strings.TrimSpace(classification)

if classification != "order_analysis" {
    return "暂时无法识别该请求类型。", nil
}
```

**为什么需要处理 <think> 标签？**
- 推理类模型（如 DeepSeek-R1、MiniMax-M3 thinking mode）会在 answer 之前输出推理过程
- 这些推理过程不应该出现在分类结果里
- **这是真实项目中的常见坑**

### 阶段 2：执行 Workflow

```go
// LLM 判断出来是订单分析任务，
// 接下来交给确定性的 Workflow。
workflowResult := RunOrderWorkflow3()
```

**关键**：LLM 决策完之后，**业务流程由 Go 代码控制**，不再让 LLM 介入。

### 阶段 3：总结 Prompt

```go
summaryPrompt := fmt.Sprintf(`请根据下面的业务查询结果，为用户生成最终回答。

要求：
1. 使用中文
2. 简洁清晰
3. 总结用户、订单、支付、物流状态
4. 不要提及 Workflow、Tool、LLM 等内部实现

业务查询结果：
%s`, workflowResult)
```

**关键**：要求 LLM 「不要提及内部实现」，避免泄露 Agent 架构细节给最终用户。

---

## ⚠️ 常见误区

### 误区 1：Workflow 和 Agent 二选一

❌ **错**：要么用 Workflow 要么用 Agent
✅ **对**：**实际是组合使用**——LLM 决策 + Workflow 执行

### 误区 2：上来就上 LangGraph / DAG

❌ **错**：复杂任务用 LangGraph 的 StateGraph
✅ **对**：**多数业务用 if/else + 几次 LLM 调用就够**

LangGraph 适合：
- 极复杂的状态流转
- 需要可视化调试
- 多人协作的大型项目

**对当前阶段**（Demo / 学习）：**根本不需要**。

### 误区 3：Workflow = 死板

❌ **错**：Fixed Workflow 不灵活
✅ **对**：Conditional + Agentic 已经能覆盖大部分场景

---

## 📊 三种模式对比

| 维度 | Fixed | Conditional | Agentic |
|------|-------|-----------|---------|
| 决策者 | 程序 | 程序 | LLM + 程序 |
| LLM 调用次数 | 0 | 0 | 2（分类 + 总结）|
| Token 成本 | 0 | 0 | 中等 |
| 响应延迟 | 最低 | 最低 | 中等（2 次 LLM 调用）|
| 灵活性 | 最低 | 中 | 高 |
| 适用场景 | 流程固定 | 状态机驱动 | 任务类型多样 |

---

## 🗺️ 跟 Anthropic《Building Effective Agents》的对应

Anthropic 把所有 Agent 模式分成 **Workflow**（确定性） 和 **Autonomous Agent**（自主性） 两大类。

| Anthropic 模式 | Demo 对应 |
|--------------|---------|
| Prompt Chaining | `RunOrderWorkflow` (Fixed) |
| Routing | `RunOrderWorkflow2/3` (Conditional) |
| Parallelization | （未实现，下一步可加）|
| Orchestrator-Workers | `RunAgenticWorkflow` (Agentic) |
| Autonomous Agent | `agent/agent.go` (Agent Loop) |

**对应关系**：Demo 里的三种 Workflow 覆盖了 Anthropic 提到的核心 Workflow 模式。

---

## 🔄 实际项目中的进阶模式

### 进阶 1：多 Workflow 并行

```go
// 不串行执行所有查询
// 而是并行调用多个独立 tool
var wg sync.WaitGroup
results := make(chan string, 4)

wg.Add(4)
go func() { defer wg.Done(); results <- tools.GetUser("1001") }()
go func() { defer wg.Done(); results <- tools.QueryOrder("10001") }()
go func() { defer wg.Done(); results <- tools.QueryPayment("10001") }()
go func() { defer wg.Done(); results <- tools.QueryLogistics("10001") }()
wg.Wait()
```

### 进阶 2：动态 Workflow 选择

```go
classification := LLM(classifyPrompt)
switch classification {
case "order_analysis":
    return RunOrderWorkflow()
case "user_query":
    return RunUserWorkflow()
case "refund":
    return RunRefundWorkflow()
default:
    return LLM(defaultResponse)
}
```

### 进阶 3：Workflow 嵌套

```
RunAgenticWorkflow
  └→ RunOrderWorkflow3
       └→ tools.QueryOrder
       └→ tools.QueryLogistics
```

---

## 🔗 跟其他模块的关系

| 模块 | 关系 |
|------|------|
| `tools/` | Workflow 调用具体 Tool |
| `config/` | LLM client 配置 |
| `llm/` | LLM 调用（分类 + 总结）|
| `agent/` | **不是 Workflow**，而是另一种形态（自主 Agent）|

**关键区别**：
- **Workflow**：业务流程由代码控制，LLM 只在边界处介入
- **Agent** (`agent/agent.go`)：业务流程由 LLM 决定，循环调用直到完成

**实际项目**：两者**组合使用**——LLM 判断走哪个 Workflow，Workflow 内是确定性代码。

---

## ✅ 自检

- [ ] 1. 三种 Workflow 模式各自由谁决策？
- [ ] 2. 为什么 LLM 决策 + 代码执行 是最常见的组合？
- [ ] 3. `RunAgenticWorkflow` 三个阶段分别是什么？
- [ ] 4. 为什么要处理 <think> 标签？
- [ ] 5. 跟 Anthropic 的 5 种模式怎么对应？

---

## 📂 关键代码

- `workflow/order.go` (239 行)
  - `RunOrderWorkflow` - Fixed (Level 1)
  - `RunOrderWorkflow2` - Conditional if (Level 2)
  - `RunOrderWorkflow3` - Conditional switch (Level 2)
  - `RunAgenticWorkflow` - Agentic (Level 3)
- `main.go` - 演示如何调用 Workflow（注释掉的代码块）

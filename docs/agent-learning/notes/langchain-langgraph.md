# LangChain + LangGraph 学习笔记

> 📚 对应学习计划：[README §Step 10](../README.md)（Multi-Agent 与框架）
> 🗓️ 时间：今天（基于 `langchain-agent-demo/` 反推）
> 💻 关键代码：[`code/06-agent/langchain-agent-demo/`](../../../code/06-agent/langchain-agent-demo/)（15 个 demo，共 1889 行 Python）

> ⚠️ **来源说明**：本笔记由代码反推而成（沙箱无网络访问 ChatGPT 链接）。可能遗漏对话里的讨论细节。

---

## 🎯 核心问题

> 我已经手写过完整的 Go Agent。现在用 LangChain/LangGraph 框架能学到什么？是简单的 API 重写，还是有本质区别？

学完 15 个 demo 后的结论：**LangChain 是封装层，LangGraph 是底层图引擎**。两者配合使用 = 工业标准答案。

---

## 📊 15 个 Demo 一览

### LangChain 基础（01-06）

| Demo | 主题 | 行数 | 学到的 |
|------|------|------|--------|
| `01_model` | LLM 调用 | 31 | `ChatOpenAI` + `ChatPromptTemplate` |
| `03_tool` | `@tool` 装饰器 | 52 | 文档字符串自动生成 Schema |
| `04_agent` | `create_agent` 高层 API | 85 | 5 行代码一个 Agent |
| `05_multi_tool` | 多 Tool | 166 | 多个 `@tool` 装饰器组合 |
| `06_state` | Checkpointer | 129 | `InMemorySaver` + `thread_id` |

### LangGraph 底层（07-09）

| Demo | 主题 | 行数 | 学到的 |
|------|------|------|--------|
| `07_langgraph_basic` | StateGraph 入门 | 75 | `TypedDict` + Node + Edge |
| `08_langgraph_agent` | 完整 Agent | 265 | LLM Node + Tool Node + Conditional Edge |
| `09_create_agent` | 高级 API 再看 | 124 | 对比高层 vs 底层 |

### LangGraph 模式（10-14）

| Demo | 主题 | 学到的模式 |
|------|------|----------|
| `10_react` | ReAct 循环 | Thought / Act / Observation |
| `11_planning` | Plan-and-Execute | Planner → Executor |
| `12_conditional` | 条件路由 | 简单 if/else 抽象 |
| `13_workflow_router` | 多意图工作流 | order / refund / chat 三选一 |
| `14_loop` | 循环节点 | Conditional self-loop |

### LangGraph 高级（15-16）

| Demo | 主题 | 学到的 |
|------|------|--------|
| `15_checkpoint` | 状态持久化 | `InMemorySaver` + `get_state` |
| `16_human_in_loop` | 人在环中 | `interrupt()` + `Command(resume=...)` |

---

## 🧱 Part 1 · LangChain 基础（5 个 demo）

### Demo 01 · LLM 调用（31 行）

```python
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)

prompt = ChatPromptTemplate.from_messages([
    ("system", "你是一名专业的订单客服，只根据用户提供的信息回答问题。"),
    ("human", "{question}"),
])

messages = prompt.invoke({"question": "帮我查询订单10001"})
response = model.invoke(messages)
print(response.content)
```

**3 个关键点**：

| 概念 | 类比（Go） |
|------|----------|
| `ChatOpenAI` | `openai.NewClient(option.WithBaseURL(...))` |
| `ChatPromptTemplate` | 手工拼 system + user message |
| `prompt.invoke(...)` | 直接构造 messages 数组 |

**LangChain 的封装价值**：把「Prompt 模板 + 模型调用」标准化，跨模型切换零成本。

### Demo 03 · `@tool` 装饰器（52 行）

```python
from langchain_core.tools import tool

@tool
def query_order(order_id: str) -> str:
    """查询订单信息。"""

    orders = {"10001": {...}, "10002": {...}}
    order = orders.get(order_id)
    if not order:
        return f"订单 {order_id} 不存在"
    return f"订单号：{order['order_id']}, 状态：{order['status']}, ..."

# 自动从函数签名 + 文档字符串生成 Schema
print(query_order.name)        # "query_order"
print(query_order.description) # "查询订单信息。"
print(query_order.args_schema.schema())
```

**核心魔法**：`@tool` 装饰器**自动从函数签名和 docstring 生成 JSON Schema**。

| Go 手动写 Schema | Python `@tool` |
|------------------|----------------|
| 50+ 行 JSON Schema 声明 | 1 行装饰器 |
| Schema 和函数实现分离 | 一起定义（docstring = description） |
| 修改函数签名要同步改 Schema | 自动跟随 |

**这是 LangChain 最香的设计**：把 Tool 的 schema 定义从「手写 JSON」解放到「代码即 schema」。

### Demo 04 · `create_agent` 高层 API（85 行）

```python
from langchain.agents import create_agent

agent = create_agent(
    model=model,
    tools=[query_order],
    system_prompt="你是一名专业的订单客服。",
)

result = agent.invoke({
    "messages": [{"role": "user", "content": "帮我查一下订单10002"}]
})

print(result["messages"][-1].content)
```

**5 行代码**就完成你的 Go Agent `agent.New(...)` + 整个 `Agent.Run` 循环。

| 对比项 | Go Agent | LangChain `create_agent` |
|--------|---------|-------------------------|
| Agent Loop 代码 | 70 行 | 0 行（封装好）|
| Tool 注册 | 手动 `Registry` | 直接传 `tools` 数组 |
| 调 LLM | 自己写 | 自己写（封装 `model.invoke`）|
| Tool Call 处理 | 自己写 | 自己写（封装）|
| 错误处理 | 自己写 | 自己写 |
| **总代码量** | **~100 行** | **5 行** |

### Demo 06 · Checkpointer（129 行）

```python
from langgraph.checkpoint.memory import InMemorySaver

checkpointer = InMemorySaver()

agent = create_agent(
    model=model,
    tools=[query_order, query_logistics],
    system_prompt="...",
    checkpointer=checkpointer,
)

config = {"configurable": {"thread_id": "user-1001"}}

# 第一轮
result1 = agent.invoke(
    {"messages": [{"role": "user", "content": "帮我查询订单10001"}]},
    config=config,
)

# 第二轮 - 自动延续 thread_id 上下文
result2 = agent.invoke(
    {"messages": [{"role": "user", "content": "那它什么时候送到？"}]},
    config=config,  # ← 同一个 thread_id
)
```

**核心**：`thread_id` + `InMemorySaver` = 多轮对话状态管理。

**等价于你的 Go `ContextManager`**，但 LangGraph **开箱即用**。

---

## 🧠 Part 2 · LangGraph 底层（3 个 demo）

### 为什么要学底层？

`create_agent` 看起来够用，但**不够灵活**：
- 想加自定义 Node（前置检查 / 后置处理）→ ❌
- 想做 Plan-and-Execute → ❌
- 想要条件分支 → ❌
- 想要 Human-in-the-Loop → ❌

**学 LangGraph = 获得完全控制能力**。

### Demo 07 · StateGraph 入门（75 行）

```python
from typing_extensions import TypedDict
from langgraph.graph import StateGraph, START, END

class State(TypedDict):
    question: str
    answer: str
    formatted_answer: str

def call_model(state: State):
    response = model.invoke(state["question"])
    return {"answer": response.content}

def format_answer(state: State):
    return {"formatted_answer": f"AI回答：{state['answer']}"}

builder = StateGraph(State)

builder.add_node("llm", call_model)
builder.add_node("format_answer", format_answer)

builder.add_edge(START, "llm")
builder.add_edge("llm", "format_answer")
builder.add_edge("format_answer", END)

graph = builder.compile()
```

**4 个核心概念**：

```
┌─────────────────────────────────────────┐
│ StateGraph（整个图）                      │
│   ├─ State（TypedDict，节点间传递的数据）  │
│   ├─ Node（处理函数，接收 State 返回 State 增量）│
│   ├─ Edge（节点之间的连线，定义执行顺序）    │
│   └─ START / END（特殊节点：入口/出口）     │
└─────────────────────────────────────────┘
```

| 概念 | 类比（Go） |
|------|----------|
| `StateGraph` | 整个 `Agent.Run` 循环 |
| `State` | `ContextManager.messages` |
| `Node` | 函数：`buildMemoryContext` / `extractAndSaveMemory` |
| `Edge` | `if/else` 流程控制 |
| `START` / `END` | 循环的进入/退出条件 |

### Demo 08 · 完整 LangGraph Agent（265 行）⭐ 核心

这是最复杂的 demo，把所有 LangGraph 概念串起来：

```python
from typing import Annotated
from langgraph.graph.message import add_messages

class State(TypedDict):
    messages: Annotated[list, add_messages]  # ← 自动合并消息

# 1. LLM Node
def call_model(state: State):
    response = model_with_tools.invoke(state["messages"])
    return {"messages": [response]}

# 2. Tool Node
def call_tools(state: State):
    last_message = state["messages"][-1]
    tool_messages = []
    for tool_call in last_message.tool_calls:
        result = tools_by_name[tool_call["name"]].invoke(tool_call["args"])
        tool_messages.append(ToolMessage(content=result, tool_call_id=tool_call["id"]))
    return {"messages": tool_messages}

# 3. Conditional Edge（决定下一步）
def should_continue(state: State):
    last_message = state["messages"][-1]
    if last_message.tool_calls:
        return "tools"
    return "end"

# 4. 构建图
builder = StateGraph(State)
builder.add_node("llm", call_model)
builder.add_node("tools", call_tools)
builder.add_edge(START, "llm")
builder.add_conditional_edges("llm", should_continue, {"tools": "tools", "end": END})
builder.add_edge("tools", "llm")  # Tool → 回到 LLM
```

**图的形状**：

```
   ┌──────────┐
   │  START   │
   └────┬─────┘
        ↓
   ┌──────────┐
   │   LLM    │ ← call_model
   └────┬─────┘
        ↓
   should_continue?
    /        \
  有 tool     无 tool
   ↓           ↓
┌────────┐  ┌──────┐
│ Tools  │  │ END  │
└───┬────┘  └──────┘
    │
    └──→ 回到 LLM
```

**这是经典的 ReAct 循环图**：
- LLM 决策 → 调用 Tool → Tool 结果 → 回到 LLM → ... → END

### 与你的 Go Agent 对照

| 你的 Go 代码 | LangGraph |
|-------------|-----------|
| `Agent.Run` for 循环 | `StateGraph` + Edge 循环 |
| `contextManager.Add(msg)` | `add_messages` Reducer |
| `model.Chat.Completions.New(...)` | `model.invoke(state["messages"])` |
| `message.ToolCalls` 判断 | `should_continue()` Conditional Edge |
| `toolRegistry.Get(name).Handler(...)` | `tools_by_name[name].invoke(args)` |
| `openai.ToolMessage(...)` | `ToolMessage(content, tool_call_id)` |

**结论**：LangGraph 把你的 200 行 Go Agent Loop **完全映射**成 50 行的图定义。

---

## 🔄 Part 3 · LangGraph 模式（5 个 demo）

### 模式 1 · ReAct（Demo 10，169 行）

**LangGraph 写法**（Demo 08）vs **手动 while 循环**（Demo 10）：

```python
# Demo 10：手动 while 循环（最贴近 Go 写法）
while True:
    response = model_with_tools.invoke(messages)
    messages.append(response)
    if not response.tool_calls:
        break
    for tool_call in response.tool_calls:
        result = tool_map[tool_call["name"]].invoke(tool_call["args"])
        messages.append(ToolMessage(content=result, tool_call_id=tool_call["id"]))
```

**对比 Go**：

```go
for {
    resp, _ := client.Chat.Completions.New(ctx, ...)
    messages = append(messages, resp.Choices[0].Message.ToParam())
    if len(resp.Choices[0].Message.ToolCalls) == 0 {
        break
    }
    for _, tc := range resp.Choices[0].Message.ToolCalls {
        result, _ := toolRegistry.Get(tc.Function.Name).Handler(tc.Function.Arguments)
        messages = append(messages, openai.ToolMessage(result, tc.ID))
    }
}
```

**LangGraph 的优势**：图能**可视化**，while 循环不行。

### 模式 2 · Plan-and-Execute（Demo 11，148 行）

```python
def planner(state):
    response = model.invoke(f"为这个任务制定计划：{state['user_task']}")
    return {"plan": response.content}

def executor(state):
    response = model.invoke(f"按计划执行：{state['plan']}")
    return {"result": response.content}

builder = StateGraph(AgentState)
builder.add_node("planner", planner)
builder.add_node("executor", executor)
builder.add_edge(START, "planner")
builder.add_edge("planner", "executor")
builder.add_edge("executor", END)
```

**图形状**：

```
START → Planner → Executor → END
```

**这是 Anthropic 说的 5 种模式之一的 Orchestrator-Workers** 的最简实现。

### 模式 3 · Conditional Routing（Demo 12，149 行）

```python
def router(state):
    # LLM 判断要 query_order 还是 direct_answer
    return {"route": "query" if "订单" in state["user_task"] else "direct"}

def query_order(state):
    return {"result": "订单信息..."}

def direct_answer(state):
    return {"result": model.invoke(state["user_task"]).content}

builder.add_conditional_edges(
    "router",
    lambda state: state["route"],
    {"query": "query_order", "direct": "direct_answer"}
)
```

**对比你的 Go Router**：本质完全相同，只是用图的形式表达。

### 模式 4 · Multi-Intent Workflow Router（Demo 13，217 行）⭐ 重要

**3 个意图，3 个 Workflow**：

```python
def intent_router(state):
    intent = classify(state["user_task"])  # LLM 分类
    return {"intent": intent}  # "order" / "refund" / "chat"

def order_workflow(state): ...
def refund_workflow(state): ...
def chat_workflow(state): ...

builder.add_conditional_edges(
    "intent_router",
    lambda s: s["intent"],
    {
        "order": "order_workflow",
        "refund": "refund_workflow",
        "chat": "chat_workflow",
    }
)
```

**图形状**：

```
                    ┌→ order_workflow  ─┐
START → router ─────┼→ refund_workflow ─┼→ END
                    └→ chat_workflow    ─┘
```

**这就是 Step 7 的 Workflow + Step 9 的 Router 的 LangGraph 实现**。

### 模式 5 · Loop（Demo 14，90 行）

```python
def should_continue(state):
    return "continue" if state["current_step"] <= 3 else "finish"

builder.add_conditional_edges(
    "executor",
    should_continue,
    {"continue": "executor", "finish": END}  # ← self-loop
)
```

**核心**：`continue` 指向自己 = 循环节点。

---

## 🛡️ Part 4 · 高级特性（2 个 demo）

### Demo 15 · Checkpoint（81 行）

```python
from langgraph.checkpoint.memory import InMemorySaver

checkpointer = InMemorySaver()
graph = builder.compile(checkpointer=checkpointer)

config = {"configurable": {"thread_id": "user-001"}}

# 第一次执行
result1 = graph.invoke({"message": "hello", "count": 0}, config)

# 查看保存的 State
saved_state = graph.get_state(config)
print(saved_state)
```

**等价于你的 Go `ContextManager`**，但 LangGraph 提供了：
- `get_state(config)` - 主动查看保存的状态
- 自动序列化（任何可序列化的对象）
- `thread_id` 隔离不同对话

### Demo 16 · Human-in-the-Loop（107 行）⭐ 重要

```python
from langgraph.types import Command, interrupt

def process_task(state):
    # 暂停执行，等待人工确认
    approval = interrupt({
        "message": "该操作需要人工确认，是否继续？",
        "task": state["task"],
    })
    
    if approval == "yes":
        return {"result": "任务已执行"}
    return {"result": "任务已取消"}

# 第一次执行（会暂停）
result = graph.invoke({"task": "执行退款操作"}, config)

# 人工确认后恢复
result = graph.invoke(Command(resume="yes"), config)
```

**关键 API**：
- `interrupt(payload)` - **暂停** Graph 执行，返回 payload 给前端
- `Command(resume=value)` - **恢复** Graph，传入人工决定

**这是 LangGraph 独有的强大特性**：你的 Go Agent 要实现这个需要写大量代码。

---

## 🔄 Go Agent vs LangChain/LangGraph 对照表

### 概念映射

| Go Agent 模块 | LangChain / LangGraph | 备注 |
|--------------|----------------------|------|
| `openai.Client` | `ChatOpenAI` | OpenAI 兼容 |
| `tools.Registry` | `@tool` 装饰器列表 | LangChain 用装饰器更简洁 |
| `BuildToolDefinitions()` | `query_order.args_schema` | LangChain 自动生成 |
| `agent.New(...)` | `create_agent(...)` | LangChain 一行搞定 |
| `Agent.Run` for 循环 | `StateGraph` + Edge | LangGraph 用图表达 |
| `ContextManager.messages` | `State["messages"]` + `add_messages` | LangGraph 自动合并 |
| `Router.Decide` | `add_conditional_edges` | 本质相同 |
| `ContextManager.Summarize` | Checkpointer | LangGraph 更通用 |
| 无 | `interrupt()` / `Command(resume)` | LangGraph 独有 |
| 无 | `graph.get_graph().draw_mermaid_png()` | 可视化 |

### 代码量对比

| 功能 | Go Agent | LangChain 高层 | LangGraph 底层 |
|------|---------|---------------|---------------|
| 单 Tool Agent | ~150 行 | **5 行** | ~50 行 |
| 多 Tool Agent | ~200 行 | ~10 行 | ~80 行 |
| ReAct 循环 | ~70 行 | 0 行（封装）| ~40 行 |
| 条件路由 | ~30 行（if/else）| N/A | ~20 行（add_conditional_edges）|
| Plan-and-Execute | ~80 行 | N/A | ~50 行 |
| Multi-Intent | ~120 行 | N/A | ~80 行 |
| Checkpoint | ~50 行 | 1 行 | ~5 行 |
| Human-in-the-Loop | **不支持** | N/A | **~20 行** |
| 图可视化 | **不支持** | N/A | **1 行** |

**结论**：
- 简单场景 → LangChain 高层省 80%+ 代码
- 复杂场景 → LangGraph 跟 Go 代码量相当，但**可视化 + 高级特性**完胜
- HITL / 可视化 → LangGraph 独有，Go 完全没这能力

---

## 🎯 关键洞察

### 1. `create_agent` vs `StateGraph`：高层 vs 底层

```
create_agent ─→ 适合 80% 场景，5 行搞定
StateGraph  ─→ 适合 20% 复杂场景（条件路由 / HITL / 子图）
```

**学习策略**：先 `create_agent` 出活，遇到瓶颈再下钻 `StateGraph`。

### 2. `@tool` 装饰器的本质

**Schema 跟代码定义在一起**：

```python
@tool
def query_order(order_id: str) -> str:
    """查询订单信息。"""
    # ↓ 自动生成
    # {
    #   "name": "query_order",
    #   "description": "查询订单信息。",
    #   "parameters": {
    #     "type": "object",
    #     "properties": {"order_id": {"type": "string"}},
    #     "required": ["order_id"]
    #   }
    # }
```

**好处**：函数签名改了，Schema 自动同步。**这是 LangChain 最值得借鉴的设计**。

### 3. `State` + Reducer 是 LangGraph 的灵魂

```python
class State(TypedDict):
    messages: Annotated[list, add_messages]
```

`add_messages` 是 **Reducer**：决定新消息如何合并到旧列表。

**类比 React 的 `useReducer`**：状态变更不直接修改，而是通过 reducer 计算新值。

### 4. `interrupt()` 是 LangGraph 的杀手锏

Go Agent 要做 HITL：

```go
// 1. 检测是否需要人工确认
if isDangerous(toolName) {
    // 2. 保存当前状态
    saveCheckpoint(state)
    // 3. 暂停，等待人工输入
    <-humanInputChannel
    // 4. 恢复执行
    state = loadCheckpoint()
    // 5. 继续
}
// ... 继续执行
```

LangGraph 3 行：

```python
approval = interrupt({"question": "是否继续？"})
result = graph.invoke(Command(resume=approval), config)
```

**差 10 倍代码量，LangGraph 直接帮你搞定持久化 + 暂停 + 恢复**。

### 5. 图可视化 = Debug 神器

```python
graph.get_graph().draw_mermaid_png(output_file_path="graph.png")
```

**Go Agent**：要在脑子里跑 for 循环
**LangGraph**：直接看 PNG

---

## ⚠️ 当前实现的局限

### 局限 1：`create_agent` 的灵活性有限

如果想：
- 在 Tool Call 前后插入自定义逻辑（如日志、风控）
- 多个 LLM 协作
- 复杂的条件分支

→ 必须下钻 `StateGraph`。

### 局限 2：`InMemorySaver` 不持久化

进程重启，Checkpoint 丢失。

**生产应该**：`PostgresSaver` / `RedisSaver`。

### 局限 3：HITL 在生产中的 UX

当前 `interrupt()` 返回 dict，前端怎么展示？

**生产需要**：跟前端 WebSocket / SSE 集成，把 payload 推给 UI。

### 局限 4：没有评估

LangGraph 不带 Eval。生产要接 LangSmith / LangFuse。

---

## 🎓 面试角度：LangChain/LangGraph 怎么聊

### 30 秒版本

> 我系统学过 LangChain 和 LangGraph。LangChain 的 `create_agent` 一行能搞定 80% 的 Agent；LangGraph 的 `StateGraph` 能处理剩下的复杂场景，包括 Plan-and-Execute、Multi-Intent Routing、Human-in-the-Loop。LangGraph 的 Checkpointer 是开箱即用的多轮对话管理，`interrupt()` 是 HITL 的杀手锏。

### 5 分钟版本

按 5 个 demo 重点讲：
1. `create_agent` 快速搭建（04 / 05）
2. StateGraph 底层（07 / 08）
3. 条件路由（12 / 13）
4. Checkpoint（06 / 15）
5. Human-in-the-Loop（16）

### 加分项

- 提 `@tool` 自动生成 Schema（vs Go 手动写）
- 提 `add_messages` Reducer 设计
- 提 LangGraph 与 Go Agent 的对照（说明你真的懂底层）

---

## ✅ 自检

- [ ] 1. LangChain 高层 API 和 LangGraph 底层有什么区别？
- [ ] 2. `@tool` 装饰器为什么比手动写 JSON Schema 好？
- [ ] 3. `StateGraph` 的 4 个核心概念是什么？
- [ ] 4. `add_conditional_edges` 怎么用？等价于 Go 代码的什么？
- [ ] 5. `interrupt()` / `Command(resume=...)` 解决了什么问题？
- [ ] 6. Checkpointer 跟你的 Go `ContextManager` 是什么关系？
- [ ] 7. Plan-and-Execute / Multi-Intent Routing 怎么用 LangGraph 实现？

---

## 📂 关键代码

- `langchain-agent-demo/01_model/` - LLM 基础
- `langchain-agent-demo/03_tool/` - @tool 装饰器
- `langchain-agent-demo/04_agent/` - 第一个 Agent
- `langchain-agent-demo/06_state/` - Checkpointer
- `langchain-agent-demo/07_langgraph_basic/` - StateGraph 入门
- `langchain-agent-demo/08_langgraph_agent/` - 完整 Agent（最复杂 265 行）
- `langchain-agent-demo/10_react/` - ReAct 手动循环
- `langchain-agent-demo/11_planning/` - Plan-and-Execute
- `langchain-agent-demo/13_workflow_router/` - 多意图路由
- `langchain-agent-demo/15_checkpoint/` - Checkpoint
- `langchain-agent-demo/16_human_in_loop/` - HITL

**总计 15 个 demo · 1889 行 Python**

---

## 🔗 相关笔记

- [Memory 完整体系](./memory.md) - 对比 LangGraph `add_messages` Reducer
- [Router 详解](./router.md) - 对比 LangGraph `add_conditional_edges`
- [Agent 架构总览](./agent-architecture.md) - 整体架构跟 LangGraph 的对应关系
- [MCP 详解](./mcp.md) - LangGraph + MCP 的集成方式

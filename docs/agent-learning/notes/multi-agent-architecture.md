# Multi-Agent 架构实战：Sub-Agent as Tool / 分布式 HTTP / Trace ID

> 📚 配套笔记：[agent-tool-security-and-reliability.md](./agent-tool-security-and-reliability.md)（Multi-Agent 概念部分）/ [langgraph-production-extensions.md](./langgraph-production-extensions.md)（HITL 基础）
> 🗓️ 时间：今天（基于 ChatGPT 分享 [6ab48013](https://chatgpt.com/share/6ab48013-59ec-83ee-9ca7-742295ca6b53) + `langchain-agent-demo/39_multi_agent`）
> 🎯 主题：企业 Multi-Agent 完整架构

> 💡 **本笔记覆盖**：
> 1. Multi-Agent vs Multi-Tool（本质区别）
> 2. **Sub-Agent as Tool 模式**（LangChain 官方推荐）
> 3. **分布式 Multi-Agent HTTP**（FastAPI 多服务）
> 4. **Trace ID 透传**（分布式追踪）
> 5. Multi-Agent 通信模式（4 种）
> 6. **Multi-Agent vs DDD/Microservices**（架构对比）
> 7. 面试完整回答模板

---

## 🎯 核心问题

> 什么时候需要 Multi-Agent？怎么实现？分布式部署怎么做？

之前 `agent-tool-security-and-reliability.md` 只涉及了概念，这一章**用 4 个 Demo 完整实现**企业 Multi-Agent 架构。

---

## 1️⃣ Multi-Agent vs Multi-Tool（再强调）

### 本质区别

| 维度 | Multi-Tool | Multi-Agent |
|------|-----------|------------|
| **决策中心** | 1 个 LLM 决定调哪些 Tool | 每个 Agent 独立决策 |
| **状态** | 共享同一 State | 每个 Agent 独立 State |
| **专长** | 共享一个 System Prompt | 每个 Agent 独立 Prompt |
| **通信** | Tool 直接返回字符串 | Agent 间可交互 |
| **典型实现** | 1 个 Agent + N 个 Tool | Supervisor + N 个 Worker Agent |

### 关键认知

> **Multi-Agent = 每个 Agent 是一个独立的「思考 + 决策」单元**。
> 不仅是「多几个 Tool」。

---

## 2️⃣ Sub-Agent as Tool 模式（LangChain 官方推荐）⭐⭐⭐

### 核心思想（Demo 39-02，253 行）

> **把 Sub-Agent 包装成 Tool，让 Supervisor Agent 像调用普通 Tool 一样调用 Sub-Agent**。

### 完整代码骨架

```python
from langchain.agents import create_agent
from langchain_core.tools import tool

# 1. 创建 3 个专业 Sub-Agent
order_agent = create_agent(
    model=model,
    tools=[],
    system_prompt="你是订单专家 Agent。只处理订单相关问题。...",
)

payment_agent = create_agent(
    model=model,
    tools=[],
    system_prompt="你是支付专家 Agent。只处理支付相关问题。...",
)

logistics_agent = create_agent(
    model=model,
    tools=[],
    system_prompt="你是物流专家 Agent。只处理物流相关问题。...",
)

# 2. 把每个 Agent 包装成 Tool
@tool("order_agent", description="订单领域专家。处理订单状态、金额、用户、发货等。")
def call_order_agent(query: str) -> str:
    result = order_agent.invoke({"messages": [{"role": "user", "content": query}]})
    return result["messages"][-1].content

@tool("payment_agent", description="支付领域专家。处理支付状态、方式、退款等。")
def call_payment_agent(query: str) -> str:
    result = payment_agent.invoke({"messages": [{"role": "user", "content": query}]})
    return result["messages"][-1].content

@tool("logistics_agent", description="物流领域专家。处理物流状态、快递、运单等。")
def call_logistics_agent(query: str) -> str:
    result = logistics_agent.invoke({"messages": [{"role": "user", "content": query}]})
    return result["messages"][-1].content

# 3. 创建 Supervisor Agent（这些 Tool 对它来说就是普通 Tool）
supervisor = create_agent(
    model=model,
    tools=[call_order_agent, call_payment_agent, call_logistics_agent],
    system_prompt="你是 Supervisor。理解用户任务，调用合适的专业 Agent。",
)

# 4. Supervisor 自己决定调谁（LLM 路由）
result = supervisor.invoke({"messages": [{
    "role": "user", 
    "content": "帮我查询订单10001的状态和支付情况"
}]})
```

### 为什么这是 LangChain 官方推荐？

| 优势 | 说明 |
|------|------|
| **简单** | Sub-Agent 用 `@tool` 一行包装 |
| **统一** | Supervisor 视角：「所有 Sub-Agent 都是 Tool」 |
| **灵活** | LLM 自己决定调谁、调几个、什么顺序 |
| **复用** | Sub-Agent 可以独立测试、独立部署（Demo 03）|

**Supervisor 的视野里没有 Sub-Agent 的概念**，只有 Tool。这降低了认知复杂度。

### 跟传统 if/else 路由的区别

```python
# 传统 if/else（代码决定路由）
if "订单" in query: 
    return order_agent.invoke(...)
elif "支付" in query:
    return payment_agent.invoke(...)

# Sub-Agent as Tool（LLM 决定路由）
# Supervisor Agent 自己根据 Tool 描述 + 用户问题决定调谁
```

**优势**：不写死规则，能处理模糊/多领域问题（如"订单和支付一起查"）。

---

## 3️⃣ 分布式 Multi-Agent（FastAPI + HTTP）

### 何时需要分布式（Demo 39-03）

```
同进程调用（Demo 39-02）     │  分布式调用（Demo 39-03）
─────────────────────────  │  ─────────────────────────
所有 Agent 在一个进程      │  每个 Agent 独立服务
共享 Python 运行时         │  独立进程 / 独立服务器
适合 Demo / 单机部署        │  适合生产 / 多实例部署
```

### 完整架构（Demo 39-03）

```
┌──────────────────────────────────────────────────────┐
│                 Supervisor (port 8000)                 │
│                                                       │
│   接收用户请求                                        │
│   ↓                                                    │
│   Supervisor Agent 判断要调哪个 Worker                │
│   ↓                                                    │
│   HTTP POST /agent/order  →  远程调用 Order Agent    │
│   HTTP POST /agent/payment → 远程调用 Payment Agent  │
└────────────┬──────────────────────┬──────────────────┘
             │                      │
             ▼                      ▼
┌──────────────────────┐  ┌──────────────────────┐
│  Order Agent          │  │  Payment Agent        │
│  (port 8001)          │  │  (port 8002)          │
│                       │  │                       │
│  FastAPI 服务          │  │  FastAPI 服务          │
│  独立进程              │  │  独立进程              │
│  可以独立部署 / 扩缩容 │  │  可以独立部署 / 扩缩容 │
└──────────────────────┘  └──────────────────────┘
```

### Worker Agent 的 FastAPI 骨架

```python
# 03_distributed_http/order_agent/main.py
from fastapi import FastAPI
from pydantic import BaseModel
from langchain.agents import create_agent

# 1. 独立进程独立 Agent
order_agent = create_agent(
    model=model,
    tools=[],
    system_prompt="你是订单领域专家 Agent。...",
)

# 2. FastAPI 包装
app = FastAPI()

class AgentRequest(BaseModel):
    task_id: str
    user_id: str
    task: str

@app.post("/agent/order")
async def handle_order(req: AgentRequest):
    result = order_agent.invoke({"messages": [{"role": "user", "content": req.task}]})
    return {
        "task_id": req.task_id,
        "user_id": req.user_id,
        "data": {"answer": result["messages"][-1].content},
    }
```

### Supervisor 通过 HTTP 调用 Worker

```python
# 03_distributed_http/supervisor/main.py
import httpx

@tool("order_agent", description="调用远程订单领域 Agent")
def call_order_agent(task: str) -> str:
    response = httpx.post(
        "http://127.0.0.1:8001/agent/order",
        headers={"X-Trace-ID": TRACE_ID},  # ← Trace ID 透传
        json={"task_id": "...", "user_id": "...", "task": task},
        timeout=30,
    )
    data = response.json()
    return data["data"]["answer"]
```

### 跟 Go 后端对比

| Go 微服务 | Agent 微服务 |
|---------|-------------|
| gRPC / HTTP | FastAPI / HTTP |
| Service 注册 | Agent 注册为 Tool |
| Trace ID（OpenTelemetry） | X-Trace-ID Header |
| 服务发现 | 静态 URL / DNS |

**本质相同**，只是 Service 变成了 Agent。

---

## 4️⃣ Trace ID 透传（分布式追踪）⭐

### 核心问题（Demo 39-04）

```
用户请求 → Supervisor → HTTP → Order Agent → 调 LLM
                                ↓
                Payment Agent → 调 LLM
                
这些日志是同一个请求吗？怎么关联？
```

### 解决方案：Trace ID 透传

```python
# Supervisor：整个用户请求只生成一次 trace_id
import uuid
TRACE_ID = str(uuid.uuid4())  # ← 同一个请求链全程不变

@tool("order_agent")
def call_order_agent(task: str) -> str:
    task_id = str(uuid.uuid4())  # ← task_id 每调用一次都不同
    
    response = httpx.post(
        "http://127.0.0.1:8001/agent/order",
        headers={"X-Trace-ID": TRACE_ID},  # ← HTTP Header 透传
        json={"task_id": task_id, "task": task},
    )
    return response.json()["data"]["answer"]
```

```python
# Worker Agent：从 Header 读取
from fastapi import Header

@app.post("/agent/order")
async def handle_order(
    req: AgentRequest,
    x_trace_id: str = Header(None),  # ← 读 Header
):
    print(f"Trace ID: {x_trace_id}, Task: {req.task}")
    # ... 处理 ...
```

### Trace ID vs Task ID

| 概念 | 生命周期 | 数量 |
|------|---------|------|
| **Trace ID** | 整个用户请求 | 1 个 |
| **Task ID** | 每次 Agent 调用 | 每个调用 1 个 |

```
用户请求
└── Trace ID: abc-123        ← 整个链路同一 ID
    ├── Supervisor 调用 Order Agent
    │   └── Task ID: task-001    ← Supervisor 内部任务
    └── Supervisor 调用 Payment Agent
        └── Task ID: task-002
```

**为什么需要两个 ID**：
- Trace ID：关联**整个用户请求**的所有日志
- Task ID：区分**同一请求内的不同子任务**

### 生产建议

> Demo 用简单的 HTTP Header。
> 生产用 **OpenTelemetry** 完整分布式追踪体系。

---

## 5️⃣ Multi-Agent 通信模式（4 种）⭐⭐

来自 ChatGPT 总结：

### 模式 1：同一 Workflow 内的 State 共享

```python
class State(TypedDict):
    task: str
    order_info: dict   # ← Sub-Agent 写入
    payment_info: dict  # ← Sub-Agent 写入
```

**适用**：所有 Sub-Agent 在同一 LangGraph Workflow 内。

### 模式 2：跨服务的同步 HTTP

```
Supervisor → HTTP POST → Worker Agent
            ↑              ↓
        返回结果 JSON    处理任务
```

**适用**：短任务（< 30 秒）、结果立即需要。

### 模式 3：异步 MQ（Kafka / RabbitMQ）

```
Supervisor → MQ (发布任务)  →  Worker Agent 监听
                                              ↓
                                          处理任务
                                              ↓
                              MQ (发布结果)  → Supervisor 监听
```

**适用**：长任务（几分钟 ~ 几小时）、解耦、事件驱动。

### 模式 4：共享存储（DB / Redis）

```
Agent A 写数据库
    ↓
Agent B 读数据库
```

**适用**：大数据结果（几 MB 不适合放 MQ）、跨 Agent 状态共享。

### 如何选择？

| 场景 | 推荐模式 |
|------|---------|
| 同一 LangGraph Workflow | 模式 1（State）|
| 短任务 + 立即要结果 | 模式 2（HTTP）|
| 长任务 / 解耦 / 事件驱动 | 模式 3（MQ）|
| 大数据 / 跨 Agent 状态 | 模式 4（DB/Redis）|

**关键**：不要把大数据塞进 MQ（传输慢）。MQ 传递引用，大数据走 DB/对象存储。

---

## 6️⃣ Multi-Agent vs DDD / Microservices（架构对比）

来自 ChatGPT 的关键洞察：

> **Multi-Agent 在架构思想上类似 DDD 和微服务的领域拆分**，
> **但每个 Agent 不只是传统 Service，还具备基于 LLM 的任务理解和动态决策能力**。

### 对比表

| 维度 | 传统 Microservice | Multi-Agent |
|------|------------------|-------------|
| **拆分依据** | 业务边界（DDD Bounded Context）| 任务领域（Order / Payment / Logistics）|
| **决策方式** | 硬编码 if/else / 配置 | LLM 动态决定 |
| **通信** | HTTP / gRPC / MQ | HTTP / State / MQ |
| **状态** | 共享 DB | 独立 State（LangGraph）|
| **可观测** | OpenTelemetry | OpenTelemetry + LangSmith |
| **部署** | K8s 多 Pod | K8s 多 Pod（同 LangChain）|

### 关键洞察

> **Multi-Agent 没有让传统分布式系统知识失效，反而把传统问题带到了更复杂的层面**。
> 
> 后端经验（分布式系统、Trace ID、Retry、Circuit Breaker）**直接复用**。

### 现实企业架构（ChatGPT 给的图）

```
┌──────────────────────────────────────────────────────┐
│              API Gateway / Load Balancer              │
└──────────────────────┬───────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        ▼              ▼              ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  Supervisor   │ │  Order Agent │ │ Payment Agent │
│  Agent       │ │  Service     │ │  Service     │
└──────┬───────┘ └──────────────┘ └──────────────┘
       │
       │ (异步长任务)
       ▼
┌──────────────┐
│  Kafka / MQ  │ ←── Worker Agent 监听
└──────────────┘

┌──────────────────────────────────────────────────────┐
│           Observability (OpenTelemetry + LangSmith)    │
│   Trace ID 串联所有调用 / Agent 决策可追溯              │
└──────────────────────────────────────────────────────┘
```

---

## 7️⃣ 关键认知（来自 ChatGPT）

### 1. Multi-Agent 的核心价值 = 职责隔离

> 每个 Sub-Agent 在**独立上下文中**完成复杂任务，
> 再把结果返回给主 Agent。

**Order Agent 不需要知道 Payment Agent 的几十轮内部推理**。

### 2. 不要共享巨大的 State

> 生产环境通常不会让所有 Agent 共享一个巨大的上下文，
> 而是强调 Agent 之间的**职责和上下文隔离**。

跟微服务里共享一个数据库的问题非常类似。

### 3. 不要过度上复杂通信协议

> 暂时不要研究 Agent-to-Agent Protocol、复杂消息协议。
> 
> 这会马上进入比较深的 Multi-Agent 工程领域。

**当前阶段掌握「Sub-Agent as Tool」+ 「HTTP + Trace ID」就够了**。

### 4. Multi-Agent 不会取代 DDD

> 但 DDD 思想**本身不会因为 AI 消失**。
> 
> 继续负责定义业务边界，而 Agent 成为这些业务边界中的智能执行者。

**Supervisor Agent 还是需要业务边界做路由**。

---

## 8️⃣ 面试回答模板（30 秒版）

> **企业 Multi-Agent 通常采用 Supervisor + Sub-Agent 架构**：
> - Supervisor Agent 是入口，根据用户任务调用合适的专业 Sub-Agent
> - Sub-Agent 用 LangChain `create_agent` 创建，并通过 `@tool` 装饰器暴露给 Supervisor
> - Sub-Agent 之间**职责隔离**（独立的 System Prompt、State、上下文）
> - 分布式部署时通过 **HTTP + Trace ID** 通信，跟微服务架构一致
> - 长任务用 **MQ**（Kafka / RabbitMQ），大结果用 **DB / 对象存储**
> - 整体可观测用 **OpenTelemetry + LangSmith**

### 5 分钟版

按 4 个真实场景讲（参考上一份 `agent-tool-security-and-reliability.md` 的 4 场景），加上 Multi-Agent 的部分。

---

## ✅ 自检

- [ ] 1. Multi-Agent 和 Multi-Tool 的本质区别是什么？
- [ ] 2. Sub-Agent as Tool 模式相比传统 if/else 路由有什么优势？
- [ ] 3. 分布式 Multi-Agent 怎么实现？需要哪些基础设施？
- [ ] 4. Trace ID 和 Task ID 的区别是什么？为什么需要两个？
- [ ] 5. Multi-Agent 通信模式有哪 4 种？各自适用什么场景？
- [ ] 6. Multi-Agent 和 DDD/Microservices 是什么关系？
- [ ] 7. 企业 Multi-Agent 通常怎么部署？跟 Go 微服务有什么异同？

---

## 📂 相关代码

- `langchain-agent-demo/39_multi_agent/01_basic/` - Multi-Agent 基础
- `langchain-agent-demo/39_multi_agent/02_agent_as_tool/` - Sub-Agent as Tool 模式（253 行）
- `langchain-agent-demo/39_multi_agent/03_distributed_http/` - 分布式 HTTP（3 个 FastAPI 服务）
- `langchain-agent-demo/39_multi_agent/04_trace_id/` - Trace ID 透传

---

## 🔗 相关笔记

- [Agent 工具安全 + 可靠性](./agent-tool-security-and-reliability.md) - Multi-Agent 概念 + Tool 安全链
- [LangGraph 生产扩展](./langgraph-production-extensions.md) - HITL / Checkpointer（Sub-Agent 内部用）
- [Agent 架构总览](./agent-architecture.md) - 整体架构视角

---

## 📅 路线现状

```
✅ Step 1-12  全部完成
✅ Multi-Agent 架构实战（39_multi_agent 4 个 demo）
⬜ Long-term Memory 实战（即将开始）
⬜ 可观测性（LangSmith / LangFuse）
⬜ Evaluation（Trajectory / Outcome）
```

**ChatGPT 路线图**：
> 不再深入 Multi-Agent 通信协议 → 进入 **Long-term Memory** → **可观测性** → **Evaluation**

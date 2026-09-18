# LangGraph 生产扩展：HITL / Checkpointer / FastAPI

> 📚 配套笔记：[langchain-langgraph.md](./langchain-langgraph.md)（基础）
> 🗓️ 时间：今天（基于 ChatGPT 分享 [6aab992e](https://chatgpt.com/share/6aab992e-ce3c-83e8-ad68-51d82403781f) / [6aab99e2](https://chatgpt.com/share/6aab99e2-aa34-83e8-97dc-ee1a2149d3b5) 补充）
> 🎯 主题：从 Demo 走向生产落地

> 💡 **本笔记覆盖**：之前 `langchain-langgraph.md` 没深入的 4 个生产话题：
> 1. HITL 真实 API 形态（interrupt + resume API）
> 2. PostgreSQL/Redis Checkpointer（生产持久化）
> 3. Memory vs Checkpoint 的本质区别
> 4. FastAPI 定位（不是 Agent 框架）

---

## 🎯 核心问题

> Demo 跑通了，怎么把它变成一个生产可用的服务？

Demo 阶段只关心：图能跑通、Tool 能调用。
生产阶段必须关心：用户从 HTTP 来、审批流怎么走、服务重启怎么恢复。

---

## 1️⃣ HITL 真实 API 形态

### Demo 阶段的简化（`16_human_in_loop/main.py`）

Demo 直接在 Python 控制台用 `Command(resume="yes")` 恢复，太简单：

```python
# 第一次：暂停
result = graph.invoke({"task": "执行退款操作"}, config)

# 第二次：直接传 resume
result = graph.invoke(Command(resume="yes"), config)
```

**这不能上线**——没有 HTTP 层、没有前端交互。

### 生产阶段的真实 API 形态

```
┌─────────┐         ┌──────────┐         ┌─────────┐
│  前端   │         │ FastAPI  │         │ LangGraph│
│         │         │ Service  │         │ Workflow │
└────┬────┘         └────┬─────┘         └────┬────┘
     │  POST /refund    │                     │
     │  {order_id}      │                     │
     ├─────────────────→│ graph.invoke(...)    │
     │                  ├────────────────────→│
     │                  │                     │ interrupt(...)
     │                  │                     │ ↓ 暂停
     │  200 OK          │                     │
     │  {status:        │                     │
     │   "waiting_for_  │                     │
     │    approval",    │                     │
     │   thread_id,     │                     │
     │   task}          │                     │
     │←─────────────────┤                     │
     │                  │                     │
     │  ... 用户审批 ...                      │
     │                  │                     │
     │  POST /refund/   │                     │
     │   {thread_id}/   │                     │
     │   resume         │                     │
     │  {decision:      │                     │
     │   "yes"}         │                     │
     ├─────────────────→│ graph.invoke(       │
     │                  │   Command(resume=   │
     │                  │     "yes"), config) │
     │                  ├────────────────────→│
     │                  │                     │ 继续执行
     │  200 OK          │                     │
     │  {status:        │                     │
     │   "completed",   │                     │
     │   result}        │                     │
     │←─────────────────┤                     │
```

### 关键 API 状态

```python
# 第一次调用
{
    "status": "waiting_for_approval",
    "thread_id": "conversation-001",
    "task": "退款订单10001",
    "approval_message": "该操作需要人工确认，是否继续？"
}

# 第二次调用（resume）
{
    "status": "completed",
    "result": "退款已执行，订单状态：已退款"
}
```

### 设计要点

| 阶段 | HTTP 状态 | Body 字段 |
|------|----------|----------|
| 首次调用 | 200 OK | `status: waiting_for_approval` |
| Resume 成功 | 200 OK | `status: completed` |
| Resume 拒绝 | 200 OK | `status: cancelled` |
| Thread 不存在 | 404 | error |
| Resume 值非法 | 400 | error |

**关键**：**两个 API 都要带 `thread_id`**，否则 LangGraph 不知道恢复哪条 Workflow。

---

## 2️⃣ PostgreSQL/Redis Checkpointer

### Demo 用的 `InMemorySaver` 是什么

```python
from langgraph.checkpoint.memory import InMemorySaver

checkpointer = InMemorySaver()  # ← 进程内 Map
```

**问题**：进程重启 = 所有 Checkpoint 丢失 = 用户必须从头来。

### 生产 Checkpointer 选项

```
┌─────────────────┬────────────────────────────────┐
│ Checkpointer    │ 适用场景                        │
├─────────────────┼────────────────────────────────┤
│ InMemorySaver   │ Demo / 单进程测试              │
│ SqliteSaver     │ 单机 / 中小规模                 │
│ PostgresSaver   │ 分布式生产（首选）             │
│ RedisSaver      │ 短期 Session / 高并发          │
└─────────────────┴────────────────────────────────┘
```

### PostgreSQL Checkpointer 使用

```python
from langgraph.checkpoint.postgres import PostgresSaver

# 官方 Checkpointer 会自动初始化所需表结构
# 不需要自己设计 checkpoint 表
checkpointer = PostgresSaver.from_conn_string(
    "postgresql://user:password@localhost:5432/langgraph"
)

graph = builder.compile(checkpointer=checkpointer)
```

**关键认知**（来自 ChatGPT 分享）：

> **不要把 Checkpointer 本身当数据库**。
> 它的职责是 **保存 Graph 执行状态**（checkpoint、write 记录、关联信息）。
> 具体的表名/字段由 LangGraph 版本决定，跟着版本 migration 走就行。

### Redis Checkpointer 的特点

```
┌─────────────────────────────────────────────┐
│ Redis Checkpointer 适合：                    │
│                                               │
│ - 短期 Agent Session                          │
│ - 高并发读（Redis 性能优势）                  │
│ - 状态丢失可接受（或者有其他持久化兜底）      │
│                                               │
│ Redis Checkpointer 不适合：                    │
│                                               │
│ - 需要长期保存 Workflow 状态                 │
│ - 复杂的 checkpoint 查询                      │
└─────────────────────────────────────────────┘
```

### Checkpointer 存什么

```
thread_id: "conversation-001"
  └─ checkpoint_1 (START 状态)
  └─ checkpoint_2 (Planner 完成)
  └─ checkpoint_3 (Executor 完成)
  └─ checkpoint_4 (Human-in-the-Loop 暂停)
  └─ write records (中断期间人工输入的额外数据)
```

**thread_id 是关键**——它把 "一次完整 Workflow" 的所有 checkpoint 串起来。

### PostgreSQL vs Redis 选型

| 维度 | PostgreSQL | Redis |
|------|-----------|-------|
| 长期保存 | ✅ 强 | ⚠️ 一般 |
| 并发读 | ⚠️ 一般 | ✅ 极强 |
| 复杂查询 | ✅ SQL | ❌ 弱 |
| 运维成本 | ⚠️ 中 | ✅ 低 |
| **适合场景** | **生产 Workflow 状态** | **短期 Session / 缓存层** |

---

## 3️⃣ Memory vs Checkpoint（面试高频区分）⭐⭐⭐

### 这是两个完全不同的概念

| 维度 | Checkpoint | Memory |
|------|-----------|--------|
| **存什么** | 当前 Workflow 的执行状态 | 用户的长期信息（user_name, job 等）|
| **存哪** | LangGraph Checkpointer（Postgres/Redis）| Memory Store / Vector DB |
| **生命周期** | Workflow 结束可以清理 | 长期保留（直到主动删除）|
| **谁读写** | LangGraph 框架自动 | Agent 业务代码主动 |
| **Schema** | LangGraph 框架定义 | 业务自己定义 |

### 三层数据分离（来自 ChatGPT 总结）

```
┌─────────────────────────────────────────────────┐
│ Layer 1: 当前 Workflow 数据                        │
│ → Checkpointer 管理                               │
│ → 存当前 State、节点输出                          │
│ → Workflow 结束可清理                             │
├─────────────────────────────────────────────────┤
│ Layer 2: Session 数据                              │
│ → 业务代码管理（ConversationBufferMemory 等）     │
│ → 短期上下文                                       │
├─────────────────────────────────────────────────┤
│ Layer 3: 长期 Memory                                │
│ → Memory Store 管理                                │
│ → 跨 Session 保留                                  │
│ → 用户画像、偏好、历史                             │
└─────────────────────────────────────────────────┘
```

### 常见错误

❌ 把 Checkpointer 当成 Memory 用
- Checkpoint 是 Workflow 临时状态，Workflow 结束就清
- Memory 是用户长期事实，要一直留

❌ 把 Memory 当成 Checkpointer 用
- Memory 不参与 Workflow 恢复
- 业务代码主动调用 Memory.Get/Set

❌ 三层数据混在一起
- 数据生命周期不同、存储策略不同
- 混在一起会导致：要么全部保留浪费空间，要么全部清理丢失信息

### 面试回答模板

> Agent 中有三层不同生命周期的数据：
> 1. **Checkpoint**：当前 Workflow 的执行状态，由 LangGraph Checkpointer 管理（Postgres/Redis），用于服务重启后的状态恢复
> 2. **Session**：单次对话的短期上下文，由业务代码管理
> 3. **Memory**：跨 Session 的长期用户事实，由 Memory Store 管理，用于记住用户偏好和历史
> 
> 三者**职责不同、生命周期不同、存储不同**，不能混用。

---

## 4️⃣ FastAPI 定位：HTTP 包装层，不是 Agent 框架

### 为什么 Agent 项目经常看到 FastAPI？

Agent 需要暴露成 HTTP 服务被其他系统调用：

```
┌──────────────────────────────────────────┐
│            Python Agent Service            │
│                                           │
│   ┌─────────────────────────────────┐    │
│   │       FastAPI (HTTP 层)          │    │
│   │                                   │    │
│   │   POST /chat                     │    │
│   │   POST /chat/{thread_id}/resume  │    │
│   │   GET  /chat/{thread_id}         │    │
│   └────────────┬─────────────────────┘    │
│                │                             │
│   ┌────────────▼─────────────────────┐    │
│   │  LangGraph Agent 核心             │    │
│   │  (与 HTTP 完全解耦)                │    │
│   └─────────────────────────────────┘    │
│                                           │
└──────────────────────────────────────────┘
```

### 关键认知

> **FastAPI 不是 Agent 框架**。
> 它只是把 Agent 核心逻辑包成 HTTP API。
> 你也可以用 Flask / gRPC / 消息队列 / CLI。

### 典型 FastAPI + LangGraph 集成

```python
from fastapi import FastAPI
from langgraph.graph import StateGraph
from langgraph.checkpoint.memory import InMemorySaver

app = FastAPI()

# Agent 核心（与 HTTP 解耦）
checkpointer = InMemorySaver()
graph = builder.compile(checkpointer=checkpointer)


@app.post("/chat")
async def chat(request: ChatRequest):
    config = {"configurable": {"thread_id": request.thread_id}}
    result = graph.invoke(
        {"messages": [HumanMessage(content=request.message)]},
        config=config,
    )
    return {
        "thread_id": request.thread_id,
        "result": result["messages"][-1].content,
    }


@app.post("/chat/{thread_id}/resume")
async def resume(thread_id: str, request: ResumeRequest):
    """HITL 恢复 API"""
    config = {"configurable": {"thread_id": thread_id}}
    result = graph.invoke(
        Command(resume=request.decision),
        config=config,
    )
    return {
        "thread_id": thread_id,
        "result": result,
    }
```

### 设计原则

| 层 | 职责 | 技术栈 |
|---|------|-------|
| HTTP 层 | 请求路由、参数校验、响应序列化 | FastAPI / Flask / gRPC |
| Agent 层 | 业务逻辑、Tool 调用、State 管理 | LangChain / LangGraph |
| 存储层 | State 持久化、Memory 存储 | PostgreSQL / Redis / Vector DB |

**三层解耦**：HTTP 层换 Flask 不影响 Agent，Agent 换 LangChain 不影响 HTTP。

---

## 5️⃣ 两级审批 Workflow（实战模式）

### 业务场景

```
用户提交退款 → 审核 1 → 审核 2 → 执行退款
                ↓           ↓
              任一失败 → 回到起点
```

### LangGraph 实现

```python
def should_continue_after_audit_1(state):
    if state["audit_1_passed"]:
        return "audit_2"
    return "restart"

def should_continue_after_audit_2(state):
    if state["audit_2_passed"]:
        return "execute_refund"
    return "restart"

builder.add_conditional_edges(
    "audit_1",
    should_continue_after_audit_1,
    {"audit_2": "audit_2", "restart": START}
)

builder.add_conditional_edges(
    "audit_2",
    should_continue_after_audit_2,
    {"execute_refund": "execute_refund", "restart": START}
)

builder.add_edge("execute_refund", END)
```

**关键区分**：
- **执行路径**回到起点 = 用 Edge 跳回 START
- **State 是否重置** = 另一个问题（可能保留 audit 1 通过的事实）

**生产提示**：真实审批系统会用 Checkpointer + thread_id 跟踪每次申请，避免重新填写。

---

## 6️⃣ 总结：LangGraph Demo → 生产

```
Demo 阶段                              生产阶段
─────────────────────────────────────────────────────
InMemorySaver                     →   PostgresSaver / RedisSaver
直接 graph.invoke()               →   FastAPI 封装成 HTTP API
Command(resume="yes")              →   POST /chat/{thread_id}/resume
单 Workflow 串                     →   thread_id 多用户隔离
单次审批                          →   多级审批 + Loop 回起点
错误打印 log                      →   错误回流 + 监控告警
没有评估                          →   LangSmith / LangFuse
本地 Mock 数据                     →   真实数据库 + 缓存层
```

---

## 7️⃣ 关键洞察

### 1. `interrupt()` 不是「卡住」，而是「返回 waiting 状态」

❌ 误解：`interrupt()` 让 Graph 卡住，要等人工
✅ 实际：`interrupt()` 把控制权交还给调用方，调用方可以返回 HTTP 响应给前端

**这是为什么 HITL 必须要 Checkpoint**——因为 Graph 暂停时是 Python 进程已经返回响应了，等前端审批完再重新调用 `graph.invoke(Command(resume=...))`，这时需要 Checkpoint 找回之前暂停的状态。

### 2. `thread_id` 是 HTTP / Workflow / Checkpoint 三者的桥梁

```
HTTP API 调用 → 传入 thread_id
   ↓
LangGraph Workflow 用 thread_id 找对应的 Checkpoint
   ↓
恢复 Graph 执行
   ↓
返回结果
```

**没有 thread_id，三个组件没法串起来**。

### 3. PostgreSQL Checkpointer 自动建表，不需要设计

**反直觉但正确**：Checkpoint 表结构由 LangGraph 版本决定，不要自己设计。
- ✅ 升级 LangGraph 时跟着 migration 走
- ❌ 自己设计 → 升级时 100% 冲突

### 4. FastAPI 的存在意义是「HTTP 入口」，不是「Agent 框架」

初学者误区：以为 FastAPI 是 Agent 的一部分
实际：FastAPI 只是 transport，跟 Agent 逻辑完全解耦

**判断标准**：
- 如果 Agent 只在内部用（脚本 / 批处理）→ 不需要 FastAPI
- 如果 Agent 要给外部系统调用 → 需要 FastAPI / gRPC / MQ

---

## ✅ 自检

- [ ] 1. HITL 中 `interrupt()` 实际做了什么？跟 Demo 中的"卡住"有什么区别？
- [ ] 2. 为什么 HITL 必须配合 Checkpoint？
- [ ] 3. PostgreSQL 和 Redis Checkpointer 各自适合什么场景？
- [ ] 4. Memory 和 Checkpoint 的本质区别是什么？
- [ ] 5. FastAPI 在 Agent 项目里的角色是什么？
- [ ] 6. 两级审批 Workflow 在 LangGraph 里怎么实现？

---

## 📂 相关代码参考

- `langchain-agent-demo/15_checkpoint/` - Checkpointer 基础（InMemorySaver）
- `langchain-agent-demo/16_human_in_loop/` - HITL 基础（interrupt + Command）

---

## 🔗 相关笔记

- [LangChain + LangGraph 基础](./langchain-langgraph.md) - 本笔记的配套基础
- [Memory 完整体系](./memory.md) - Memory 跟 Checkpoint 的对比
- [Context Management](./context-management.md) - Go 中的 ContextManager 等价物
- [Router 详解](./router.md) - 能力路由（在生产 Agent 中也需要）

---

## 📅 下一步主线：Agentic RAG

根据 ChatGPT 分享，下一阶段主线是 **Agentic RAG**：
- 之前学的 RAG（08）是「文档 → 知识库 → 检索 → LLM」
- Agentic RAG 是「Agent 决定要不要查 RAG、查什么、查完怎么用」
- 会涉及：LangGraph + RAG + Knowledge Base + 智能路由

# Agent 工具安全 + 可靠性：Permission / HITL / Idempotency

> 📚 配套笔记：[langgraph-production-extensions.md](./langgraph-production-extensions.md)（HITL 基础）/ [rag.md](./rag.md)（RAG 基础）
> 🗓️ 时间：今天（基于 ChatGPT 分享 [6ab340ea](https://chatgpt.com/share/6ab340ea-f644-83e8-aa26-add0f336b6a9) + `langchain-agent-demo/31-38`）
> 🎯 主题：企业 Agent 的**安全 + 可靠性**完整链路

> 💡 **本笔记覆盖**：
> 1. Tool Permission（用户级权限控制）
> 2. Tool Security（普通操作 vs 高风险 + HITL 集成）
> 3. Retry + Idempotency（可靠性 + Redis）
> 4. Multi-Agent 概念 + Subgraph
> 5. 企业 Agent 完整安全链

---

## 🎯 核心问题

> 企业 Agent 不能让 LLM 自己决定所有事：
> - **谁能调用哪个 Tool**？→ 不是 **LLM 决定**
> - **高风险操作要不要审批**？→ 不是 **LLM 决定**
> - **网络断了要重试吗？** → 不会双扣事务
> 
> 这些都是**传统后端能力**接入 Agent 的问题。

---

## 1️⃣ Tool Permission（用户级权限控制）

### 核心洞察（来自 ChatGPT）

> **能调用一个 Tool，不代表当前用户就有权限调用这个 Tool。**
> 
> **不能依赖 LLM 自己判断权限。** 必须由**程序控制**。

### 为什么 LLM 不能管权限？

| 问题 | 说明 |
|------|------|
| **Prompt Injection** | 攻击者可能通过输入绕过 LLM 权限判断 |
| **黑盒决策** | LLM 决策不可解释、不可审计 |
| **不一致** | 同样的输入，LLM 可能给出不同判断 |
| **合规要求** | 金融、医疗等场景要求**显式权限日志** |

### 实现（Demo 35，441 行）

```python
# 1. 模拟当前用户
current_user = {
    "user_id": "user1001",
    "username": "张三",
    "roles": ["USER"],
    "permissions": [
        "order:query",
        "order:create",
        # 注意：没有 "order:refund"
    ],
}

# 2. 权限检查函数（自己定义，不是 LangChain API）
def has_permission(permission: str) -> bool:
    return permission in current_user["permissions"]

def require_permission(permission: str):
    if not has_permission(permission):
        raise ToolException(f"没有权限执行操作：{permission}")

# 3. Tool 内调用权限检查
@tool("refund_order", description="申请订单退款。需要 order:refund 权限。")
def refund_order(order_id: str) -> dict:
    require_permission("order:refund")  # ← 在 Tool 里强制检查
    return {"order_id": order_id, "status": "已退款"}
```

### 真实企业链路

```
用户登录
   ↓
JWT / Session
   ↓
UserContext（含 roles / permissions）
   ↓
Tool Permission 检查
   ↓
业务逻辑
```

**关键**：UserContext 从 Session 来，**不在 Prompt 里**。

### 跟 Go 后端的关系

```
传统 Web 后端        Agent Tool
─────────────────    ──────────
HTTP Request    →   Tool Call
Session User    →   current_user
Role / Permission → Permission Check
Audit Log       →   Tool Execution Log
```

**本质相同**，只是把后端能力**接到 Tool 执行链路上**。

---

## 2️⃣ Tool Security（普通操作 vs 高风险）

### 风险分层（来自 ChatGPT）

```
┌──────────────────────────────────────────────────┐
│  普通操作（低风险）                                  │
│  - query_order（查询订单）                          │
│  - get_user（查询用户）                            │
│  - search_knowledge（搜索知识库）                   │
│  → 直接执行，无需审批                                │
├──────────────────────────────────────────────────┤
│  高风险操作                                         │
│  - refund_order（退款）                            │
│  - update_salary（改工资）                            │
│  - delete_account（删账号）                          │
│  - execute_sql（执行 SQL）                           │
│  → 需要 Human Approval（HITL）                       │
└──────────────────────────────────────────────────┘
```

### 三层安全链（Demo 36-38）

```
用户请求
   ↓
[1] Permission Check        ← 你有权调这个 Tool 吗？
   ↓  (没权限 → 直接拒绝)
[2] Risk Classification     ← 这是高风险操作吗？
   ↓  (低风险 → 直接执行)
[3] HITL Approval          ← 高风险 → 人工确认
   ↓  (审批通过 → 执行)
[4] Tool Execution + Audit Log
```

### HITL 集成（Demo 36-37）

```python
# Demo 36: HITL 在 LangGraph 里
from langgraph.types import interrupt, Command

def refund_node(state: ApprovalState):
    # 检测是否是高风险操作
    if state["amount"] > 10000:
        # 暂停等审批
        approval = interrupt({
            "question": f"退款 {state['amount']} 元是否批准？",
            "order_id": state["order_id"],
        })
        if approval != "approve":
            return {"status": "cancelled"}
    
    # 真正执行
    return {"status": "refunded"}
```

```python
# Demo 37: HITL 在 create_agent 里（LangChain 1.x 新方式）
from langchain.agents.middleware import HumanInTheLoopMiddleware

agent = create_agent(
    model=model,
    tools=[refund_order],
    middleware=[
        HumanInTheLoopMiddleware(
            tool_config={
                "refund_order": {"require_approval": True},
            },
        ),
    ],
)
```

**LangChain 1.x 新增的 Middleware 机制**，把 HITL 配置化：
- 不需要手写 interrupt()
- 通过 `tool_config` 声明哪些 Tool 需要审批

---

## 3️⃣ Retry + Idempotency（可靠性）⭐

### 核心问题

```
Agent 调用 Tool → 网络异常 → 重试？
                ↓
              重试 → 业务被执行两次（双扣钱、多次下单）
              不重试 → 用户体验差
```

**Idempotency** = 多次调用结果相同。

### 实现（Demo 34，515 行）

```python
import redis
import uuid

redis_client = redis.Redis(host="localhost", port=6379, db=0)

@tool("create_order", args_schema=CreateOrderInput)
def create_order(request_id: str, amount: int) -> dict:
    """
    创建订单。
    request_id 用于保证请求幂等。
    Retry 时必须保持不变。
    """
    
    # 1. Redis SETNX 检查 + 设置
    # 等价于 SET if not exists
    is_first = redis_client.set(
        f"order:idempotency:{request_id}",
        "processing",
        nx=True,        # ← 只在 Key 不存在时设置
        ex=60,          # ← 60 秒过期
    )
    
    if not is_first:
        # 已经处理过
        existing = redis_client.get(f"order:idempotency:{request_id}")
        if existing == "completed":
            return {"status": "already_created", "request_id": request_id}
        else:
            raise ToolException(f"请求 {request_id} 正在处理中")
    
    # 2. 真正执行业务逻辑
    try:
        order_id = "10001"
        # ... 写数据库 ...
        
        redis_client.set(
            f"order:idempotency:{request_id}",
            "completed",
            ex=3600,  # 完成后保留 1 小时
        )
        return {"order_id": order_id, "amount": amount}
    except Exception as e:
        redis_client.delete(f"order:idempotency:{request_id}")
        raise
```

### Redis SETNX 是 Idempotency 的经典实现

```
请求 1 进来 → SETNX idempotency:req001 = processing
   ↓
请求 2 进来（重试） → SETNX 失败（Key 已存在）
   ↓
请求 2 查 Status → 已处理 / 处理中
   ↓
避免双扣
```

**关键**：
- `request_id` 由客户端生成，每次重试必须保持一致
- Redis TTL 要合理（过短可能 Key 提前过期，过长浪费内存）

### 跟 Go 后端的关系

这就是**传统幂等设计**：
- Web 后端防止表单重复提交
- 支付系统防止双扣款
- 消息队列防止重复消费

**Agent Tool 的 Idempot** = 同一思路，只是载体换成 Tool Call。

---

## 4️⃣ Multi-Agent 完整思路

### 什么时候需要 Multi-Agent？（来自 ChatGPT）

```
单 Agent 不够的场景：
├ 任务复杂，一个 Agent 管不了
├ 不同任务需要不同专长（如「订单 Agent」+「物流 Agent」+「支付 Agent」）
├ 需要并行处理多个独立子任务
└ 需要在 Agent 间隔离状态/权限

单 Agent 够用的场景：
├ 单一任务（如客服、查询）
├ 工具集不多（< 10 个）
└ 不需要并行
```

### Multi-Agent vs Multi-Tool（关键区分）

| 维度 | Multi-Tool | Multi-Agent |
|------|-----------|------------|
| **决策中心** | 1 个 LLM 决定调哪些 Tool | 每个 Agent 独立决策 |
| **状态** | 共享同一 State | 每个 Agent 独立 State |
| **专长** | 所有 Tool 共享一个 System Prompt | 每个 Agent 有自己的 Prompt |
| **适用** | 任务相对单一 | 任务复杂、需要分工 |

**最常见的误区**：以为加几个 Tool 就是 Multi-Agent。

### Multi-Agent 两种架构

**Supervisor 模式**：
```
         Supervisor Agent
        /       |       \
   Order Agent  Logistics  Payment
   (查询订单)    (查物流)   (查支付)
```

Supervisor 决定调哪个 Worker。

**Swarm 模式**（对等）：
```
Order Agent ←→ Logistics Agent
     ↕                ↕
  Payment Agent ←→ User Agent
```

Agent 之间互相通信，没有中央调度。

### Subgraph（LangGraph 概念）

```
┌──────────────────────────────────┐
│  主 Graph                        │
│   ┌──────────────────┐           │
│   │   Subgraph       │           │
│   │  (Agent A)       │           │
│   │  ┌────┐ ┌────┐   │           │
│   │  │ N1 │→│ N2 │   │           │
│   │  └────┘ └────┘   │           │
│   └──────────────────┘           │
│   ┌──────────────────┐           │
│   │   Subgraph       │           │
│   │  (Agent B)       │           │
│   └──────────────────┘           │
└──────────────────────────────────┘
```

**关键洞察**：一个 LangGraph Workflow 本身可以成为另一个 Workflow 的 Node。

**企业价值**：
- 模块化（每个 Subgraph 一个团队维护）
- 复用（同一 Subgraph 在多个主 Graph 里用）
- 可观测（每个 Subgraph 单独的 trace）

### ChatGPT 给的整体 Roadmap

```
你现在的覆盖度：约 70%

不要再继续深入 LangGraph 高级 API

下一阶段：
├ Multi-Agent（Subgraph + Supervisor）
├ 可观测性（LangSmith / LangFuse）
└ Evaluation（Trajectory / Outcome）

企业 Tool 安全：
├ Tool Permission
├ Tool Security（HITL 集成）
└ Retry + Idempotency

不要学：
├ 各种 Vector DB 都跑一遍
├ 复杂 Agent-to-Agent 协议
├ Multi-Agent 底层通信
└ LangChain 源码细节
```

---

## 5️⃣ 完整企业 Agent 安全链（综合 Demo 34-38）

```
                            用户请求
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 1] Authentication / Session                    │
│   - JWT / Cookie 解析                                │
│   - current_user = {user_id, roles, permissions}    │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 2] Tool Permission Check                       │
│   - require_permission("order:refund")               │
│   - 没权限 → 拒绝                                    │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 3] Risk Classification                         │
│   - 普通操作 → 直接执行                              │
│   - 高风险 → HITL Approval                          │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 4] Idempotency (request_id + Redis)            │
│   - SETNX 检查                                       │
│   - 已处理 → 返回上次                                │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 5] Retry + Circuit Breaker                     │
│   - 网络异常 → Retry                            │
│   - 多次失败 → 熔断                                  │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 6] Tool Execution + Audit Log                  │
│   - 执行业务                                         │
│   - 记录谁、什么时候、调了什么、结果                  │
└─────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────┐
│ [Layer 7] Observability (LangSmith / LangFuse)        │
│   - Trace 整条链路                                   │
│   - 监控 Token / 延迟 / 错误                         │
└─────────────────────────────────────────────────────┘
```

**每一层都是后端工程师的老本行**，只是叠加在 Agent 上。

---

## 6️⃣ 关键洞察

### 1. Tool Permission 必须由程序控制

❌ 让 LLM 决定「我能不能调这个 Tool」
✅ 程序检查 `current_user.permissions` 是否包含所需权限

**理由**：合规、可审计、不可被 Prompt Injection 绕过。

### 2. HITL 集成有两条路线

| 路线 | 适用 |
|------|------|
| LangGraph 底层（手写 interrupt）| 复杂 Workflow |
| LangChain 1.x Middleware | 标准 create_agent 场景 |

### 3. Retry 必须配合 Idempotency

❌ 网络异常 → 直接重试（可能双扣）
✅ 网络异常 → 重试 + request_id 不变 + Redis 幂等

### 4. Multi-Agent ≠ Multi-Tool

**最大误区**：以为加了几个 Tool 就是 Multi-Agent。

**真正的 Multi-Agent**：
- 每个 Agent 独立 State
- 每个 Agent 有自己的 System Prompt / 专长
- Agent 间需要通信协议

### 5. Subgraph 是企业 Multi-Agent 的工程实现

一个 LangGraph Workflow = 一个 Node = 一个 Sub-Agent。
可复用、可独立测试、可观测。

### 6. 企业 Agent 的核心 = 传统后端 + Agent 思维

| 传统后端 | Agent |
|---------|-------|
| Session / JWT | current_user |
| Role / RBAC | Permission Check |
| Audit Log | Tool Execution Log |
| Retry / Circuit Breaker | Idempotency |
| Trace | LangSmith trace |

**后端经验是你的护城河**。

---

## 7️⃣ 面试角度：怎么聊 Tool 安全 + 可靠性

### 30 秒版本

> 企业 Agent 的 Tool 安全链分 7 层：Authentication → Permission → Risk Classification → HITL → Idempotency → Retry → Audit。每一层都是传统后端能力，只是叠加在 Tool 调用上。LLM 不管权限、不管高风险审批，只负责意图理解和工具决策。

### 5 分钟版本

按 ChatGPT 给的 4 个真实场景讲：

**场景 1：客服 Agent**
- query_order / get_user：普通用户可调
- refund_order：需要「order:refund」权限
- 大额退款（> 10000）：HITL 审批

**场景 2：内部 IT Agent**
- 普通员工：只能看自己机器
- 运维工程师：可重启服务
- 修改生产配置：HITL 审批 + 二次确认

**场景 3：数据 Agent**
- 权限控制 Tool 还不够
- 还要控制 Tool 返回的数据（用户只能查自己的数据）

### 加分项

- 提 Idempotency + Redis SETNX 的经典实现
- 提 Middleware 抽象 HITL（vs 手写 interrupt）
- 提 Subgraph 是 Multi-Agent 的工程落地

---

## ✅ 自检

- [ ] 1. 为什么 Tool Permission 必须由程序控制，而不是 LLM？
- [ ] 2. 普通操作和高风险操作的分层是怎么做的？
- [ ] 3. HITL 在 create_agent 中如何用 Middleware 抽象？
- [ ] 4. Idempotency 为什么必须配合 Retry？经典实现是什么？
- [ ] 5. Multi-Agent 和 Multi-Tool 的本质区别是什么？
- [ ] 6. Subgraph 在 Multi-Agent 里是什么角色？
- [ ] 7. 企业 Agent 安全链分几层？哪一层跟你的 Go 后端经验最相关？

---

## 📂 相关代码

- `langchain-agent-demo/31_rag_hybrid_search/` - Hybrid Search
- `langchain-agent-demo/32_rag_query_rewrite/` - Query Rewrite
- `langchain-agent-demo/33_mcp/` - LangChain + MCP
- `langchain-agent-demo/34_tool_design/` - Tool 设计 + Idempotency + Redis (515 行)
- `langchain-agent-demo/35_tool_permission/` - Tool 权限 (441 行)
- `langchain-agent-demo/36_tool_security/02_human_approval/` - HITL 集成
- `langchain-agent-demo/37_agent_hitl/` - Agent + HITL (361 行)
- `langchain-agent-demo/38_agent_workflow/` - 完整企业 Workflow (523 行)

**总计 8 个新增 · ~3000 行 Python**

---

## 🔗 相关笔记

- [LangGraph 生产扩展](./langgraph-production-extensions.md) - HITL / Checkpointer / FastAPI 基础
- [Agent 架构总览](./agent-architecture.md) - 整体架构视角
- [Memory 完整体系](./memory.md) - 跟 Tool Permission 不同的权限维度

---

## 📅 接下来的优先级（按 ChatGPT 路线）

按 ChatGPT 的整体 Roadmap：

1. **Multi-Agent 实战**（Subgraph + Supervisor）⭐⭐⭐
2. **可观测性**（LangSmith / LangFuse 接入）⭐⭐⭐
3. **Evaluation**（Trajectory Eval / Outcome Eval）⭐⭐⭐

企业面试高频题，也是企业 Agent 落地必备能力。

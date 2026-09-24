import os
from typing import TypedDict

from dotenv import load_dotenv

from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, START, END
from langgraph.checkpoint.memory import InMemorySaver
from langgraph.types import interrupt, Command


# ============================================================
# 1. 初始化 LLM
# ============================================================
# 本 Demo 暂时不会真正调用 LLM。
# 保留统一的项目初始化方式，后面接 Agent 时继续使用。

load_dotenv()

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 2. Graph State
# ============================================================
# LangGraph 官方 State
#
# 保存整个 Workflow 当前状态。

class ApprovalState(TypedDict):
    order_id: str
    amount: int
    status: str
    approval_result: str


# ============================================================
# 3. 退款申请节点
# ============================================================
# 用户自定义业务逻辑
#
# 这里模拟：
#
# Agent 判断需要退款
#       ↓
# 创建退款申请
#       ↓
# 进入人工审批
#

def create_refund_request(state: ApprovalState):

    print("\n[Business]")
    print(
        f"创建退款申请："
        f"order={state['order_id']}, "
        f"amount={state['amount']}"
    )

    return {
        "status": "PENDING_APPROVAL"
    }


# ============================================================
# 4. Human Approval Node
# ============================================================
# LangGraph 官方 interrupt()
#
# 这里是整个 Demo 最重要的地方。
#
# interrupt() 会：
#
# 1. 暂停 Graph
# 2. 通过 Checkpointer 保存状态
# 3. 把审批信息返回给调用方
# 4. 等待 Command(resume=...)
#

def human_approval(state: ApprovalState):

    print("\n[HITL]")
    print("准备请求人工审批...")

    decision = interrupt(
        {
            "type": "refund_approval",
            "question": "是否批准这笔退款？",
            "order_id": state["order_id"],
            "amount": state["amount"],
        }
    )

    print("\n[HITL]")
    print(f"收到人工审批结果：{decision}")

    if decision == "approve":
        return {
            "approval_result": "APPROVED",
            "status": "APPROVED",
        }

    return {
        "approval_result": "REJECTED",
        "status": "REJECTED",
    }


# ============================================================
# 5. 真正执行退款
# ============================================================
# 用户自定义业务逻辑
#
# 只有审批通过以后，才会执行。

def execute_refund(state: ApprovalState):

    print("\n[Business]")
    print(
        f"真正执行退款："
        f"order={state['order_id']}, "
        f"amount={state['amount']}"
    )

    return {
        "status": "REFUND_SUCCESS"
    }


# ============================================================
# 6. 拒绝退款
# ============================================================

def reject_refund(state: ApprovalState):

    print("\n[Business]")
    print("退款申请被拒绝")

    return {
        "status": "REFUND_REJECTED"
    }


# ============================================================
# 7. 构建 Graph
# ============================================================

builder = StateGraph(ApprovalState)

builder.add_node(
    "create_refund_request",
    create_refund_request,
)

builder.add_node(
    "human_approval",
    human_approval,
)

builder.add_node(
    "execute_refund",
    execute_refund,
)

builder.add_node(
    "reject_refund",
    reject_refund,
)


# ============================================================
# 8. Graph 路由
# ============================================================

builder.add_edge(
    START,
    "create_refund_request",
)

builder.add_edge(
    "create_refund_request",
    "human_approval",
)


def route_after_approval(state: ApprovalState):

    if state["approval_result"] == "APPROVED":
        return "execute_refund"

    return "reject_refund"


builder.add_conditional_edges(
    "human_approval",
    route_after_approval,
)

builder.add_edge(
    "execute_refund",
    END,
)

builder.add_edge(
    "reject_refund",
    END,
)


# ============================================================
# 9. Checkpointer
# ============================================================
# LangGraph 官方能力
#
# InMemorySaver：
#     教学 / Demo 使用
#
# 生产环境：
#     使用持久化 Checkpointer
#     例如 PostgreSQL 等。

checkpointer = InMemorySaver()

graph = builder.compile(
    checkpointer=checkpointer
)


# ============================================================
# 10. thread_id
# ============================================================
# thread_id：
# 标识这一条 Agent / Workflow 状态链。
#
# interrupt 后必须使用同一个 thread_id 恢复。

config = {
    "configurable": {
        "thread_id": "refund-10001"
    }
}


# ============================================================
# 11. 第一次执行
# ============================================================

print("=" * 70)
print("第一次执行")
print("=" * 70)

result = graph.invoke(
    {
        "order_id": "10001",
        "amount": 50000,
        "status": "NEW",
        "approval_result": "",
    },
    config=config,
)


# ============================================================
# 12. 检查是否进入 interrupt
# ============================================================

print("\n========== Graph Result ==========")

print(result)


if "__interrupt__" in result:

    print("\n========== Human Approval Required ==========")

    interrupt_info = result["__interrupt__"][0]

    print("审批信息：")
    print(interrupt_info.value)

    print("\nGraph 已暂停，等待人工确认。")


# ============================================================
# 13. 模拟人工操作
# ============================================================
#
# 真实项目这里不是 input()。
#
# 通常可能是：
#
# Web UI
# 管理后台
# 企业微信
# 钉钉
# 审批系统
# 人工客服系统
#
# 这里为了 Demo 简单模拟：
#

human_decision = "approve"

print("\n")
print("=" * 70)
print(f"人工审批：{human_decision}")
print("=" * 70)


# ============================================================
# 14. 恢复 Graph
# ============================================================
#
# 非常重要：
#
# 使用同一个 config
# 使用同一个 thread_id
#
# Command(resume=...)
# 会把结果传回 interrupt()

result = graph.invoke(
    Command(
        resume=human_decision
    ),
    config=config,
)


# ============================================================
# 15. 最终结果
# ============================================================

print("\n========== Final Result ==========")

print(result)

print("\n最终状态：")
print(result["status"])
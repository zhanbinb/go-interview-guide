from typing import TypedDict

from langgraph.graph import (
    StateGraph,
    START,
    END,
)

from langgraph.checkpoint.memory import InMemorySaver

from langgraph.types import (
    interrupt,
    Command,
)


# ============================================================
# 1. Workflow State
# ============================================================
# LangGraph 官方 State
#
# State 就是整个 Workflow 在执行过程中携带的数据。
#
# 可以理解成：
#
#     Workflow Context
#
# 每一个 Node：
#
#     读取 State
#     ↓
#     执行业务逻辑
#     ↓
#     修改 State
#

class RefundState(TypedDict):

    # 用户请求
    order_id: str
    amount: int

    # 订单信息
    order_status: str
    order_amount: int

    # 风险判断
    risk_level: str

    # 审批结果
    approval_result: str

    # 最终结果
    result: str


# ============================================================
# 2. Node：查询订单
# ============================================================
# 用户自定义业务逻辑
#
# 真实项目这里可能调用：
#
#     Order Service
#     REST API
#     gRPC
#     Database
#

def query_order(state: RefundState):

    print("\n[Node] query_order")

    order_id = state["order_id"]

    print(
        f"查询订单：{order_id}"
    )

    # 模拟订单服务返回
    return {
        "order_status": "PAID",
        "order_amount": 50000,
    }


# ============================================================
# 3. Node：检查退款请求
# ============================================================
# 用户自定义业务逻辑
#
# 判断：
#
#     用户申请退款金额
#     是否超过订单金额
#

def validate_refund(state: RefundState):

    print("\n[Node] validate_refund")

    request_amount = state["amount"]
    order_amount = state["order_amount"]

    print(
        f"申请退款：{request_amount}"
    )

    print(
        f"订单金额：{order_amount}"
    )

    if request_amount <= 0:

        return {
            "result": "退款金额必须大于 0"
        }

    if request_amount > order_amount:

        return {
            "result": "退款金额不能超过订单金额"
        }

    return {
        "result": ""
    }


# ============================================================
# 4. Node：风险判断
# ============================================================
# 用户自定义业务规则
#
# 企业真实项目可能来自：
#
#     风控服务
#     Policy Engine
#     配置中心
#     数据库
#

RISK_THRESHOLD = 5000


def check_risk(state: RefundState):

    print("\n[Node] check_risk")

    amount = state["amount"]

    if amount > RISK_THRESHOLD:

        print(
            f"退款金额 {amount} "
            f"> {RISK_THRESHOLD}"
        )

        return {
            "risk_level": "HIGH"
        }

    print("普通退款")

    return {
        "risk_level": "NORMAL"
    }


# ============================================================
# 5. Node：人工审批
# ============================================================
# LangGraph 官方 interrupt()
#
# 只有 HIGH 风险才会进入这里。
#

def human_approval(state: RefundState):

    print("\n[Node] human_approval")

    decision = interrupt(
        {
            "type": "refund_approval",

            "order_id": state["order_id"],

            "amount": state["amount"],

            "message": (
                "该退款属于高风险操作，"
                "是否批准？"
            ),
        }
    )

    print(
        f"人工审批结果：{decision}"
    )

    return {
        "approval_result": decision
    }


# ============================================================
# 6. Node：执行退款
# ============================================================
# 用户自定义业务逻辑
#
# 真实项目这里可能：
#
#     调用 Payment Service
#     写数据库
#     发布 MQ Event
#     写审计日志
#

def execute_refund(state: RefundState):

    print("\n[Node] execute_refund")

    print(
        f"执行退款："
        f"order={state['order_id']}, "
        f"amount={state['amount']}"
    )

    return {
        "result": (
            f"订单 {state['order_id']} "
            f"退款 {state['amount']} 元成功"
        )
    }


# ============================================================
# 7. Node：拒绝退款
# ============================================================

def reject_refund(state: RefundState):

    print("\n[Node] reject_refund")

    return {
        "result": (
            f"订单 {state['order_id']} "
            f"退款申请被拒绝"
        )
    }


# ============================================================
# 8. Node：失败
# ============================================================

def refund_failed(state: RefundState):

    print("\n[Node] refund_failed")

    return {
        "result": state["result"]
    }


# ============================================================
# 9. 构建 Workflow
# ============================================================

builder = StateGraph(RefundState)


# 注册 Nodes

builder.add_node(
    "query_order",
    query_order,
)

builder.add_node(
    "validate_refund",
    validate_refund,
)

builder.add_node(
    "check_risk",
    check_risk,
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

builder.add_node(
    "refund_failed",
    refund_failed,
)


# ============================================================
# 10. Workflow 主流程
# ============================================================

builder.add_edge(
    START,
    "query_order",
)

builder.add_edge(
    "query_order",
    "validate_refund",
)


# ============================================================
# 11. validate_refund 路由
# ============================================================

def route_after_validation(
    state: RefundState
):

    if state["result"]:

        return "refund_failed"

    return "check_risk"


builder.add_conditional_edges(
    "validate_refund",
    route_after_validation,
)


# ============================================================
# 12. 风险路由
# ============================================================

def route_after_risk(
    state: RefundState
):

    if state["risk_level"] == "HIGH":

        return "human_approval"

    return "execute_refund"


builder.add_conditional_edges(
    "check_risk",
    route_after_risk,
)


# ============================================================
# 13. 人工审批路由
# ============================================================

def route_after_approval(
    state: RefundState
):

    if state["approval_result"] == "approve":

        return "execute_refund"

    return "reject_refund"


builder.add_conditional_edges(
    "human_approval",
    route_after_approval,
)


# ============================================================
# 14. 结束节点
# ============================================================

builder.add_edge(
    "execute_refund",
    END,
)

builder.add_edge(
    "reject_refund",
    END,
)

builder.add_edge(
    "refund_failed",
    END,
)


# ============================================================
# 15. Checkpointer
# ============================================================

checkpointer = InMemorySaver()


graph = builder.compile(
    checkpointer=checkpointer
)


# ============================================================
# 16. 第一次执行
# ============================================================

config = {
    "configurable": {
        "thread_id": "refund-workflow-10001"
    }
}


print("=" * 70)
print("第一次执行 Workflow")
print("=" * 70)


result = graph.invoke(
    {
        "order_id": "10001",

        "amount": 50000,

        "order_status": "",

        "order_amount": 0,

        "risk_level": "",

        "approval_result": "",

        "result": "",
    },

    config=config,
)


# ============================================================
# 17. 检查是否暂停
# ============================================================

print("\n")
print("=" * 70)
print("第一次执行结果")
print("=" * 70)

print(result)


if "__interrupt__" in result:

    print("\n")
    print("=" * 70)
    print("Workflow 暂停")
    print("=" * 70)

    for interrupt_info in result["__interrupt__"]:

        print(
            interrupt_info.value
        )


# ============================================================
# 18. 模拟人工审批
# ============================================================

print("\n")
print("=" * 70)
print("人工审批")
print("=" * 70)

decision = "approve"

print(
    f"人工决定：{decision}"
)


# ============================================================
# 19. 恢复 Workflow
# ============================================================

result = graph.invoke(
    Command(
        resume=decision
    ),

    config=config,
)


# ============================================================
# 20. 最终结果
# ============================================================

print("\n")
print("=" * 70)
print("Workflow 最终结果")
print("=" * 70)

print(
    result["result"]
)
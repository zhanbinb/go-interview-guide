import os

from dotenv import load_dotenv

from langchain_openai import ChatOpenAI
from langchain_core.tools import tool

from langchain.agents import create_agent
from langchain.agents.middleware import HumanInTheLoopMiddleware

from langgraph.checkpoint.memory import InMemorySaver
from langgraph.types import Command


# ============================================================
# 1. 初始化 LLM
# LangChain 官方 API
# ============================================================

load_dotenv()

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 2. 模拟当前用户
# 用户自定义业务数据
# ============================================================

current_user = {
    "user_id": "user1001",
    "username": "张三",
    "roles": ["CUSTOMER_SERVICE"],
    "permissions": [
        "order:query",
        "order:refund",
    ],
}


# ============================================================
# 3. 权限检查
# 用户自定义业务逻辑
# ============================================================

def require_permission(permission: str):

    if permission not in current_user["permissions"]:
        raise PermissionError(
            f"没有权限执行操作：{permission}"
        )


# ============================================================
# 4. 查询订单 Tool
# LangChain 官方 @tool
# ============================================================

@tool(
    "query_order",
    description=(
        "查询订单信息。"
        "当用户需要查询订单状态、金额、用户等信息时使用。"
        "需要提供订单号。"
    ),
)
def query_order(order_id: str) -> dict:

    require_permission("order:query")

    print(
        f"\n[Tool] query_order("
        f"order_id={order_id})"
    )

    return {
        "order_id": order_id,
        "user": "张三",
        "amount": 3999,
        "status": "已支付",
    }


# ============================================================
# 5. 退款 Tool
# LangChain 官方 @tool
#
# 注意：
# Tool 本身不负责 Human Approval。
#
# Human Approval 由：
#
# HumanInTheLoopMiddleware
#
# 在 Tool 真正执行之前拦截。
# ============================================================

@tool(
    "refund_order",
    description=(
        "执行订单退款。"
        "需要提供订单号和退款金额。"
        "这是高风险业务操作，执行前需要人工确认。"
    ),
)
def refund_order(
    order_id: str,
    amount: int,
) -> dict:

    require_permission("order:refund")

    print(
        f"\n[Business] 真正执行退款："
        f"order={order_id}, "
        f"amount={amount}"
    )

    return {
        "status": "SUCCESS",
        "order_id": order_id,
        "amount": amount,
        "message": "退款成功",
    }


# ============================================================
# 6. Checkpointer
# LangGraph 官方能力
#
# HITL 必须保存 Agent 状态。
#
# Demo：
#     InMemorySaver
#
# Production：
#     PostgreSQL / 其他持久化 Checkpointer
# ============================================================

checkpointer = InMemorySaver()


# ============================================================
# 7. Human-in-the-loop Middleware
# LangChain 官方 API
# ============================================================

hitl_middleware = HumanInTheLoopMiddleware(
    interrupt_on={
        # 查询订单：
        # 不需要人工审批
        "query_order": False,

        # 退款订单：
        # 必须人工审批
        "refund_order": {
            "allowed_decisions": [
                "approve",
                "reject",
            ],
            "description": (
                "这是一个退款操作，"
                "请确认是否允许执行。"
            ),
        },
    },

    description_prefix="Tool execution pending approval",
)


# ============================================================
# 8. 创建 Agent
# LangChain 官方 API
# ============================================================

agent = create_agent(
    model=model,

    tools=[
        query_order,
        refund_order,
    ],

    middleware=[
        hitl_middleware,
    ],

    checkpointer=checkpointer,
)


# ============================================================
# 9. 打印 Agent Trace
# 用户自定义辅助函数
# ============================================================

def print_trace(result):

    print("\n")
    print("=" * 70)
    print("Agent Trace")
    print("=" * 70)

    for message in result["messages"]:

        print("--------------------------------")

        print(
            f"Message Type: "
            f"{type(message).__name__}"
        )

        if message.content:

            print("Content:")
            print(message.content)

        if getattr(message, "tool_calls", None):

            print("Tool Calls:")

            for tool_call in message.tool_calls:

                print(tool_call)


# ============================================================
# 10. 第一次调用
# ============================================================

thread_id = "conversation-refund-10001"

config = {
    "configurable": {
        "thread_id": thread_id
    }
}


print("=" * 70)
print("Step 1：用户发起请求")
print("=" * 70)

result = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": (
                    "帮我把订单10001退款50000元"
                ),
            }
        ]
    },
    config=config,
)


# ============================================================
# 11. 查看第一次执行结果
# ============================================================

print_trace(result)


# ============================================================
# 12. 检查 HITL interrupt
# ============================================================

if "__interrupt__" in result:

    print("\n")
    print("=" * 70)
    print("Step 2：Agent 暂停，等待人工审批")
    print("=" * 70)

    for interrupt_info in result["__interrupt__"]:

        print("\nInterrupt Value:")

        print(interrupt_info.value)


# ============================================================
# 13. 模拟人工审批
#
# 真实项目：
#
#     前端审批页面
#     管理后台
#     企业微信
#     钉钉
#     审批系统
#
# 这里直接模拟：
# ============================================================

human_decision = {
    "type": "approve"
}


print("\n")
print("=" * 70)
print("Step 3：人工审批")
print("=" * 70)

print(
    f"人工决定：{human_decision}"
)


# ============================================================
# 14. 恢复 Agent
#
# 注意：
#
# 必须：
#
#     Command(resume=...)
#
#     + 相同 thread_id
#
# ============================================================

result = agent.invoke(
    Command(
        resume={
            "decisions": [
                {
                    "type": "approve"
                }
            ]
        }
    ),
    config=config,
)


# ============================================================
# 15. 最终结果
# ============================================================

print("\n")
print("=" * 70)
print("Step 4：Agent 恢复完成")
print("=" * 70)

print_trace(result)

print("\n")
print("=" * 70)
print("最终 Agent 回复")
print("=" * 70)

print(result["messages"][-1].content)
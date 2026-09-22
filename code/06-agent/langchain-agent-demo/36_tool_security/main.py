import os
from dotenv import load_dotenv

from langchain_openai import ChatOpenAI
from langchain_core.tools import tool, ToolException
from langchain.agents import create_agent


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
# 2. 当前用户
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

def has_permission(permission: str) -> bool:
    return permission in current_user["permissions"]


def require_permission(permission: str):
    if not has_permission(permission):
        raise ToolException(
            f"没有权限执行操作：{permission}"
        )


# ============================================================
# 4. 风险检查
# 用户自定义业务逻辑
# ============================================================

REFUND_APPROVAL_THRESHOLD = 5000


def requires_approval(amount: int) -> bool:
    """
    判断退款是否需要人工审批。
    """
    return amount > REFUND_APPROVAL_THRESHOLD


# ============================================================
# 5. 查询订单
# LangChain Tool
# ============================================================

@tool(
    "query_order",
    description=(
        "查询订单信息。"
        "当用户需要查看订单状态、金额、用户信息时使用。"
    ),
)
def query_order(order_id: str) -> dict:
    require_permission("order:query")

    print(f"[Tool] 查询订单：{order_id}")

    return {
        "order_id": order_id,
        "user": "张三",
        "amount": 3999,
        "status": "已支付",
    }


# ============================================================
# 6. 退款 Tool
# LangChain Tool + 我们自己的权限/风险控制
# ============================================================

@tool(
    "refund_order",
    description=(
        "申请订单退款。"
        "需要提供订单号和退款金额。"
        "对于高金额退款，需要人工确认。"
    ),
)
def refund_order(
    order_id: str,
    amount: int,
) -> dict:

    # --------------------------------------------------------
    # 第一层：权限
    # --------------------------------------------------------

    require_permission("order:refund")

    print(
        f"[Security] 用户 {current_user['username']} "
        f"拥有 order:refund 权限"
    )

    # --------------------------------------------------------
    # 第二层：业务规则
    # --------------------------------------------------------

    if amount <= 0:
        raise ToolException("退款金额必须大于 0")

    # --------------------------------------------------------
    # 第三层：风险判断
    # --------------------------------------------------------

    if requires_approval(amount):

        print(
            f"[Security] 退款金额 {amount} 元 "
            f"超过 {REFUND_APPROVAL_THRESHOLD} 元"
        )

        return {
            "status": "PENDING_APPROVAL",
            "order_id": order_id,
            "amount": amount,
            "message": (
                f"退款金额 {amount} 元较高，"
                "需要人工审批后才能执行。"
            ),
        }

    # --------------------------------------------------------
    # 第四层：真正执行业务
    # --------------------------------------------------------

    print(
        f"[Business] 执行退款："
        f"order={order_id}, amount={amount}"
    )

    return {
        "status": "SUCCESS",
        "order_id": order_id,
        "amount": amount,
        "message": "退款成功",
    }


# ============================================================
# 7. Tool 错误处理
# LangChain Tool 能力 + 用户配置
# ============================================================

query_order.handle_tool_error = True
refund_order.handle_tool_error = True


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
)


# ============================================================
# 9. Agent 调用
# ============================================================

def run_agent(user_input: str):

    print("\n")
    print("=" * 70)
    print(f"用户：{user_input}")
    print("=" * 70)

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": user_input,
                }
            ]
        }
    )

    print("\n========== Agent Trace ==========")

    for message in result["messages"]:

        print("--------------------------------")

        print(type(message).__name__)

        if message.content:
            print("Content:")
            print(message.content)

        if getattr(message, "tool_calls", None):
            print("Tool Calls:")

            for tool_call in message.tool_calls:
                print(tool_call)


# ============================================================
# 10. 测试
# ============================================================

if __name__ == "__main__":

    # --------------------------------------------------------
    # 场景 1：普通查询
    # --------------------------------------------------------

    run_agent(
        "帮我查询一下订单10001的信息"
    )

    # --------------------------------------------------------
    # 场景 2：普通退款
    # --------------------------------------------------------

    run_agent(
        "帮我把订单10001退款500元"
    )

    # --------------------------------------------------------
    # 场景 3：高风险退款
    # --------------------------------------------------------

    run_agent(
        "帮我把订单10001退款50000元"
    )
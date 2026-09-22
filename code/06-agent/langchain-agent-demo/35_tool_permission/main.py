import os

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import ToolException, tool
from langchain_openai import ChatOpenAI
from pydantic import BaseModel, Field


# ============================================================
# 1. 初始化环境
# ============================================================

load_dotenv()


# ============================================================
# 2. 初始化 LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 3. 模拟当前用户
# ============================================================
#
# 实际企业系统中：
#
# 用户登录
#   ↓
# JWT / Session
#   ↓
# UserContext
#   ↓
# Tool Permission
#
# 这里为了 Demo 简单模拟。
#

current_user = {
    "user_id": "user1001",
    "username": "张三",
    "roles": [
        "USER",
    ],
    "permissions": [
        "order:query",
        "order:create",
    ],
}


# ============================================================
# 4. Permission Check
# ============================================================
#
# 这是我们自己定义的权限检查函数。
#
# 不是 LangChain API。
#

def has_permission(
    permission: str,
) -> bool:

    return permission in current_user["permissions"]


def require_permission(
    permission: str,
):
    """
    检查当前用户是否拥有指定权限。
    """

    if not has_permission(permission):

        raise ToolException(
            f"没有权限执行操作：{permission}"
        )


# ============================================================
# 5. Query Order Tool
# ============================================================

class QueryOrderInput(BaseModel):

    order_id: str = Field(
        description="订单号，例如：10001"
    )


@tool(
    "query_order",
    description=(
        "查询订单信息。"
        "普通用户可以使用。"
    ),
    args_schema=QueryOrderInput,
)
def query_order(
    order_id: str,
) -> dict:

    # --------------------------------------------------------
    # Permission Check
    # --------------------------------------------------------

    require_permission(
        "order:query"
    )

    # --------------------------------------------------------
    # Business Logic
    # --------------------------------------------------------

    return {
        "order_id": order_id,
        "status": "已发货",
        "amount": 3999,
    }


query_order.handle_tool_error = True


# ============================================================
# 6. Create Order Tool
# ============================================================

class CreateOrderInput(BaseModel):

    amount: int = Field(
        description="订单金额，例如：3999"
    )


@tool(
    "create_order",
    description=(
        "创建订单。"
        "需要 order:create 权限。"
    ),
    args_schema=CreateOrderInput,
)
def create_order(
    amount: int,
) -> dict:

    # --------------------------------------------------------
    # Permission Check
    # --------------------------------------------------------

    require_permission(
        "order:create"
    )

    # --------------------------------------------------------
    # Business Logic
    # --------------------------------------------------------

    return {
        "order_id": "10001",
        "amount": amount,
        "status": "已创建",
    }


create_order.handle_tool_error = True


# ============================================================
# 7. Refund Order Tool
# ============================================================

class RefundOrderInput(BaseModel):

    order_id: str = Field(
        description="需要退款的订单号，例如：10001"
    )


@tool(
    "refund_order",
    description=(
        "申请订单退款。"
        "需要 order:refund 权限。"
    ),
    args_schema=RefundOrderInput,
)
def refund_order(
    order_id: str,
) -> dict:

    # --------------------------------------------------------
    # Permission Check
    # --------------------------------------------------------

    require_permission(
        "order:refund"
    )

    # --------------------------------------------------------
    # Business Logic
    # --------------------------------------------------------

    return {
        "order_id": order_id,
        "status": "退款处理中",
    }


refund_order.handle_tool_error = True


# ============================================================
# 8. Tool Permission Information
# ============================================================

def print_permission_info():

    print(
        "\n========== Current User =========="
    )

    print(
        "user_id:"
    )

    print(
        current_user["user_id"]
    )

    print(
        "username:"
    )

    print(
        current_user["username"]
    )

    print(
        "roles:"
    )

    print(
        current_user["roles"]
    )

    print(
        "permissions:"
    )

    print(
        current_user["permissions"]
    )


# ============================================================
# 9. Direct Tool Test
# ============================================================

def test_tools():

    print(
        "\n========== Direct Tool Test =========="
    )


    # --------------------------------------------------------
    # Query
    # --------------------------------------------------------

    print(
        "\n[1] Query Order"
    )

    result = query_order.invoke(
        {
            "order_id": "10001"
        }
    )

    print(
        result
    )


    # --------------------------------------------------------
    # Create
    # --------------------------------------------------------

    print(
        "\n[2] Create Order"
    )

    result = create_order.invoke(
        {
            "amount": 3999
        }
    )

    print(
        result
    )


    # --------------------------------------------------------
    # Refund
    # --------------------------------------------------------

    print(
        "\n[3] Refund Order"
    )

    try:

        result = refund_order.invoke(
            {
                "order_id": "10001"
            }
        )

        print(
            result
        )

    except Exception as e:

        print(
            "Tool 执行失败："
        )

        print(
            e
        )


# ============================================================
# 10. Agent Test
# ============================================================

def test_agent():

    print(
        "\n========== Agent Test =========="
    )

    agent = create_agent(
        model=model,
        tools=[
            query_order,
            create_order,
            refund_order,
        ],
    )


    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": (
                        "帮我查询一下订单10001的信息，"
                        "然后帮我申请退款。"
                    ),
                }
            ]
        }
    )


    # ========================================================
    # Agent Trace
    # ========================================================

    print(
        "\n========== Agent Trace =========="
    )

    for message in result["messages"]:

        print(
            "\n--------------------"
        )

        print(
            type(message).__name__
        )

        print(
            "content:"
        )

        print(
            message.content
        )

        if getattr(
            message,
            "tool_calls",
            None,
        ):

            print(
                "Tool Calls:"
            )

            for tool_call in message.tool_calls:

                print(
                    tool_call
                )


# ============================================================
# 11. Main
# ============================================================

def main():

    print_permission_info()

    test_tools()

    test_agent()


# ============================================================
# Entry
# ============================================================

if __name__ == "__main__":
    main()
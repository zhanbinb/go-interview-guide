import os

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI


load_dotenv()


# =========================
# 1. Model
# =========================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# =========================
# 2. Tools
# =========================

@tool
def query_order(order_id: str) -> str:
    """查询订单信息，包括订单金额和订单状态。"""

    orders = {
        "10001": {
            "amount": "3999元",
            "status": "已发货",
        }
    }

    order = orders.get(order_id)

    if not order:
        return f"没有找到订单 {order_id}"

    return (
        f"订单号：{order_id}，"
        f"金额：{order['amount']}，"
        f"状态：{order['status']}"
    )


@tool
def query_logistics(order_id: str) -> str:
    """查询订单物流信息。"""

    logistics = {
        "10001": {
            "company": "顺丰",
            "tracking_no": "123456",
            "status": "运输中",
        }
    }

    result = logistics.get(order_id)

    if not result:
        return f"没有找到订单 {order_id} 的物流信息"

    return (
        f"订单号：{order_id}，"
        f"物流公司：{result['company']}，"
        f"运单号：{result['tracking_no']}，"
        f"物流状态：{result['status']}"
    )


# =========================
# 3. Create Agent
# =========================

agent = create_agent(
    model=model,
    tools=[
        query_order,
        query_logistics,
    ],
    system_prompt=(
        "你是一名专业的订单客服。"
        "根据用户的问题选择合适的工具查询信息。"
        "不要编造工具没有返回的数据。"
    ),
)


# =========================
# 4. Run
# =========================

result = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": (
                    "帮我查询订单10001的金额、状态和物流信息，"
                    "最后给我一个完整的结果。"
                ),
            }
        ]
    }
)


# =========================
# 5. Print
# =========================

for message in result["messages"]:
    print("================================")
    print(type(message).__name__)
    print(message.content)

    if getattr(message, "tool_calls", None):
        print("Tool Calls:")
        for tool_call in message.tool_calls:
            print(tool_call)
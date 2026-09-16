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
def get_user(user_id: str) -> str:
    """查询用户信息。"""

    users = {
        "1001": {
            "user_id": "1001",
            "name": "张三",
            "level": "VIP",
        }
    }

    user = users.get(user_id)

    if not user:
        return f"用户 {user_id} 不存在"

    return (
        f"用户ID：{user['user_id']}，"
        f"姓名：{user['name']}，"
        f"等级：{user['level']}"
    )


@tool
def query_order(order_id: str) -> str:
    """查询订单信息。"""

    orders = {
        "10001": {
            "order_id": "10001",
            "status": "已发货",
            "amount": 3999,
        }
    }

    order = orders.get(order_id)

    if not order:
        return f"订单 {order_id} 不存在"

    return (
        f"订单号：{order['order_id']}，"
        f"状态：{order['status']}，"
        f"金额：{order['amount']} 元"
    )


@tool
def query_payment(order_id: str) -> str:
    """查询订单支付信息。"""

    payments = {
        "10001": {
            "status": "已支付",
            "amount": 3999,
            "method": "微信支付",
        }
    }

    payment = payments.get(order_id)

    if not payment:
        return f"订单 {order_id} 没有支付记录"

    return (
        f"订单号：{order_id}，"
        f"支付状态：{payment['status']}，"
        f"支付金额：{payment['amount']} 元，"
        f"支付方式：{payment['method']}"
    )


@tool
def query_logistics(order_id: str) -> str:
    """查询订单物流信息。"""

    logistics = {
        "10001": {
            "status": "运输中",
            "tracking_no": "123456",
            "expected": "明天送达",
        }
    }

    info = logistics.get(order_id)

    if not info:
        return f"订单 {order_id} 没有物流信息"

    return (
        f"订单号：{order_id}，"
        f"物流状态：{info['status']}，"
        f"物流单号：{info['tracking_no']}，"
        f"预计：{info['expected']}"
    )


# =========================
# 3. Agent
# =========================

agent = create_agent(
    model=model,
    tools=[
        get_user,
        query_order,
        query_payment,
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

result = agent.invoke({
    "messages": [
        {
            "role": "user",
            "content": (
                "订单10001现在什么状态？"
                # "包括用户信息、订单状态和金额、支付信息以及物流信息。"
            ),
        }
    ]
})


print("=== Agent Messages ===")

for message in result["messages"]:
    print()
    print(f"Type: {message.__class__.__name__}")
    print(message)
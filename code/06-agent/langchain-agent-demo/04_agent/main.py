import os

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_openai import ChatOpenAI
from langchain_core.tools import tool

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
# 2. Tool
# =========================

@tool
def query_order(order_id: str) -> str:
    """查询订单信息。"""

    orders = {
        "10001": {
            "order_id": "10001",
            "status": "已发货",
            "amount": 3999,
        },
        "10002": {
            "order_id": "10002",
            "status": "待付款",
            "amount": 1999,
        },
    }

    order = orders.get(order_id)

    if not order:
        return f"订单 {order_id} 不存在"

    return (
        f"订单号：{order['order_id']}，"
        f"状态：{order['status']}，"
        f"金额：{order['amount']} 元"
    )


# =========================
# 3. Agent
# =========================

agent = create_agent(
    model=model,
    tools=[query_order],
    system_prompt="你是一名专业的订单客服。",
)


# =========================
# 4. Run Agent
# =========================

result = agent.invoke({
    "messages": [
        {
            "role": "user",
            "content": "帮我查一下订单10002",
        }
    ]
})


print("=== Agent Result ===")
print(result)

print()
print("=== Final Message ===")
print(result["messages"][-1].content)
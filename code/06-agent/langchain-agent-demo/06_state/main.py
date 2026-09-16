import os

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import InMemorySaver

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
def query_logistics(order_id: str) -> str:
    """查询订单物流信息。"""

    return (
        f"订单号：{order_id}，"
        "物流状态：运输中，"
        "物流单号：123456，"
        "预计明天送达"
    )


# =========================
# 3. Checkpointer
# =========================

checkpointer = InMemorySaver()


# =========================
# 4. Agent
# =========================

agent = create_agent(
    model=model,
    tools=[
        query_order,
        query_logistics,
    ],
    system_prompt="你是一名专业的订单客服。",
    checkpointer=checkpointer,
)


# =========================
# 5. First conversation
# =========================

config = {
    "configurable": {
        "thread_id": "user-1001"
    }
}

result1 = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": "帮我查询订单10001",
            }
        ]
    },
    config=config,
)

print("=== 第一次对话 ===")
print(result1["messages"][-1].content)


# =========================
# 6. Second conversation
# =========================

result2 = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": "那它什么时候送到？",
            }
        ]
    },
    config=config,
)

print()
print("=== 第二次对话 ===")
print(result2["messages"][-1].content)
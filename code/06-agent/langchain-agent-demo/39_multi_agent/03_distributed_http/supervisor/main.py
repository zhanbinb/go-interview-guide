import uuid

import httpx

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 3. 远程调用 Order Agent
# ============================================================

@tool(
    "order_agent",
    description=(
        "调用远程订单领域 Agent。"
        "处理订单状态、订单金额、订单用户、订单发货等问题。"
    ),
)
def call_order_agent(task: str) -> str:

    print("\n>>> Supervisor 调用 Order Agent")

    response = httpx.post(
        "http://127.0.0.1:8001/agent/order",
        json={
            "task_id": str(uuid.uuid4()),
            "user_id": "user1001",
            "task": task,
        },
        timeout=30,
    )

    response.raise_for_status()

    data = response.json()

    return data["data"]["answer"]


# ============================================================
# 4. 远程调用 Payment Agent
# ============================================================

@tool(
    "payment_agent",
    description=(
        "调用远程支付领域 Agent。"
        "处理支付状态、支付方式、退款等支付问题。"
    ),
)
def call_payment_agent(task: str) -> str:

    print("\n>>> Supervisor 调用 Payment Agent")

    response = httpx.post(
        "http://127.0.0.1:8002/agent/payment",
        json={
            "task_id": str(uuid.uuid4()),
            "user_id": "user1001",
            "task": task,
        },
        timeout=30,
    )

    response.raise_for_status()

    data = response.json()

    return data["data"]["answer"]


# ============================================================
# 5. Supervisor Agent
# ============================================================

supervisor_agent = create_agent(
    model=model,
    tools=[
        call_order_agent,
        call_payment_agent,
    ],
    system_prompt=(
        "你是一个电商客服 Supervisor Agent。\n\n"
        "你负责理解用户任务，并调用合适的专业 Agent。\n\n"
        "可用 Agent：\n"
        "- order_agent：订单领域\n"
        "- payment_agent：支付领域\n\n"
        "如果问题涉及订单，请调用 order_agent。\n"
        "如果问题涉及支付，请调用 payment_agent。\n"
        "如果同时涉及多个领域，可以调用多个 Agent。\n\n"
        "专业 Agent 返回结果后，"
        "请综合这些结果回答用户。"
    ),
)


# ============================================================
# 6. 测试
# ============================================================

if __name__ == "__main__":

    user_task = (
        "帮我查询订单10001的状态，"
        "然后看看这个订单有没有付款。"
    )

    print("=" * 70)
    print("用户：")
    print(user_task)

    print("=" * 70)

    result = supervisor_agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": user_task,
                }
            ]
        }
    )

    print("=" * 70)
    print("最终回答：")

    print(result["messages"][-1].content)

    print("=" * 70)
    print("Supervisor 消息链：")

    for message in result["messages"]:

        print("----------------------------------------")

        print(type(message).__name__)

        if message.content:
            print("Content:")
            print(message.content)

        if getattr(message, "tool_calls", None):

            print("Tool Calls:")

            for tool_call in message.tool_calls:
                print(tool_call)
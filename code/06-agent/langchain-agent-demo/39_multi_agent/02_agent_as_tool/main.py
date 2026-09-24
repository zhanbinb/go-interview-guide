from dotenv import load_dotenv

from langchain.agents import create_agent
from langchain.tools import tool
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 加载环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. 创建 LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 3. Order Agent
# ============================================================

order_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是订单专家 Agent。\n"
        "你只负责处理订单相关问题。\n\n"
        "订单数据：\n"
        "- 订单号：10001\n"
        "- 用户：张三\n"
        "- 金额：3999 元\n"
        "- 状态：已发货\n\n"
        "请根据任务返回准确、简洁的订单信息。"
    ),
)


# ============================================================
# 4. Payment Agent
# ============================================================

payment_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是支付专家 Agent。\n"
        "你只负责处理支付相关问题。\n\n"
        "支付数据：\n"
        "- 订单号：10001\n"
        "- 支付状态：已支付\n"
        "- 支付方式：微信支付\n\n"
        "请根据任务返回准确、简洁的支付信息。"
    ),
)


# ============================================================
# 5. Logistics Agent
# ============================================================

logistics_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是物流专家 Agent。\n"
        "你只负责处理物流相关问题。\n\n"
        "物流数据：\n"
        "- 订单号：10001\n"
        "- 快递公司：顺丰\n"
        "- 运单号：123456\n"
        "- 当前状态：运输中\n\n"
        "请根据任务返回准确、简洁的物流信息。"
    ),
)


# ============================================================
# 6. 把 Order Agent 包装成 Tool
# ============================================================

@tool(
    "order_agent",
    description=(
        "订单领域专家 Agent。"
        "当用户需要查询订单状态、订单金额、订单用户、"
        "订单是否发货等订单相关信息时调用。"
    ),
)
def call_order_agent(query: str) -> str:
    """
    调用 Order Agent。
    """

    result = order_agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": query,
                }
            ]
        }
    )

    return result["messages"][-1].content


# ============================================================
# 7. 把 Payment Agent 包装成 Tool
# ============================================================

@tool(
    "payment_agent",
    description=(
        "支付领域专家 Agent。"
        "当用户需要查询订单支付状态、支付方式、"
        "退款等支付相关问题时调用。"
    ),
)
def call_payment_agent(query: str) -> str:
    """
    调用 Payment Agent。
    """

    result = payment_agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": query,
                }
            ]
        }
    )

    return result["messages"][-1].content


# ============================================================
# 8. 把 Logistics Agent 包装成 Tool
# ============================================================

@tool(
    "logistics_agent",
    description=(
        "物流领域专家 Agent。"
        "当用户需要查询物流状态、快递公司、"
        "运单号、物流进度等物流相关信息时调用。"
    ),
)
def call_logistics_agent(query: str) -> str:
    """
    调用 Logistics Agent。
    """

    result = logistics_agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": query,
                }
            ]
        }
    )

    return result["messages"][-1].content


# ============================================================
# 9. Supervisor Agent
# ============================================================

supervisor_agent = create_agent(
    model=model,
    tools=[
        call_order_agent,
        call_payment_agent,
        call_logistics_agent,
    ],
    system_prompt=(
        "你是一个电商客服 Supervisor Agent。\n\n"
        "你负责理解用户的问题，并协调专业领域 Agent 完成任务。\n\n"
        "可使用的专业 Agent：\n"
        "- order_agent：负责订单领域\n"
        "- payment_agent：负责支付领域\n"
        "- logistics_agent：负责物流领域\n\n"
        "如果一个问题涉及多个领域，可以调用多个 Agent。\n"
        "根据专业 Agent 返回的结果，给用户生成最终回答。\n"
        "不要编造专业 Agent 没有提供的信息。"
    ),
)


# ============================================================
# 10. 执行测试
# ============================================================

if __name__ == "__main__":

    user_task = (
        "帮我查询一下订单10001的状态，"
        "看看有没有付款，"
        "再告诉我现在物流到哪里了。"
    )

    print("=" * 70)
    print("用户任务：")
    print(user_task)

    print("=" * 70)
    print("Agent 执行过程：")

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
    print("最终结果：")

    print(result["messages"][-1].content)

    print("=" * 70)
    print("完整消息链：")

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
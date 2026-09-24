from dotenv import load_dotenv

from langchain.agents import create_agent
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
# 3. 创建 Order Agent
# ============================================================

order_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是订单专家 Agent。\n"
        "你只负责处理订单相关问题。\n"
        "已知订单数据：\n"
        "- 订单号：10001\n"
        "- 用户：张三\n"
        "- 金额：3999 元\n"
        "- 状态：已发货\n"
        "\n"
        "请根据用户任务返回简洁、准确的订单信息。"
    ),
)


# ============================================================
# 4. 创建 Payment Agent
# ============================================================

payment_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是支付专家 Agent。\n"
        "你只负责处理支付相关问题。\n"
        "已知支付数据：\n"
        "- 订单号：10001\n"
        "- 支付状态：已支付\n"
        "- 支付方式：微信支付\n"
        "\n"
        "请根据用户任务返回简洁、准确的支付信息。"
    ),
)


# ============================================================
# 5. 创建 Logistics Agent
# ============================================================

logistics_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是物流专家 Agent。\n"
        "你只负责处理物流相关问题。\n"
        "已知物流数据：\n"
        "- 订单号：10001\n"
        "- 快递公司：顺丰\n"
        "- 运单号：123456\n"
        "- 当前状态：运输中\n"
        "\n"
        "请根据用户任务返回简洁、准确的物流信息。"
    ),
)


# ============================================================
# 6. Supervisor Agent
# ============================================================

def call_agent(agent, task: str) -> str:
    """
    调用一个专业 Agent，并提取最终回答。
    """

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": task,
                }
            ]
        }
    )

    return result["messages"][-1].content


def supervisor(task: str) -> str:
    """
    Supervisor：

    根据用户任务判断应该调用哪些专业 Agent。
    """

    decision = model.invoke(
        [
            {
                "role": "system",
                "content": (
                    "你是 Supervisor Agent。\n"
                    "你负责分析用户任务，并决定需要哪些专业 Agent。\n\n"
                    "可用 Agent：\n"
                    "1. order：订单相关问题\n"
                    "2. payment：支付相关问题\n"
                    "3. logistics：物流相关问题\n\n"
                    "请严格按照以下格式返回：\n"
                    "order,payment,logistics\n"
                    "只返回需要的 Agent 名称，用逗号分隔。\n"
                    "例如：\n"
                    "order\n"
                    "或者：\n"
                    "payment,logistics"
                ),
            },
            {
                "role": "user",
                "content": task,
            },
        ]
    )

    decision_text = decision.content.strip()

    print("Supervisor 决策：", decision_text)

    selected_agents = [
        item.strip()
        for item in decision_text.split(",")
        if item.strip()
    ]

    results = []

    # ========================================================
    # 7. Supervisor 委派任务
    # ========================================================

    if "order" in selected_agents:
        result = call_agent(
            order_agent,
            task,
        )

        results.append(
            f"【订单 Agent】\n{result}"
        )

    if "payment" in selected_agents:
        result = call_agent(
            payment_agent,
            task,
        )

        results.append(
            f"【支付 Agent】\n{result}"
        )

    if "logistics" in selected_agents:
        result = call_agent(
            logistics_agent,
            task,
        )

        results.append(
            f"【物流 Agent】\n{result}"
        )

    # ========================================================
    # 8. Supervisor 汇总结果
    # ========================================================

    final_result = model.invoke(
        [
            {
                "role": "system",
                "content": (
                    "你是 Supervisor Agent。\n"
                    "请根据多个专业 Agent 的结果，"
                    "给用户生成一个简洁、自然的最终回答。\n"
                    "不要编造信息。"
                ),
            },
            {
                "role": "user",
                "content": (
                    f"原始任务：\n{task}\n\n"
                    f"专业 Agent 结果：\n"
                    f"{chr(10).join(results)}"
                ),
            },
        ]
    )

    return final_result.content


# ============================================================
# 9. 测试
# ============================================================

if __name__ == "__main__":

    user_task = (
        "帮我查询一下订单10001的订单状态，"
        "再看看这个订单有没有付款，"
        "以及现在物流到哪里了。"
    )

    print("=" * 60)
    print("用户任务：")
    print(user_task)

    print("=" * 60)

    result = supervisor(user_task)

    print("最终结果：")
    print(result)


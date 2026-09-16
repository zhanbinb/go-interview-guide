import os

from dotenv import load_dotenv
from langchain_core.messages import HumanMessage, ToolMessage
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
        return f"订单 {order_id} 不存在"

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

    logistics_info = logistics.get(order_id)

    if not logistics_info:
        return f"订单 {order_id} 暂无物流信息"

    return (
        f"订单号：{order_id}，"
        f"物流公司：{logistics_info['company']}，"
        f"运单号：{logistics_info['tracking_no']}，"
        f"物流状态：{logistics_info['status']}"
    )


# =========================
# 3. Tool Registry
# =========================

tools = [
    query_order,
    query_logistics,
]

tool_map = {
    tool.name: tool
    for tool in tools
}


# =========================
# 4. Bind Tools
# =========================

model_with_tools = model.bind_tools(tools)


# =========================
# 5. ReAct Loop
# =========================

messages = [
    HumanMessage(
        content=(
            "帮我查询订单10001的金额、状态和物流信息，"
            "最后给我一个完整的结果。"
        )
    )
]


while True:

    print("\n==============================")
    print("Reason")
    print("==============================")

    # LLM 决策
    response = model_with_tools.invoke(messages)

    messages.append(response)

    # 如果没有 Tool Call
    # 说明 Agent 已经得到了足够的信息
    # 可以直接结束
    if not response.tool_calls:

        print("Final Answer:")
        print(response.content)

        break

    # =========================
    # Act
    # =========================

    print("\n==============================")
    print("Act")
    print("==============================")

    for tool_call in response.tool_calls:

        tool_name = tool_call["name"]
        tool_args = tool_call["args"]

        print(f"调用工具：{tool_name}")
        print(f"参数：{tool_args}")

        tool = tool_map.get(tool_name)

        if not tool:
            tool_result = f"工具 {tool_name} 不存在"
        else:
            tool_result = tool.invoke(tool_args)

        print(f"结果：{tool_result}")

        # =========================
        # Observation
        # =========================

        messages.append(
            ToolMessage(
                content=tool_result,
                tool_call_id=tool_call["id"],
            )
        )

        print("\nObservation:")
        print(tool_result)
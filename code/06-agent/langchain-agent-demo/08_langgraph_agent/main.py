import os
from typing import Annotated

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, ToolMessage
from langchain_core.tools import tool
from langgraph.graph import StateGraph, START, END
from langgraph.graph.message import add_messages
from typing_extensions import TypedDict


load_dotenv()


# =========================================================
# 1. Model
# =========================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# =========================================================
# 2. Tools
# =========================================================

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


@tool
def query_logistics(order_id: str) -> str:
    """查询订单物流信息。"""

    return (
        f"订单号：{order_id}，"
        "物流状态：运输中，"
        "物流单号：123456，"
        "预计明天送达"
    )


tools = [
    query_order,
    query_logistics,
]


# =========================================================
# 3. Tool Registry
# =========================================================

tools_by_name = {
    tool.name: tool
    for tool in tools
}


# =========================================================
# 4. State
# =========================================================

class State(TypedDict):
    messages: Annotated[list, add_messages]


# =========================================================
# 5. Bind Tools
# =========================================================

model_with_tools = model.bind_tools(tools)


# =========================================================
# 6. LLM Node
# =========================================================

def call_model(state: State):

    print("\n========== LLM Node ==========")

    response = model_with_tools.invoke(
        state["messages"]
    )

    print("LLM Response:")

    if response.tool_calls:
        for tool_call in response.tool_calls:
            print(
                f"Tool Call: "
                f"{tool_call['name']} "
                f"{tool_call['args']}"
            )
    else:
        print(response.content)

    return {
        "messages": [response]
    }


# =========================================================
# 7. Tool Node
# =========================================================

def call_tools(state: State):

    print("\n========== Tool Node ==========")

    last_message = state["messages"][-1]

    tool_messages = []

    for tool_call in last_message.tool_calls:

        tool_name = tool_call["name"]
        tool_args = tool_call["args"]

        print(
            f"Calling Tool: "
            f"{tool_name} "
            f"{tool_args}"
        )

        tool = tools_by_name.get(tool_name)

        if not tool:
            result = f"未知 Tool：{tool_name}"
        else:
            result = tool.invoke(tool_args)

        print(f"Tool Result: {result}")

        tool_messages.append(
            ToolMessage(
                content=result,
                tool_call_id=tool_call["id"],
            )
        )

    return {
        "messages": tool_messages
    }


# =========================================================
# 8. Conditional Edge
# =========================================================

def should_continue(state: State):

    last_message = state["messages"][-1]

    if last_message.tool_calls:
        return "tools"

    return "end"


# =========================================================
# 9. Build Graph
# =========================================================

builder = StateGraph(State)


builder.add_node(
    "llm",
    call_model,
)

builder.add_node(
    "tools",
    call_tools,
)


# START → LLM

builder.add_edge(
    START,
    "llm",
)


# LLM → Tool / END

builder.add_conditional_edges(
    "llm",
    should_continue,
    {
        "tools": "tools",
        "end": END,
    },
)


# Tool → LLM

builder.add_edge(
    "tools",
    "llm",
)


graph = builder.compile()


# =========================================================
# 10. Run
# =========================================================

result = graph.invoke(
    {
        "messages": [
            HumanMessage(
                content="帮我查询一下订单10001的状态、金额和物流"
            )
        ]
    }
)


# =========================================================
# 11. Final Result
# =========================================================

print("\n========== Final Result ==========")

for message in result["messages"]:

    print(
        f"{message.__class__.__name__}: "
        f"{message.content}"
    )
import os
from typing import TypedDict

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, START, END


load_dotenv()


# ==================================================
# LangChain 官方 API
# ==================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ==================================================
# 我们自己定义的 State
# ==================================================

class AgentState(TypedDict):
    user_task: str
    intent: str
    result: str


# ==================================================
# Node：Intent Router
# ==================================================

def intent_router(state: AgentState):

    prompt = f"""
你是一个客服系统的意图分类器。

请判断用户的问题属于哪一种：

1. order：订单查询
2. refund：退款问题
3. chat：普通聊天

用户问题：
{state["user_task"]}

只输出：
order
或者：
refund
或者：
chat
"""

    response = model.invoke(prompt)

    intent = response.content.strip().lower()

    if "order" in intent:
        intent = "order"
    elif "refund" in intent:
        intent = "refund"
    else:
        intent = "chat"

    return {
        "intent": intent
    }


# ==================================================
# Order Workflow
# ==================================================

def order_workflow(state: AgentState):

    print("进入：订单流程")

    # 这里暂时模拟订单查询
    order_info = "订单10001：金额3999元，状态：已发货"

    logistics_info = "物流单号：123456，预计明天送达"

    response = model.invoke(
        f"""
请根据下面的信息回答用户问题。

用户问题：
{state["user_task"]}

订单信息：
{order_info}

物流信息：
{logistics_info}
"""
    )

    return {
        "result": response.content
    }


# ==================================================
# Refund Workflow
# ==================================================

def refund_workflow(state: AgentState):

    print("进入：退款流程")

    # 暂时模拟查询结果
    order_info = "订单10001：金额3999元，状态：已发货"

    response = model.invoke(
        f"""
你是退款客服。

用户问题：
{state["user_task"]}

订单信息：
{order_info}

请判断用户是否可以申请退款，并说明原因。
"""
    )

    return {
        "result": response.content
    }


# ==================================================
# Chat Workflow
# ==================================================

def chat_workflow(state: AgentState):

    print("进入：普通问答流程")

    response = model.invoke(
        state["user_task"]
    )

    return {
        "result": response.content
    }


# ==================================================
# Conditional Routing
# ==================================================

def route_by_intent(state: AgentState):
    return state["intent"]


# ==================================================
# 创建 Graph
# ==================================================

builder = StateGraph(AgentState)

builder.add_node("intent_router", intent_router)
builder.add_node("order_workflow", order_workflow)
builder.add_node("refund_workflow", refund_workflow)
builder.add_node("chat_workflow", chat_workflow)

builder.add_edge(START, "intent_router")


# ==================================================
# LangGraph 官方 API：
# Conditional Routing
# ==================================================

builder.add_conditional_edges(
    "intent_router",
    route_by_intent,
    {
        "order": "order_workflow",
        "refund": "refund_workflow",
        "chat": "chat_workflow",
    }
)


builder.add_edge("order_workflow", END)
builder.add_edge("refund_workflow", END)
builder.add_edge("chat_workflow", END)


graph = builder.compile()


# ==================================================
# 执行
# ==================================================

result = graph.invoke({
    "user_task": "帮我查一下订单10001什么时候可以收到",
})

print("\n最终 State：")
print(result)

print("\n最终结果：")
print(result["result"])

print(graph.get_graph().draw_mermaid())
# 或生成图片
graph.get_graph().draw_mermaid_png(output_file_path="graph.png")
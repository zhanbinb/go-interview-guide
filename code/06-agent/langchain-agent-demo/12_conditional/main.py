import os
from typing import TypedDict

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, START, END


load_dotenv()


# =========================
# LangChain 官方 API
# =========================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# =========================
# 我们自己定义的 State
# =========================

class AgentState(TypedDict):
    user_task: str
    route: str
    result: str


# =========================
# Node 1：Router
# =========================

def router(state: AgentState):
    user_task = state["user_task"]

    prompt = f"""
你是一个任务路由器。

请判断用户的问题是否需要查询订单信息。

用户问题：
{user_task}

如果需要查询订单，只输出：
query

如果不需要查询订单，只输出：
direct
"""

    response = model.invoke(prompt)

    route = response.content.strip()

    if "query" in route.lower():
        route = "query"
    else:
        route = "direct"

    return {
        "route": route
    }


# =========================
# Node 2：查询订单
# =========================

def query_order(state: AgentState):
    print("执行：query_order")

    return {
        "result": "订单10001：金额3999元，状态：已发货"
    }


# =========================
# Node 3：直接回答
# =========================

def direct_answer(state: AgentState):
    print("执行：direct_answer")

    response = model.invoke(
        f"""
请直接回答用户的问题。

用户问题：
{state["user_task"]}
"""
    )

    return {
        "result": response.content
    }


# =========================
# 条件路由函数
# =========================

def route_after_router(state: AgentState):
    return state["route"]


# =========================
# 创建 Graph
# =========================

builder = StateGraph(AgentState)

builder.add_node("router", router)
builder.add_node("query_order", query_order)
builder.add_node("direct_answer", direct_answer)

builder.add_edge(START, "router")

builder.add_conditional_edges(
    "router",
    route_after_router,
    {
        "query": "query_order",
        "direct": "direct_answer",
    }
)

builder.add_edge("query_order", END)
builder.add_edge("direct_answer", END)

graph = builder.compile()


# =========================
# 执行
# =========================

result = graph.invoke({
    "user_task": "帮我查询一下订单10001",
})

print("\n最终 State：")
print(result)

print("\n最终结果：")
print(result["result"])
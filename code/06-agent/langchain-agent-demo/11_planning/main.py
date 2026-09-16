import os

from dotenv import load_dotenv
from typing import TypedDict

from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, START, END


load_dotenv()


# ========================================
# 1. Model
# ========================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ========================================
# 2. State
# ========================================

class AgentState(TypedDict):
    user_task: str
    plan: str
    result: str


# ========================================
# 3. Planner Node
# ========================================

def planner(state: AgentState):

    prompt = f"""
你是一名任务规划器。

请分析下面的用户任务，并制定一个简单的执行计划。

用户任务：
{state["user_task"]}

要求：
1. 将任务拆分成明确的步骤。
2. 每一步说明需要做什么。
3. 只输出执行计划。
"""

    response = model.invoke(prompt)

    return {
        "plan": response.content
    }


# ========================================
# 4. Executor Node
# ========================================

def executor(state: AgentState):

    prompt = f"""
你是一名订单客服 Agent。

用户任务：
{state["user_task"]}

执行计划：
{state["plan"]}

请根据这个计划，给出一个完整的执行结果。

目前只是学习 Planning 和 Workflow，
不需要真正调用工具。

请模拟执行过程，并给出最终结果。
"""

    response = model.invoke(prompt)

    return {
        "result": response.content
    }


# ========================================
# 5. Build Graph
# ========================================

builder = StateGraph(AgentState)


builder.add_node(
    "planner",
    planner,
)

builder.add_node(
    "executor",
    executor,
)


builder.add_edge(
    START,
    "planner",
)

builder.add_edge(
    "planner",
    "executor",
)

builder.add_edge(
    "executor",
    END,
)


graph = builder.compile()


# ========================================
# 6. Invoke
# ========================================

result = graph.invoke(
    {
        "user_task": (
            "帮我分析订单10001的情况，"
            "需要先获取订单信息，再分析结果。"
        ),
        "plan": "",
        "result": "",
    }
)


print("\n==============================")
print("Final Result")
print("==============================")

print(result["result"])
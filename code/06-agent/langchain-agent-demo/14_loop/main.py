from typing import TypedDict

from langgraph.graph import StateGraph, START, END


# ==========================================
# 我们自己定义的 State
# ==========================================

class AgentState(TypedDict):
    task: str
    current_step: int
    result: str


# ==========================================
# Node：执行一步任务
# ==========================================

def executor(state: AgentState):

    current_step = state["current_step"]

    print(f"执行第 {current_step} 步")

    new_result = (
        state["result"]
        + f"完成步骤 {current_step}\n"
    )

    return {
        "current_step": current_step + 1,
        "result": new_result,
    }


# ==========================================
# Node：判断是否继续
# ==========================================

def should_continue(state: AgentState):

    if state["current_step"] <= 3:
        return "continue"

    return "finish"


# ==========================================
# 创建 Graph
# ==========================================

builder = StateGraph(AgentState)

builder.add_node("executor", executor)

builder.add_edge(START, "executor")


# ==========================================
# Conditional Edge
# ==========================================

builder.add_conditional_edges(
    "executor",
    should_continue,
    {
        "continue": "executor",
        "finish": END,
    }
)


graph = builder.compile()


# ==========================================
# 执行
# ==========================================

result = graph.invoke({
    "task": "执行一个三步骤任务",
    "current_step": 1,
    "result": "",
})


print("\n最终 State：")
print(result)

graph.get_graph().draw_mermaid_png(output_file_path="graph.png")
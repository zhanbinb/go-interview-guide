from typing import TypedDict

from langgraph.graph import StateGraph, START, END
from langgraph.checkpoint.memory import InMemorySaver


# ==========================================
# 我们自己定义的 State
# ==========================================

class AgentState(TypedDict):
    message: str
    count: int


# ==========================================
# Node
# ==========================================

def process(state: AgentState):

    print("当前 State：", state)

    return {
        "message": state["message"] + " → processed",
        "count": state["count"] + 1,
    }


# ==========================================
# 创建 Graph
# ==========================================

builder = StateGraph(AgentState)

builder.add_node("process", process)

builder.add_edge(START, "process")
builder.add_edge("process", END)


# ==========================================
# 创建 Checkpointer
# ==========================================

checkpointer = InMemorySaver()

graph = builder.compile(
    checkpointer=checkpointer
)


# ==========================================
# 第一次执行
# ==========================================

config = {
    "configurable": {
        "thread_id": "user-001"
    }
}

result1 = graph.invoke(
    {
        "message": "hello",
        "count": 0,
    },
    config,
)

print("\n第一次执行：")
print(result1)


# ==========================================
# 查看保存的 State
# ==========================================

saved_state = graph.get_state(config)

print("\nCheckpoint 中保存的 State：")
print(saved_state)
from typing import TypedDict

from langgraph.checkpoint.memory import InMemorySaver
from langgraph.graph import StateGraph, START, END
from langgraph.types import Command, interrupt


# =========================
# 1. State：我们自己定义
# =========================

class AgentState(TypedDict):
    task: str
    result: str


# =========================
# 2. Node：普通业务节点
# =========================

def process_task(state: AgentState):
    print("开始处理任务：", state["task"])

    # 模拟一个需要人工确认的操作
    approval = interrupt(
        {
            "message": "该操作需要人工确认，是否继续？",
            "task": state["task"],
        }
    )

    print("人工确认结果：", approval)

    if approval == "yes":
        return {
            "result": "任务已执行"
        }

    return {
        "result": "任务已取消"
    }


# =========================
# 3. 创建 Graph
# =========================

builder = StateGraph(AgentState)

builder.add_node("process_task", process_task)

builder.add_edge(START, "process_task")
builder.add_edge("process_task", END)


# =========================
# 4. Checkpoint
# =========================

checkpointer = InMemorySaver()

graph = builder.compile(
    checkpointer=checkpointer
)


# =========================
# 5. thread_id
# =========================

config = {
    "configurable": {
        "thread_id": "conversation-001"
    }
}


# =========================
# 6. 第一次执行
# =========================

print("\n=== 第一次执行 ===")

result = graph.invoke(
    {
        "task": "执行退款操作",
        "result": "",
    },
    config,
)

print("第一次执行结果：")
print(result)


# =========================
# 7. 人工确认后恢复
# =========================

print("\n=== 人工确认：yes ===")

result = graph.invoke(
    Command(resume="yes"),
    config,
)

print("最终结果：")
print(result)
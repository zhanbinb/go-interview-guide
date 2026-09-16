import os

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, START, END
from typing_extensions import TypedDict

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
# 2. State
# =========================

class State(TypedDict):
    question: str
    answer: str
    formatted_answer: str


# =========================
# 3. Node
# =========================

def call_model(state: State):
    response = model.invoke(state["question"])

    return {
        "answer": response.content
    }


# =========================
# 4. Build Graph
# =========================

builder = StateGraph(State)

builder.add_node("llm", call_model)

builder.add_edge(START, "llm")
builder.add_edge("llm", "format_answer")
builder.add_edge("format_answer", END)

graph = builder.compile()


# =========================
# 5. Run
# =========================

result = graph.invoke({
    "question": "用一句话解释什么是 Go 的 Goroutine",
    "answer": "",
})

print("=== Result ===")
print(result)


def format_answer(state: State):
    return {
        "formatted_answer": f"AI回答：{state['answer']}"
    }
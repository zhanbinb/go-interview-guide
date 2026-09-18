from pathlib import Path
import os
from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain.agents import create_agent
from langchain.tools import tool

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
# 我们自己的代码
# =========================

KNOWLEDGE_DIR = Path(__file__).parent / "knowledge"


@tool
def search_knowledge(query: str) -> str:
    """Search the local knowledge base for information relevant to the query."""

    results = []

    for file in KNOWLEDGE_DIR.glob("*.md"):
        content = file.read_text(encoding="utf-8")

        if any(keyword in query.lower() for keyword in ["go", "channel", "goroutine"]):
            if file.name == "go.md":
                results.append(content)

        if any(keyword in query.lower() for keyword in ["redis", "setnx", "rdb", "aof"]):
            if file.name == "redis.md":
                results.append(content)

        if any(keyword in query.lower() for keyword in ["mysql", "mvcc", "b+tree", "索引"]):
            if file.name == "mysql.md":
                results.append(content)

    if not results:
        return "知识库中没有找到相关信息。"

    return "\n\n---\n\n".join(results)


# =========================
# LangChain 官方 API
# =========================

agent = create_agent(
    model=model,
    tools=[search_knowledge],
    system_prompt="""
你是一个技术知识助手。

当用户的问题涉及知识库中的技术知识时，
可以调用 search_knowledge 工具获取资料。

如果问题不需要查询知识库，可以直接回答。

回答时优先使用知识库中的内容。
""",
)


# =========================
# 运行 Agent
# =========================

result = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": "你好，今天心情不错。",
            }
        ]
    }
)

# =========================
# 5. Print
# =========================

for message in result["messages"]:
    print("================================")
    print(type(message).__name__)
    print(message.content)

    if getattr(message, "tool_calls", None):
        print("Tool Calls:")
        for tool_call in message.tool_calls:
            print(tool_call)
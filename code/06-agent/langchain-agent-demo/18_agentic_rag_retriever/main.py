from pathlib import Path
import os

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain.agents import create_agent
from langchain.tools import tool
from langchain_core.documents import Document
from langchain_core.retrievers import BaseRetriever

from pydantic import Field


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


# ============================================================
# 2. Knowledge
#    我们自己的代码
# ============================================================

KNOWLEDGE_DIR = Path(__file__).parent / "knowledge"


# ============================================================
# 3. Retriever
#    LangChain 官方 BaseRetriever
#    + 我们自己的 SimpleRetriever
# ============================================================

class SimpleRetriever(BaseRetriever):

    knowledge_dir: Path = Field(exclude=True)

    def _get_relevant_documents(self, query: str) -> list[Document]:

        print(f"[Retriever] query = {query}")
        print(f"[Retriever] knowledge_dir = {self.knowledge_dir}")

        results = []

        query_lower = query.lower()

        for file in self.knowledge_dir.glob("*.md"):

            print(f"[Retriever] file = {file.name}")

            content = file.read_text(encoding="utf-8")

            if file.name == "go.md":
                keywords = ["go", "channel", "goroutine"]

            elif file.name == "redis.md":
                keywords = ["redis", "setnx", "rdb", "aof"]

            elif file.name == "mysql.md":
                keywords = ["mysql", "mvcc", "b+tree", "索引"]

            else:
                keywords = []

            if any(keyword in query_lower for keyword in keywords):

                print(f"[Retriever] matched = {file.name}")

                results.append(
                    Document(
                        page_content=content,
                        metadata={
                            "source": file.name
                        }
                    )
                )

        print(f"[Retriever] result count = {len(results)}")

        return results

retriever = SimpleRetriever(
    knowledge_dir=KNOWLEDGE_DIR
)


# ============================================================
# 4. Tool
#    我们自己的 Tool
# ============================================================

@tool
def search_knowledge(query: str) -> str:
    """Search the technical knowledge base for relevant information."""

    documents = retriever.invoke(query)

    if not documents:
        return "知识库中没有找到相关信息。"

    results = []

    for document in documents:

        results.append(
            f"来源：{document.metadata.get('source')}\n"
            f"{document.page_content}"
        )

    return "\n\n---\n\n".join(results)


# ============================================================
# 5. Agent
#    LangChain 官方 create_agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[search_knowledge],
system_prompt="""
你是一个技术知识助手。

当用户的问题涉及技术知识时，
优先调用 search_knowledge 工具查询知识库。

如果第一次检索没有找到相关信息，
可以修改查询关键词并再次调用 search_knowledge。

如果经过检索仍然没有找到相关资料，
可以明确告诉用户知识库中没有相关内容，
不要把没有检索到的内容伪装成知识库内容。

如果问题不需要查询知识库，可以直接回答。
"""
)


# ============================================================
# 6. Run
# ============================================================

result = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": "Go 的 channel 有什么作用？",
            }
        ]
    }
)


# ============================================================
# 7. Print
# ============================================================

for message in result["messages"]:

    print("================================")
    print(type(message).__name__)
    print(message.content)

    if getattr(message, "tool_calls", None):

        print("Tool Calls:")

        for tool_call in message.tool_calls:
            print(tool_call)
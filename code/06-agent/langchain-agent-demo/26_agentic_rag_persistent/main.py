from pathlib import Path
import os

from dotenv import load_dotenv
from sentence_transformers import SentenceTransformer

from langchain.agents import create_agent
from langchain.tools import tool
from langchain_openai import ChatOpenAI
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS


# ============================================================
# 1. Environment
# ============================================================

load_dotenv()


# ============================================================
# 2. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    # ---------- 【重点】MiniMax M3 自定义参数放 extra_body ----------
    extra_body={
        "thinking": {"type": "disabled"},
        "reasoning_split": True
    },
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 3. Path
# ============================================================

BASE_DIR = Path(__file__).parent

VECTORSTORE_DIR = BASE_DIR / "vectorstore"


# ============================================================
# 4. Embedding
# ============================================================

class LocalEmbedding(Embeddings):

    def __init__(self):

        self.model = SentenceTransformer(
            "BAAI/bge-small-zh-v1.5"
        )

    def embed_documents(
        self,
        texts: list[str],
    ) -> list[list[float]]:

        vectors = self.model.encode(texts)

        return vectors.tolist()

    def embed_query(
        self,
        text: str,
    ) -> list[float]:

        vector = self.model.encode(text)

        return vector.tolist()


embedding = LocalEmbedding()


# ============================================================
# 5. Load Persistent VectorStore
# ============================================================

vectorstore = FAISS.load_local(
    str(VECTORSTORE_DIR),
    embedding,
    allow_dangerous_deserialization=True,
)

print("VectorStore 加载完成")


# ============================================================
# 6. Retriever
# ============================================================

retriever = vectorstore.as_retriever(
    search_kwargs={
        "k": 3
    }
)


# ============================================================
# 7. Search Tool
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
# 8. Agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[
        search_knowledge
    ],
    system_prompt="""
你是一个技术知识助手。

当用户的问题涉及技术知识时，
优先调用 search_knowledge 工具查询知识库。

根据工具返回的知识回答用户。

不要编造知识库中不存在的信息。

如果知识库没有相关内容，
明确告诉用户知识库中没有找到相关资料。
""",
)


# ============================================================
# 9. Invoke Agent
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
# 10. Print Trace
# ============================================================

for message in result["messages"]:

    print("\n================================")

    print(type(message).__name__)

    print(message.content)

    if getattr(message, "tool_calls", None):

        print("Tool Calls:")

        for tool_call in message.tool_calls:

            print(tool_call)
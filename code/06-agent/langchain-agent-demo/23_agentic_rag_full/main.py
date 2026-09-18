from pathlib import Path
import os

from dotenv import load_dotenv
from sentence_transformers import SentenceTransformer

from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS
from langchain_text_splitters import RecursiveCharacterTextSplitter

from langchain_openai import ChatOpenAI
from langchain.agents import create_agent
from langchain.tools import tool


# ============================================================
# 1. Model
# ============================================================

load_dotenv()

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
        # ---------- 【重点】MiniMax M3 自定义参数放 extra_body ----------
    extra_body={
        "thinking": {"type": "disabled"},
        "reasoning_split": True
    }
)


# ============================================================
# 2. Load Documents
# ============================================================

KNOWLEDGE_DIR = Path(__file__).parent / "knowledge"

documents = []

for file in KNOWLEDGE_DIR.glob("*.md"):

    content = file.read_text(encoding="utf-8")

    documents.append(
        Document(
            page_content=content,
            metadata={
                "source": file.name
            }
        )
    )

print("Documents:", len(documents))


# ============================================================
# 3. Chunking
# ============================================================

text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=80,
    chunk_overlap=20,
)

chunks = text_splitter.split_documents(
    documents
)

print("Chunks:", len(chunks))


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
# 5. VectorStore
# ============================================================

vectorstore = FAISS.from_documents(
    chunks,
    embedding,
)

print("VectorStore 创建完成")


# ============================================================
# 6. Retriever
# ============================================================

retriever = vectorstore.as_retriever(
    search_kwargs={
        "k": 3
    }
)


# ============================================================
# 7. Tool
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
    tools=[search_knowledge],
    system_prompt="""
你是一个技术知识助手。

当用户的问题涉及技术知识时，
优先调用 search_knowledge 工具查询知识库。

如果知识库返回了相关资料，
优先根据知识库内容回答。

不要把没有检索到的内容伪装成知识库内容。

如果问题不需要查询知识库，可以直接回答。
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
# 10. Print Message Trace
# ============================================================

for message in result["messages"]:

    print("\n================================")

    print(type(message).__name__)

    print(message.content)

    if getattr(message, "tool_calls", None):

        print("Tool Calls:")

        for tool_call in message.tool_calls:

            print(tool_call)
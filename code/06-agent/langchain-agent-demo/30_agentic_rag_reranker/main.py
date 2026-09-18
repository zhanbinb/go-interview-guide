from pathlib import Path
import os

from dotenv import load_dotenv
from sentence_transformers import SentenceTransformer

from langchain.agents import create_agent
from langchain.tools import tool
from langchain_openai import ChatOpenAI
from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS
from langchain_text_splitters import RecursiveCharacterTextSplitter


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
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 3. Path
# ============================================================

BASE_DIR = Path(__file__).parent

KNOWLEDGE_DIR = BASE_DIR / "knowledge"


# ============================================================
# 4. Load Documents + Metadata
# ============================================================

documents = []

for file in KNOWLEDGE_DIR.glob("*.md"):

    content = file.read_text(
        encoding="utf-8"
    )

    if file.stem == "go":
        category = "golang"
        title = "Go 并发编程"

    elif file.stem == "redis":
        category = "redis"
        title = "Redis 基础"

    elif file.stem == "mysql":
        category = "mysql"
        title = "MySQL 基础"

    else:
        category = "unknown"
        title = file.stem

    documents.append(
        Document(
            page_content=content,
            metadata={
                "source": file.name,
                "category": category,
                "title": title,
            },
        )
    )


print("Documents:", len(documents))


# ============================================================
# 5. Chunking
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
# 6. Embedding
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
# 7. VectorStore
# ============================================================

vectorstore = FAISS.from_documents(
    chunks,
    embedding,
)


# ============================================================
# 8. Reranker
# ============================================================

def rerank(
    query: str,
    documents: list[Document],
) -> list[Document]:

    query_keywords = [
        "channel",
        "goroutine",
        "通信",
        "并发",
        "缓冲",
        "发送",
        "接收",
    ]

    scored_documents = []

    for document in documents:

        content = document.page_content.lower()

        score = 0

        for keyword in query_keywords:

            if keyword.lower() in query.lower():

                if keyword.lower() in content:

                    score += 1

        scored_documents.append(
            (
                score,
                document,
            )
        )

    scored_documents.sort(
        key=lambda item: item[0],
        reverse=True,
    )

    return [
        document
        for score, document
        in scored_documents
    ]


# ============================================================
# 9. Search Tool
# ============================================================

@tool
def search_knowledge(
    query: str,
    category: str = "",
) -> str:
    """
    Search the technical knowledge base.

    category can be:
    golang, redis, mysql
    """

    # --------------------------------------------------------
    # Step 1: Vector Search
    # --------------------------------------------------------

    if category:

        candidates = vectorstore.similarity_search(
            query,
            k=5,
            filter={
                "category": category
            },
        )

    else:

        candidates = vectorstore.similarity_search(
            query,
            k=5,
        )

    print("\n[Vector Search] candidates:", len(candidates))


    # --------------------------------------------------------
    # Step 2: Rerank
    # --------------------------------------------------------

    reranked = rerank(
        query,
        candidates,
    )

    print(
        "[Reranker] results:",
        len(reranked),
    )


    # --------------------------------------------------------
    # Step 3: Take Top 3
    # --------------------------------------------------------

    results = reranked[:3]


    if not results:

        return "知识库中没有找到相关信息。"


    # --------------------------------------------------------
    # Step 4: Build Tool Result
    # --------------------------------------------------------

    output = []

    for document in results:

        output.append(
            f"来源：{document.metadata.get('source')}\n"
            f"分类：{document.metadata.get('category')}\n"
            f"{document.page_content}"
        )

    return "\n\n---\n\n".join(output)


# ============================================================
# 10. Agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[
        search_knowledge
    ],
    system_prompt="""
你是一个技术知识助手。

当用户的问题涉及技术知识时，
优先调用 search_knowledge 工具。

如果能够明确判断技术分类，
可以将 category 设置为：

golang
redis
mysql

如果无法确定分类，
category 可以留空。

根据知识库返回的内容回答问题。

不要编造知识库中不存在的信息。
""",
)


# ============================================================
# 11. Invoke Agent
# ============================================================

result = agent.invoke(
    {
        "messages": [
            {
                "role": "user",
                "content": "Go 的 channel 有什么作用?",
            }
        ]
    }
)


# ============================================================
# 12. Print Trace
# ============================================================

for message in result["messages"]:

    print("\n================================")

    print(type(message).__name__)

    print(message.content)

    if getattr(message, "tool_calls", None):

        print("Tool Calls:")

        for tool_call in message.tool_calls:

            print(tool_call)
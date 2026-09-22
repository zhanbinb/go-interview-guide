from pathlib import Path
import os

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI

from sentence_transformers import SentenceTransformer
from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS


# ========================================
# 1. Model
# ========================================

load_dotenv()

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ========================================
# 2. Embedding
# ========================================

class LocalEmbedding(Embeddings):

    def __init__(self):
        self.model = SentenceTransformer(
            "BAAI/bge-small-zh-v1.5"
        )

    def embed_documents(self, texts):

        return self.model.encode(
            texts,
            normalize_embeddings=True
        ).tolist()

    def embed_query(self, text):

        return self.model.encode(
            text,
            normalize_embeddings=True
        ).tolist()


embedding = LocalEmbedding()


# ========================================
# 3. Load Documents
# ========================================

BASE_DIR = Path(__file__).parent
KNOWLEDGE_DIR = BASE_DIR / "knowledge"

documents = []

for file in KNOWLEDGE_DIR.glob("*.md"):

    content = file.read_text(
        encoding="utf-8"
    )

    documents.append(
        Document(
            page_content=content,
            metadata={
                "source": file.name
            }
        )
    )


# ========================================
# 4. VectorStore
# ========================================

vectorstore = FAISS.from_documents(
    documents,
    embedding
)


# ========================================
# 5. Query Rewrite
# ========================================

def rewrite_query(query: str) -> str:

    prompt = f"""
你是一个 RAG 检索 Query 优化器。

请把用户的问题改写成更适合知识库检索的搜索 Query。

要求：
1. 保留用户原始意图
2. 补充必要的技术上下文
3. 使用明确的技术术语
4. 不要回答问题
5. 只输出改写后的 Query

用户问题：

{query}
"""

    response = model.invoke(prompt)

    return response.content.strip()


# ========================================
# 6. Retrieval
# ========================================

def search(query: str):

    results = vectorstore.similarity_search(
        query,
        k=3
    )

    return results


# ========================================
# 7. Test
# ========================================

query = "channel 怎么用？"

print("\n========== 原始 Query ==========")

print(query)


rewritten_query = rewrite_query(query)

print("\n========== Rewrite Query ==========")

print(rewritten_query)


results = search(rewritten_query)

print("\n========== Retrieval ==========")

for index, doc in enumerate(
    results,
    start=1
):

    print(
        f"\n--- Result {index} ---"
    )

    print(
        "source:",
        doc.metadata["source"]
    )

    print(
        doc.page_content
    )
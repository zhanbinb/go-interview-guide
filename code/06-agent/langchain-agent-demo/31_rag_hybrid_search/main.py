from pathlib import Path
import re

from sentence_transformers import SentenceTransformer
from langchain_core.documents import Document
from langchain_community.vectorstores import FAISS
from langchain_core.embeddings import Embeddings


BASE_DIR = Path(__file__).parent
KNOWLEDGE_DIR = BASE_DIR / "knowledge"


# =========================
# 1. Embedding
# =========================

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


# =========================
# 2. 加载 Document
# =========================

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


# =========================
# 3. 创建 VectorStore
# =========================

vectorstore = FAISS.from_documents(
    documents,
    embedding
)


# =========================
# 4. Keyword Search
# =========================

def keyword_search(query, documents, k=3):

    # 非严格的教学版关键词拆分
    keywords = re.findall(
        r"[\u4e00-\u9fff]+|[a-zA-Z0-9_]+",
        query.lower()
    )

    results = []

    for document in documents:

        content = document.page_content.lower()

        score = 0

        for keyword in keywords:

            if keyword in content:
                score += 1

        if score > 0:

            results.append(
                (score, document)
            )

    results.sort(
        key=lambda x: x[0],
        reverse=True
    )

    return results[:k]


# =========================
# 5. Vector Search
# =========================

def vector_search(query, k=3):

    return vectorstore.similarity_search(
        query,
        k=k
    )


# =========================
# 6. Hybrid Search
# =========================

def hybrid_search(query):

    keyword_results = keyword_search(
        query,
        documents,
        k=3
    )

    vector_results = vector_search(
        query,
        k=3
    )

    print("\n========== Keyword Search ==========")

    for rank, (score, doc) in enumerate(
        keyword_results,
        start=1
    ):
        print(
            f"{rank}. "
            f"{doc.metadata['source']} "
            f"keyword_score={score}"
        )

    print("\n========== Vector Search ==========")

    for rank, doc in enumerate(
        vector_results,
        start=1
    ):
        print(
            f"{rank}. "
            f"{doc.metadata['source']}"
        )

    # =========================
    # 简单合并
    # =========================

    merged = {}

    # Keyword 排名贡献
    for rank, (score, doc) in enumerate(
        keyword_results,
        start=1
    ):

        key = doc.metadata["source"]

        merged[key] = merged.get(key, 0) + (
            1.0 / rank
        )

    # Vector 排名贡献
    for rank, doc in enumerate(
        vector_results,
        start=1
    ):

        key = doc.metadata["source"]

        merged[key] = merged.get(key, 0) + (
            1.0 / rank
        )

    final_results = sorted(
        merged.items(),
        key=lambda x: x[1],
        reverse=True
    )

    print("\n========== Hybrid Search ==========")

    for rank, (source, score) in enumerate(
        final_results,
        start=1
    ):

        print(
            f"{rank}. "
            f"{source} "
            f"hybrid_score={score:.3f}"
        )


# =========================
# 7. Test
# =========================

query = "Go channel goroutine 通信"

hybrid_search(query)
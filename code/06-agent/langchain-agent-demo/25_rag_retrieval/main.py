from pathlib import Path

from sentence_transformers import SentenceTransformer

from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS


# ============================================================
# 1. Path
# ============================================================

BASE_DIR = Path(__file__).parent

VECTORSTORE_DIR = BASE_DIR / "vectorstore"


# ============================================================
# 2. Embedding
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
# 3. Load VectorStore
# ============================================================

vectorstore = FAISS.load_local(
    str(VECTORSTORE_DIR),
    embedding,
    allow_dangerous_deserialization=True,
)

print("VectorStore 加载完成")


# ============================================================
# 4. Create Retriever
# ============================================================

retriever = vectorstore.as_retriever(
    search_kwargs={
        "k": 3
    }
)


# ============================================================
# 5. Retrieval
# ============================================================

query = "Go 的 channel 有什么作用？"

results = retriever.invoke(query)


# ============================================================
# 6. Print Results
# ============================================================

print("\nQuery:", query)

for index, document in enumerate(results):

    print("\n================================")
    print(f"Rank: {index + 1}")
    print(
        f"Source: {document.metadata.get('source')}"
    )
    print(
        f"Length: {len(document.page_content)}"
    )
    print(document.page_content)
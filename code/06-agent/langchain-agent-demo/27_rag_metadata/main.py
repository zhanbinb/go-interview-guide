from pathlib import Path

from sentence_transformers import SentenceTransformer

from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS


# ============================================================
# 1. Path
# ============================================================

BASE_DIR = Path(__file__).parent

KNOWLEDGE_DIR = BASE_DIR / "knowledge"


# ============================================================
# 2. Load Documents
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

for document in documents:

    print("\n--------------------")

    print(
        "Source:",
        document.metadata["source"]
    )

    print(
        "Category:",
        document.metadata["category"]
    )

    print(
        "Title:",
        document.metadata["title"]
    )


# ============================================================
# 3. Chunking
# ============================================================

from langchain_text_splitters import RecursiveCharacterTextSplitter

text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=80,
    chunk_overlap=20,
)

chunks = text_splitter.split_documents(
    documents
)

print("\nChunks:", len(chunks))


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
# 5. Build VectorStore
# ============================================================

vectorstore = FAISS.from_documents(
    chunks,
    embedding,
)


# ============================================================
# 6. Vector Search
# ============================================================

query = "Go 的 channel 有什么作用？"

results = vectorstore.similarity_search(
    query,
    k=3,
    filter={"category": "golang"},
)

print("\n\n================ Vector Search ================")

for index, document in enumerate(results):

    print("\n--------------------------------")

    print(
        f"Rank: {index + 1}"
    )

    print(
        "Source:",
        document.metadata.get("source")
    )

    print(
        "Category:",
        document.metadata.get("category")
    )

    print(
        document.page_content
    )
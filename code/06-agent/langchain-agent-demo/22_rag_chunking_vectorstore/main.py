from pathlib import Path

from sentence_transformers import SentenceTransformer

from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS
from langchain_text_splitters import RecursiveCharacterTextSplitter


# =========================
# 1. Knowledge Base
# =========================

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


# =========================
# 2. Document → Chunks
# =========================

text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=80,
    chunk_overlap=20,
)

chunks = text_splitter.split_documents(
    documents
)

print("Chunks:", len(chunks))

for index, chunk in enumerate(chunks):

    print("\n================================")
    print(f"Chunk: {index + 1}")
    print(f"Source: {chunk.metadata.get('source')}")
    print(f"Length: {len(chunk.page_content)}")
    print(chunk.page_content)


# =========================
# 3. Embedding
# =========================

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


# =========================
# 4. Chunks → VectorStore
# =========================

vectorstore = FAISS.from_documents(
    chunks,
    embedding,
)

print("\nVectorStore 创建完成")


# =========================
# 5. VectorStore → Retriever
# =========================

retriever = vectorstore.as_retriever(
    search_kwargs={
        "k": 3
    }
)


# =========================
# 6. Query
# =========================

query = "Go 的 channel 有什么作用？"


# =========================
# 7. Similarity Search
# =========================

results = retriever.invoke(query)


# =========================
# 8. Print Results
# =========================

print("\n========== 检索结果 ==========")

for index, document in enumerate(results):

    print("\n================================")
    print(f"Rank: {index + 1}")
    print(
        f"Source: {document.metadata.get('source')}"
    )
    print(document.page_content)
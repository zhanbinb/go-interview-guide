from pathlib import Path

from sentence_transformers import SentenceTransformer

from langchain_core.documents import Document
from langchain_core.embeddings import Embeddings
from langchain_community.vectorstores import FAISS
from langchain_text_splitters import RecursiveCharacterTextSplitter


# ============================================================
# 1. Path
# ============================================================

BASE_DIR = Path(__file__).parent

KNOWLEDGE_DIR = BASE_DIR / "knowledge"

VECTORSTORE_DIR = BASE_DIR / "vectorstore"


# ============================================================
# 2. Load Documents
# ============================================================

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
# 5. Build VectorStore
# ============================================================

vectorstore = FAISS.from_documents(
    chunks,
    embedding,
)

print("VectorStore 创建完成")


# ============================================================
# 6. Save VectorStore
# ============================================================

VECTORSTORE_DIR.mkdir(
    parents=True,
    exist_ok=True,
)

vectorstore.save_local(
    str(VECTORSTORE_DIR)
)

print(
    "VectorStore 已保存到:",
    VECTORSTORE_DIR
)
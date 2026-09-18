from pathlib import Path

from langchain_core.documents import Document
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
# 2. Text Splitter
# =========================

text_splitter = RecursiveCharacterTextSplitter(
    chunk_size=200,
    chunk_overlap=50,
)


# =========================
# 3. Document → Chunks
# =========================

chunks = text_splitter.split_documents(
    documents
)


# =========================
# 4. Print Chunks
# =========================

print("Chunks:", len(chunks))

for index, chunk in enumerate(chunks):

    print("\n================================")
    print(f"Chunk: {index + 1}")
    print(f"Source: {chunk.metadata.get('source')}")
    print(f"Length: {len(chunk.page_content)}")
    print(chunk.page_content)
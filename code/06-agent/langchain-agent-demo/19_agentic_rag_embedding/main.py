from sentence_transformers import SentenceTransformer
import numpy as np


# ============================================================
# 1. Embedding Model
# ============================================================

model = SentenceTransformer(
    "BAAI/bge-small-zh-v1.5"
)


# ============================================================
# 2. Documents
# ============================================================

documents = [
    "Go 使用 goroutine 实现并发，channel 用于 goroutine 之间的数据通信。",
    "Redis 是一个基于内存的高性能键值数据库，支持 String、Hash、List、Set 和 Sorted Set。",
    "MySQL InnoDB 使用 B+Tree 索引，并支持 MVCC 多版本并发控制。",
]

query = "Go 的 channel 有什么作用？"


# ============================================================
# 3. Embedding
# ============================================================

document_vectors = model.encode(documents)
query_vector = model.encode(query)


# ============================================================
# 4. Similarity
# ============================================================

def cosine_similarity(a, b):

    return np.dot(a, b) / (
        np.linalg.norm(a) * np.linalg.norm(b)
    )


# ============================================================
# 5. Print
# ============================================================

print("Query:")
print(query)

print("\nVector Dimension:")
print(len(query_vector))

print("\nSimilarity:")

for document, vector in zip(documents, document_vectors):

    score = cosine_similarity(
        query_vector,
        vector
    )

    print(f"{score:.4f} -> {document}")
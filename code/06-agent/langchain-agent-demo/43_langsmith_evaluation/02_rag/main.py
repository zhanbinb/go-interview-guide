
from dotenv import load_dotenv
from langsmith import Client, evaluate
from langsmith.schemas import Example, Run
from langchain_openai import ChatOpenAI
from langsmith import traceable


# ============================================================
# 1. 加载环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. LangSmith
# ============================================================

client = Client()

DATASET_NAME = "agent-learning-rag-eval-v1"


# ============================================================
# 3. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 4. Knowledge Base
#
# 教学 Demo。
#
# 实际生产环境可以替换为：
#
# BM25
# Vector Search
# Hybrid Search
# Reranker
# Elasticsearch
# FAISS
# Milvus
# pgvector
# ...
# ============================================================

documents = [
    {
        "id": "doc-go-channel",
        "title": "Go Channel",
        "content": (
            "Go 的 channel 用于 goroutine 之间进行通信。"
            "channel 可以是无缓冲 channel，也可以是有缓冲 channel。"
        ),
        "keywords": [
            "Go",
            "channel",
            "goroutine",
        ],
    },
    {
        "id": "doc-go-goroutine",
        "title": "Go Goroutine",
        "content": (
            "goroutine 是 Go 中轻量级的并发执行单元，"
            "通常通过 go 关键字启动。"
        ),
        "keywords": [
            "goroutine",
            "并发",
        ],
    },
    {
        "id": "doc-redis-lock",
        "title": "Redis SETNX",
        "content": (
            "Redis SETNX 可以用于实现分布式锁和幂等控制。"
            "通过 NX 参数可以保证 Key 不存在时才写入。"
        ),
        "keywords": [
            "Redis",
            "SETNX",
            "分布式锁",
            "幂等",
        ],
    },
    {
        "id": "doc-mysql-mvcc",
        "title": "MySQL MVCC",
        "content": (
            "MySQL InnoDB 通过 MVCC 实现多版本并发控制，"
            "用于提高事务并发能力。"
        ),
        "keywords": [
            "MySQL",
            "MVCC",
            "事务",
        ],
    },
]


# ============================================================
# 5. RAG Retrieval
#
# @traceable：
#
# LangSmith 会把这个函数记录到 Trace 中。
#
# 之后你可以在 Trace 中看到：
#
# RAG Pipeline
#      ↓
# Retrieval
#      ↓
# top-k documents
# ============================================================

@traceable(name="RAG Retrieval")
def retrieve_documents(
    query: str,
    top_k: int = 3,
) -> list[dict]:

    results = []

    for document in documents:

        score = 0

        for keyword in document["keywords"]:

            if keyword.lower() in query.lower():
                score += 1

        if score > 0:

            results.append(
                {
                    "id": document["id"],
                    "title": document["title"],
                    "content": document["content"],
                    "score": score,
                }
            )

    # --------------------------------------------------------
    # 按分数排序
    # --------------------------------------------------------

    results.sort(
        key=lambda item: item["score"],
        reverse=True,
    )

    return results[:top_k]


# ============================================================
# 6. RAG Answer
# ============================================================

def generate_answer(
    question: str,
    retrieved_documents: list[dict],
) -> str:

    # --------------------------------------------------------
    # 构建 Context
    # --------------------------------------------------------

    context_parts = []

    for document in retrieved_documents:

        context_parts.append(
            f"【{document['title']}】\n"
            f"{document['content']}"
        )

    context = "\n\n".join(
        context_parts
    )

    prompt = f"""
你是一个企业内部知识助手。

请根据下面的知识库内容回答用户问题。

要求：

1. 优先根据知识库回答。
2. 不要编造知识库中不存在的信息。
3. 如果知识库没有相关信息，明确说明。
4. 回答简洁。

知识库：

{context}

用户问题：

{question}
"""

    response = model.invoke(prompt)

    return response.content


# ============================================================
# 7. Evaluation Dataset
# ============================================================

EVALUATION_EXAMPLES = [
    {
        "inputs": {
            "question": "Go 的 channel 有什么作用？",
        },
        "outputs": {
            "answer": (
                "Go 的 channel 用于 goroutine 之间进行通信。"
            ),
            "expected_doc_ids": [
                "doc-go-channel",
            ],
        },
        "metadata": {
            "category": "go",
        },
    },
    {
        "inputs": {
            "question": "Redis SETNX 可以做什么？",
        },
        "outputs": {
            "answer": (
                "Redis SETNX 可以用于分布式锁和幂等控制。"
            ),
            "expected_doc_ids": [
                "doc-redis-lock",
            ],
        },
        "metadata": {
            "category": "redis",
        },
    },
    {
        "inputs": {
            "question": "MySQL MVCC 是什么？",
        },
        "outputs": {
            "answer": (
                "MVCC 是 MySQL InnoDB 的多版本并发控制机制。"
            ),
            "expected_doc_ids": [
                "doc-mysql-mvcc",
            ],
        },
        "metadata": {
            "category": "mysql",
        },
    },
]


# ============================================================
# 8. 创建 Dataset
# ============================================================

def ensure_dataset():

    if client.has_dataset(
        dataset_name=DATASET_NAME
    ):

        print(
            f"Dataset 已存在："
            f"{DATASET_NAME}"
        )

        return

    dataset = client.create_dataset(
        dataset_name=DATASET_NAME,
        description=(
            "RAG Retrieval Evaluation Dataset"
        ),
    )

    client.create_examples(
        dataset_id=dataset.id,
        examples=EVALUATION_EXAMPLES,
    )

    print()
    print("=" * 70)
    print("创建 RAG Dataset")
    print("=" * 70)

    print(
        f"Dataset: {dataset.name}"
    )

    print(
        f"Examples: "
        f"{len(EVALUATION_EXAMPLES)}"
    )


# ============================================================
# 9. Target System
#
# LangSmith Evaluation 会执行：
#
# Dataset Input
#       ↓
# target()
#       ↓
# Retrieval
#       ↓
# LLM
#
# 最终返回：
#
# answer
# retrieved_doc_ids
# ============================================================

def target(
    inputs: dict,
) -> dict:

    question = inputs["question"]

    # --------------------------------------------------------
    # Retrieval
    # --------------------------------------------------------

    retrieved_documents = retrieve_documents(
        query=question,
        top_k=3,
    )

    retrieved_doc_ids = [
        document["id"]
        for document in retrieved_documents
    ]

    # --------------------------------------------------------
    # Generation
    # --------------------------------------------------------

    answer = generate_answer(
        question=question,
        retrieved_documents=retrieved_documents,
    )

    return {
        "answer": answer,
        "retrieved_doc_ids": retrieved_doc_ids,
    }


# ============================================================
# 10. Recall@K Evaluator
# ============================================================

def recall_at_k_evaluator(
    run: Run,
    example: Example,
) -> dict:

    # --------------------------------------------------------
    # 实际召回
    # --------------------------------------------------------

    actual_doc_ids = (
        run.outputs.get(
            "retrieved_doc_ids",
            [],
        )
        if run.outputs
        else []
    )

    # --------------------------------------------------------
    # 正确文档
    # --------------------------------------------------------

    expected_doc_ids = (
        example.outputs.get(
            "expected_doc_ids",
            [],
        )
        if example.outputs
        else []
    )

    # --------------------------------------------------------
    # 计算 Recall@K
    #
    # Recall =
    #
    # 命中的正确文档数量
    # ------------------
    # 期望正确文档总数
    # --------------------------------------------------------

    if not expected_doc_ids:

        return {
            "key": "recall_at_k",
            "score": 0.0,
            "comment": "没有 expected_doc_ids",
        }

    actual_set = set(
        actual_doc_ids
    )

    expected_set = set(
        expected_doc_ids
    )

    hit_count = len(
        actual_set & expected_set
    )

    recall = (
        hit_count
        / len(expected_set)
    )

    return {
        "key": "recall_at_k",
        "score": recall,
        "comment": (
            f"actual={actual_doc_ids}, "
            f"expected={expected_doc_ids}, "
            f"hits={hit_count}"
        ),
    }


# ============================================================
# 11. Answer Keyword Evaluator
#
# 这是一个非常简单的确定性答案检查。
#
# 重点仍然是演示：
#
# 一个 Dataset
#      ↓
# 多个 Evaluator
#
# ============================================================

def answer_keyword_evaluator(
    run: Run,
    example: Example,
) -> dict:

    actual_answer = (
        run.outputs.get(
            "answer",
            "",
        )
        if run.outputs
        else ""
    )

    reference_answer = (
        example.outputs.get(
            "answer",
            "",
        )
        if example.outputs
        else ""
    )

    # --------------------------------------------------------
    # 提取一些关键业务词
    #
    # 教学 Demo。
    # 实际项目可使用：
    #
    # Exact Match
    # Semantic Similarity
    # LLM Judge
    # Business Rule
    # --------------------------------------------------------

    keywords = []

    for keyword in [
        "channel",
        "goroutine",
        "分布式锁",
        "幂等",
        "MVCC",
        "多版本并发控制",
    ]:

        if keyword in reference_answer:
            keywords.append(keyword)

    if not keywords:

        return {
            "key": "answer_keyword_match",
            "score": 1.0,
            "comment": "没有可检查关键词",
        }

    matched = [
        keyword
        for keyword in keywords
        if keyword in actual_answer
    ]

    score = (
        len(matched)
        / len(keywords)
    )

    return {
        "key": "answer_keyword_match",
        "score": score,
        "comment": (
            f"matched={matched}, "
            f"expected={keywords}"
        ),
    }


# ============================================================
# 12. Run Evaluation
# ============================================================

def run_evaluation():

    ensure_dataset()

    print()
    print("=" * 70)
    print("LangSmith RAG Evaluation")
    print("=" * 70)

    results = evaluate(
        target,
        data=DATASET_NAME,
        evaluators=[
            recall_at_k_evaluator,
            answer_keyword_evaluator,
        ],
        experiment_prefix=(
            "rag-evaluation"
        ),
        description=(
            "RAG Retrieval + Answer Evaluation"
        ),
        metadata={
            "model": "MiniMax-M3",
            "retrieval": "keyword-demo",
            "top_k": 3,
            "version": "v1",
        },
        max_concurrency=1,
    )

    print()
    print("=" * 70)
    print("Evaluation 完成")
    print("=" * 70)

    print(results)


# ============================================================
# 13. Main
# ============================================================

if __name__ == "__main__":
    run_evaluation()

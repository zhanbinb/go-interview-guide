
import uuid

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from langsmith import traceable


# ============================================================
# 1. 加载环境变量
#
# .env 中已经配置：
#
# LANGSMITH_TRACING=true
# LANGSMITH_API_KEY=...
# LANGSMITH_PROJECT=agent-learning
#
# MiniMax：
#
# OPENAI_API_KEY=...
# OPENAI_BASE_URL=...
# ============================================================

load_dotenv()


# ============================================================
# 2. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 3. Knowledge Base
#
# 这里为了演示 Observability，
# 不再重新搭建 FAISS。
#
# 实际项目中可以替换成：
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
]


# ============================================================
# 4. RAG Retrieval
#
# 重点：
#
# @traceable
#
# 这个函数不是 LangChain Tool。
# 它只是我们自己的业务代码。
#
# 通过 @traceable，
# LangSmith 会把它记录成一个子 Trace。
# ============================================================

@traceable(name="Knowledge Retrieval")
def retrieve_knowledge(
    query: str,
    top_k: int = 2,
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
    # 根据匹配分数排序
    # --------------------------------------------------------

    results.sort(
        key=lambda item: item["score"],
        reverse=True,
    )

    return results[:top_k]


# ============================================================
# 5. RAG Tool
#
# Agent 调用的是这个 Tool。
#
# Tool 内部：
#
# search_knowledge
#       ↓
# retrieve_knowledge
#
# 所以 LangSmith 中会形成：
#
# Tool
#   └── Knowledge Retrieval
# ============================================================

@tool
def search_knowledge(query: str) -> str:
    """
    搜索内部技术知识库。

    当用户询问 Go、Redis、MySQL、RAG 等
    技术知识时使用。
    """

    documents_found = retrieve_knowledge(
        query=query,
        top_k=2,
    )

    if not documents_found:

        return "知识库中没有找到相关信息。"

    result = []

    for document in documents_found:

        result.append(
            f"【{document['title']}】\n"
            f"{document['content']}"
        )

    return "\n\n".join(result)


# ============================================================
# 6. Order Tool
# ============================================================

@tool
def query_order(order_id: str) -> dict:
    """
    查询订单信息。
    """

    orders = {
        "10001": {
            "order_id": "10001",
            "status": "已发货",
            "amount": 3999,
        },
        "10002": {
            "order_id": "10002",
            "status": "已完成",
            "amount": 1999,
        },
    }

    return orders.get(
        order_id,
        {
            "error": "订单不存在",
        },
    )


# ============================================================
# 7. Agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[
        query_order,
        search_knowledge,
    ],
    system_prompt="""
你是一个企业内部智能助手。

工具规则：

1. 用户询问订单状态、订单金额等信息时，
   使用 query_order。

2. 用户询问 Go、Redis、MySQL、RAG
   等技术知识时，
   使用 search_knowledge。

3. 不要编造业务数据。

4. 对知识库问题，优先依据 search_knowledge
   返回的信息回答。

5. 如果问题包含多个任务，
   可以根据需要调用多个 Tool。
""",
)


# ============================================================
# 8. Agent Run
# ============================================================

def run_agent(user_message: str):

    # --------------------------------------------------------
    # 业务请求 ID
    #
    # 注意：
    #
    # LangSmith 自己会有 Run / Trace 信息。
    #
    # request_id 是我们业务系统自己的标识。
    # --------------------------------------------------------

    request_id = str(
        uuid.uuid4()
    )

    print()
    print("=" * 70)
    print("Agent Request")
    print("=" * 70)

    print(
        f"request_id: {request_id}"
    )

    print(
        f"user_message: {user_message}"
    )

    # --------------------------------------------------------
    # Agent Invoke
    #
    # metadata / tags 会进入 LangSmith，
    # 方便后面筛选 Trace。
    # --------------------------------------------------------

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": user_message,
                }
            ]
        },
        config={
            "metadata": {
                "request_id": request_id,
                "user_id": "user1001",
                "environment": "local",
            },
            "tags": [
                "agent-learning",
                "42-3",
            ],
        },
    )

    print()
    print("=" * 70)
    print("Final Answer")
    print("=" * 70)

    print(
        result["messages"][-1].content
    )

    return result


# ============================================================
# 9. Main
# ============================================================

if __name__ == "__main__":

    run_agent(
        """
        请帮我查询订单10001的状态，
        另外解释一下 Go 的 channel 是什么？
        """
    )

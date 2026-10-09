
from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 加载环境变量
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
# 3. Tools
# ============================================================

@tool
def query_order(order_id: str) -> dict:
    """
    查询订单信息。
    当用户询问订单状态、订单金额等订单信息时使用。
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


@tool
def query_payment(order_id: str) -> dict:
    """
    查询订单支付信息。
    当用户询问支付状态、支付方式等信息时使用。
    """

    payments = {
        "10001": {
            "order_id": "10001",
            "status": "已支付",
            "method": "微信支付",
        },
        "10002": {
            "order_id": "10002",
            "status": "已支付",
            "method": "支付宝",
        },
    }

    return payments.get(
        order_id,
        {
            "error": "支付记录不存在",
        },
    )


# ============================================================
# 4. Agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[
        query_order,
        query_payment,
    ],
    system_prompt="""
你是一个订单业务助手。

规则：

1. 用户询问订单状态、订单金额等信息时，使用 query_order。
2. 用户询问支付状态、支付方式等信息时，使用 query_payment。
3. 不要编造业务数据。
4. 工具参数必须来自用户问题。
""",
)


# ============================================================
# ============================================================
#                    第一部分：Tool Evaluation
# ============================================================
# ============================================================


# ============================================================
# 5. Tool Evaluation Test Cases
# ============================================================

tool_test_cases = [
    {
        "name": "订单状态查询",
        "input": "请帮我查询订单10001现在是什么状态？",
        "expected_tool": "query_order",
        "expected_args": {
            "order_id": "10001",
        },
    },
    {
        "name": "支付状态查询",
        "input": "订单10001支付成功了吗？",
        "expected_tool": "query_payment",
        "expected_args": {
            "order_id": "10001",
        },
    },
]


# ============================================================
# 6. 执行 Agent
# ============================================================

def run_agent(user_message: str):
    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": user_message,
                }
            ]
        }
    )

    return result


# ============================================================
# 7. 提取所有 Tool Calls
# ============================================================

def extract_tool_calls(result) -> list[dict]:
    """
    从 Agent 最终结果中提取 AIMessage 的 tool_calls。
    """

    tool_calls = []

    for message in result["messages"]:

        message_tool_calls = getattr(
            message,
            "tool_calls",
            None,
        )

        if not message_tool_calls:
            continue

        tool_calls.extend(message_tool_calls)

    return tool_calls


# ============================================================
# 8. Tool Evaluation
# ============================================================

def evaluate_tool_call(
    result,
    expected_tool: str,
    expected_args: dict,
) -> dict:

    tool_calls = extract_tool_calls(result)

    # --------------------------------------------------------
    # 没有调用 Tool
    # --------------------------------------------------------

    if not tool_calls:
        return {
            "passed": False,
            "reason": "Agent 没有调用任何 Tool",
        }

    # --------------------------------------------------------
    # 这里只检查第一个 Tool Call
    #
    # 教学 Demo 简化处理。
    # 实际 Agent 可能有多个 Tool Call。
    # --------------------------------------------------------

    actual_tool_call = tool_calls[0]

    actual_tool = actual_tool_call.get(
        "name"
    )

    actual_args = actual_tool_call.get(
        "args",
        {},
    )

    # --------------------------------------------------------
    # Tool Name
    # --------------------------------------------------------

    tool_name_correct = (
        actual_tool == expected_tool
    )

    # --------------------------------------------------------
    # Tool Args
    # --------------------------------------------------------

    args_correct = True

    for key, expected_value in expected_args.items():

        actual_value = actual_args.get(key)

        if actual_value != expected_value:
            args_correct = False
            break

    passed = (
        tool_name_correct
        and args_correct
    )

    return {
        "passed": passed,
        "actual_tool": actual_tool,
        "expected_tool": expected_tool,
        "actual_args": actual_args,
        "expected_args": expected_args,
        "tool_name_correct": tool_name_correct,
        "args_correct": args_correct,
    }


# ============================================================
# 9. Run Tool Evaluation
# ============================================================

def run_tool_evaluation():

    print()
    print("=" * 70)
    print("Tool Calling Evaluation")
    print("=" * 70)

    total = len(tool_test_cases)
    passed = 0

    for index, test_case in enumerate(
        tool_test_cases,
        start=1,
    ):

        print()
        print("-" * 70)
        print(
            f"Test Case {index}: "
            f"{test_case['name']}"
        )
        print("-" * 70)

        print("Input:")
        print(test_case["input"])

        result = run_agent(
            test_case["input"]
        )

        tool_calls = extract_tool_calls(result)

        print()
        print("Tool Calls:")

        if not tool_calls:
            print("没有 Tool Call")

        for tool_call in tool_calls:
            print(
                f"Tool: {tool_call.get('name')}"
            )
            print(
                f"Args: {tool_call.get('args')}"
            )

        evaluation = evaluate_tool_call(
            result=result,
            expected_tool=test_case["expected_tool"],
            expected_args=test_case["expected_args"],
        )

        print()
        print("Evaluation:")

        print(
            "Tool Name:",
            "PASS"
            if evaluation.get("tool_name_correct")
            else "FAIL",
        )

        print(
            "Tool Args:",
            "PASS"
            if evaluation.get("args_correct")
            else "FAIL",
        )

        print()

        if evaluation["passed"]:
            print("Result: PASS")
            passed += 1
        else:
            print("Result: FAIL")
            print(
                "Reason:",
                evaluation.get(
                    "reason",
                    "Tool 名称或参数错误",
                ),
            )

    score = passed / total

    print()
    print("=" * 70)
    print("Tool Evaluation Summary")
    print("=" * 70)

    print(f"Total : {total}")
    print(f"Passed: {passed}")
    print(f"Failed: {total - passed}")
    print(f"Score : {score:.2%}")


# ============================================================
# ============================================================
#                     第二部分：RAG Evaluation
# ============================================================
# ============================================================


# ============================================================
# 10. 模拟知识库
#
# 这里不引入 FAISS。
#
# 目的不是重新学习 RAG，
# 而是学习：
#
#     Retriever
#         ↓
#     Evaluation
#
# 实际项目中，这里可以直接替换成你前面学过的
# BM25 / FAISS / Hybrid Search / Reranker。
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
# 11. 简单 Retriever
#
# 教学版本：
#
# 根据 query 与文档 keywords 的匹配数量排序。
#
# 注意：
# 这不是生产 Retriever，只是为了演示 Evaluation。
# ============================================================

def retrieve(
    query: str,
    top_k: int = 3,
) -> list[dict]:

    results = []

    for document in documents:

        score = 0

        for keyword in document["keywords"]:

            if keyword.lower() in query.lower():
                score += 1

        results.append(
            {
                "id": document["id"],
                "title": document["title"],
                "content": document["content"],
                "score": score,
            }
        )

    # 分数高的排前面
    results.sort(
        key=lambda item: item["score"],
        reverse=True,
    )

    return results[:top_k]


# ============================================================
# 12. RAG Evaluation Test Cases
# ============================================================

rag_test_cases = [
    {
        "name": "Go Channel",
        "query": "Go 的 channel 有什么作用？",
        "expected_doc_ids": [
            "doc-go-channel",
        ],
    },
    {
        "name": "Redis SETNX",
        "query": "Redis SETNX 可以做什么？",
        "expected_doc_ids": [
            "doc-redis-lock",
        ],
    },
    {
        "name": "MySQL MVCC",
        "query": "MySQL MVCC 是什么？",
        "expected_doc_ids": [
            "doc-mysql-mvcc",
        ],
    },
]


# ============================================================
# 13. Recall@K
# ============================================================

def recall_at_k(
    retrieved_docs: list[dict],
    expected_doc_ids: list[str],
) -> float:

    retrieved_ids = {
        document["id"]
        for document in retrieved_docs
    }

    hit_count = 0

    for expected_id in expected_doc_ids:

        if expected_id in retrieved_ids:
            hit_count += 1

    if not expected_doc_ids:
        return 0.0

    return (
        hit_count
        / len(expected_doc_ids)
    )


# ============================================================
# 14. Run RAG Evaluation
# ============================================================

def run_rag_evaluation():

    print()
    print("=" * 70)
    print("RAG Retrieval Evaluation")
    print("=" * 70)

    total = len(rag_test_cases)

    recall_values = []

    for index, test_case in enumerate(
        rag_test_cases,
        start=1,
    ):

        print()
        print("-" * 70)
        print(
            f"Test Case {index}: "
            f"{test_case['name']}"
        )
        print("-" * 70)

        query = test_case["query"]

        expected_doc_ids = (
            test_case["expected_doc_ids"]
        )

        print("Query:")
        print(query)

        print()
        print("Expected Documents:")
        print(expected_doc_ids)

        # ----------------------------------------------------
        # Retrieval
        # ----------------------------------------------------

        retrieved_docs = retrieve(
            query=query,
            top_k=3,
        )

        print()
        print("Retrieved Documents:")

        for rank, document in enumerate(
            retrieved_docs,
            start=1,
        ):

            print(
                f"{rank}. "
                f"{document['id']} "
                f"(score={document['score']})"
            )

        # ----------------------------------------------------
        # Evaluation
        # ----------------------------------------------------

        recall = recall_at_k(
            retrieved_docs=retrieved_docs,
            expected_doc_ids=expected_doc_ids,
        )

        recall_values.append(recall)

        print()
        print(
            f"Recall@3: {recall:.2%}"
        )

        if recall == 1.0:
            print("Result: PASS")
        else:
            print("Result: FAIL")

    # --------------------------------------------------------
    # Summary
    # --------------------------------------------------------

    average_recall = (
        sum(recall_values)
        / total
        if total > 0
        else 0
    )

    print()
    print("=" * 70)
    print("RAG Evaluation Summary")
    print("=" * 70)

    print(
        f"Test Cases: {total}"
    )

    print(
        f"Average Recall@3: "
        f"{average_recall:.2%}"
    )


# ============================================================
# 15. Main
# ============================================================

if __name__ == "__main__":

    # --------------------------------------------------------
    # Tool Evaluation
    # --------------------------------------------------------

    run_tool_evaluation()

    # --------------------------------------------------------
    # RAG Evaluation
    # --------------------------------------------------------

    run_rag_evaluation()

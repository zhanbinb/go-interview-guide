
import re

from dotenv import load_dotenv
from langsmith import Client
from langsmith.schemas import Example, Run
from langchain.agents import create_agent
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 加载环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. LangSmith Client
# ============================================================

langsmith_client = Client()


# ============================================================
# 3. Dataset
# ============================================================

DATASET_NAME = "agent-learning-order-eval-v1"


EVALUATION_EXAMPLES = [
    {
        "inputs": {
            "question": "请帮我查询订单10001现在是什么状态？",
        },
        "outputs": {
            "answer": "订单10001当前状态是已发货。",
        },
        "metadata": {
            "category": "order_status",
        },
    },
    {
        "inputs": {
            "question": "订单10002多少钱？",
        },
        "outputs": {
            "answer": "订单10002的金额是1999元。",
        },
        "metadata": {
            "category": "order_amount",
        },
    },
    {
        "inputs": {
            "question": "订单10001支付成功了吗？",
        },
        "outputs": {
            "answer": "订单10001已经支付成功，支付方式是微信支付。",
        },
        "metadata": {
            "category": "payment",
        },
    },
]


def ensure_dataset():
    """
    如果 Dataset 不存在，就创建。

    已经存在：
        直接使用。

    这样脚本可以重复运行，
    不会每次都创建一个新的 Dataset。
    """

    if langsmith_client.has_dataset(
        dataset_name=DATASET_NAME
    ):
        print(
            f"Dataset 已存在: {DATASET_NAME}"
        )
        return

    dataset = langsmith_client.create_dataset(
        dataset_name=DATASET_NAME,
        description=(
            "Agent 订单查询 Evaluation Dataset"
        ),
    )

    langsmith_client.create_examples(
        dataset_id=dataset.id,
        examples=EVALUATION_EXAMPLES,
    )

    print()
    print("=" * 70)
    print("创建 Dataset")
    print("=" * 70)

    print(
        f"Dataset: {dataset.name}"
    )

    print(
        f"Examples: {len(EVALUATION_EXAMPLES)}"
    )


# ============================================================
# 4. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 5. Tools
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


@tool
def query_payment(order_id: str) -> dict:
    """
    查询订单支付信息。
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
# 6. Agent
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

1. 用户询问订单状态、订单金额等信息时，
   使用 query_order。

2. 用户询问支付状态、支付方式等信息时，
   使用 query_payment。

3. 不要编造订单数据。

4. 必须根据 Tool 返回的数据回答。
""",
)


# ============================================================
# 7. Target System
#
# LangSmith evaluate() 会自动把 Dataset 中的：
#
# inputs
#    ↓
# target()
#
# target() 返回：
#
# outputs
# ============================================================

def target(inputs: dict) -> dict:

    question = inputs["question"]

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": question,
                }
            ]
        }
    )

    answer = (
        result["messages"][-1].content
    )

    return {
        "answer": answer,
    }


# ============================================================
# 8. Deterministic Evaluator
#
# LangSmith 会把：
#
# run    = Agent 实际运行结果
# example = Dataset 中的参考答案
#
# 传给 evaluator。
# ============================================================

def keyword_evaluator(
    run: Run,
    example: Example,
) -> dict:

    actual_answer = (
        run.outputs.get("answer", "")
        if run.outputs
        else ""
    )

    reference_answer = (
        example.outputs.get("answer", "")
        if example.outputs
        else ""
    )

    # --------------------------------------------------------
    # 为了教学简单起见，
    # 这里检查参考答案中的关键业务词。
    #
    # 真实项目可以替换成：
    # Tool Name
    # Tool Args
    # JSON Schema
    # Business Rule
    # Recall@K
    # --------------------------------------------------------

    keywords = []

    if "已发货" in reference_answer:
        keywords.append("已发货")

    if "1999" in reference_answer:
        keywords.append("1999")

    if "已支付" in reference_answer:
        keywords.append("已支付")

    if "微信支付" in reference_answer:
        keywords.append("微信支付")

    matched = [
        keyword
        for keyword in keywords
        if keyword in actual_answer
    ]

    score = (
        1.0
        if keywords
        and len(matched) == len(keywords)
        else 0.0
    )

    return {
        "key": "keyword_match",
        "score": score,
        "comment": (
            f"matched={matched}, "
            f"expected={keywords}"
        ),
    }


# ============================================================
# 9. LLM-as-a-Judge Evaluator
#
# 这里仍然使用我们自己的 MiniMax。
#
# 但是：
#
# Agent 执行
# Dataset 管理
# Evaluation 实验
# Score 保存
#
# 都由 LangSmith evaluate() 管理。
# ============================================================

JUDGE_PROMPT = """
你是一个 Agent Evaluation Judge。

请比较：

用户问题：
{question}

参考答案：
{reference_answer}

Agent 实际回答：
{actual_answer}

请评价：

correctness：
回答事实是否正确。

relevance：
回答是否真正针对用户问题。

只输出下面四行：

correctness: 数字
relevance: 数字
overall: 数字
reason: 一句话原因

评分范围：

0 = 完全错误
1 = 很差
2 = 较差
3 = 基本正确
4 = 正确
5 = 完全正确

不要输出其它内容。
不要输出 JSON。
不要输出 Markdown。
"""


def llm_judge_evaluator(
    run: Run,
    example: Example,
) -> dict:

    question = (
        example.inputs.get(
            "question",
            "",
        )
    )

    reference_answer = (
        example.outputs.get(
            "answer",
            "",
        )
    )

    actual_answer = (
        run.outputs.get(
            "answer",
            "",
        )
        if run.outputs
        else ""
    )

    prompt = JUDGE_PROMPT.format(
        question=question,
        reference_answer=reference_answer,
        actual_answer=actual_answer,
    )

    response = model.invoke(prompt)

    text = response.content.strip()

    # --------------------------------------------------------
    # 兼容 MiniMax <think>...</think>
    # --------------------------------------------------------

    text = re.sub(
        r"<think>.*?</think>",
        "",
        text,
        flags=re.DOTALL,
    ).strip()

    correctness_match = re.search(
        r"correctness\s*[:：]\s*(\d+(?:\.\d+)?)",
        text,
        re.IGNORECASE,
    )

    relevance_match = re.search(
        r"relevance\s*[:：]\s*(\d+(?:\.\d+)?)",
        text,
        re.IGNORECASE,
    )

    overall_match = re.search(
        r"overall\s*[:：]\s*(\d+(?:\.\d+)?)",
        text,
        re.IGNORECASE,
    )

    reason_match = re.search(
        r"reason\s*[:：]\s*(.*)",
        text,
        re.IGNORECASE | re.DOTALL,
    )

    if not (
        correctness_match
        and relevance_match
        and overall_match
    ):
        return {
            "key": "llm_judge",
            "score": 0.0,
            "comment": (
                "Judge 输出解析失败："
                f"{text}"
            ),
        }

    correctness = float(
        correctness_match.group(1)
    )

    relevance = float(
        relevance_match.group(1)
    )

    overall = float(
        overall_match.group(1)
    )

    reason = (
        reason_match.group(1).strip()
        if reason_match
        else text
    )

    return {
        "key": "llm_judge_overall",
        "score": overall / 5.0,
        "comment": (
            f"correctness={correctness}/5, "
            f"relevance={relevance}/5, "
            f"overall={overall}/5, "
            f"reason={reason}"
        ),
    }


# ============================================================
# 10. Run Evaluation
# ============================================================

def run_evaluation():

    ensure_dataset()

    print()
    print("=" * 70)
    print("开始 LangSmith Evaluation")
    print("=" * 70)

    results = langsmith_client.evaluate(
        target,
        data=DATASET_NAME,
        evaluators=[
            keyword_evaluator,
            llm_judge_evaluator,
        ],
        experiment_prefix=(
            "agent-learning-order"
        ),
        description=(
            "订单 Agent Evaluation"
        ),
        metadata={
            "model": "MiniMax-M3",
            "version": "v1",
            "environment": "local",
        },
        max_concurrency=1,
    )

    print()
    print("=" * 70)
    print("Evaluation 完成")
    print("=" * 70)

    print(results)


# ============================================================
# 11. Main
# ============================================================

if __name__ == "__main__":
    run_evaluation()

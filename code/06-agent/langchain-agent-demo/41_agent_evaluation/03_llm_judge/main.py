
import json
import re

from dotenv import load_dotenv
from pydantic import BaseModel, ValidationError
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
4. 必须根据 Tool 返回的数据回答。
""",
)


# ============================================================
# 5. Evaluation Dataset
# ============================================================

evaluation_dataset = [
    {
        "id": "case-001",
        "input": "请帮我查询订单10001现在是什么状态？",
        "reference_answer": "订单10001当前状态是已发货。",
    },
    {
        "id": "case-002",
        "input": "订单10002多少钱？",
        "reference_answer": "订单10002的金额是1999元。",
    },
    {
        "id": "case-003",
        "input": "订单10001支付成功了吗？",
        "reference_answer": "订单10001已经支付成功，支付方式是微信支付。",
    },
]


# ============================================================
# 6. Agent Run
# ============================================================

def run_agent(user_input: str) -> str:

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": user_input,
                }
            ]
        }
    )

    return result["messages"][-1].content


# ============================================================
# 7. Judge Result
# ============================================================

class JudgeResult(BaseModel):
    correctness: float
    relevance: float
    completeness: float
    overall: float
    reason: str


# ============================================================
# 8. 清理 <think>...</think>
# ============================================================

def remove_think_block(text: str) -> str:
    """
    去掉 MiniMax 等模型可能返回的：

    <think>
    ...
    </think>

    只保留正式回答部分。
    """

    text = re.sub(
        r"<think>.*?</think>",
        "",
        text,
        flags=re.DOTALL,
    )

    return text.strip()


# ============================================================
# 9. 从文本中提取 JSON
# ============================================================

def extract_json(text: str) -> dict | None:
    """
    尝试从 LLM 输出中提取 JSON。

    支持：

    {
        "correctness": 5,
        ...
    }

    也支持：

    ```json
    {
        ...
    }
    ```

    如果没有 JSON，返回 None。
    """

    text = remove_think_block(text)

    if not text:
        return None

    # --------------------------------------------------------
    # 去掉 Markdown code block
    # --------------------------------------------------------

    if text.startswith("```"):

        lines = text.splitlines()

        if len(lines) >= 3:
            lines = lines[1:-1]

        text = "\n".join(lines).strip()

    # --------------------------------------------------------
    # 查找 JSON
    # --------------------------------------------------------

    start = text.find("{")
    end = text.rfind("}")

    if start == -1 or end == -1:
        return None

    json_text = text[start:end + 1]

    try:
        return json.loads(json_text)

    except json.JSONDecodeError:
        return None


# ============================================================
# 10. 解析纯文本 Judge 输出
# ============================================================

def extract_text_score(
    text: str,
) -> JudgeResult | None:
    """
    兼容 Judge LLM 没有返回 JSON，而是：

    correctness: 5
    relevance: 5
    completeness: 4
    overall: 5
    reason: ...

    """

    text = remove_think_block(text)

    if not text:
        return None

    # --------------------------------------------------------
    # 提取数字
    # --------------------------------------------------------

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

    completeness_match = re.search(
        r"completeness\s*[:：]\s*(\d+(?:\.\d+)?)",
        text,
        re.IGNORECASE,
    )

    overall_match = re.search(
        r"overall\s*[:：]\s*(\d+(?:\.\d+)?)",
        text,
        re.IGNORECASE,
    )

    if not (
        correctness_match
        and relevance_match
        and completeness_match
        and overall_match
    ):
        return None

    correctness = float(
        correctness_match.group(1)
    )

    relevance = float(
        relevance_match.group(1)
    )

    completeness = float(
        completeness_match.group(1)
    )

    overall = float(
        overall_match.group(1)
    )

    # --------------------------------------------------------
    # Reason
    # --------------------------------------------------------

    reason_match = re.search(
        r"reason\s*[:：]\s*(.*)",
        text,
        re.IGNORECASE | re.DOTALL,
    )

    if reason_match:
        reason = reason_match.group(1).strip()
    else:
        reason = text.strip()

    return JudgeResult(
        correctness=correctness,
        relevance=relevance,
        completeness=completeness,
        overall=overall,
        reason=reason,
    )


# ============================================================
# 11. Judge Prompt
# ============================================================

JUDGE_PROMPT = """
你是一个专业的 AI Agent Evaluation Judge。

请评价 Agent 的回答。

你需要比较：

1. 用户问题
2. 参考答案
3. Agent 实际回答

评价三个维度：

correctness：
回答是否正确。
0 = 完全错误
1 = 非常差
2 = 较差
3 = 基本正确但存在明显问题
4 = 正确
5 = 完全正确

relevance：
回答是否紧扣用户问题。
0 = 完全无关
1 = 很不相关
2 = 有较多无关内容
3 = 基本相关
4 = 相关
5 = 高度相关

completeness：
回答是否包含回答问题所需要的关键信息。
0 = 完全缺失
1 = 严重缺失
2 = 缺少重要信息
3 = 基本完整
4 = 较完整
5 = 非常完整

overall：
综合评分，0-5。

评价原则：

- 不要求 Agent 必须和参考答案使用完全相同的文字。
- 表达方式不同，但事实和意思正确，可以给高分。
- 不要凭空增加参考答案之外的事实。
- 只判断回答质量。
- 不要输出额外解释。

优先按照下面格式输出：

correctness: 5
relevance: 5
completeness: 5
overall: 5
reason: 回答准确、相关且完整。

用户问题：
{user_input}

参考答案：
{reference_answer}

Agent 实际回答：
{actual_answer}
"""


# ============================================================
# 12. LLM Judge
# ============================================================

def judge_answer(
    user_input: str,
    reference_answer: str,
    actual_answer: str,
) -> JudgeResult:

    prompt = JUDGE_PROMPT.format(
        user_input=user_input,
        reference_answer=reference_answer,
        actual_answer=actual_answer,
    )

    response = model.invoke(prompt)

    print()
    print("=" * 70)
    print("Judge 原始输出")
    print("=" * 70)

    print(response.content)

    raw_output = response.content

    # ========================================================
    # 第一优先级：JSON
    # ========================================================

    data = extract_json(raw_output)

    if data is not None:

        try:

            return JudgeResult.model_validate(
                data
            )

        except ValidationError as e:

            print()
            print("JSON 数据结构校验失败:")
            print(e)

    # ========================================================
    # 第二优先级：纯文本格式
    # ========================================================

    result = extract_text_score(
        raw_output
    )

    if result is not None:
        return result

    # ========================================================
    # 最终失败
    # ========================================================

    raise ValueError(
        "Judge LLM 输出无法解析。\n\n"
        f"原始输出：\n{raw_output}"
    )


# ============================================================
# 13. Deterministic Evaluation
# ============================================================

def deterministic_check(
    actual_answer: str,
) -> bool:

    return bool(
        actual_answer
        and actual_answer.strip()
    )


# ============================================================
# 14. Run Evaluation
# ============================================================

def run_evaluation():

    total = len(
        evaluation_dataset
    )

    deterministic_passed = 0
    judge_results = []

    print()
    print("=" * 70)
    print("Agent Evaluation")
    print("=" * 70)

    for case in evaluation_dataset:

        print()
        print("-" * 70)
        print(
            f"Case ID: {case['id']}"
        )
        print("-" * 70)

        user_input = case["input"]

        reference_answer = (
            case["reference_answer"]
        )

        print("User:")
        print(user_input)

        print()
        print("Reference:")
        print(reference_answer)

        # ----------------------------------------------------
        # Agent
        # ----------------------------------------------------

        actual_answer = run_agent(
            user_input
        )

        print()
        print("Agent:")
        print(actual_answer)

        # ----------------------------------------------------
        # Deterministic Check
        # ----------------------------------------------------

        deterministic_result = (
            deterministic_check(
                actual_answer
            )
        )

        print()
        print(
            "Deterministic Check:",
            "PASS"
            if deterministic_result
            else "FAIL",
        )

        if deterministic_result:
            deterministic_passed += 1

        # ----------------------------------------------------
        # Judge
        # ----------------------------------------------------

        judge_result = judge_answer(
            user_input=user_input,
            reference_answer=reference_answer,
            actual_answer=actual_answer,
        )

        judge_results.append(
            judge_result
        )

        print()
        print("Judge Result:")

        print(
            f"Correctness : "
            f"{judge_result.correctness}/5"
        )

        print(
            f"Relevance   : "
            f"{judge_result.relevance}/5"
        )

        print(
            f"Completeness: "
            f"{judge_result.completeness}/5"
        )

        print(
            f"Overall     : "
            f"{judge_result.overall}/5"
        )

        print(
            f"Reason      : "
            f"{judge_result.reason}"
        )

    # ========================================================
    # Summary
    # ========================================================

    average_correctness = (
        sum(
            item.correctness
            for item in judge_results
        )
        / total
    )

    average_relevance = (
        sum(
            item.relevance
            for item in judge_results
        )
        / total
    )

    average_completeness = (
        sum(
            item.completeness
            for item in judge_results
        )
        / total
    )

    average_overall = (
        sum(
            item.overall
            for item in judge_results
        )
        / total
    )

    print()
    print("=" * 70)
    print("Evaluation Summary")
    print("=" * 70)

    print(
        f"Dataset Size            : {total}"
    )

    print(
        f"Deterministic Pass Rate : "
        f"{deterministic_passed / total:.2%}"
    )

    print(
        f"Avg Correctness         : "
        f"{average_correctness:.2f}/5"
    )

    print(
        f"Avg Relevance           : "
        f"{average_relevance:.2f}/5"
    )

    print(
        f"Avg Completeness        : "
        f"{average_completeness:.2f}/5"
    )

    print(
        f"Avg Overall             : "
        f"{average_overall:.2f}/5"
    )


# ============================================================
# 15. Main
# ============================================================

if __name__ == "__main__":
    run_evaluation()


import os

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
# 4. 创建 Agent
# ============================================================

agent = create_agent(
    model=model,
    tools=[query_order],
    system_prompt="""
你是一个订单查询助手。

当用户询问订单信息时，必须调用 query_order 工具。

不要编造订单信息。
""",
)


# ============================================================
# 5. Agent 执行
# ============================================================

def run_agent(user_message: str) -> str:

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

    return result["messages"][-1].content


# ============================================================
# 6. Evaluation Test Case
# ============================================================

test_cases = [
    {
        "name": "查询订单状态",
        "input": "请帮我查询订单10001的状态。",
        "expected": "已发货",
    },
    {
        "name": "查询订单金额",
        "input": "订单10002多少钱？",
        "expected": "1999",
    },
    {
        "name": "查询不存在订单",
        "input": "请查询订单99999。",
        "expected": "订单不存在",
    },
]


# ============================================================
# 7. Evaluation
# ============================================================

def evaluate(
    actual: str,
    expected: str,
) -> bool:

    return expected in actual


# ============================================================
# 8. 执行 Evaluation
# ============================================================

def run_evaluation():

    total = len(test_cases)
    passed = 0

    print()
    print("=" * 70)
    print("Agent Evaluation")
    print("=" * 70)

    for index, test_case in enumerate(test_cases, start=1):

        print()
        print("-" * 70)
        print(f"Test Case {index}: {test_case['name']}")
        print("-" * 70)

        user_input = test_case["input"]
        expected = test_case["expected"]

        print("Input:")
        print(user_input)

        print()
        print("Expected:")
        print(expected)

        # ----------------------------------------------------
        # Agent
        # ----------------------------------------------------

        actual = run_agent(user_input)

        print()
        print("Actual:")
        print(actual)

        # ----------------------------------------------------
        # Evaluation
        # ----------------------------------------------------

        passed_case = evaluate(
            actual=actual,
            expected=expected,
        )

        print()

        if passed_case:
            print("Result: PASS")
            passed += 1
        else:
            print("Result: FAIL")


    # ========================================================
    # Summary
    # ========================================================

    score = passed / total

    print()
    print("=" * 70)
    print("Evaluation Summary")
    print("=" * 70)

    print(f"Total : {total}")
    print(f"Passed: {passed}")
    print(f"Failed: {total - passed}")
    print(f"Score : {score:.2%}")


# ============================================================
# 9. Main
# ============================================================

if __name__ == "__main__":
    run_evaluation()

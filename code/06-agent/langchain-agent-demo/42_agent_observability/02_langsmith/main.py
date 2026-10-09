
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
#
# LangSmith tracing 不需要在这里额外写代码。
# 因为 LANGSMITH_TRACING=true 后，
# LangChain 会自动记录相关运行信息。
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
# 5. Agent Run
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
# 6. Main
# ============================================================

if __name__ == "__main__":

    user_message = (
        "请帮我查询订单10001现在是什么状态？"
    )

    result = run_agent(
        user_message
    )

    print()
    print("=" * 70)
    print("Agent Result")
    print("=" * 70)

    print(
        result["messages"][-1].content
    )

from fastapi import FastAPI
from pydantic import BaseModel

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 环境变量
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
# 3. Payment Agent
# ============================================================

payment_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是支付领域专家 Agent。\n"
        "你只负责处理支付相关问题。\n\n"
        "支付数据：\n"
        "- 订单号：10001\n"
        "- 支付状态：已支付\n"
        "- 支付方式：微信支付\n\n"
        "根据任务返回准确、简洁的支付信息。"
    ),
)


# ============================================================
# 4. HTTP Request Model
# ============================================================

class AgentRequest(BaseModel):
    task_id: str
    user_id: str
    task: str


# ============================================================
# 5. FastAPI
# ============================================================

app = FastAPI(
    title="Payment Agent",
)


# ============================================================
# 6. Agent API
# ============================================================

@app.post("/agent/payment")
def handle_payment(request: AgentRequest):

    print("=" * 60)
    print("Payment Agent 收到请求")
    print("task_id:", request.task_id)
    print("user_id:", request.user_id)
    print("task:", request.task)

    result = payment_agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": request.task,
                }
            ]
        }
    )

    answer = result["messages"][-1].content

    print("Payment Agent 返回：")
    print(answer)

    return {
        "task_id": request.task_id,
        "agent": "payment",
        "status": "SUCCESS",
        "data": {
            "answer": answer,
        },
    }
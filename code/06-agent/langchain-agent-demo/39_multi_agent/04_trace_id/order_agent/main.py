from fastapi import FastAPI, Header
from pydantic import BaseModel

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_openai import ChatOpenAI


load_dotenv()


model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


order_agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是订单领域专家 Agent。\n"
        "你只负责处理订单相关问题。\n\n"
        "订单数据：\n"
        "- 订单号：10001\n"
        "- 用户：张三\n"
        "- 金额：3999 元\n"
        "- 状态：已发货\n\n"
        "根据任务返回准确、简洁的订单信息。"
    ),
)


class AgentRequest(BaseModel):
    task_id: str
    user_id: str
    task: str


app = FastAPI(title="Order Agent")


@app.post("/agent/order")
def handle_order(
    request: AgentRequest,
    x_trace_id: str = Header(...),
):
    print("=" * 60)
    print("Order Agent 收到请求")
    print("trace_id:", x_trace_id)
    print("task_id:", request.task_id)
    print("user_id:", request.user_id)
    print("task:", request.task)

    result = order_agent.invoke(
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

    print("Order Agent 返回：")
    print(answer)

    return {
        "trace_id": x_trace_id,
        "task_id": request.task_id,
        "agent": "order",
        "status": "SUCCESS",
        "data": {
            "answer": answer,
        },
    }
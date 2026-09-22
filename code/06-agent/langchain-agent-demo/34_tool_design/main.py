import json
import os
import uuid

import redis

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.tools import ToolException, tool
from langchain_openai import ChatOpenAI
from pydantic import BaseModel, Field


# ============================================================
# 1. 初始化环境
# ============================================================

load_dotenv()


# ============================================================
# 2. 初始化 LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


# ============================================================
# 3. Redis
# ============================================================

redis_client = redis.Redis(
    host="localhost",
    port=6379,
    db=0,
    decode_responses=True,
)


# ============================================================
# 4. Create Order Input
# ============================================================

class CreateOrderInput(BaseModel):
    """
    创建订单输入参数。
    """

    request_id: str = Field(
        description=(
            "本次业务请求的唯一 ID。"
            "Retry 时必须保持不变。"
        )
    )

    amount: int = Field(
        description="订单金额，例如：3999"
    )


# ============================================================
# 5. Create Order Tool
# ============================================================

@tool(
    "create_order",
    description=(
        "创建订单。"
        "request_id 用于保证请求幂等。"
        "如果请求因为网络问题需要 Retry，"
        "Retry 时必须继续使用相同 request_id。"
    ),
    args_schema=CreateOrderInput,
)
def create_order(
    request_id: str,
    amount: int,
) -> dict:
    """
    创建订单。

    使用 Redis 实现简单的幂等控制。
    """

    status_key = (
        f"idempotency:create_order:status:{request_id}"
    )

    result_key = (
        f"idempotency:create_order:result:{request_id}"
    )

    # ========================================================
    # Step 1：检查状态
    # ========================================================

    status = redis_client.get(
        status_key
    )

    # ========================================================
    # 第一次请求
    # ========================================================

    if status is None:

        print(
            f"[CreateOrder] "
            f"第一次请求：{request_id}"
        )

        # ----------------------------------------------------
        # SETNX
        # ----------------------------------------------------

        acquired = redis_client.set(
            status_key,
            "PROCESSING",
            nx=True,
            ex=60,
        )

        if not acquired:

            raise ToolException(
                "订单请求正在处理中，请稍后重试"
            )

        # ----------------------------------------------------
        # 模拟真正的订单创建
        # ----------------------------------------------------

        order = {
            "order_id": str(uuid.uuid4()),
            "request_id": request_id,
            "amount": amount,
            "status": "已创建",
        }

        print(
            "[CreateOrder] "
            "订单已经成功创建："
        )

        print(
            order
        )

        # ----------------------------------------------------
        # 保存订单结果
        # ----------------------------------------------------

        redis_client.set(
            result_key,
            json.dumps(
                order,
                ensure_ascii=False,
            ),
            ex=3600,
        )

        # ----------------------------------------------------
        # 修改状态
        # ----------------------------------------------------

        redis_client.set(
            status_key,
            "SUCCESS",
            ex=3600,
        )

        # ----------------------------------------------------
        # 注意：
        #
        # 这里模拟“订单创建成功，但是响应丢失”
        # ----------------------------------------------------

        raise ToolException(
            "模拟网络异常：订单实际已经创建成功，但响应丢失"
        )


    # ========================================================
    # 第二次请求 / Retry
    # ========================================================

    if status == "SUCCESS":

        print(
            f"[CreateOrder] "
            f"发现请求已经成功：{request_id}"
        )

        result = redis_client.get(
            result_key
        )

        if result:

            print(
                "[CreateOrder] "
                "返回第一次创建的订单"
            )

            return json.loads(result)

        raise ToolException(
            "订单已经成功，但订单结果暂时无法获取"
        )


    # ========================================================
    # 正在处理中
    # ========================================================

    if status == "PROCESSING":

        raise ToolException(
            "订单正在处理中，请稍后重试"
        )


    # ========================================================
    # 之前失败
    # ========================================================

    if status == "FAILED":

        raise ToolException(
            "订单之前执行失败，可以重新发起"
        )


    # ========================================================
    # 未知状态
    # ========================================================

    raise ToolException(
        "订单当前状态未知，请稍后重试"
    )


# Tool 错误转成 ToolMessage
create_order.handle_tool_error = True


# ============================================================
# 6. Agent
# ============================================================

def create_order_agent():

    return create_agent(
        model=model,
        tools=[
            create_order,
        ],
    )


# ============================================================
# 7. 模拟 Retry
# ============================================================

def retry_create_order(
    request_id: str,
    amount: int,
    max_retries: int = 2,
):
    """
    模拟 Client / Service 层 Retry。

    注意：
    Retry 时 request_id 必须保持一致。
    """

    print(
        "\n========== Retry =========="
    )

    for attempt in range(
        1,
        max_retries + 1,
    ):

        print(
            f"\n第 {attempt} 次请求"
        )

        print(
            f"request_id = {request_id}"
        )

        try:

            result = create_order.invoke(
                {
                    "request_id": request_id,
                    "amount": amount,
                }
            )

            print(
                "\n请求成功："
            )

            print(
                result
            )

            return result

        except Exception as e:

            print(
                "\n请求失败："
            )

            print(
                e
            )

            if attempt >= max_retries:

                print(
                    "\nRetry 次数已经用完"
                )

                raise

            print(
                "\n准备 Retry..."
            )

    return None


# ============================================================
# 8. Main
# ============================================================

def main():

    # ========================================================
    # 清理测试数据
    # ========================================================

    request_id = (
        f"req-retry-{uuid.uuid4()}"
    )

    status_key = (
        f"idempotency:create_order:status:{request_id}"
    )

    result_key = (
        f"idempotency:create_order:result:{request_id}"
    )

    redis_client.delete(
        status_key,
        result_key,
    )


    # ========================================================
    # 1. Agent
    # ========================================================

    print(
        "========== Agent =========="
    )

    agent = create_order_agent()


    # ========================================================
    # 2. 第一次 Agent 请求
    # ========================================================

    print(
        "\n========== First Request =========="
    )

    result = agent.invoke(
        {
            "messages": [
                {
                    "role": "user",
                    "content": (
                        "帮我创建一个3999元的订单。"
                        f"request_id={request_id}"
                    ),
                }
            ]
        }
    )


    # ========================================================
    # Agent Trace
    # ========================================================

    print(
        "\n========== Agent Trace =========="
    )

    for message in result["messages"]:

        print(
            "\n--------------------"
        )

        print(
            type(message).__name__
        )

        print(
            "content:"
        )

        print(
            message.content
        )

        if getattr(
            message,
            "tool_calls",
            None,
        ):

            print(
                "Tool Calls:"
            )

            for tool_call in message.tool_calls:

                print(
                    tool_call
                )


    # ========================================================
    # 3. Retry
    # ========================================================
    #
    # 注意：
    #
    # Retry 使用完全相同的 request_id。
    #
    # request_id 不变
    # amount 不变
    #
    # 这样 Idempotency 才能生效。
    #

    retry_result = retry_create_order(
        request_id=request_id,
        amount=3999,
        max_retries=2,
    )


    # ========================================================
    # 4. Redis 最终状态
    # ========================================================

    print(
        "\n========== Redis Final State =========="
    )

    print(
        "\nStatus:"
    )

    print(
        redis_client.get(
            status_key
        )
    )

    print(
        "\nOrder:"
    )

    print(
        redis_client.get(
            result_key
        )
    )


    # ========================================================
    # 5. Final Result
    # ========================================================

    print(
        "\n========== Final Result =========="
    )

    print(
        retry_result
    )


# ============================================================
# Entry
# ============================================================

if __name__ == "__main__":
    main()
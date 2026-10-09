import time
import uuid
from typing import Any

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain_core.callbacks import BaseCallbackHandler
from langchain_core.tools import tool
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 加载环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. Observable Callback
#
# 用于观察：
#
# LLM Start
# LLM End
# LLM Error
# Tool Start
# Tool End
# Tool Error
#
# 当前只是教学 Demo。
# 生产环境会进一步接 OpenTelemetry / tracing system。
# ============================================================

class ObservationCallbackHandler(BaseCallbackHandler):

    def __init__(self, trace_id: str):
        super().__init__()

        self.trace_id = trace_id

        # run_id -> start_time
        self.start_times: dict[str, float] = {}

        self.llm_count = 0
        self.tool_count = 0

    # ========================================================
    # Common
    # ========================================================

    def _start_span(
        self,
        span_type: str,
        run_id: Any,
        name: str,
    ):
        run_id_str = str(run_id)

        self.start_times[run_id_str] = time.perf_counter()

        print(
            f"[SPAN START] "
            f"trace_id={self.trace_id} "
            f"run_id={run_id_str} "
            f"type={span_type} "
            f"name={name}"
        )

    def _end_span(
        self,
        span_type: str,
        run_id: Any,
        name: str,
        status: str,
    ):
        run_id_str = str(run_id)

        start_time = self.start_times.pop(
            run_id_str,
            None,
        )

        duration_ms = None

        if start_time is not None:
            duration_ms = (
                time.perf_counter()
                - start_time
            ) * 1000

        if duration_ms is None:

            print(
                f"[SPAN END] "
                f"trace_id={self.trace_id} "
                f"run_id={run_id_str} "
                f"type={span_type} "
                f"name={name} "
                f"status={status}"
            )

        else:

            print(
                f"[SPAN END] "
                f"trace_id={self.trace_id} "
                f"run_id={run_id_str} "
                f"type={span_type} "
                f"name={name} "
                f"status={status} "
                f"duration={duration_ms:.2f}ms"
            )

    # ========================================================
    # LLM
    # ========================================================

    def on_llm_start(
        self,
        serialized: dict[str, Any],
        prompts: list[str],
        *,
        run_id,
        parent_run_id=None,
        tags=None,
        metadata=None,
        **kwargs: Any,
    ) -> None:

        self.llm_count += 1

        model_name = (
            serialized.get("name")
            or serialized.get("id", ["Unknown"])[-1]
        )

        self._start_span(
            span_type="LLM",
            run_id=run_id,
            name=model_name,
        )

    def on_llm_end(
        self,
        response,
        *,
        run_id,
        parent_run_id=None,
        **kwargs: Any,
    ) -> None:

        self._end_span(
            span_type="LLM",
            run_id=run_id,
            name="LLM",
            status="SUCCESS",
        )

    def on_llm_error(
        self,
        error,
        *,
        run_id,
        parent_run_id=None,
        **kwargs: Any,
    ) -> None:

        self._end_span(
            span_type="LLM",
            run_id=run_id,
            name="LLM",
            status="ERROR",
        )

        print(
            f"[LLM ERROR] "
            f"trace_id={self.trace_id} "
            f"error={error}"
        )

    # ========================================================
    # Tool
    # ========================================================

    def on_tool_start(
        self,
        serialized: dict[str, Any],
        input_str: str,
        *,
        run_id,
        parent_run_id=None,
        tags=None,
        metadata=None,
        inputs=None,
        **kwargs: Any,
    ) -> None:

        self.tool_count += 1

        tool_name = (
            serialized.get("name")
            or "UnknownTool"
        )

        self._start_span(
            span_type="TOOL",
            run_id=run_id,
            name=tool_name,
        )

        print(
            f"[TOOL INPUT] "
            f"trace_id={self.trace_id} "
            f"tool={tool_name} "
            f"input={input_str}"
        )

    def on_tool_end(
        self,
        output,
        *,
        run_id,
        parent_run_id=None,
        **kwargs: Any,
    ) -> None:

        self._end_span(
            span_type="TOOL",
            run_id=run_id,
            name="Tool",
            status="SUCCESS",
        )

    def on_tool_error(
        self,
        error,
        *,
        run_id,
        parent_run_id=None,
        **kwargs: Any,
    ) -> None:

        self._end_span(
            span_type="TOOL",
            run_id=run_id,
            name="Tool",
            status="ERROR",
        )

        print(
            f"[TOOL ERROR] "
            f"trace_id={self.trace_id} "
            f"error={error}"
        )


# ============================================================
# 3. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 4. Tools
# ============================================================

@tool
def query_order(order_id: str) -> dict:
    """
    查询订单信息。
    """

    time.sleep(0.2)

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

    time.sleep(0.1)

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
# 5. Agent
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
# 6. Agent Run
# ============================================================

def run_agent(user_message: str):

    # --------------------------------------------------------
    # 一个请求生成一个 Trace ID
    # --------------------------------------------------------

    trace_id = str(
        uuid.uuid4()
    )

    callback = ObservationCallbackHandler(
        trace_id=trace_id
    )

    request_start = time.perf_counter()

    print()
    print("=" * 70)
    print("Agent Request")
    print("=" * 70)

    print(
        f"trace_id={trace_id}"
    )

    print(
        f"user_message={user_message}"
    )

    # --------------------------------------------------------
    # Agent
    # --------------------------------------------------------

    try:

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
                "callbacks": [
                    callback
                ],
                "metadata": {
                    "trace_id": trace_id,
                },
            },
        )

        duration_ms = (
            time.perf_counter()
            - request_start
        ) * 1000

        print()
        print("=" * 70)
        print("Agent Request Summary")
        print("=" * 70)

        print(
            f"trace_id={trace_id}"
        )

        print(
            f"status=SUCCESS"
        )

        print(
            f"total_duration={duration_ms:.2f}ms"
        )

        print(
            f"llm_count={callback.llm_count}"
        )

        print(
            f"tool_count={callback.tool_count}"
        )

        print()
        print(
            "Final Answer:"
        )

        print(
            result["messages"][-1].content
        )

        return result

    except Exception as e:

        duration_ms = (
            time.perf_counter()
            - request_start
        ) * 1000

        print()
        print("=" * 70)
        print("Agent Request Summary")
        print("=" * 70)

        print(
            f"trace_id={trace_id}"
        )

        print(
            f"status=ERROR"
        )

        print(
            f"total_duration={duration_ms:.2f}ms"
        )

        print(
            f"llm_count={callback.llm_count}"
        )

        print(
            f"tool_count={callback.tool_count}"
        )

        print(
            f"error={e}"
        )

        raise


# ============================================================
# 7. Main
# ============================================================

if __name__ == "__main__":

    run_agent(
        "请帮我查询订单10001现在是什么状态？"
    )
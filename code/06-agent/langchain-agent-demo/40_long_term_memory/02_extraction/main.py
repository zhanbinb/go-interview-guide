
import json

from pydantic import BaseModel, ValidationError
from langchain_openai import ChatOpenAI

from dotenv import load_dotenv

load_dotenv()
# ============================================================
# 1. Memory 数据结构
# ============================================================

class UserMemory(BaseModel):
    name: str | None = None
    job: str | None = None
    learning_focus: str | None = None
    preference: str | None = None


# ============================================================
# 2. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 3. Memory Extraction Prompt
# ============================================================

EXTRACTION_PROMPT = """
你是一个长期记忆提取器。

请从用户消息中提取适合长期保存的信息。

只保存真正具有长期价值的信息，例如：

- 用户姓名
- 用户职业
- 用户长期学习方向
- 用户长期偏好

不要保存：

- 临时问题
- 一次性的任务
- 当前正在讨论的具体代码
- 无长期价值的信息

需要提取的字段：

name：用户姓名
job：用户职业
learning_focus：长期学习方向
preference：长期偏好

如果某个字段无法确定，就使用 null。

要求：

1. 只输出 JSON
2. 不要输出解释
3. 不要输出 Markdown
4. 不要使用 ```json
5. 不要输出 <think>

用户消息：

{user_message}
"""


# ============================================================
# 4. Memory Store
#
# 当前只是教学 Demo。
# 后面会换成 Redis / PostgreSQL 等持久化存储。
# ============================================================

memory_store = {}


# ============================================================
# 5. 从 LLM 输出中提取 JSON
# ============================================================

def extract_json(text: str) -> dict:
    """
    从 LLM 原始输出中提取 JSON。

    兼容：

    {
        "name": "张三"
    }

    以及：

    ```json
    {
        "name": "张三"
    }
    ```
    """

    text = text.strip()

    # 去掉 Markdown code block
    if text.startswith("```"):
        lines = text.splitlines()

        if len(lines) >= 3:
            lines = lines[1:-1]

        text = "\n".join(lines).strip()

    # 找到第一个 {
    start = text.find("{")

    # 找到最后一个 }
    end = text.rfind("}")

    if start == -1 or end == -1:
        raise ValueError(
            f"LLM 没有返回有效 JSON：\n{text}"
        )

    json_text = text[start:end + 1]

    return json.loads(json_text)


# ============================================================
# 6. Memory Extraction
# ============================================================

def extract_memory(user_message: str) -> UserMemory:
    """
    用户消息
        ↓
    LLM
        ↓
    JSON
        ↓
    json.loads
        ↓
    Pydantic
        ↓
    UserMemory
    """

    prompt = EXTRACTION_PROMPT.format(
        user_message=user_message
    )

    response = model.invoke(prompt)

    print()
    print("=" * 70)
    print("LLM 原始输出")
    print("=" * 70)

    print(response.content)

    try:
        # ----------------------------------------------------
        # 第一步：解析 JSON
        # ----------------------------------------------------
        data = extract_json(response.content)

        # ----------------------------------------------------
        # 第二步：验证数据结构
        # ----------------------------------------------------
        memory = UserMemory.model_validate(data)

        return memory

    except (
        json.JSONDecodeError,
        ValidationError,
        ValueError,
    ) as e:

        print()
        print("=" * 70)
        print("Memory Extraction 解析失败")
        print("=" * 70)

        print(e)

        raise


# ============================================================
# 7. 保存 Memory
# ============================================================

def save_memory(
    user_id: str,
    memory: UserMemory,
):
    memory_store[user_id] = memory.model_dump()


# ============================================================
# 8. 主程序
# ============================================================

if __name__ == "__main__":

    user_id = "user1001"

    user_message = (
        "我叫张三，目前主要做 Go 后端开发。"
        "最近一段时间主要在学习 Agent 和 RAG。"
        "我平时比较喜欢看完整的代码示例，"
        "不太喜欢特别长的理论解释。"
    )

    print("=" * 70)
    print("用户消息")
    print("=" * 70)

    print(user_message)

    # --------------------------------------------------------
    # Memory Extraction
    # --------------------------------------------------------

    memory = extract_memory(user_message)

    # --------------------------------------------------------
    # 打印结构化 Memory
    # --------------------------------------------------------

    print()
    print("=" * 70)
    print("提取后的 Memory")
    print("=" * 70)

    print(memory.model_dump())

    # --------------------------------------------------------
    # 保存 Memory
    # --------------------------------------------------------

    save_memory(
        user_id=user_id,
        memory=memory,
    )

    # --------------------------------------------------------
    # 查看 Memory Store
    # --------------------------------------------------------

    print()
    print("=" * 70)
    print("Memory Store")
    print("=" * 70)

    print(memory_store)


import json
import os

from dotenv import load_dotenv
from pydantic import BaseModel, ValidationError
from langchain_openai import ChatOpenAI


# ============================================================
# 1. 加载环境变量
# ============================================================

load_dotenv()


# ============================================================
# 2. Memory 数据结构
# ============================================================

class UserMemory(BaseModel):
    name: str | None = None
    job: str | None = None
    learning_focus: str | None = None
    preference: str | None = None


# ============================================================
# 3. LLM
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


# ============================================================
# 4. Memory Store
#
# 当前使用 dict 模拟数据库。
#
# 生产环境可以替换成：
# Redis / PostgreSQL / MongoDB / Vector DB
# ============================================================

memory_store: dict[str, dict] = {}


# ============================================================
# 5. Memory Extraction Prompt
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
# 6. JSON 提取
# ============================================================

def extract_json(text: str) -> dict:
    """
    从 LLM 原始输出中提取 JSON。
    """

    text = text.strip()

    # 兼容 ```json ... ```
    if text.startswith("```"):
        lines = text.splitlines()

        if len(lines) >= 3:
            lines = lines[1:-1]

        text = "\n".join(lines).strip()

    start = text.find("{")
    end = text.rfind("}")

    if start == -1 or end == -1:
        raise ValueError(
            f"LLM 没有返回有效 JSON：\n{text}"
        )

    json_text = text[start:end + 1]

    return json.loads(json_text)


# ============================================================
# 7. Memory Extraction
# ============================================================

def extract_memory(user_message: str) -> UserMemory:
    """
    用户消息
        ↓
    LLM
        ↓
    JSON
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
    print("LLM Memory Extraction 原始输出")
    print("=" * 70)

    print(response.content)

    try:
        data = extract_json(response.content)

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
# 8. 保存 Memory
# ============================================================

def save_memory(
    user_id: str,
    memory: UserMemory,
):
    memory_store[user_id] = memory.model_dump()


# ============================================================
# 9. 读取 Memory
# ============================================================

def load_memory(user_id: str) -> UserMemory:
    """
    根据 user_id 读取长期记忆。
    """

    data = memory_store.get(user_id)

    if not data:
        return UserMemory()

    return UserMemory.model_validate(data)


# ============================================================
# 10. Memory Update
# ============================================================

def update_memory(
    user_id: str,
    new_memory: UserMemory,
):
    """
    将新的 Memory 合并到旧 Memory。

    规则：

    新 Memory 有值
        ↓
    更新旧值

    新 Memory 为 None
        ↓
    保留旧值
    """

    old_memory = load_memory(user_id)

    old_data = old_memory.model_dump()
    new_data = new_memory.model_dump()

    for key, value in new_data.items():

        if value is not None:
            old_data[key] = value

    updated_memory = UserMemory.model_validate(old_data)

    save_memory(
        user_id=user_id,
        memory=updated_memory,
    )

    return updated_memory


# ============================================================
# 11. Memory Retrieval
# ============================================================

def retrieve_memory(user_id: str) -> UserMemory:
    """
    当前 Demo 直接根据 user_id 查询。

    后面如果 Memory 数量非常大，
    可以进一步做：
    
    user_id
        ↓
    Memory Retrieval
        ↓
    相关 Memory
    """

    return load_memory(user_id)


# ============================================================
# 12. Agent 使用 Memory
# ============================================================

def ask_agent(
    user_id: str,
    user_message: str,
):
    """
    新 Session 中：

    读取长期 Memory
        ↓
    注入 Context
        ↓
    调用 LLM
    """

    memory = retrieve_memory(user_id)

    memory_context = f"""
你正在帮助一位用户。

这是系统保存的用户长期记忆：

{memory.model_dump()}

请结合这些长期记忆回答用户的问题。

注意：
- 不要机械重复 Memory
- 如果 Memory 与当前问题无关，可以忽略
- 不要编造 Memory 中不存在的信息
"""

    messages = [
        {
            "role": "system",
            "content": memory_context,
        },
        {
            "role": "user",
            "content": user_message,
        },
    ]

    response = model.invoke(messages)

    return response


# ============================================================
# 13. 主程序
# ============================================================

if __name__ == "__main__":

    user_id = "user1001"


    # ========================================================
    # Session 1
    # ========================================================

    thread_id_1 = "session-001"

    user_message_1 = "我主要做 Go 后端开发。"

    print()
    print("=" * 70)
    print("Session 1")
    print("=" * 70)

    print("thread_id:", thread_id_1)
    print("user_id:", user_id)
    print("用户:", user_message_1)

    memory_1 = extract_memory(user_message_1)

    print()
    print("提取 Memory:")
    print(memory_1.model_dump())

    save_memory(
        user_id=user_id,
        memory=memory_1,
    )

    print()
    print("Memory Store:")
    print(memory_store)


    # ========================================================
    # Session 2
    # ========================================================

    thread_id_2 = "session-002"

    user_message_2 = (
        "我最近开始学习 Agent 和 RAG，"
        "平时比较喜欢看完整的代码示例，"
        "不太喜欢特别长的理论解释。"
    )

    print()
    print("=" * 70)
    print("Session 2")
    print("=" * 70)

    print("thread_id:", thread_id_2)
    print("user_id:", user_id)
    print("用户:", user_message_2)

    memory_2 = extract_memory(user_message_2)

    print()
    print("本次提取的新 Memory:")
    print(memory_2.model_dump())

    updated_memory = update_memory(
        user_id=user_id,
        new_memory=memory_2,
    )

    print()
    print("更新后的 Memory:")
    print(updated_memory.model_dump())


    # ========================================================
    # Session 3
    # ========================================================

    thread_id_3 = "session-003"

    user_message_3 = "根据我的情况，我接下来应该重点学习什么？"

    print()
    print("=" * 70)
    print("Session 3")
    print("=" * 70)

    print("thread_id:", thread_id_3)
    print("user_id:", user_id)
    print("用户:", user_message_3)

    # --------------------------------------------------------
    # Retrieval
    # --------------------------------------------------------

    memory = retrieve_memory(user_id)

    print()
    print("Retrieved Memory:")
    print(memory.model_dump())


    # --------------------------------------------------------
    # Agent 使用 Memory
    # --------------------------------------------------------

    response = ask_agent(
        user_id=user_id,
        user_message=user_message_3,
    )

    print()
    print("=" * 70)
    print("Agent 最终回答")
    print("=" * 70)

    print(response.content)

from langchain.agents import create_agent
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv


load_dotenv()


# ============================================================
# 1. 模拟 Long-term Memory Store
# ============================================================
#
# 真实生产环境这里通常不会是 Python dict。
# 后面我们会逐步替换成 Redis / PostgreSQL / Vector Store。
#
memory_store = {}


# ============================================================
# 2. Memory 操作
# ============================================================


def save_memory(user_id: str, memory: dict):
    """
    保存用户长期记忆。
    """
    memory_store[user_id] = memory

    print("\n[Memory Store] 保存记忆：")
    print(memory)


def load_memory(user_id: str) -> dict:
    """
    读取用户长期记忆。
    """
    memory = memory_store.get(user_id, {})

    print("\n[Memory Store] 读取记忆：")
    print(memory)

    return memory


# ============================================================
# 3. 创建 Agent
# ============================================================

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
)


agent = create_agent(
    model=model,
    tools=[],
    system_prompt=(
        "你是一个企业级 Agent。\n"
        "你可以使用系统提供的用户长期记忆。\n\n"
        "回答用户问题时，如果长期记忆中存在相关信息，"
        "可以结合这些信息回答。\n"
        "不要虚构长期记忆中不存在的信息。"
    ),
)


# ============================================================
# 4. Session 001
# ============================================================

print("=" * 70)
print("Session 001")
print("=" * 70)


user_id = "user1001"
thread_id_1 = "session-001"


user_message_1 = "我叫张三，我是一名 Go 后端开发。"


print("\n用户：")
print(user_message_1)

# ------------------------------------------------------------
# 这里为了教学，直接模拟 Memory Extraction。
#
# 后面 40-2 我们会让 LLM 自动判断：
# “哪些信息值得保存成长期记忆？”
# ------------------------------------------------------------

memory = {
    "name": "张三",
    "job": "Go 后端开发",
}


save_memory(
    user_id=user_id,
    memory=memory,
)


print("\nSession 001 完成")
print("thread_id:", thread_id_1)


# ============================================================
# 5. Session 002
# ============================================================

print("\n")
print("=" * 70)
print("Session 002")
print("=" * 70)


thread_id_2 = "session-002"

user_message_2 = "推荐一下适合我的学习方向。"


print("\n用户：")
print(user_message_2)

print("\n当前 Session：")
print(thread_id_2)

# ------------------------------------------------------------
# 注意：
#
# Session 002 和 Session 001 是两个不同的 Thread。
#
# 但是我们通过 user_id 找到了之前保存的长期记忆。
# ------------------------------------------------------------

long_term_memory = load_memory(user_id)


# ============================================================
# 6. 把 Long-term Memory 注入 Agent Context
# ============================================================

memory_context = (
    f"用户长期记忆：\n{long_term_memory}\n\n请结合这些长期记忆回答用户问题。"
)


result = agent.invoke(
    {
        "messages": [
            {
                "role": "system",
                "content": memory_context,
            },
            {
                "role": "user",
                "content": user_message_2,
            },
        ]
    }
)


# ============================================================
# 7. 输出 Agent 最终回答
# ============================================================

print("\nAgent：")
print(result["messages"][-1].content)

print("\n")
print("=" * 70)
print("Demo 完成")
print("=" * 70)

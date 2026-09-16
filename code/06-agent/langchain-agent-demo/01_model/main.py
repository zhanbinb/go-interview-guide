import os

from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain_core.prompts import ChatPromptTemplate

load_dotenv()

model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)

prompt = ChatPromptTemplate.from_messages([
    ("system", "你是一名专业的订单客服，只根据用户提供的信息回答问题。"),
    ("human", "{question}"),
])

messages = prompt.invoke({
    "question": "帮我查询订单10001"
})

print("=== Prompt ===")
print(messages)

response = model.invoke(messages)

print()
print("=== LLM Response ===")
print(response.content)
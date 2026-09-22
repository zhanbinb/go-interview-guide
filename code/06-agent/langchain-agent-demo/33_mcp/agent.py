import asyncio
import os
from pathlib import Path

from dotenv import load_dotenv
from langchain.agents import create_agent
from langchain.mcp import MCPAdapter
from langchain_openai import ChatOpenAI


load_dotenv()


model = ChatOpenAI(
    model="MiniMax-M3",
    temperature=0,
    base_url=os.getenv("OPENAI_BASE_URL"),
    api_key=os.getenv("OPENAI_API_KEY"),
)


async def main():

    base_dir = Path(__file__).resolve().parent
    server_path = base_dir / "server.py"

    async with MCPAdapter(
        {
            "order_server": {
                "transport": "stdio",
                "command": "python",
                "args": [str(server_path)],
            }
        }
    ) as adapter:

        # MCP Server → LangChain Tools
        tools = await adapter.list_tools()

        print("========== MCP Tools ==========")

        for tool in tools:

            print("name:", tool.name)
            print("description:", tool.description)

            print("args_schema:")
            print(tool.args_schema)

            print()

        # LangChain Agent
        agent = create_agent(
            model=model,
            tools=tools,
        )

        print("========== Agent Start ==========")

        result = await agent.ainvoke(
            {
                "messages": [
                    {
                        "role": "user",
                        "content": "帮我查询一下订单10001的信息",
                    }
                ]
            }
        )

        print("\n========== Agent Trace ==========")

        for message in result["messages"]:

            print("\n--------------------")

            print(type(message).__name__)

            print(message.content)

            if getattr(message, "tool_calls", None):

                print("Tool Calls:")

                for tool_call in message.tool_calls:
                    print(tool_call)


if __name__ == "__main__":
    asyncio.run(main())
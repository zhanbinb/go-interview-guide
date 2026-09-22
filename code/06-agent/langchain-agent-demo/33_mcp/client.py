import asyncio
from pathlib import Path

from mcp import Client, StdioServerParameters


async def main():
    base_dir = Path(__file__).resolve().parent

    server = StdioServerParameters(
        command="python",
        args=[str(base_dir / "server.py")],
    )

    async with Client(server) as client:

        print("========== MCP Client ==========")

        result = await client.list_tools()

        print("\n========== Tools ==========")

        for tool in result.tools:
            print("name:", tool.name)
            print("description:", tool.description)
            print("input_schema:", tool.input_schema)

            if hasattr(tool, "output_schema"):
                print("output_schema:", tool.output_schema)

            print()

        print("========== Call Tool ==========")

        result = await client.call_tool(
            "query_order",
            {
                "order_id": "10001",
            },
        )

        print("\n========== Tool Result ==========")

        print("result:")
        print(result)

        print("\ncontent:")
        print(result.content)

        if hasattr(result, "structuredContent"):
            print("\nstructuredContent:")
            print(result.structuredContent)

        if hasattr(result, "structured_content"):
            print("\nstructured_content:")
            print(result.structured_content)


if __name__ == "__main__":
    asyncio.run(main())
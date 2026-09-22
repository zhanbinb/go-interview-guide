from mcp.server import MCPServer
from pydantic import BaseModel
mcp = MCPServer("Order Server")


class Order(BaseModel):
    order_id: str
    user: str
    status: str
    amount: int

# =========================
# Tool
# =========================

@mcp.tool()
def query_order(order_id: str) -> Order:
    """
    查询订单信息。
    """

    orders = {
        "10001": {
            "order_id": "10001",
            "user": "张三",
            "status": "已发货",
            "amount": 3999,
        },
        "10002": {
            "order_id": "10002",
            "user": "李四",
            "status": "待付款",
            "amount": 1999,
        },
    }

    order = orders.get(order_id)

    if order is None:
        raise ValueError(f"订单 {order_id} 不存在")

    return Order(**order)


# =========================
# Resource
# =========================

@mcp.resource("order://{order_id}")
def get_order_resource(order_id: str) -> str:
    """
    获取订单资源。
    """

    orders = {
        "10001": {
            "order_id": "10001",
            "user": "张三",
            "status": "已发货",
            "amount": 3999,
        },
        "10002": {
            "order_id": "10002",
            "user": "李四",
            "status": "待付款",
            "amount": 1999,
        },
    }

    order = orders.get(order_id)

    if not order:
        return f"订单 {order_id} 不存在"

    return (
        f"订单号：{order['order_id']}\n"
        f"用户：{order['user']}\n"
        f"状态：{order['status']}\n"
        f"金额：{order['amount']}"
    )


# =========================
# Prompt
# =========================

@mcp.prompt()
def order_summary(order_id: str) -> str:
    """
    生成订单总结 Prompt。
    """

    return (
        f"请查询订单 {order_id} 的相关信息，"
        f"并从订单状态、用户、金额三个方面进行简洁总结。"
    )


# =========================
# Start MCP Server
# =========================

if __name__ == "__main__":
    mcp.run()
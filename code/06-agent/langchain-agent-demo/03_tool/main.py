from langchain_core.tools import tool


@tool
def query_order(order_id: str) -> str:
    """查询订单信息。"""

    orders = {
        "10001": {
            "order_id": "10001",
            "status": "已发货",
            "amount": 3999,
        },
        "10002": {
            "order_id": "10002",
            "status": "待付款",
            "amount": 1999,
        },
    }

    order = orders.get(order_id)

    if not order:
        return f"订单 {order_id} 不存在"

    return (
        f"订单号：{order['order_id']}，"
        f"状态：{order['status']}，"
        f"金额：{order['amount']} 元"
    )


print("=== Tool Name ===")
print(query_order.name)

print()
print("=== Tool Description ===")
print(query_order.description)

print()
print("=== Tool Schema ===")
print(query_order.args_schema.schema())


print()
print("=== Tool Call ===")

result = query_order.invoke({
    "order_id": "10001"
})

print(result)

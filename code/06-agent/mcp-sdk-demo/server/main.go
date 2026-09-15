package main

import (
	"context"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// =========================
// greet Tool
// =========================

type GreetInput struct {
	Name string `json:"name" jsonschema:"the name of the person"`
}

type GreetOutput struct {
	Message string `json:"message"`
}

func Greet(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GreetInput,
) (
	*mcp.CallToolResult,
	GreetOutput,
	error,
) {
	return nil, GreetOutput{
		Message: "你好，" + input.Name,
	}, nil
}

// =========================
// get_user Tool
// =========================

type GetUserInput struct {
	UserID string `json:"user_id" jsonschema:"the user ID"`
}

type GetUserOutput struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Level  string `json:"level"`
}

func GetUser(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input GetUserInput,
) (
	*mcp.CallToolResult,
	GetUserOutput,
	error,
) {
	if input.UserID == "" {
		return nil, GetUserOutput{}, nil
	}

	return nil, GetUserOutput{
		UserID: input.UserID,
		Name:   "张三",
		Level:  "VIP",
	}, nil
}

// =========================
// query_order Tool
// =========================

type QueryOrderInput struct {
	OrderID string `json:"order_id" jsonschema:"the order ID"`
}

type QueryOrderOutput struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Amount  int    `json:"amount"`
}

func QueryOrder(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input QueryOrderInput,
) (
	*mcp.CallToolResult,
	QueryOrderOutput,
	error,
) {
	return nil, QueryOrderOutput{
		OrderID: input.OrderID,
		Status:  "已发货",
		Amount:  3999,
	}, nil
}

// =========================
// MCP Server
// =========================

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "mcp-sdk-demo-server",
			Version: "v1.0.0",
		},
		nil,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "greet",
			Description: "向用户打招呼",
		},
		Greet,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_user",
			Description: "查询用户信息",
		},
		GetUser,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "query_order",
			Description: "查询订单信息",
		},
		QueryOrder,
	)

	// 创建 Streamable HTTP Handler
	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return server
		},
		&mcp.StreamableHTTPOptions{
			JSONResponse: true,
		},
	)

	http.Handle("/mcp", handler)

	log.Println("MCP Server listening on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

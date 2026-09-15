package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// =========================
	// 1. 创建 MCP Client
	// =========================

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "mcp-sdk-demo-client",
			Version: "v1.0.0",
		},
		nil,
	)

	// =========================
	// 2. 创建 Transport
	// =========================

	transport := &mcp.StreamableClientTransport{
		Endpoint: "http://localhost:8080/mcp",
	}

	// =========================
	// 3. 连接 MCP Server
	// =========================

	session, err := client.Connect(
		ctx,
		transport,
		nil,
	)

	if err != nil {
		log.Fatalf(
			"connect MCP server failed: %v",
			err,
		)
	}

	defer session.Close()

	fmt.Println("=========================================================")
	fmt.Println("                    MCP SDK Demo")
	fmt.Println("=========================================================")

	// =========================
	// 4. 发现 Server 提供的 Tools
	// =========================

	fmt.Println()
	fmt.Println("=== Available Tools ===")

	toolsResult, err := session.ListTools(
		ctx,
		&mcp.ListToolsParams{},
	)

	if err != nil {
		log.Fatalf(
			"list MCP tools failed: %v",
			err,
		)
	}

	for _, tool := range toolsResult.Tools {
		fmt.Printf(
			"- %s: %s\n",
			tool.Name,
			tool.Description,
		)
	}

	// =========================
	// 5. 调用 get_user
	// =========================

	fmt.Println()
	fmt.Println("=== Call get_user ===")

	userResult, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name: "get_user",
			Arguments: map[string]any{
				"user_id": "1001",
			},
		},
	)

	if err != nil {
		log.Fatalf(
			"call get_user failed: %v",
			err,
		)
	}

	fmt.Println("Result:")

	for _, content := range userResult.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			fmt.Println(text.Text)
		}
	}

	// =========================
	// 6. 调用 query_order
	// =========================

	fmt.Println()
	fmt.Println("=== Call query_order ===")

	orderResult, err := session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name: "query_order",
			Arguments: map[string]any{
				"order_id": "10001",
			},
		},
	)

	if err != nil {
		log.Fatalf(
			"call query_order failed: %v",
			err,
		)
	}

	fmt.Println("Result:")

	for _, content := range orderResult.Content {
		if text, ok := content.(*mcp.TextContent); ok {
			fmt.Println(text.Text)
		}
	}

	fmt.Println()
	fmt.Println("=========================================================")
}

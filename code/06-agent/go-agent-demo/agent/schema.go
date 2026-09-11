package agent

import "github.com/openai/openai-go"

// BuildToolDefinitions 构造给 LLM 使用的 Tool Schema。
//
// 注意：
//
// Tool Schema
//
//	↓
//
// 告诉 LLM “有哪些工具、工具需要什么参数”
//
// Tool Registry
//
//	↓
//
// 真正告诉程序“如何执行这个工具”
//
// 两者不是一回事。
func BuildToolDefinitions() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		// =====================================================
		// query_order
		// =====================================================
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "query_order",
				Description: openai.String("查询订单信息"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"order_id": map[string]any{
							"type":        "string",
							"description": "订单号",
						},
					},
					"required": []string{
						"order_id",
					},
				},
			},
		},

		// =====================================================
		// get_user
		// =====================================================
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "get_user",
				Description: openai.String("查询用户信息"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"user_id": map[string]any{
							"type":        "string",
							"description": "用户ID",
						},
					},
					"required": []string{
						"user_id",
					},
				},
			},
		},

		// =====================================================
		// query_payment
		// =====================================================
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "query_payment",
				Description: openai.String("查询订单支付信息"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"order_id": map[string]any{
							"type":        "string",
							"description": "订单号",
						},
					},
					"required": []string{
						"order_id",
					},
				},
			},
		},

		// =====================================================
		// query_logistics
		// =====================================================
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "query_logistics",
				Description: openai.String("查询订单物流信息"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]any{
						"order_id": map[string]any{
							"type":        "string",
							"description": "订单号",
						},
					},
					"required": []string{
						"order_id",
					},
				},
			},
		},
	}
}

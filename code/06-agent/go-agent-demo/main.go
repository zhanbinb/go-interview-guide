package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// preset 描述一个模型供应商的连接配置，修改 presets 即可增删供应商
type preset struct {
	apiKeyEnv string // API Key 的环境变量名（从 .env 读取）
	baseURL   string // OpenAI 兼容的 Base URL
	model     string // 模型 ID
}

type Tool struct {
	Name        string
	Description string
	Handler     func(args string) (string, error)
}

func main() {
	// ============================================================
	// 模型预设：切换 provider 即可换接口，不用改其它代码
	// ============================================================
	presets := map[string]preset{
		"deepseek": {
			apiKeyEnv: "DEEPSEEK_API_KEY",
			baseURL:   "https://api.deepseek.com",
			model:     "deepseek-v4-flash",
		},
		"minimax": {
			// 官方 OpenAI 兼容端点（海外: https://api.minimax.io/v1）
			apiKeyEnv: "MINIMAX_API_KEY",
			baseURL:   "https://api.minimax.cn/v1",
			model:     "MiniMax-M3",
		},
		"openai": {
			apiKeyEnv: "OPENAI_API_KEY",
			baseURL:   "https://api.openai.com",
			model:     "gpt-4o-mini",
		},
	}

	// ★ 切换模型只需改这一行 ★
	provider := "minimax"
	const promptMsg = "帮我查询用户1001，并查询他的订单10001"
	// ==========================================

	p, ok := presets[provider]
	if !ok {
		log.Fatalf("未知 provider: %s（可选: %v）", provider, keysOf(presets))
	}
	fmt.Printf("当前使用: provider=%s, model=%s, baseURL=%s\n", provider, p.model, p.baseURL)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("加载 .env 失败:", err)
	}
	apiKey := os.Getenv(p.apiKeyEnv)
	if apiKey == "" {
		log.Fatalf("%s is not set", p.apiKeyEnv)
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(p.baseURL),
	)

	ctx := context.Background()

	//1. 用户消息
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(promptMsg),
	}

	queryOrderTool := Tool{
		Name:        "query_order",
		Description: "查询订单信息",
		Handler: func(args string) (string, error) {
			var params struct {
				OrderID string `json:"order_id"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			return queryOrder(params.OrderID), nil
		},
	}

	getUserTool := Tool{
		Name:        "get_user",
		Description: "查询用户信息",
		Handler: func(args string) (string, error) {
			var params struct {
				UserID string `json:"user_id"`
			}

			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			return getUser(params.UserID), nil
		},
	}

	toolRegistry := map[string]Tool{
		"query_order": queryOrderTool,
		"get_user":    getUserTool,
	}

	//2. 定义tool
	tools := []openai.ChatCompletionToolParam{
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
					"required": []string{"order_id"},
				},
			},
		},
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
					"required": []string{"user_id"},
				},
			},
		},
	}
	// ======AGENT 单次调用======================================================
	{

		// //3. 第一次调用模型
		// resp, err := client.Chat.Completions.New(
		// 	ctx,
		// 	openai.ChatCompletionNewParams{
		// 		Model:    p.model,
		// 		Messages: messages,
		// 		Tools:    tools,
		// 	},
		// )

		// if err != nil {
		// 	log.Fatal(err)
		// }

		// if len(resp.Choices) == 0 {
		// 	log.Fatal("No choices returned")
		// }

		// //fmt.Println(resp.Choices[0].Message.Content)
		// message := resp.Choices[0].Message
		// //4. 把模型的tool call 加入消息历史
		// messages = append(messages, message.ToParam())

		// //5. 执行Tool
		// for _, toolCall := range message.ToolCalls {
		// 	fmt.Printf("Tool Name: %s\n", toolCall.Function.Name)
		// 	fmt.Printf("Arguments: %s\n", toolCall.Function.Arguments)

		// 	// if toolCall.Function.Name == "query_order" {

		// 	// 	var args struct {
		// 	// 		OrderID string `json:"order_id"`
		// 	// 	}

		// 	// 	err := json.Unmarshal(
		// 	// 		[]byte(toolCall.Function.Arguments),
		// 	// 		&args,
		// 	// 	)

		// 	// 	if err != nil {
		// 	// 		log.Fatal(err)
		// 	// 	}

		// 	// 	result := queryOrder(args.OrderID)

		// 	// 	fmt.Println("Tool Result:", result)

		// 	// 	//6. 把Tool Result 加入消息历史
		// 	// 	messages = append(messages, openai.ToolMessage(result, toolCall.ID))
		// 	// }
		// 	toolName := toolCall.Function.Name
		// 	tool := toolRegistry[toolName]
		// 	if !ok {
		// 		log.Fatalf("未知 tool: %s", toolName)
		// 		continue
		// 	}
		// 	result, err := tool.Handler(toolCall.Function.Arguments)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	fmt.Println("Tool Result:", result)
		// 	messages = append(messages, openai.ToolMessage(result, toolCall.ID))
		// }

		// if err != nil {
		// 	log.Fatal(err)
		// }

		// if len(resp.Choices) == 0 {
		// 	log.Fatal("No choices returned")
		// }

		// // 7. 第二次调用 DeepSeek
		// finalResp, err := client.Chat.Completions.New(
		// 	ctx,
		// 	openai.ChatCompletionNewParams{
		// 		Model:    p.model,
		// 		Messages: messages,
		// 	},
		// )

		// if err != nil {
		// 	log.Fatal(err)
		// }

		// // 8. 输出最终答案
		// fmt.Println("Final Answer:")
		// fmt.Println(finalResp.Choices[0].Message.Content)

		// ============================================================
	}

	// ======AGENT LOOP=====================================================
	for {
		resp, err := client.Chat.Completions.New(
			ctx,
			openai.ChatCompletionNewParams{
				Model:    p.model,
				Messages: messages,
				Tools:    tools,
			},
		)

		if err != nil {
			log.Fatal(err)
		}

		message := resp.Choices[0].Message

		// 保存 Assistant 的消息
		messages = append(messages, message.ToParam())

		// 没有 Tool Call
		if len(message.ToolCalls) == 0 {
			fmt.Println("Final Answer:")
			fmt.Println(message.Content)
			break
		}

		// 有 Tool Call，执行 Tool
		for _, toolCall := range message.ToolCalls {

			toolName := toolCall.Function.Name

			tool, ok := toolRegistry[toolName]
			if !ok {
				log.Printf("unknown tool: %s", toolName)
				continue
			}

			result, err := tool.Handler(
				toolCall.Function.Arguments,
			)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Tool:", toolName)
			fmt.Println("Result:", result)

			messages = append(
				messages,
				openai.ToolMessage(
					result,
					toolCall.ID,
				),
			)
		}
	}
}

// keysOf 返回 map 的所有 key，用于错误提示时列出可选项
func keysOf(m map[string]preset) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func queryOrder(orderID string) string {
	return fmt.Sprintf(
		"订单 %s 已发货，金额3999元，预计明天送达",
		orderID,
	)
}

func getUser(userID string) string {
	return fmt.Sprintf("用户 %s 张三,VIP用户，账户状态正常", userID)
}

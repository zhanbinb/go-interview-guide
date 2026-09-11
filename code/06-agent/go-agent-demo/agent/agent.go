package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go"

	"go-agent-demo/memory"
	"go-agent-demo/tools"
)

// Agent 是 Agent Loop 的核心执行器。
//
// Agent 负责：
// 1. 调用 LLM
// 2. 判断 LLM 是否要求调用 Tool
// 3. 执行 Tool
// 4. 将 Tool Result 放回上下文
// 5. 循环调用 LLM
// 6. 最终得到 Final Answer
type Agent struct {
	client       *openai.Client
	model        string
	toolRegistry *tools.Registry
	memory       *memory.ContextManager
}

// New 创建 Agent。
func New(
	client *openai.Client,
	model string,
	toolRegistry *tools.Registry,
	memory *memory.ContextManager,
) *Agent {
	return &Agent{
		client:       client,
		model:        model,
		toolRegistry: toolRegistry,
		memory:       memory,
	}
}

// Run 执行 Agent Loop。
//
// 基本流程：
//
//	User Message
//	    ↓
//	LLM
//	    ↓
//	是否需要 Tool？
//	   /     \
//	  是      否
//	  ↓       ↓
//	Tool    Final Answer
//	  ↓
//	Tool Result
//	  ↓
//	LLM
//	  ↓
//	...
func (a *Agent) Run(ctx context.Context) (string, error) {
	// Tool Schema 告诉 LLM：
	// 当前 Agent 有哪些 Tool 可以调用。
	toolDefinitions := BuildToolDefinitions()

	for {
		// =====================================================
		// 1. 检查 Context 是否需要摘要
		// =====================================================

		if err := a.memory.MaybeSummarize(
			ctx,
			a.client,
			a.model,
		); err != nil {
			return "", fmt.Errorf(
				"maybe summarize context: %w",
				err,
			)
		}

		// =====================================================
		// 2. 构造当前上下文
		// =====================================================

		messages := a.memory.BuildMessages()

		// =====================================================
		// 3. 调用 LLM
		// =====================================================

		resp, err := a.client.Chat.Completions.New(
			ctx,
			openai.ChatCompletionNewParams{
				Model:    a.model,
				Messages: messages,
				Tools:    toolDefinitions,
			},
		)
		if err != nil {
			return "", fmt.Errorf(
				"call LLM: %w",
				err,
			)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf(
				"LLM returned no choices",
			)
		}

		message := resp.Choices[0].Message

		// =====================================================
		// 4. 保存 Assistant Message
		// =====================================================

		a.memory.Add(
			message.ToParam(),
		)

		// =====================================================
		// 5. 没有 Tool Call
		//
		// 说明 LLM 已经可以直接回答用户。
		// =====================================================

		if len(message.ToolCalls) == 0 {
			fmt.Println()
			fmt.Println("================================")
			fmt.Println("Final Answer:")
			fmt.Println(message.Content)
			fmt.Println("================================")

			return message.Content, nil
		}

		// =====================================================
		// 6. 处理 Tool Calls
		//
		// 一次 Assistant Message 可能要求调用多个 Tool。
		// =====================================================

		for _, toolCall := range message.ToolCalls {
			toolName := toolCall.Function.Name
			arguments := toolCall.Function.Arguments

			fmt.Println()
			fmt.Println("Tool Call:")
			fmt.Println("Name:", toolName)
			fmt.Println("Arguments:", arguments)

			// -------------------------------------------------
			// 6.1 从 Registry 查找 Tool
			// -------------------------------------------------

			tool, ok := a.toolRegistry.Get(toolName)

			if !ok {
				log.Printf(
					"unknown tool: %s",
					toolName,
				)

				result := fmt.Sprintf(
					"unknown tool: %s",
					toolName,
				)

				a.memory.Add(
					openai.ToolMessage(
						result,
						toolCall.ID,
					),
				)

				continue
			}

			// -------------------------------------------------
			// 6.2 执行 Tool
			// -------------------------------------------------

			result, err := tool.Handler(arguments)

			if err != nil {
				log.Printf(
					"tool %s failed: %v",
					toolName,
					err,
				)

				// Tool 执行失败也作为 Tool Result
				// 返回给 LLM。
				result = fmt.Sprintf(
					"tool %s failed: %v",
					toolName,
					err,
				)
			}

			fmt.Println("Tool Result:")
			fmt.Println(result)

			// -------------------------------------------------
			// 6.3 把 Tool Result 放回 Context
			// -------------------------------------------------

			a.memory.Add(
				openai.ToolMessage(
					result,
					toolCall.ID,
				),
			)
		}

		// =====================================================
		// 7. 继续下一轮
		//
		// 下一轮 LLM 可以看到：
		//
		// Assistant Tool Call
		//        +
		// Tool Result
		//
		// 然后决定：
		// - 再调用 Tool
		// - 或者直接生成最终答案
		// =====================================================
	}
}

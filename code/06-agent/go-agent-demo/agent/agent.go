package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go-agent-demo/llm"
	"go-agent-demo/rag"

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
	client         *openai.Client
	model          string
	toolRegistry   *tools.Registry
	contextManager *memory.ContextManager
	longTermMemory *memory.Memory
	queryRewriter  *QueryRewriter
	rag            *rag.RAG
}

// New 创建 Agent。
func New(
	client *openai.Client,
	model string,
	toolRegistry *tools.Registry,
	contextManager *memory.ContextManager,
	longTermMemory *memory.Memory,
	queryRewriter *QueryRewriter,
	ragService *rag.RAG,
) *Agent {
	return &Agent{
		client: client,
		model:  model,

		toolRegistry:   toolRegistry,
		contextManager: contextManager,

		longTermMemory: longTermMemory,
		queryRewriter:  queryRewriter,

		rag: ragService,
	}
}

// Chat 开始一次新的用户对话。
//
// Chat 负责把用户输入加入当前 Context，
// 然后交给 Agent.Run() 执行完整 Agent Loop。
func (a *Agent) Chat(
	ctx context.Context,
	prompt string,
) (string, error) {

	a.contextManager.AddUserMessage(prompt)

	return a.Run(ctx)
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

		if err := a.contextManager.MaybeSummarize(
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

		messages := a.contextManager.BuildMessages()

		query := a.contextManager.LastUserMessage()

		memoryContext := a.buildMemoryContext(ctx, query)
		ragContext, err := a.buildRAGContext(ctx, query)
		if err != nil {
			return "", fmt.Errorf("build rag context: %w", err)
		}
		if ragContext != "" {
			messages = append(
				messages,
				openai.UserMessage(ragContext),
			)
		}
		if memoryContext != "" {
			messages = append(
				messages,
				openai.UserMessage(memoryContext),
			)
		}

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

		a.contextManager.Add(
			message.ToParam(),
		)

		// =====================================================
		// 5. 没有 Tool Call
		//
		// 说明 LLM 已经可以直接回答用户。
		// =====================================================

		if len(message.ToolCalls) == 0 {
			// 移除 <think>...</think> 标签
			content := llm.CleanThinking(message.Content)

			fmt.Println()
			fmt.Println("================================")
			fmt.Println("Final Answer:")
			fmt.Println(content)
			fmt.Println("================================")

			// 当前轮对话已经完成。
			//
			// 现在让 LLM 判断：
			// 用户刚才说的话中有没有值得长期保存的信息。
			if err := a.extractAndSaveMemory(
				ctx,
				query,
				content,
			); err != nil {
				log.Printf("memory extraction failed: %v", err)
			}

			return content, nil
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

				a.contextManager.Add(
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

			a.contextManager.Add(
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

func (a *Agent) buildRAGContext(
	ctx context.Context,
	query string,
) (string, error) {
	if a.rag == nil {
		return "", nil
	}

	results, err := a.rag.Retrieve(ctx, query)
	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "", nil
	}

	var builder strings.Builder

	builder.WriteString(
		"以下是知识库中与当前问题相关的参考资料：\n",
	)

	for _, result := range results {
		builder.WriteString(
			fmt.Sprintf(
				"\n[%s]\n%s\n",
				result.Chunk.ID,
				result.Chunk.Content,
			),
		)
	}

	builder.WriteString(
		"\n回答问题时，如果使用知识库内容，请以知识库为依据，不要编造知识库中不存在的信息。",
	)

	return builder.String(), nil
}

// func (a *Agent) buildMemoryContext(query string) string {
// 	results := a.longTermMemory.Search("Go")

// 	if len(results) == 0 {
// 		return ""
// 	}

// 	var builder strings.Builder

// 	builder.WriteString("以下是与当前请求相关的长期记忆：\n")

// 	for _, item := range results {
// 		builder.WriteString("- ")
// 		builder.WriteString(item.Key)
// 		builder.WriteString(": ")
// 		builder.WriteString(item.Value)
// 		builder.WriteString("\n")
// 	}

// 	return builder.String()
// }

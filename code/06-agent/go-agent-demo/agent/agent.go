package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"go-agent-demo/llm"
	"go-agent-demo/memory"
	"go-agent-demo/rag"
	"go-agent-demo/tools"

	mcpclient "go-agent-demo/mcp/client"

	"github.com/openai/openai-go"
)

// Agent 是 Agent Loop 的核心执行器。
//
// Agent 负责：
// 1. 调用 Router 判断当前问题需要哪些能力
// 2. 调用 LLM
// 3. 判断 LLM 是否要求调用 Tool
// 4. 执行 Tool
// 5. 将 Tool Result 放回上下文
// 6. 循环调用 LLM
// 7. 最终得到 Final Answer
type Agent struct {
	client *openai.Client
	model  string

	// Router 负责判断当前用户问题需要哪些能力：
	// Memory / RAG / Tools
	router *Router

	toolRegistry   *tools.Registry
	mcpClient      *mcpclient.Client
	contextManager *memory.ContextManager
	longTermMemory *memory.Memory
	queryRewriter  *QueryRewriter
	rag            *rag.RAG
}

// New 创建 Agent。
func New(
	client *openai.Client,
	model string,
	router *Router,
	toolRegistry *tools.Registry,
	contextManager *memory.ContextManager,
	longTermMemory *memory.Memory,
	queryRewriter *QueryRewriter,
	ragService *rag.RAG,
	mcpClient *mcpclient.Client,
) *Agent {
	return &Agent{
		client: client,
		model:  model,

		router: router,

		toolRegistry:   toolRegistry,
		contextManager: contextManager,

		longTermMemory: longTermMemory,
		queryRewriter:  queryRewriter,

		rag:       ragService,
		mcpClient: mcpClient,
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
// 整体流程：
//
//	User Query
//	    ↓
//	 Router
//	    ↓
//	Route Decision
//	   / | \
//	  /  |  \
//	Memory RAG Tools
//	  \   |   /
//	   \  |  /
//	    ↓
//	   LLM
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

	// =====================================================
	// 1. Router
	//
	// Router 只针对当前 User Query 判断一次。
	//
	// 注意：
	// Router 不应该放到下面的 Agent Loop for 中。
	//
	// 因为 Tool Call -> Tool Result -> LLM
	// 都属于同一次用户请求。
	// =====================================================

	query := a.contextManager.LastUserMessage()

	decision, err := a.router.Decide(
		ctx,
		query,
		BuildCapabilities(),
	)
	if err != nil {
		return "", fmt.Errorf(
			"router decide: %w",
			err,
		)
	}

	fmt.Println()
	fmt.Println("Router Decision:")
	fmt.Printf(
		"Memory=%v, RAG=%v, Tool=%v\n",
		decision.UseMemory,
		decision.UseRAG,
		decision.UseTool,
	)

	// =====================================================
	// 2. 根据 Router Decision 准备能力
	// =====================================================
	//
	// Router 只是做决定。
	//
	// 真正执行能力的是 Agent：
	//
	// UseMemory=true
	//        ↓
	// buildMemoryContext()
	//
	// UseRAG=true
	//        ↓
	// buildRAGContext()
	//
	// UseTool=true
	//        ↓
	// 给 LLM 提供 Tool Schema
	// =====================================================

	var memoryContext string

	if decision.UseMemory {
		memoryContext = a.buildMemoryContext(
			ctx,
			query,
		)
	}

	var ragContext string

	if decision.UseRAG {
		ragContext, err = a.buildRAGContext(
			ctx,
			query,
		)
		if err != nil {
			return "", fmt.Errorf(
				"build rag context: %w",
				err,
			)
		}
	}

	// =====================================================
	// 3. 根据 Router Decision 决定是否向 LLM 提供 Tool
	// =====================================================

	var toolDefinitions []openai.ChatCompletionToolParam

	if decision.UseTool {
		toolDefinitions = BuildToolDefinitions()

		if a.mcpClient != nil {
			mcpTools, err := a.mcpClient.ListTools(ctx)
			if err != nil {
				return "", fmt.Errorf("list MCP tools failed: %w", err)
			}

			mcpDefinitions := mcpclient.ConvertTools(mcpTools)

			toolDefinitions = append(
				toolDefinitions,
				mcpDefinitions...,
			)
		}
	}

	// =====================================================
	// 4. Agent Loop
	// =====================================================

	for {

		// =====================================================
		// 4.1 检查 Context 是否需要摘要
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
		// 4.2 构造当前上下文
		// =====================================================

		messages := a.contextManager.BuildMessages()

		// 如果 Router 判断需要 RAG，
		// 则把 RAG 检索结果加入当前上下文。
		if ragContext != "" {
			messages = append(
				messages,
				openai.UserMessage(ragContext),
			)
		}

		// 如果 Router 判断需要 Memory，
		// 则把长期记忆加入当前上下文。
		if memoryContext != "" {
			messages = append(
				messages,
				openai.UserMessage(memoryContext),
			)
		}

		// =====================================================
		// 4.3 调用 LLM
		// =====================================================

		resp, err := a.client.Chat.Completions.New(
			ctx,
			openai.ChatCompletionNewParams{
				Model:    a.model,
				Messages: messages,

				// 如果 Router 判断不需要 Tool，
				// 这里就是 nil。
				//
				// LLM 就不会知道当前 Agent 有哪些 Tool。
				Tools: toolDefinitions,
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
		// 4.4 保存 Assistant Message
		// =====================================================

		a.contextManager.Add(
			message.ToParam(),
		)

		// =====================================================
		// 4.5 没有 Tool Call
		//
		// 说明 LLM 已经可以直接回答用户。
		// =====================================================

		if len(message.ToolCalls) == 0 {

			// 移除 <think>...</think> 标签
			content := llm.CleanThinking(
				message.Content,
			)

			fmt.Println()
			fmt.Println("================================")
			fmt.Println("Final Answer:")
			fmt.Println(content)
			fmt.Println("================================")

			// =================================================
			// 当前轮对话已经完成。
			//
			// 让 LLM 判断：
			// 用户刚才说的话中有没有值得长期保存的信息。
			// =================================================

			if err := a.extractAndSaveMemory(
				ctx,
				query,
				content,
			); err != nil {
				log.Printf(
					"memory extraction failed: %v",
					err,
				)
			}

			return content, nil
		}

		// =====================================================
		// 4.6 处理 Tool Calls
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
			// 4.6.1 判断 Tool 来源
			//
			// Tool 可能来自：
			//
			// 1. 本地 Tool Registry
			// 2. MCP Server
			// -------------------------------------------------

			isMCP, err := a.isMCPTool(
				ctx,
				toolName,
			)

			if err != nil {
				return "", fmt.Errorf(
					"check MCP tool: %w",
					err,
				)
			}

			var result string

			// -------------------------------------------------
			// 4.6.2 MCP Tool
			// -------------------------------------------------

			if isMCP {

				fmt.Println("Tool Source: MCP")

				mcpResult, err := a.mcpClient.CallTool(
					ctx,
					toolName,
					parseToolArguments(arguments),
				)

				if err != nil {

					log.Printf(
						"MCP tool %s failed: %v",
						toolName,
						err,
					)

					result = fmt.Sprintf(
						"MCP tool %s failed: %v",
						toolName,
						err,
					)

				} else {

					resultBytes, err := json.Marshal(
						mcpResult,
					)

					if err != nil {
						result = fmt.Sprintf(
							"MCP tool %s result marshal failed: %v",
							toolName,
							err,
						)
					} else {
						result = string(resultBytes)
					}
				}

			} else {

				// -------------------------------------------------
				// 4.6.3 本地 Tool
				// -------------------------------------------------

				fmt.Println("Tool Source: Local")

				tool, ok := a.toolRegistry.Get(
					toolName,
				)

				if !ok {

					result = fmt.Sprintf(
						"unknown tool: %s",
						toolName,
					)

				} else {

					resultValue, err := tool.Handler(
						arguments,
					)

					if err != nil {

						log.Printf(
							"tool %s failed: %v",
							toolName,
							err,
						)

						result = fmt.Sprintf(
							"tool %s failed: %v",
							toolName,
							err,
						)

					} else {
						result = resultValue
					}
				}
			}

			// -------------------------------------------------
			// 4.6.4 Tool Result
			// -------------------------------------------------

			fmt.Println("Tool Result:")
			fmt.Println(result)

			// -------------------------------------------------
			// 4.6.5 把 Tool Result 放回 Context
			// -------------------------------------------------

			a.contextManager.Add(
				openai.ToolMessage(
					result,
					toolCall.ID,
				),
			)
		}

		// =====================================================
		// 4.7 继续下一轮 Agent Loop
		//
		// 下一轮 LLM 可以看到：
		//
		// Assistant Tool Call
		//        +
		// Tool Result
		//
		// 然后决定：
		//
		// - 再调用 Tool
		// - 或者直接生成最终答案
		// =====================================================
	}
}

// buildRAGContext 根据用户问题从 RAG 中检索知识。
func (a *Agent) buildRAGContext(
	ctx context.Context,
	query string,
) (string, error) {

	if a.rag == nil {
		return "", nil
	}

	results, err := a.rag.Retrieve(
		ctx,
		query,
	)
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

func (a *Agent) isMCPTool(
	ctx context.Context,
	name string,
) (bool, error) {
	if a.mcpClient == nil {
		return false, nil
	}

	mcpTools, err := a.mcpClient.ListTools(ctx)
	if err != nil {
		return false, err
	}

	for _, tool := range mcpTools {
		if tool.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func parseToolArguments(
	arguments string,
) map[string]any {

	var result map[string]any

	if err := json.Unmarshal(
		[]byte(arguments),
		&result,
	); err != nil {
		return map[string]any{}
	}

	return result
}

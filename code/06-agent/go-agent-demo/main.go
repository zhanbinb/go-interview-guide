package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"

	"go-agent-demo/agent"
	"go-agent-demo/config"
	"go-agent-demo/embedding"
	"go-agent-demo/llm"
	mcpclient "go-agent-demo/mcp/client"
	"go-agent-demo/memory"
	"go-agent-demo/rag"
	"go-agent-demo/tools"
	"go-agent-demo/workflow"
)

func main_backup() {
	ctx := context.Background()

	// =========================================================
	// 1. 初始化所有依赖
	// =========================================================

	client,
		model,
		router,
		queryRewriter,
		toolRegistry,
		mcpClient,
		longTermMemory,
		ragService := initialize()

	// =========================================================
	// 2. MCP Demo
	// =========================================================

	runMCPDemo(
		ctx,
		client,
		model,
		mcpClient,
	)

	// =========================================================
	// 3. Memory + Session Demo
	// =========================================================

	agent2 := runMemoryDemo(
		ctx,
		client,
		model,
		router,
		toolRegistry,
		longTermMemory,
		queryRewriter,
		ragService,
		mcpClient,
	)

	// =========================================================
	// 4. Router Test
	// =========================================================

	runRouterTest(
		ctx,
		router,
	)

	// =========================================================
	// 5. RAG Test
	// =========================================================

	runRAGTest(
		ctx,
		agent2,
	)

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                    Demo Finished")
	fmt.Println("=========================================================")
}

func main() {
	ctx := context.Background()

	// =========================================================
	// 当前学习主题：MCP + Agent
	// =========================================================

	client,
		model,
		router,
		queryRewriter,
		toolRegistry,
		mcpClient,
		longTermMemory,
		ragService := initialize()

	contextManager := memory.NewContextManager(nil)

	agentService := agent.New(
		client,
		model,
		router,
		toolRegistry,
		contextManager,
		longTermMemory,
		queryRewriter,
		ragService,
		mcpClient,
	)

	answer, err := agentService.Chat(
		ctx,
		"帮我查询一下订单10001的状态和金额",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("=== Agent Final Answer ===")
	fmt.Println(answer)
}

// =========================================================
// Application Initialization
// =========================================================

func initialize() (
	*openai.Client,
	string,
	*agent.Router,
	*agent.QueryRewriter,
	*tools.Registry,
	*mcpclient.Client,
	*memory.Memory,
	*rag.RAG,
) {

	// =========================================================
	// 1. 加载配置
	// =========================================================

	presets, err := config.Load()
	if err != nil {
		log.Fatalf(
			"load config failed: %v",
			err,
		)
	}

	provider := "minimax"

	p, ok := presets[provider]
	if !ok {
		log.Fatalf(
			"provider %s not found, available providers: %v",
			provider,
			workflow.KeysOf(presets),
		)
	}

	apiKey := os.Getenv(p.APIKey)

	if apiKey == "" {
		log.Fatalf(
			"environment variable %s is not set",
			p.APIKey,
		)
	}

	// =========================================================
	// 2. 创建 LLM Client
	// =========================================================

	client := llm.NewClient(
		apiKey,
		p.BaseURL,
	)

	model := p.Model

	// =========================================================
	// 3. Router
	// =========================================================

	router := agent.NewRouter(
		&client,
		model,
	)

	// =========================================================
	// 4. Query Rewriter
	// =========================================================

	queryRewriter := agent.NewQueryRewriter(
		&client,
		model,
	)

	// =========================================================
	// 5. Tool Registry
	// =========================================================

	toolRegistry := tools.NewRegistry()

	// =========================================================
	// 6. MCP Server / Client
	// =========================================================

	// 官方 MCP Client
	mcpClient, err := mcpclient.New(
		context.Background(),
		"http://localhost:8080/mcp",
	)

	if err != nil {
		log.Fatalf(
			"connect MCP server failed: %v",
			err,
		)
	}

	// =========================================================
	// 7. Long-term Memory
	// =========================================================

	embedder := embedding.NewFakeEmbedder(8)

	longTermMemory := memory.NewMemory(
		embedder,
	)

	// =========================================================
	// 8. RAG
	// =========================================================

	knowledgeBase := rag.NewKnowledgeBase(
		embedder,
	)

	err = knowledgeBase.AddDocument(
		rag.Document{
			ID: "refund-policy",
			Content: `退款规则

普通商品支持购买后7天内申请退款。

如果商品已经发货，用户仍然可以提交退款申请，
但需要根据物流状态进行处理。

虚拟商品一旦完成交付，通常不支持退款。

退款金额原则上按照实际支付金额计算。`,
		},
	)

	if err != nil {
		log.Fatalf(
			"add knowledge document failed: %v",
			err,
		)
	}

	retriever := rag.NewRetriever(
		knowledgeBase,
		embedder,
	)

	ragService := rag.NewRAG(
		retriever,
		&client,
		model,
	)

	return &client,
		model,
		router,
		queryRewriter,
		toolRegistry,
		mcpClient,
		longTermMemory,
		ragService

}

// =========================================================
// MCP Demo
// =========================================================

func runMCPDemo(
	ctx context.Context,
	client *openai.Client,
	model string,
	mcpClient *mcpclient.Client,
) {

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                    MCP Demo")
	fmt.Println("=========================================================")

	// =========================================================
	// 1. MCP Client 获取 Tool
	// =========================================================

	toolInfos, err := mcpClient.ListTools(ctx)
	if err != nil {
		log.Printf(
			"mcp list tools failed: %v",
			err,
		)
		return
	}

	fmt.Println()
	fmt.Println("=== MCP Client Tools ===")

	for _, tool := range toolInfos {
		fmt.Printf(
			"- %s: %s\n",
			tool.Name,
			tool.Description,
		)
	}

	// =========================================================
	// 2. MCP Tool → LLM Tool
	// =========================================================

	llmTools := mcpclient.ConvertTools(
		toolInfos,
	)

	fmt.Println()
	fmt.Println("=== LLM Tools ===")

	for _, tool := range llmTools {
		fmt.Printf(
			"- %s: %s\n",
			tool.Function.Name,
			tool.Function.Description.Value,
		)
	}

	// =========================================================
	// 3. 第一次调用 LLM
	// =========================================================

	fmt.Println()
	fmt.Println("=== LLM Tool Calling ===")

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(
			"请查询订单10001的订单状态和金额。",
		),
	}

	completion, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:    model,
			Messages: messages,
			Tools:    llmTools,
		},
	)

	if err != nil {
		log.Printf(
			"llm chat failed: %v",
			err,
		)
		return
	}

	message := completion.Choices[0].Message

	fmt.Println("LLM Response:")
	fmt.Println(message.Content)

	// =========================================================
	// 4. LLM Tool Call
	// =========================================================

	fmt.Println()
	fmt.Println("Tool Calls:")

	if len(message.ToolCalls) == 0 {
		fmt.Println("LLM 没有请求调用工具。")
		return
	}

	// =========================================================
	// 5. 构造 Assistant Tool Call Message
	// =========================================================

	assistantToolCalls := make(
		[]openai.ChatCompletionMessageToolCallParam,
		0,
		len(message.ToolCalls),
	)

	for _, toolCall := range message.ToolCalls {

		fmt.Printf(
			"- name=%s arguments=%s\n",
			toolCall.Function.Name,
			toolCall.Function.Arguments,
		)

		assistantToolCalls = append(
			assistantToolCalls,
			openai.ChatCompletionMessageToolCallParam{
				ID: toolCall.ID,

				Function: openai.ChatCompletionMessageToolCallFunctionParam{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},

				Type: toolCall.Type,
			},
		)
	}

	assistantMessage := openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			// Content:   message.Content,
			ToolCalls: assistantToolCalls,
		},
	}

	// =========================================================
	// 6. 执行 MCP Tool
	// =========================================================

	for _, toolCall := range message.ToolCalls {

		var arguments map[string]any

		if err := json.Unmarshal(
			[]byte(toolCall.Function.Arguments),
			&arguments,
		); err != nil {
			log.Printf(
				"parse tool arguments failed: %v",
				err,
			)
			return
		}

		result, err := mcpClient.CallTool(
			ctx,
			toolCall.Function.Name,
			arguments,
		)

		if err != nil {
			log.Printf(
				"mcp call tool failed: %v",
				err,
			)
			return
		}

		fmt.Println()
		fmt.Println("MCP Tool Result:")
		fmt.Println(result)

		// =====================================================
		// 7. Tool Result → JSON
		// =====================================================

		toolResultBytes, err := json.Marshal(result)
		if err != nil {
			log.Printf(
				"marshal tool result failed: %v",
				err,
			)
			return
		}

		fmt.Println()
		fmt.Println("=== Tool Result JSON ===")
		fmt.Println(string(toolResultBytes))

		// =====================================================
		// 8. Assistant Tool Call + Tool Result
		// =====================================================

		messages = append(
			messages,
			assistantMessage,
			openai.ToolMessage(
				string(toolResultBytes),
				toolCall.ID,
			),
		)
	}

	// =========================================================
	// 9. 第二次调用 LLM
	// =========================================================

	finalCompletion, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		log.Printf(
			"llm final answer failed: %v",
			err,
		)
		return
	}

	finalAnswer := finalCompletion.Choices[0].Message.Content

	// =========================================================
	// 10. 最终回答
	// =========================================================

	fmt.Println()
	fmt.Println("=== Final Answer ===")
	fmt.Println(
		llm.CleanThinking(finalAnswer),
	)
}

// =========================================================
// Memory + Session Demo
// =========================================================

func runMemoryDemo(
	ctx context.Context,
	client *openai.Client,
	model string,
	router *agent.Router,
	toolRegistry *tools.Registry,
	longTermMemory *memory.Memory,
	queryRewriter *agent.QueryRewriter,
	ragService *rag.RAG,
	mcpClient *mcpclient.Client,
) *agent.Agent {

	// =========================================================
	// Session 1
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                 Session 1")
	fmt.Println("=========================================================")

	ctxManager1 := memory.NewContextManager(
		nil,
	)

	agent1 := agent.New(
		client,
		model,
		router,
		toolRegistry,
		ctxManager1,
		longTermMemory,
		queryRewriter,
		ragService,
		mcpClient,
	)

	session1Prompts := []string{
		"我叫张三，我是一名 Go 后端开发工程师。",
		"我最近主要想找 Go 后端开发相关的工作。",
	}

	runSession(
		ctx,
		agent1,
		"Session 1",
		session1Prompts,
	)

	// =========================================================
	// 查看 Session 1 产生的 Long-term Memory
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("       Long-term Memory After Session 1")
	fmt.Println("=========================================================")

	printMemory(
		longTermMemory,
	)

	// =========================================================
	// Vector Search Test
	// =========================================================

	fmt.Println()
	fmt.Println("========== Vector Search Test ==========")

	vectorResults := longTermMemory.SearchVectorTopK(
		"我的职业方向是什么？",
		5,
	)

	for _, result := range vectorResults {
		fmt.Printf(
			"key=%s similarity=%.4f value=%s\n",
			result.Item.Key,
			result.Similarity,
			result.Item.Value,
		)
	}

	fmt.Println("========================================")

	// =========================================================
	// Session 2
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                 Session 2")
	fmt.Println("=========================================================")

	ctxManager2 := memory.NewContextManager(
		nil,
	)

	agent2 := agent.New(
		client,
		model,
		router,
		toolRegistry,
		ctxManager2,
		longTermMemory,
		queryRewriter,
		ragService,
		mcpClient,
	)

	session2Prompts := []string{
		"你还记得我的职业吗？",
		"你还记得我之前说过我想找什么工作吗？",
		"你还记得我的名字吗？",
		"你还记得我的年龄吗？",
	}

	runSession(
		ctx,
		agent2,
		"Session 2",
		session2Prompts,
	)

	// =========================================================
	// 最终 Long-term Memory
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("             Final Long-term Memory")
	fmt.Println("=========================================================")

	printMemory(
		longTermMemory,
	)

	return agent2
}

// =========================================================
// Session
// =========================================================

func runSession(
	ctx context.Context,
	a *agent.Agent,
	sessionName string,
	prompts []string,
) {

	for i, prompt := range prompts {

		fmt.Println()
		fmt.Println("-----------------------------------------")
		fmt.Printf(
			"%s - Turn %d\n",
			sessionName,
			i+1,
		)

		fmt.Println("User:")
		fmt.Println(prompt)

		fmt.Println("-----------------------------------------")

		answer, err := a.Chat(
			ctx,
			prompt,
		)

		if err != nil {
			log.Printf(
				"%s agent chat failed: %v",
				sessionName,
				err,
			)
			continue
		}

		fmt.Println()
		fmt.Println("Agent:")
		fmt.Println(answer)
	}
}

// =========================================================
// Memory
// =========================================================

func printMemory(
	longTermMemory *memory.Memory,
) {

	for _, item := range longTermMemory.All() {
		fmt.Printf(
			"%s = %s\n",
			item.Key,
			item.Value,
		)
	}
}

// =========================================================
// Router Test
// =========================================================

func runRouterTest(
	ctx context.Context,
	router *agent.Router,
) {

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                    Router Test")
	fmt.Println("=========================================================")

	// 使用 Router 自己的能力定义。
	//
	// 不再在 main.go 重复维护 capabilities。
	capabilities := agent.BuildCapabilities()

	routerQueries := []string{
		"你还记得我的名字吗？",
		"普通商品购买后多少天可以申请退款？",
		"查询订单10001现在是什么状态？",
		"什么是 Go 的 Goroutine？",
		"根据退款规则，帮我判断订单10001能不能退款？",
	}

	for i, query := range routerQueries {

		fmt.Println()
		fmt.Println("-----------------------------------------")
		fmt.Printf(
			"Router Test - %d\n",
			i+1,
		)

		fmt.Println("User:")
		fmt.Println(query)

		decision, err := router.Decide(
			ctx,
			query,
			capabilities,
		)

		if err != nil {
			log.Printf(
				"router decide failed: %v",
				err,
			)
			continue
		}

		fmt.Println("Decision:")

		fmt.Printf(
			"Memory=%v, RAG=%v, Tool=%v\n",
			decision.UseMemory,
			decision.UseRAG,
			decision.UseTool,
		)
	}

	fmt.Println()
	fmt.Println("=========================================================")
}

// =========================================================
// RAG Test
// =========================================================

func runRAGTest(
	ctx context.Context,
	a *agent.Agent,
) {

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                    RAG Test")
	fmt.Println("=========================================================")

	ragPrompt := "普通商品购买后多少天可以申请退款？"

	fmt.Println()
	fmt.Println("User:")
	fmt.Println(ragPrompt)

	answer, err := a.Chat(
		ctx,
		ragPrompt,
	)

	if err != nil {
		log.Printf(
			"rag test failed: %v",
			err,
		)
		return
	}

	fmt.Println()
	fmt.Println("Agent:")
	fmt.Println(answer)

	fmt.Println()
	fmt.Println("=========================================================")
}

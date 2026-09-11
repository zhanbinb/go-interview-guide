package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"

	"go-agent-demo/agent"
	"go-agent-demo/config"
	"go-agent-demo/llm"
	"go-agent-demo/memory"
	"go-agent-demo/tools"
	"go-agent-demo/workflow"
)

func main() {
	// =========================================================
	// 1. 加载配置
	// =========================================================

	presets, err := config.Load()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
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

	ctx := context.Background()

	// =========================================================
	// 3. 创建 Tool Registry
	// =========================================================

	toolRegistry := tools.NewRegistry()

	// =========================================================
	// 4. 创建 ContextManager
	// =========================================================

	prompt := `
请帮我分析用户1001的订单10001。

需要查询：
1. 用户信息
2. 订单信息
3. 支付信息
4. 物流信息

最后请总结：
- 用户是谁
- 订单金额和状态
- 支付状态
- 物流状态
`

	promptMsg := openai.UserMessage(prompt)

	ctxManager := memory.NewContextManager(
		[]openai.ChatCompletionMessageParamUnion{
			promptMsg,
		},
	)

	// =========================================================
	// 5. Memory Demo
	// =========================================================

	testMemory()

	// =========================================================
	// 6. Workflow Demo
	// =========================================================

	// ---------------------------------------------------------
	// 固定顺序 Workflow
	// ---------------------------------------------------------

	// fmt.Println()
	// fmt.Println("===== Order Workflow =====")
	// fmt.Println(workflow.RunOrderWorkflow())

	// ---------------------------------------------------------
	// Conditional Workflow
	// ---------------------------------------------------------

	// fmt.Println()
	// fmt.Println("===== Conditional Workflow =====")
	// fmt.Println(workflow.RunOrderWorkflow2())

	// ---------------------------------------------------------
	// Agentic Workflow
	// ---------------------------------------------------------

	// fmt.Println()
	// fmt.Println("===== Agentic Workflow =====")
	//
	// result, err := workflow.RunAgenticWorkflow(
	// 	ctx,
	// 	&client,
	// 	model,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// fmt.Println(result)

	// =========================================================
	// 7. 创建 Agent
	// =========================================================

	agentRunner := agent.New(
		&client,
		model,
		toolRegistry,
		ctxManager,
	)

	// =========================================================
	// 8. 执行 Agent
	// =========================================================

	_, err = agentRunner.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// testMemory 演示简单的跨对话 Memory。
func testMemory() {
	fmt.Println()
	fmt.Println("===== Memory Demo =====")

	m := memory.NewMemory()

	// 第一轮对话：
	// 用户告诉 Agent 自己叫什么。
	m.Set(
		"user_name",
		"张三",
	)

	// 第二轮对话：
	// Agent 从 Memory 中获取之前保存的信息。
	name := m.Get("user_name")

	fmt.Println("Memory user_name:", name)

	fmt.Println()
}

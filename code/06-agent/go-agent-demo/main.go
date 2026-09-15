package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go-agent-demo/agent"
	"go-agent-demo/config"
	"go-agent-demo/embedding"
	"go-agent-demo/llm"
	"go-agent-demo/memory"
	"go-agent-demo/rag"
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

	queryRewriter := agent.NewQueryRewriter(
		&client,
		model,
	)

	// =========================================================
	// 3. 创建 Tool Registry
	// =========================================================

	toolRegistry := tools.NewRegistry()

	// =========================================================
	// 4. 创建 Long-term Memory
	// =========================================================
	//
	// 注意：
	//
	// Long-term Memory 的生命周期比 ContextManager 长。
	//
	// 所以这里创建一次，后面两个 Session 共用。
	//
	// Session 1：
	//
	//     写入 Memory
	//
	// Session 2：
	//
	//     读取 Memory
	//
	// 这正是 Long-term Memory 的核心意义。
	//

	embedder := embedding.NewFakeEmbedder(8)

	longTermMemory := memory.NewMemory(
		embedder,
	)

	knowledgeBase := rag.NewKnowledgeBase(embedder)

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
		log.Fatal(err)
	}

	retriever := rag.NewRetriever(
		knowledgeBase,
		embedder,
	)

	ragService := rag.NewRAG(
		retriever,
		&client,
		p.Model,
	)

	// =========================================================
	// 5. Session 1
	// =========================================================
	//
	// 第一轮 Session：
	//
	//     有自己的 ContextManager
	//     有自己的 Agent
	//
	// 用户告诉 Agent：
	//
	//     我叫张三
	//     我是一名 Go 后端开发工程师
	//     我正在找 Go 后端开发相关的工作
	//
	// Agent 会通过 Memory Extraction
	// 自动保存这些信息。
	//

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                 Session 1")
	fmt.Println("=========================================================")

	// Session 1 的短期记忆
	ctxManager1 := memory.NewContextManager(nil)

	// Session 1 的 Agent
	agent1 := agent.New(
		&client,
		model,
		toolRegistry,
		ctxManager1,
		longTermMemory,
		queryRewriter,
		ragService,
	)

	session1Prompts := []string{
		"我叫张三，我是一名 Go 后端开发工程师。",
		"我最近主要想找 Go 后端开发相关的工作。",
	}

	for i, prompt := range session1Prompts {

		fmt.Println()
		fmt.Println("-----------------------------------------")
		fmt.Printf("Session 1 - Turn %d\n", i+1)
		fmt.Println("User:")
		fmt.Println(prompt)
		fmt.Println("-----------------------------------------")

		answer, err := agent1.Chat(ctx, prompt)
		if err != nil {
			log.Printf(
				"session 1 agent chat failed: %v",
				err,
			)
			continue
		}

		fmt.Println()
		fmt.Println("Agent:")
		fmt.Println(answer)
	}

	// =========================================================
	// 6. 查看 Session 1 产生的 Long-term Memory
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("       Long-term Memory After Session 1")
	fmt.Println("=========================================================")

	for _, item := range longTermMemory.All() {
		fmt.Printf(
			"%s = %s\n",
			item.Key,
			item.Value,
		)
	}

	// =========================================================
	// 7. Session 1 结束
	// =========================================================
	//
	// 非常重要：
	//
	// 这里我们不再使用 agent1。
	//
	// 更准确地说：
	//
	//     ContextManager 1
	//     Agent 1
	//
	// 都代表 Session 1。
	//
	// Session 2 将创建全新的 ContextManager。
	//
	// 但是：
	//
	//     longTermMemory
	//
	// 仍然保留。
	//
	// 因此：
	//
	//     Short-term Memory → 清空 / 换新
	//
	//     Long-term Memory  → 保留
	//

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("             Session 1 Finished")
	fmt.Println("=========================================================")

	fmt.Println("\n========== Vector Search Test ==========")

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
	// 8. Session 2
	// =========================================================
	//
	// 现在创建一个全新的 ContextManager。
	//
	// 它里面没有 Session 1 的任何聊天记录。
	//
	// 这意味着：
	//
	//     ctxManager2 != ctxManager1
	//
	// Session 2 的 Agent 也重新创建。
	//
	// 但是它们共享：
	//
	//     longTermMemory
	//

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                 Session 2")
	fmt.Println("=========================================================")

	// 全新的短期记忆
	ctxManager2 := memory.NewContextManager(nil)

	// 全新的 Agent
	//
	// 注意：
	//
	// ContextManager 是新的
	//
	// Long-term Memory 是旧的
	//
	agent2 := agent.New(
		&client,
		model,
		toolRegistry,
		ctxManager2,
		longTermMemory,
		queryRewriter,
		ragService,
	)

	// =========================================================
	// 9. Session 2 测试 Long-term Memory
	// =========================================================
	//
	// 这里故意不再告诉 Agent：
	//
	//     我叫张三
	//
	//     我是一名 Go 后端开发工程师
	//
	// 而是直接询问 Memory 中的内容。
	//
	// 这样可以验证：
	//
	//     Session 2
	//         ↓
	//     没有 Session 1 的短期上下文
	//         ↓
	//     只能依赖 Long-term Memory
	//

	session2Prompts := []string{
		"你还记得我的 职业 吗？",
		"你还记得我的 我之前说过我想找什么工作吗？",
		"你还记得我的 名字 吗？",
		"你还记得我的 年龄 吗？",
	}

	for i, prompt := range session2Prompts {

		fmt.Println()
		fmt.Println("-----------------------------------------")
		fmt.Printf("Session 2 - Turn %d\n", i+1)
		fmt.Println("User:")
		fmt.Println(prompt)
		fmt.Println("-----------------------------------------")

		answer, err := agent2.Chat(ctx, prompt)
		if err != nil {
			log.Printf(
				"session 2 agent chat failed: %v",
				err,
			)
			continue
		}

		fmt.Println()
		fmt.Println("Agent:")
		fmt.Println(answer)
	}

	// =========================================================
	// 10. 最终 Long-term Memory
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("             Final Long-term Memory")
	fmt.Println("=========================================================")

	for _, item := range longTermMemory.All() {
		fmt.Printf(
			"%s = %s\n",
			item.Key,
			item.Value,
		)
	}

	// =========================================================
	// 11. RAG Test
	// =========================================================

	fmt.Println()
	fmt.Println("=========================================================")
	fmt.Println("                    RAG Test")
	fmt.Println("=========================================================")

	ragPrompt := "普通商品购买后多少天可以申请退款？"

	fmt.Println()
	fmt.Println("User:")
	fmt.Println(ragPrompt)

	answer, err := agent2.Chat(
		ctx,
		ragPrompt,
	)
	if err != nil {
		log.Printf("rag test failed: %v", err)
	} else {
		fmt.Println()
		fmt.Println("Agent:")
		fmt.Println(answer)
	}

	fmt.Println()
	fmt.Println("=========================================================")
}

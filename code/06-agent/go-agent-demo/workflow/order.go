package workflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"

	"go-agent-demo/config"
	"go-agent-demo/tools"
)

// RunOrderWorkflow
//
// 固定顺序 Workflow：
//
// 1. 查询用户
// 2. 查询订单
// 3. 查询支付
// 4. 查询物流
//
// 这是最典型的 Sequential Workflow。
func RunOrderWorkflow() string {
	user := tools.GetUser("1001")

	order := tools.QueryOrder("10001")

	payment := tools.QueryPayment("10001")

	logistics := tools.QueryLogistics("10001")

	return fmt.Sprintf(
		"用户信息：%s\n订单信息：%s\n支付信息：%s\n物流信息：%s",
		user,
		order,
		payment,
		logistics,
	)
}

// RunOrderWorkflow2
//
// Conditional Workflow：
//
// 先查询订单状态，
// 再根据订单状态决定下一步调用哪个工具。
func RunOrderWorkflow2() string {
	user := tools.GetUser("1001")

	order := tools.QueryOrder("10001")

	var nextResult string

	if strings.Contains(order, "已发货") {
		nextResult = tools.QueryLogistics("10001")
	} else {
		nextResult = tools.QueryPayment("10001")
	}

	return fmt.Sprintf(
		"用户信息：%s\n订单信息：%s\n后续查询结果：%s",
		user,
		order,
		nextResult,
	)
}

// RunOrderWorkflow3
//
// 一个更接近 Agentic Workflow 的示例。
//
// Workflow 本身仍然负责业务流程，
// 但是根据前一步结果进行动态决策。
func RunOrderWorkflow3() string {
	user := tools.GetUser("1001")

	order := tools.QueryOrder("10001")

	var result string

	switch {
	case strings.Contains(order, "已发货"):
		logistics := tools.QueryLogistics("10001")

		result = fmt.Sprintf(
			"订单已经发货。\n%s",
			logistics,
		)

	case strings.Contains(order, "待支付"):
		payment := tools.QueryPayment("10001")

		result = fmt.Sprintf(
			"订单尚未完成支付。\n%s",
			payment,
		)

	default:
		result = "订单状态暂时无法判断。"
	}

	return fmt.Sprintf(
		"用户：%s\n订单：%s\n结果：%s",
		user,
		order,
		result,
	)
}

// RunAgenticWorkflow
//
// 这里展示：
//
//	用户请求
//	    ↓
//	LLM 判断任务类型
//	    ↓
//	Workflow
//	    ↓
//	LLM 总结最终结果
//
// 也就是说：
//
// LLM 负责决策 / 理解
// Workflow 负责执行
// Tool 负责具体动作
func RunAgenticWorkflow(
	ctx context.Context,
	client *openai.Client,
	model string,
) (string, error) {

	classifyPrompt := `
请判断下面的用户请求属于哪一种类型。

只允许返回以下两个值之一：

order_analysis
unknown

用户请求：
请帮我分析用户1001的订单10001，包括订单、支付和物流情况。
`

	resp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(classifyPrompt),
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("classify request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("classify request: no choices returned")
	}

	classification := resp.Choices[0].Message.Content

	// 某些模型可能会输出：
	//
	// <think>
	// ...
	// </think>
	// order_analysis
	//
	// Demo 中简单去掉 think 内容。
	if start := strings.Index(classification, "<think>"); start >= 0 {
		if end := strings.Index(classification, "</think>"); end > start {
			classification =
				classification[:start] +
					classification[end+len("</think>"):]
		}
	}

	classification = strings.TrimSpace(classification)

	if classification != "order_analysis" {
		return "暂时无法识别该请求类型。", nil
	}

	// LLM 判断出来是订单分析任务，
	// 接下来交给确定性的 Workflow。
	workflowResult := RunOrderWorkflow3()

	summaryPrompt := fmt.Sprintf(
		`请根据下面的业务查询结果，为用户生成最终回答。

要求：
1. 使用中文
2. 简洁清晰
3. 总结用户、订单、支付、物流状态
4. 不要提及 Workflow、Tool、LLM 等内部实现

业务查询结果：

%s`,
		workflowResult,
	)

	finalResp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(summaryPrompt),
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("summarize workflow result: %w", err)
	}

	if len(finalResp.Choices) == 0 {
		return "", fmt.Errorf("summarize workflow result: no choices returned")
	}

	return finalResp.Choices[0].Message.Content, nil
}

// KeysOf 返回配置 Map 中所有 Key。
//
// 这个函数原来位于 main.go，
// 现在放到 workflow 中并不是必须的。
// 如果后续没有使用，可以直接删除。
func KeysOf(m map[string]config.Preset) []string {
	keys := make([]string, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

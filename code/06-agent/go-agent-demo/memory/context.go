package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

// ContextManager 管理 Agent 当前对话上下文。
//
// 主要负责：
// 1. 保存历史消息
// 2. 保存摘要
// 3. 保存当前用户 Query
// 4. 控制上下文长度
// 5. 在上下文过长时触发摘要
// 6. 构造最终发送给 LLM 的 Messages
type ContextManager struct {
	messages      []openai.ChatCompletionMessageParamUnion
	summary       string
	lastUserQuery string
}

// NewContextManager 创建 ContextManager。
func NewContextManager(
	initialMessages []openai.ChatCompletionMessageParamUnion,
) *ContextManager {
	cm := &ContextManager{
		messages: make(
			[]openai.ChatCompletionMessageParamUnion,
			0,
		),
	}

	cm.messages = append(cm.messages, initialMessages...)

	return cm
}

// SetSummary 设置摘要。
func (cm *ContextManager) SetSummary(summary string) {
	cm.summary = summary
}

// GetSummary 获取摘要。
func (cm *ContextManager) GetSummary() string {
	return cm.summary
}

// Add 添加一条普通消息。
func (cm *ContextManager) Add(
	msg openai.ChatCompletionMessageParamUnion,
) {
	cm.messages = append(cm.messages, msg)
}

// AddUserMessage 添加一条用户消息。
//
// 除了保存到 messages，
// 还会记录当前用户 Query，
// 供 Query Rewrite / Memory Retrieval / Memory Extraction 使用。
func (cm *ContextManager) AddUserMessage(
	content string,
) {
	cm.messages = append(
		cm.messages,
		openai.UserMessage(content),
	)

	cm.lastUserQuery = content
}

// Messages 获取当前历史消息。
func (cm *ContextManager) Messages() []openai.ChatCompletionMessageParamUnion {
	return cm.messages
}

// Trim 保留最近的 maxMessages 条消息。
func (cm *ContextManager) Trim(maxMessages int) {
	if maxMessages <= 0 {
		cm.messages = nil
		return
	}

	if len(cm.messages) <= maxMessages {
		return
	}

	cm.messages = cm.messages[len(cm.messages)-maxMessages:]
}

// Summarize 对当前历史消息进行摘要。
func (cm *ContextManager) Summarize(
	ctx context.Context,
	client *openai.Client,
	model string,
) error {
	if len(cm.messages) == 0 {
		return nil
	}

	// 将当前对话消息转换成 JSON，
	// 方便 LLM 理解每条消息的 role / content / tool call 等信息。
	data, err := json.Marshal(cm.messages)
	if err != nil {
		return fmt.Errorf("marshal messages: %w", err)
	}

	prompt := fmt.Sprintf(
		`请总结下面这段 Agent 对话历史。

要求：
1. 保留用户的重要需求
2. 保留已经查询到的重要业务数据
3. 保留订单、用户、支付、物流等关键状态
4. 忽略无关的中间过程
5. 输出简洁的中文摘要

对话历史：
%s`,
		string(data),
	)

	resp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("summarize messages: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("summarize: no choices returned")
	}

	cm.summary = resp.Choices[0].Message.Content

	// 摘要完成后清空短期历史消息。
	//
	// 注意：
	// lastUserQuery 不清空。
	// 因为它代表当前正在处理的用户 Query，
	// Query Rewrite / Memory Retrieval / Memory Extraction
	// 仍然需要使用它。
	cm.messages = nil

	return nil
}

// MaybeSummarize 当消息数量达到阈值时自动摘要。
func (cm *ContextManager) MaybeSummarize(
	ctx context.Context,
	client *openai.Client,
	model string,
) error {
	// 当前 Demo 设置为 6 条。
	//
	// 实际项目中通常不会简单按照消息数量，
	// 而会按照 token 数量判断。
	if len(cm.messages) < 6 {
		return nil
	}

	return cm.Summarize(ctx, client, model)
}

// BuildMessages 构造最终发送给 LLM 的上下文。
//
// 如果存在 summary，则：
//
//	历史摘要
//	+ 最近消息
//
// 一起发送给 LLM。
func (cm *ContextManager) BuildMessages() []openai.ChatCompletionMessageParamUnion {
	messages := make(
		[]openai.ChatCompletionMessageParamUnion,
		0,
		len(cm.messages)+1,
	)

	if cm.summary != "" {
		summaryMessage := fmt.Sprintf(
			"以下是之前对话的摘要，请结合摘要理解当前对话：\n%s",
			cm.summary,
		)

		messages = append(
			messages,
			openai.UserMessage(summaryMessage),
		)
	}

	messages = append(messages, cm.messages...)

	return messages
}

// LastUserMessage 返回最近一次用户输入。
//
// 不再从 openai.ChatCompletionMessageParamUnion 中解析，
// 而是直接返回 AddUserMessage() 保存的当前 Query。
func (cm *ContextManager) LastUserMessage() string {
	return cm.lastUserQuery
}

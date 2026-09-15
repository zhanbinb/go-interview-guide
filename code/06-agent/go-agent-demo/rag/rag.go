package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

// RAG 负责知识检索和生成。
type RAG struct {
	retriever *Retriever
	client    *openai.Client
	model     string
}

// NewRAG 创建 RAG。
func NewRAG(
	retriever *Retriever,
	client *openai.Client,
	model string,
) *RAG {
	return &RAG{
		retriever: retriever,
		client:    client,
		model:     model,
	}
}

// Retrieve 检索与 Query 相关的知识。
func (r *RAG) Retrieve(
	ctx context.Context,
	query string,
) ([]RetrievalResult, error) {
	return r.retriever.Retrieve(query, 3)
}

// Ask 执行完整的 RAG：
// Query → Retrieval → Context → LLM。
func (r *RAG) Ask(
	ctx context.Context,
	query string,
) (string, error) {
	results, err := r.Retrieve(ctx, query)
	if err != nil {
		return "", fmt.Errorf("retrieve knowledge: %w", err)
	}

	var contextBuilder strings.Builder

	for _, result := range results {
		contextBuilder.WriteString(
			fmt.Sprintf(
				"- %s\n",
				result.Chunk.Content,
			),
		)
	}

	prompt := fmt.Sprintf(
		`请根据下面提供的知识回答用户问题。

要求：
1. 优先使用提供的知识回答
2. 不要编造知识中不存在的信息
3. 如果知识中没有答案，请明确说明无法从知识库中找到答案

知识：
%s

用户问题：
%s`,
		contextBuilder.String(),
		query,
	)

	resp, err := r.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: r.model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("generate answer: no choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

package rag

import (
	"fmt"
	"strings"

	"go-agent-demo/embedding"
)

// Document 表示一份知识文档。
type Document struct {
	ID      string
	Content string
}

// Chunk 表示文档切分后的知识片段。
type Chunk struct {
	ID         string
	DocumentID string
	Content    string
	Vector     []float64
}

// KnowledgeBase 表示一个简单的内存知识库。
type KnowledgeBase struct {
	chunks   []Chunk
	embedder embedding.Embedder
}

// NewKnowledgeBase 创建知识库。
func NewKnowledgeBase(
	embedder embedding.Embedder,
) *KnowledgeBase {
	return &KnowledgeBase{
		chunks:   make([]Chunk, 0),
		embedder: embedder,
	}
}

// AddDocument 添加文档，并简单切分成 Chunk。
func (kb *KnowledgeBase) AddDocument(
	doc Document,
) error {
	parts := splitDocument(doc.Content)

	for index, content := range parts {
		vector, err := kb.embedder.Embed(content)
		if err != nil {
			return err
		}

		kb.chunks = append(
			kb.chunks,
			Chunk{
				ID: fmt.Sprintf(
					"%s-chunk-%d",
					doc.ID,
					index+1,
				),
				DocumentID: doc.ID,
				Content:    content,
				Vector:     vector,
			},
		)
	}

	return nil
}

// Chunks 返回知识库中的所有 Chunk。
func (kb *KnowledgeBase) Chunks() []Chunk {
	return kb.chunks
}

// splitDocument 是教学版本的简单切分逻辑。
func splitDocument(content string) []string {
	parts := strings.Split(content, "\n\n")

	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "" {
			continue
		}

		result = append(result, part)
	}

	return result
}

package rag

import (
	"sort"

	"go-agent-demo/embedding"
)

// RetrievalResult 表示一次检索结果。
type RetrievalResult struct {
	Chunk      Chunk
	Similarity float64
}

// Retriever 负责从知识库中检索相关知识。
type Retriever struct {
	knowledgeBase *KnowledgeBase
	embedder      embedding.Embedder
}

// NewRetriever 创建 Retriever。
func NewRetriever(
	knowledgeBase *KnowledgeBase,
	embedder embedding.Embedder,
) *Retriever {
	return &Retriever{
		knowledgeBase: knowledgeBase,
		embedder:      embedder,
	}
}

// Retrieve 根据 Query 检索 Top K 知识片段。
func (r *Retriever) Retrieve(
	query string,
	topK int,
) ([]RetrievalResult, error) {

	queryVector, err := r.embedder.Embed(query)
	if err != nil {
		return nil, err
	}

	results := make(
		[]RetrievalResult,
		0,
		len(r.knowledgeBase.Chunks()),
	)

	for _, chunk := range r.knowledgeBase.Chunks() {
		similarity := embedding.CosineSimilarity(
			queryVector,
			chunk.Vector,
		)

		results = append(
			results,
			RetrievalResult{
				Chunk:      chunk,
				Similarity: similarity,
			},
		)
	}

	sort.Slice(
		results,
		func(i, j int) bool {
			return results[i].Similarity > results[j].Similarity
		},
	)

	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

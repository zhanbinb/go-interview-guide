package embedding

import (
	"math"
	"strings"
)

// Embedder 是 Embedding 模型的抽象接口。
//
// Agent / Memory 不需要知道具体使用哪个模型。
// 它只关心：
//
//	text -> vector
type Embedder interface {
	Embed(text string) ([]float64, error)
}

// FakeEmbedder 是学习阶段使用的简化 Embedding。
//
// 注意：
// 它不是真正的语义 Embedding 模型。
// 这里只是为了模拟：
//
//	文本 -> 向量
//
// 后面接真实 Embedding API 时，
// Memory 层不需要改变接口。
type FakeEmbedder struct {
	Dimension int
}

// NewFakeEmbedder 创建 Fake Embedding。
func NewFakeEmbedder(dimension int) *FakeEmbedder {
	if dimension <= 0 {
		dimension = 8
	}

	return &FakeEmbedder{
		Dimension: dimension,
	}
}

// Embed 将文本转换成向量。
//
// 当前 Demo 使用简单的字符 Hash 方式生成向量。
// 它只是为了让我们能够演示 Vector Search，
// 不代表真正的语义 Embedding。
func (e *FakeEmbedder) Embed(
	text string,
) ([]float64, error) {

	vector := make(
		[]float64,
		e.Dimension,
	)

	text = strings.ToLower(
		strings.TrimSpace(text),
	)

	if text == "" {
		return vector, nil
	}

	for _, r := range text {

		index := int(r) % e.Dimension

		vector[index] += 1
	}

	// 归一化。
	//
	// 这样不同长度的文本，
	// 不会仅仅因为字符数量更多，
	// 就天然拥有更大的向量。
	normalize(vector)

	return vector, nil
}

// CosineSimilarity 计算两个向量的余弦相似度。
//
//	                     A · B
//	cos(A,B) = --------------------------
//	            ||A|| × ||B||
//
// 返回值通常位于 [-1, 1]。
// 越接近 1，表示越相似。
func CosineSimilarity(
	a []float64,
	b []float64,
) float64 {

	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	if len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := range a {

		dotProduct += a[i] * b[i]

		normA += a[i] * a[i]

		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct /
		(math.Sqrt(normA) * math.Sqrt(normB))
}

// normalize 对向量进行 L2 归一化。
func normalize(vector []float64) {

	var sum float64

	for _, value := range vector {
		sum += value * value
	}

	if sum == 0 {
		return
	}

	norm := math.Sqrt(sum)

	for i := range vector {
		vector[i] /= norm
	}
}

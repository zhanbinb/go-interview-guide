package memory

import (
	"go-agent-demo/embedding"
	"regexp"
	"sort"
	"strings"
)

// MemoryItem 表示一条长期记忆。
type MemoryItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ScoredMemoryItem 表示一条带检索分数的 Memory。
//
// Retrieval 系统通常不会只返回数据，
// 还会给候选结果计算一个 Score，
// 然后根据 Score 排序。
type ScoredMemoryItem struct {
	Item  MemoryItem
	Score int
}

type Memory struct {
	data     map[string]string
	vectors  map[string][]float64
	embedder embedding.Embedder
}

type VectorSearchResult struct {
	Item       MemoryItem
	Similarity float64
}

// HybridSearchResult 表示 Hybrid Search 的结果。
//
// 同时保存：
// 1. Keyword Score
// 2. Vector Similarity
// 3. 最终 Hybrid Score
type HybridSearchResult struct {
	Item         MemoryItem
	KeywordScore float64
	VectorScore  float64
	HybridScore  float64
}

// SearchHybridTopK 同时使用 Keyword Search 和 Vector Search。
//
// 当前策略：
//
//	Hybrid Score
//	= 0.5 * Keyword Score
//	+ 0.5 * Vector Similarity
//
// 注意：
// Keyword Score 在融合之前会进行归一化。
func (m *Memory) SearchHybridTopK(
	query string,
	topK int,
) []HybridSearchResult {

	if strings.TrimSpace(query) == "" {
		return nil
	}

	if topK <= 0 {
		return nil
	}

	// -----------------------------------------
	// 1. Keyword Search
	// -----------------------------------------

	keywords := extractKeywords(query)

	keywordScores := make(
		map[string]int,
	)

	var maxKeywordScore int

	if len(keywords) > 0 {

		for key, value := range m.data {

			lowerKey := strings.ToLower(key)
			lowerValue := strings.ToLower(value)

			score := 0

			for _, keyword := range keywords {

				if lowerKey == keyword {
					score += 10
					continue
				}

				if strings.Contains(
					lowerKey,
					keyword,
				) {
					score += 5
				}

				if strings.Contains(
					lowerValue,
					keyword,
				) {
					score += 2
				}
			}

			if score > 0 {

				keywordScores[key] = score

				if score > maxKeywordScore {
					maxKeywordScore = score
				}
			}
		}
	}

	// -----------------------------------------
	// 2. Vector Search
	// -----------------------------------------

	vectorScores := make(
		map[string]float64,
	)

	if m.embedder != nil {

		queryVector, err :=
			m.embedder.Embed(query)

		if err == nil {

			for key, vector := range m.vectors {

				similarity :=
					embedding.CosineSimilarity(
						queryVector,
						vector,
					)

				vectorScores[key] = similarity
			}
		}
	}

	// -----------------------------------------
	// 3. Merge Candidates
	// -----------------------------------------

	candidateKeys := make(
		map[string]bool,
	)

	for key := range keywordScores {
		candidateKeys[key] = true
	}

	for key := range vectorScores {
		candidateKeys[key] = true
	}

	results := make(
		[]HybridSearchResult,
		0,
		len(candidateKeys),
	)

	// -----------------------------------------
	// 4. Score Fusion
	// -----------------------------------------

	for key := range candidateKeys {

		value := m.data[key]

		var keywordScore float64

		if maxKeywordScore > 0 {

			keywordScore =
				float64(
					keywordScores[key],
				) /
					float64(maxKeywordScore)
		}

		vectorScore :=
			vectorScores[key]

		hybridScore :=
			0.5*keywordScore +
				0.5*vectorScore

		results = append(
			results,
			HybridSearchResult{
				Item: MemoryItem{
					Key:   key,
					Value: value,
				},
				KeywordScore: keywordScore,
				VectorScore:  vectorScore,
				HybridScore:  hybridScore,
			},
		)
	}

	// -----------------------------------------
	// 5. Ranking
	// -----------------------------------------

	sort.Slice(
		results,
		func(i, j int) bool {

			if results[i].HybridScore !=
				results[j].HybridScore {

				return results[i].HybridScore >
					results[j].HybridScore
			}

			return results[i].Item.Key <
				results[j].Item.Key
		},
	)

	// -----------------------------------------
	// 6. Top K
	// -----------------------------------------

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

func (m *Memory) SearchVectorTopK(
	query string,
	topK int,
) []VectorSearchResult {

	if m.embedder == nil {
		return nil
	}

	if strings.TrimSpace(query) == "" {
		return nil
	}

	if topK <= 0 {
		return nil
	}

	queryVector, err := m.embedder.Embed(query)
	if err != nil {
		return nil
	}

	results := make(
		[]VectorSearchResult,
		0,
		len(m.data),
	)

	for key, value := range m.data {

		vector, exists := m.vectors[key]

		if !exists {
			continue
		}

		similarity :=
			embedding.CosineSimilarity(
				queryVector,
				vector,
			)

		results = append(
			results,
			VectorSearchResult{
				Item: MemoryItem{
					Key:   key,
					Value: value,
				},
				Similarity: similarity,
			},
		)
	}

	sort.Slice(
		results,
		func(i, j int) bool {

			if results[i].Similarity !=
				results[j].Similarity {

				return results[i].Similarity >
					results[j].Similarity
			}

			return results[i].Item.Key <
				results[j].Item.Key
		},
	)

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

func NewMemory(
	embedder embedding.Embedder,
) *Memory {

	return &Memory{
		data: make(
			map[string]string,
		),
		vectors: make(
			map[string][]float64,
		),
		embedder: embedder,
	}
}

// Save 保存一条 Memory。
func (m *Memory) Save(
	key string,
	value string,
) {

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" || value == "" {
		return
	}

	m.data[key] = value

	// 如果配置了 Embedder，
	// 保存 Memory 的同时生成 Vector。
	if m.embedder != nil {

		vector, err := m.embedder.Embed(
			key + " " + value,
		)

		if err == nil {
			m.vectors[key] = vector
		}
	}
}

// Get 获取指定 Memory。
func (m *Memory) Get(key string) string {
	return m.data[key]
}

// Search 保留原来的简单搜索能力。
func (m *Memory) Search(query string) []MemoryItem {
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return nil
	}

	var results []MemoryItem

	for key, value := range m.data {
		keyMatch := strings.Contains(
			strings.ToLower(key),
			query,
		)

		valueMatch := strings.Contains(
			strings.ToLower(value),
			query,
		)

		if keyMatch || valueMatch {
			results = append(
				results,
				MemoryItem{
					Key:   key,
					Value: value,
				},
			)
		}
	}

	return results
}

// SearchRelevant 保留原来的接口。
//
// 现在内部统一使用 SearchRelevantTopK。
func (m *Memory) SearchRelevant(
	query string,
) []MemoryItem {

	results := m.SearchRelevantTopK(
		query,
		len(m.data),
	)

	items := make(
		[]MemoryItem,
		0,
		len(results),
	)

	for _, result := range results {
		items = append(
			items,
			result.Item,
		)
	}

	return items
}

// SearchRelevantTopK 根据 Query 检索 Memory，
// 计算 Score，并返回 Top K。
//
// 当前 Score 仍然是简单的关键词匹配。
// 后面学习 Embedding 时，
// 可以把这里替换成向量相似度。
func (m *Memory) SearchRelevantTopK(
	query string,
	topK int,
) []ScoredMemoryItem {

	query = strings.TrimSpace(query)

	if query == "" {
		return nil
	}

	if topK <= 0 {
		return nil
	}

	keywords := extractKeywords(query)

	if len(keywords) == 0 {
		return nil
	}

	var scored []ScoredMemoryItem

	for key, value := range m.data {

		lowerKey := strings.ToLower(key)
		lowerValue := strings.ToLower(value)

		score := 0

		for _, keyword := range keywords {

			// Key 完全匹配。
			//
			// 例如：
			//
			// query = user_profession
			// key   = user_profession
			//
			// 这是非常强的匹配。
			if lowerKey == keyword {
				score += 10
				continue
			}

			// Key 部分匹配。
			if strings.Contains(
				lowerKey,
				keyword,
			) {
				score += 5
			}

			// Value 匹配。
			if strings.Contains(
				lowerValue,
				keyword,
			) {
				score += 2
			}
		}

		// 没有命中任何关键词，
		// 不进入 Candidate。
		if score == 0 {
			continue
		}

		scored = append(
			scored,
			ScoredMemoryItem{
				Item: MemoryItem{
					Key:   key,
					Value: value,
				},
				Score: score,
			},
		)
	}

	// Score 从高到低排序。
	//
	// 如果 Score 相同，
	// 再按照 Key 排序，
	// 保证结果稳定。
	sort.Slice(
		scored,
		func(i, j int) bool {

			if scored[i].Score != scored[j].Score {
				return scored[i].Score > scored[j].Score
			}

			return scored[i].Item.Key <
				scored[j].Item.Key
		},
	)

	// Top K。
	if len(scored) > topK {
		scored = scored[:topK]
	}

	return scored
}

// extractKeywords 从 Query 中提取关键词。
//
// 当前 Demo 主要处理英文、数字和下划线。
// 中文语义搜索后面通过 Embedding 解决。
func extractKeywords(query string) []string {

	re := regexp.MustCompile(
		`[a-zA-Z][a-zA-Z0-9_]*`,
	)

	matches := re.FindAllString(
		strings.ToLower(query),
		-1,
	)

	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)

	var keywords []string

	for _, match := range matches {

		if seen[match] {
			continue
		}

		seen[match] = true

		keywords = append(
			keywords,
			match,
		)
	}

	return keywords
}

// All 返回所有 Memory。
func (m *Memory) All() []MemoryItem {

	results := make(
		[]MemoryItem,
		0,
		len(m.data),
	)

	for key, value := range m.data {

		results = append(
			results,
			MemoryItem{
				Key:   key,
				Value: value,
			},
		)
	}

	return results
}

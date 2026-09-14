package memory

import (
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
	data map[string]string
}

// NewMemory 创建 Memory。
func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]string),
	}
}

// Save 保存一条 Memory。
func (m *Memory) Save(key, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" || value == "" {
		return
	}

	m.data[key] = value
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

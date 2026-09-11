package memory

import "strings"

// MemoryItem 是一个记忆项。
type MemoryItem struct {
	Key   string
	Value string
}

// Memory 是一个最简单的长期记忆示例。
//
// 和 ContextManager 不同：
//
// ContextManager
//
//	-> 管理当前对话上下文
//
// Memory
//
//	-> 保存跨对话的结构化信息
//
// 例如：
// user_name -> 张三
// language  -> 中文
// preference -> 简洁回答
type Memory struct {
	data map[string]string
}

// NewMemory 创建 Memory。
func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]string),
	}
}

// Save 保存一条长期记忆
func (m *Memory) Save(key, value string) {
	m.data[key] = value
}

// Get 获取记忆。
func (m *Memory) Get(key string) string {
	return m.data[key]
}

// Search 根据关键词搜索相关记忆。
//
// 当前只是简单的字符串匹配。
// 后面我们会把这里升级成：
//
// Keyword Search
//
//	↓
//
// Semantic Search
//
//	↓
//
// Embedding + Vector DB
func (m *Memory) Search(query string) []MemoryItem {
	query = strings.ToLower(query)
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
			results = append(results, MemoryItem{
				Key:   key,
				Value: value,
			})
		}
	}
	return results
}

func (m *Memory) All() []MemoryItem {
	results := make([]MemoryItem, 0, len(m.data))
	for key, value := range m.data {
		results = append(results, MemoryItem{
			Key:   key,
			Value: value,
		})
	}
	return results
}

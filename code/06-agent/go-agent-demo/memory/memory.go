package memory

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

// Set 保存记忆。
func (m *Memory) Set(key, value string) {
	m.data[key] = value
}

// Get 获取记忆。
func (m *Memory) Get(key string) string {
	return m.data[key]
}

package tools

// Registry 管理 Agent 可以使用的所有 Tool。
type Registry struct {
	tools map[string]Tool
}

// NewRegistry 创建 Tool Registry。
func NewRegistry() *Registry {
	return &Registry{
		tools: map[string]Tool{
			"query_order":     QueryOrderTool(),
			"get_user":        GetUserTool(),
			"query_payment":   QueryPaymentTool(),
			"query_logistics": QueryLogisticsTool(),
		},
	}
}

// Get 根据 Tool Name 获取 Tool。
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// All 返回所有注册的 Tool。
func (r *Registry) All() map[string]Tool {
	return r.tools
}

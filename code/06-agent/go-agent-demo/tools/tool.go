package tools

// Tool 表示一个可以被 Agent 调用的工具。
//
// Agent 不需要知道 Tool 内部具体怎么实现，
// 只需要知道：
// 1. Tool 的名称
// 2. Tool 的描述
// 3. Tool 接收到参数后如何执行
type Tool struct {
	Name        string
	Description string
	Handler     func(args string) (string, error)
}

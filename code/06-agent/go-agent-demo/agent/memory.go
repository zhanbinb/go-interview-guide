package agent

import (
	"strings"
)

// currentUserQuery 获取当前对话中最后一条用户消息。
//
// Memory Retrieval 需要知道：
// “用户现在正在问什么？”
//
// 后面我们会用这个 Query 去搜索长期记忆。
// func (a *Agent) currentUserQuery() string {
// 	messages := a.contextManager.Messages()

// 	for i := len(messages) - 1; i >= 0; i-- {
// 		message := messages[i]

// 		// 当前 Demo 中主要处理 User Message。
// 		// openai 的消息类型是 Union，因此这里使用 JSON
// 		// 方式不太优雅；下一步可以进一步封装 ContextManager。
// 		data := message.GetUserMessage()
// 		if data.Content != "" {
// 			return data.Content
// 		}
// 	}

// 	return ""
// }

// buildMemoryContext 根据当前用户问题检索长期记忆。
func (a *Agent) buildMemoryContext(query string) string {
	if query == "" {
		return ""
	}

	results := a.longTermMemory.Search(query)

	if len(results) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("以下是与当前请求相关的长期记忆：\n")

	for _, item := range results {
		builder.WriteString("- ")
		builder.WriteString(item.Key)
		builder.WriteString(": ")
		builder.WriteString(item.Value)
		builder.WriteString("\n")
	}

	return builder.String()
}

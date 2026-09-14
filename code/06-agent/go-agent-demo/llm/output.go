package llm

import "strings"

// CleanThinking 从模型输出中移除 <think>...</think> 内容。
//
// 一些模型可能会把内部分析过程放在 <think> 标签中。
// Agent 通常只应该把最终回答交给用户。
func CleanThinking(content string) string {
	content = strings.TrimSpace(content)

	for {
		start := strings.Index(content, "<think>")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "</think>")

		if end == -1 {
			// 没有找到结束标签。
			// 为了避免把后面的正常内容误删，这里直接停止。
			break
		}

		end = start + end + len("</think>")

		content = content[:start] + content[end:]
	}

	return strings.TrimSpace(content)
}

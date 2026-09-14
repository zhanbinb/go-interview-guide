package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

// ExtractedMemory 表示一条从对话中提取出来的长期记忆。
type ExtractedMemory struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExtractMemoryResult 表示 Memory Extraction 的结果。
type ExtractMemoryResult struct {
	ShouldSave bool              `json:"should_save"`
	Memories   []ExtractedMemory `json:"memories"`
}

// ExtractMemory 使用 LLM 从用户消息中提取值得长期保存的信息。
func ExtractMemory(
	ctx context.Context,
	client *openai.Client,
	model string,
	userMessage string,
) ([]ExtractedMemory, error) {

	if userMessage == "" {
		return nil, nil
	}

	prompt := fmt.Sprintf(`
你是一个 Agent Memory 提取器。

请判断下面这条用户消息中，是否包含值得长期保存的信息。

适合保存的信息包括：
- 用户姓名
- 职业
- 技术方向
- 技能
- 长期目标
- 用户偏好
- 长期项目
- 稳定的个人背景信息

不要保存：
- 普通闲聊
- 一次性的临时问题
- 当前订单查询结果
- 当前时间
- 无长期价值的信息

请严格返回 JSON，不要输出其他内容。

JSON 格式：

{
  "should_save": true,
  "memories": [
    {
      "key": "user_name",
      "value": "张三"
    }
  ]
}

如果没有值得保存的信息：

{
  "should_save": false,
  "memories": []
}

用户消息：

%s
`, userMessage)

	resp, err := client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("extract memory: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("extract memory: no choices returned")
	}

	content := resp.Choices[0].Message.Content

	var result ExtractMemoryResult

	if err := json.Unmarshal(
		[]byte(content),
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"parse extracted memory: %w, content=%s",
			err,
			content,
		)
	}

	if !result.ShouldSave {
		return nil, nil
	}

	return result.Memories, nil
}

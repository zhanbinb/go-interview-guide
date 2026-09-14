package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-agent-demo/llm"

	"github.com/openai/openai-go"
)

type QueryRewriter struct {
	client *openai.Client
	model  string
}

func NewQueryRewriter(
	client *openai.Client,
	model string,
) *QueryRewriter {
	return &QueryRewriter{
		client: client,
		model:  model,
	}
}

type QueryRewriteResult struct {
	Queries []string `json:"queries"`
}

func (q *QueryRewriter) Rewrite(
	ctx context.Context,
	userQuery string,
) ([]string, error) {

	systemPrompt := `
你是一个 Memory Query Rewriter。

你的任务是：
把用户的自然语言问题转换成适合长期记忆检索的关键词。

目标：
帮助 Memory 系统找到用户之前保存的信息。

例如：

用户：
“你还记得我的职业吗？”

应该返回：

{
  "queries": [
    "user_profession",
    "职业",
    "工作"
  ]
}

用户：
“我之前说过我想找什么工作？”

应该返回：

{
  "queries": [
    "career_goal",
    "工作",
    "求职"
  ]
}

用户：
“你还记得我的名字吗？”

应该返回：

{
  "queries": [
    "user_name",
    "名字",
    "姓名"
  ]
}

用户：
“我主要使用什么技术？”

应该返回：

{
  "queries": [
    "tech_direction",
    "技术",
    "开发"
  ]
}

重要规则：

1. 不要回答用户的问题。
2. 只负责生成检索 Query。
3. 不要创建 Memory。
4. 不要猜测 Memory 中不存在的信息。
5. 优先生成可能对应 Memory key 的英文关键词。
6. 可以同时生成中文语义关键词。
7. 最多生成 5 个 Query。
8. 只输出 JSON。
9. 不要输出 Markdown。
10. 不要输出解释。

输出格式：

{
  "queries": [
    "query1",
    "query2"
  ]
}
`

	userPrompt := fmt.Sprintf(
		"用户问题：\n%s",
		userQuery,
	)

	response, err := q.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: q.model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
		},
	)

	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf(
			"query rewrite returned no choices",
		)
	}

	content := response.Choices[0].Message.Content

	content = llm.CleanThinking(content)
	content = cleanJSONContent(content)

	var result QueryRewriteResult

	if err := json.Unmarshal(
		[]byte(content),
		&result,
	); err != nil {
		return nil, fmt.Errorf(
			"parse query rewrite result failed: %w; content=%s",
			err,
			content,
		)
	}

	var queries []string

	seen := make(map[string]bool)

	for _, query := range result.Queries {

		query = strings.TrimSpace(query)

		if query == "" {
			continue
		}

		query = strings.ToLower(query)

		if seen[query] {
			continue
		}

		seen[query] = true

		queries = append(
			queries,
			query,
		)

		if len(queries) >= 5 {
			break
		}
	}

	return queries, nil
}

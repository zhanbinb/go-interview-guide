package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
)

// Capability 表示 Agent 的一种能力。
//
// Router 会把这些能力描述告诉 LLM，
// 让 LLM 根据用户问题判断应该使用哪些能力。
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RouteDecision 表示 Router 的决策结果。
type RouteDecision struct {
	UseMemory bool `json:"use_memory"`
	UseRAG    bool `json:"use_rag"`
	UseTool   bool `json:"use_tool"`
}

// Router 负责根据用户问题选择 Agent 能力。
type Router struct {
	client *openai.Client
	model  string
}

// NewRouter 创建 Router。
func NewRouter(
	client *openai.Client,
	model string,
) *Router {
	return &Router{
		client: client,
		model:  model,
	}
}

// Decide 根据用户问题和 Agent 能力描述做路由决策。
func (r *Router) Decide(
	ctx context.Context,
	query string,
	capabilities []Capability,
) (RouteDecision, error) {

	capabilityJSON, err := json.MarshalIndent(
		capabilities,
		"",
		"  ",
	)
	if err != nil {
		return RouteDecision{}, fmt.Errorf(
			"marshal capabilities: %w",
			err,
		)
	}

	prompt := fmt.Sprintf(
		`你是一个 Agent Router。

你的任务不是回答用户问题，
而是判断当前用户问题需要使用 Agent 的哪些能力。

当前 Agent 可用能力：

%s

判断规则：

1. Memory
   用于查询用户过去明确保存的长期记忆，
   例如姓名、职业、偏好、历史目标等。

2. RAG
   用于查询当前 Agent 提供的业务知识库，
   例如业务规则、产品政策、内部业务文档等。

   如果是普通的通用知识问题，
   且不需要查询当前 Agent 的知识库，
   不要使用 RAG。

3. Tools
   用于查询实时业务数据或执行外部操作，
   例如用户、订单、支付、物流等。

4. 如果用户问题不需要以上任何能力，
   则全部设置为 false。

5. 可以同时选择多个能力。

只返回 JSON，不要输出其他内容。

JSON 格式：

{
  "use_memory": true,
  "use_rag": false,
  "use_tool": false
}

用户问题：
%s`,
		string(capabilityJSON),
		query,
	)

	resp, err := r.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: r.model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		},
	)
	if err != nil {
		return RouteDecision{}, fmt.Errorf(
			"call router LLM: %w",
		)
	}

	if len(resp.Choices) == 0 {
		return RouteDecision{}, fmt.Errorf(
			"router LLM returned no choices",
		)
	}

	content := strings.TrimSpace(
		resp.Choices[0].Message.Content,
	)

	// MiniMax 可能返回 <think>...</think>。
	content = cleanRouterThinking(content)

	// 有些模型可能在 JSON 外面附带 Markdown。
	content = cleanRouterJSON(content)

	var decision RouteDecision

	if err := json.Unmarshal(
		[]byte(content),
		&decision,
	); err != nil {
		return RouteDecision{}, fmt.Errorf(
			"parse router decision: %w, content=%s",
			err,
			content,
		)
	}

	return decision, nil
}

func cleanRouterThinking(content string) string {
	for {
		start := strings.Index(content, "<think>")
		if start == -1 {
			break
		}

		end := strings.Index(
			content[start:],
			"</think>",
		)
		if end == -1 {
			break
		}

		end = start + end + len("</think>")

		content = content[:start] + content[end:]
	}

	return strings.TrimSpace(content)
}

func cleanRouterJSON(content string) string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	return strings.TrimSpace(content)
}
func BuildCapabilities() []Capability {
	return []Capability{
		{
			Name:        "memory",
			Description: "查询用户过去明确保存的长期记忆，例如姓名、职业、偏好、历史目标等",
		},
		{
			Name:        "rag",
			Description: "查询当前 Agent 的业务知识库，包括业务规则、产品政策、内部文档等信息",
		},
		{
			Name:        "tools",
			Description: "查询实时业务数据或执行外部操作，例如用户、订单、支付、物流",
		},
	}
}

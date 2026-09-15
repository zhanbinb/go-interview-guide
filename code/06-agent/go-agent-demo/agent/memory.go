package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/openai/openai-go"

	"go-agent-demo/llm"
	"go-agent-demo/memory"
)

// MemoryExtraction 表示 LLM 对用户长期记忆的提取结果。
//
// 例如：
//
//	{
//	    "should_save": true,
//	    "memories": [
//	        {
//	            "key": "user_name",
//	            "value": "张三"
//	        },
//	        {
//	            "key": "occupation",
//	            "value": "Go 后端开发工程师"
//	        }
//	    ]
//	}
type MemoryExtraction struct {
	ShouldSave bool                `json:"should_save"`
	Memories   []memory.MemoryItem `json:"memories"`
}

func (a *Agent) buildMemoryContext(
	ctx context.Context,
	query string,
) string {

	queries, err := a.queryRewriter.Rewrite(
		ctx,
		query,
	)

	if err != nil {

		fmt.Println(
			"Query Rewrite failed:",
			err,
		)

		queries = []string{query}
	}

	fmt.Println("\n========== Query Rewrite ==========")
	fmt.Println("Original Query:", query)
	fmt.Println("Rewritten Queries:", queries)
	fmt.Println("===================================")

	// -----------------------------------------
	// Candidate Merge
	// -----------------------------------------
	//
	// 多个 Query 可能命中同一条 Memory。
	//
	// 例如：
	//
	// Query 1 -> user_profession
	// Query 2 -> 职业
	// Query 3 -> 工作
	//
	// 最终只保留一条 Memory。
	//
	// 如果同一个 Memory 被多个 Query 命中，
	// 保留最高 Hybrid Score。
	// -----------------------------------------

	candidates := make(
		map[string]memory.HybridSearchResult,
	)

	for _, q := range queries {

		results :=
			a.longTermMemory.SearchHybridTopK(
				q,
				5,
			)

		fmt.Println(
			"\n========== Hybrid Retrieval ==========",
		)

		fmt.Println(
			"Query:",
			q,
		)

		for _, result := range results {

			fmt.Printf(
				"Candidate: key=%s keyword=%.4f vector=%.4f hybrid=%.4f value=%s\n",
				result.Item.Key,
				result.KeywordScore,
				result.VectorScore,
				result.HybridScore,
				result.Item.Value,
			)

			existing, exists :=
				candidates[result.Item.Key]

			if !exists ||
				result.HybridScore >
					existing.HybridScore {

				candidates[result.Item.Key] =
					result
			}
		}
	}

	// -----------------------------------------
	// Global Ranking
	// -----------------------------------------

	ranked := make(
		[]memory.HybridSearchResult,
		0,
		len(candidates),
	)

	for _, candidate := range candidates {

		ranked = append(
			ranked,
			candidate,
		)
	}

	sort.Slice(
		ranked,
		func(i, j int) bool {

			if ranked[i].HybridScore !=
				ranked[j].HybridScore {

				return ranked[i].HybridScore >
					ranked[j].HybridScore
			}

			return ranked[i].Item.Key <
				ranked[j].Item.Key
		},
	)

	// -----------------------------------------
	// Top K
	// -----------------------------------------

	const topK = 5

	if len(ranked) > topK {
		ranked = ranked[:topK]
	}

	fmt.Println(
		"\n========== Final Memory Ranking ==========",
	)

	for _, result := range ranked {

		fmt.Printf(
			"key=%s keyword=%.4f vector=%.4f hybrid=%.4f value=%s\n",
			result.Item.Key,
			result.KeywordScore,
			result.VectorScore,
			result.HybridScore,
			result.Item.Value,
		)
	}

	fmt.Println(
		"==========================================",
	)

	// -----------------------------------------
	// No Memory
	// -----------------------------------------

	if len(ranked) == 0 {

		return `
[长期记忆检索结果]

没有找到与当前问题相关的长期记忆。

请不要猜测用户过去的信息。
如果无法从当前上下文得到答案，应明确说明没有相关记录。
`
	}

	// -----------------------------------------
	// Build Memory Context
	// -----------------------------------------

	var builder strings.Builder

	builder.WriteString(
		"[长期记忆检索结果]\n",
	)

	builder.WriteString(
		"以下是从用户历史长期记忆中检索到的相关信息。\n",
	)

	builder.WriteString(
		"只能使用这些已有信息，不要自行猜测。\n\n",
	)

	for _, result := range ranked {

		builder.WriteString(
			fmt.Sprintf(
				"- %s: %s\n",
				result.Item.Key,
				result.Item.Value,
			),
		)
	}

	return builder.String()
}

// extractMemories 从当前对话中提取值得长期保存的信息。
//
// 这是一个辅助流程：
//
//	User
//	  ↓
//	Agent
//	  ↓
//	LLM 正常回答
//	  ↓
//	Memory Extraction
//	  ↓
//	保存 Long-term Memory
//
// Memory Extraction 失败时，
// 不应该影响用户当前请求。

func (a *Agent) extractAndSaveMemory(
	ctx context.Context,
	userQuery string,
	assistantAnswer string,
) error {

	systemPrompt := `
你是一个长期记忆提取器。

你的任务是：
从“用户消息”中提取值得长期保存的用户事实。

只允许从用户明确表达的内容中提取。

非常重要：

1. 不要根据用户的问题推测答案。
2. 不要根据助手回答推测用户事实。
3. 不要把助手自己的描述保存为用户事实。
4. 如果用户只是询问某个 Memory 是否存在，不要创建新的 Memory。
5. 如果用户没有明确提供某项信息，不要保存该信息。
6. 不要根据上下文猜测用户属性。
7. 不要把“可能”“应该”“似乎”等推测保存为事实。
8. 如果没有值得保存的信息，返回空数组。

例如：

用户：
“我叫张三，我是一名 Go 后端开发工程师。”

可以提取：

{
  "memories": [
    {
      "key": "user_name",
      "value": "张三"
    },
    {
      "key": "user_profession",
      "value": "Go 后端开发工程师"
    }
  ]
}

用户：
“我最近主要想找 Go 后端开发相关的工作。”

可以提取：

{
  "memories": [
    {
      "key": "career_goal",
      "value": "正在寻找 Go 后端开发相关的工作"
    },
    {
      "key": "tech_direction",
      "value": "Go 后端开发"
    }
  ]
}

但是：

用户：
“你还记得我的 user_role 吗？”

不能提取任何 Memory。

返回：

{
  "memories": []
}

只输出 JSON，不要输出 Markdown，不要输出解释。
`

	// ---------------------------------------------------------
	// 只把用户消息交给 Memory Extraction
	// ---------------------------------------------------------

	userPrompt := fmt.Sprintf(
		"用户消息：\n%s",
		userQuery,
	)

	resp, err := a.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Model: a.model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(systemPrompt),
				openai.UserMessage(userPrompt),
			},
		},
	)

	if err != nil {
		return fmt.Errorf("memory extraction LLM call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("memory extraction returned no choices")
	}

	// ---------------------------------------------------------
	// LLM 原始输出
	// ---------------------------------------------------------

	content := resp.Choices[0].Message.Content

	fmt.Println("\n========== Memory Extraction ==========")
	fmt.Println("User Query:")
	fmt.Println(userQuery)

	fmt.Println("\nLLM Raw Output:")
	fmt.Println(content)

	// ---------------------------------------------------------
	// 清理 <think>
	// ---------------------------------------------------------

	content = llm.CleanThinking(content)

	// ---------------------------------------------------------
	// 清理 Markdown Code Fence
	// ---------------------------------------------------------

	content = cleanJSONContent(content)

	fmt.Println("\nCleaned JSON:")
	fmt.Println(content)

	fmt.Println("=======================================")

	// ---------------------------------------------------------
	// JSON Parse
	// ---------------------------------------------------------

	var result struct {
		Memories []memory.MemoryItem `json:"memories"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return fmt.Errorf(
			"parse memory extraction result failed: %w; content=%s",
			err,
			content,
		)
	}

	// ---------------------------------------------------------
	// 保存 Memory
	// ---------------------------------------------------------

	for _, item := range result.Memories {

		key := strings.TrimSpace(item.Key)
		value := strings.TrimSpace(item.Value)

		if key == "" || value == "" {
			continue
		}

		a.longTermMemory.Save(key, value)

		fmt.Printf(
			"Memory Saved: %s = %s\n",
			key,
			value,
		)
	}

	return nil
}

// cleanJSONContent 清理模型可能返回的 Markdown Code Fence。
//
// 例如：
//
//	```json
//	{
//	    "should_save": true
//	}
//	```
//
// 转换成：
//
//	{
//	    "should_save": true
//	}
func cleanJSONContent(content string) string {

	content = strings.TrimSpace(content)

	// ```json
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(
			content,
			"```json",
		)

		content = strings.TrimSpace(content)
	}

	// ```javascript / ``` 等情况
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(
			content,
			"```",
		)

		content = strings.TrimSpace(content)
	}

	// 结尾 ```
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(
			content,
			"```",
		)

		content = strings.TrimSpace(content)
	}

	return content
}

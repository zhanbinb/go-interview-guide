package client

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go"
)

// ConvertTools 将 MCP Tool 定义转换成 LLM Tool 定义。
//
// MCP Tool:
//
//	Name
//	Description
//	InputSchema
//
// LLM Tool:
//
//	Function.Name
//	Function.Description
//	Function.Parameters
func ConvertTools(
	tools []*mcp.Tool,
) []openai.ChatCompletionToolParam {

	result := make(
		[]openai.ChatCompletionToolParam,
		0,
		len(tools),
	)

	for _, tool := range tools {
		inputSchema, err := json.Marshal(
			tool.InputSchema,
		)
		if err != nil {
			continue
		}

		var parameters map[string]any

		if err := json.Unmarshal(
			inputSchema,
			&parameters,
		); err != nil {
			continue
		}

		result = append(
			result,
			openai.ChatCompletionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        tool.Name,
					Description: openai.String(tool.Description),
					Parameters:  parameters,
				},
			},
		)
	}

	return result
}

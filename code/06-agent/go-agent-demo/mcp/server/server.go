package server

import (
	"context"
	"encoding/json"
	"fmt"

	"go-agent-demo/mcp/protocol"
	"go-agent-demo/tools"
)

// Server 表示一个 MCP Server。
//
// 当前版本先模拟 MCP Server 的核心职责：
// 1. 提供 Tool 列表
// 2. 根据 Tool Name 调用 Tool
//
// 暂时不涉及真正的 MCP 协议、JSON-RPC、网络通信。
type Server struct {
	registry *tools.Registry
}

// New 创建 MCP Server。
func New(registry *tools.Registry) *Server {
	return &Server{
		registry: registry,
	}
}

// ToolInfo 描述 MCP Server 提供的 Tool。
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListTools 返回 MCP Server 提供的所有 Tool。
func (s *Server) ListTools() []ToolInfo {
	allTools := s.registry.All()

	result := make([]ToolInfo, 0, len(allTools))

	for _, tool := range allTools {
		result = append(result, ToolInfo{
			Name:        tool.Name,
			Description: tool.Description,
		})
	}

	return result
}

// CallTool 调用指定 Tool。
func (s *Server) CallTool(
	ctx context.Context,
	name string,
	args map[string]any,
) (string, error) {

	// 当前示例暂时没有使用 context。
	// 真正 MCP Server 接入后，context 会参与请求生命周期管理。
	_ = ctx

	tool, ok := s.registry.Get(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// 当前 Tool.Handler 接收的是 JSON 字符串，
	// 所以这里把参数转换成 JSON。
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("marshal tool args: %w", err)
	}

	result, err := tool.Handler(string(argsJSON))
	if err != nil {
		return "", fmt.Errorf("call tool %s: %w", name, err)
	}

	return result, nil
}

// HandleJSONRPC 处理一个 JSON-RPC Request。
func (s *Server) HandleJSONRPC(
	requestJSON string,
) (string, error) {

	var request protocol.Request

	if err := json.Unmarshal(
		[]byte(requestJSON),
		&request,
	); err != nil {
		return "", fmt.Errorf(
			"decode json-rpc request: %w",
			err,
		)
	}

	switch request.Method {

	case "tools/list":
		return s.handleToolsList(request)

	case "tools/call":
		return s.handleToolsCall(request)

	default:
		return s.errorResponse(
			request.ID,
			-32601,
			"method not found",
		)
	}
}
func (s *Server) handleToolsList(
	request protocol.Request,
) (string, error) {

	tools := s.ListTools()

	response := protocol.Response{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}

	return marshalResponse(response)
}

func (s *Server) handleToolsCall(
	request protocol.Request,
) (string, error) {

	var params protocol.ToolsCallParams

	if err := json.Unmarshal(
		request.Params,
		&params,
	); err != nil {
		return "", fmt.Errorf(
			"decode tools/call params: %w",
			err,
		)
	}

	result, err := s.CallTool(
		context.Background(),
		params.Name,
		params.Arguments,
	)

	if err != nil {
		return s.errorResponse(
			request.ID,
			-32000,
			err.Error(),
		)
	}

	response := protocol.Response{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: protocol.ToolsCallResult{
			Content: []protocol.ContentItem{
				{
					Type: "text",
					Text: result,
				},
			},
		},
	}

	return marshalResponse(response)
}

func (s *Server) errorResponse(
	id int,
	code int,
	message string,
) (string, error) {

	response := protocol.Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &protocol.RPCError{
			Code:    code,
			Message: message,
		},
	}

	return marshalResponse(response)
}

func marshalResponse(
	response protocol.Response,
) (string, error) {

	data, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf(
			"encode json-rpc response: %w",
			err,
		)
	}

	return string(data), nil
}

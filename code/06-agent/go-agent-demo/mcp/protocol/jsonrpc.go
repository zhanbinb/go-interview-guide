package protocol

import "encoding/json"

// Request 表示一个最小 JSON-RPC Request。
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response 表示一个最小 JSON-RPC Response。
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError 表示 JSON-RPC Error。
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolsCallParams 表示 tools/call 的参数。
type ToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolsCallResult 表示 Tool 调用结果。
type ToolsCallResult struct {
	Content []ContentItem `json:"content"`
}

// ContentItem 表示 MCP 返回的内容。
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

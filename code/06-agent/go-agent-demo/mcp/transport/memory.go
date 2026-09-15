package transport

// MemoryTransport 是一个教学用的内存 Transport。
//
// 它不进行真正的网络通信。
// Client 和 Server 通过它传递 JSON 字符串。
type MemoryTransport struct {
	handler func(request string) (string, error)
}

// NewMemoryTransport 创建 Transport。
func NewMemoryTransport(
	handler func(request string) (string, error),
) *MemoryTransport {
	return &MemoryTransport{
		handler: handler,
	}
}

// Send 将 Request 发送给 Server。
func (t *MemoryTransport) Send(request string) (string, error) {
	return t.handler(request)
}

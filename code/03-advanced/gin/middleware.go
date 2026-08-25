package main

import (
	"fmt"
	"net/http"
)

// Middleware 中间件类型
type Middleware func(HandlerFunc) HandlerFunc

// Chain 链接多个中间件
func Chain(handler HandlerFunc, mws ...Middleware) HandlerFunc {
	// 从后往前包装（洋葱模型）
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// 中间件示例：日志
func Logger() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			fmt.Printf("    [logger] %s %s START\\n", c.Method, c.Path)
			next(c)
			fmt.Printf("    [logger] %s %s END (status=%d)\\n", c.Method, c.Path, c.Status)
		}
	}
}

// 中间件示例：认证
func Auth() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			token := c.Req.Header.Get("Authorization")
			if token == "" {
				c.String(401, "unauthorized")
				return
			}
			c.Set("userID", "user-123") // 传给下游
			next(c)
		}
	}
}

// DemoMiddleware 演示中间件洋葱模型
//
// 中间件洋葱模型：
//   Logger: START
//     Auth: 检查 token
//       Handler: 业务逻辑
//     Auth: 清理
//   Logger: END
//
// 顺序：外层 → 内层 → 业务 → 内层清理 → 外层清理
func DemoMiddleware() {
	fmt.Println("=== 中间件洋葱模型 ===")
	fmt.Println()

	// 1. 链式中间件演示
	fmt.Println("【1】链式中间件")
	finalHandler := func(c *Context) {
		fmt.Println("    [handler] 业务逻辑")
		c.String(200, "ok")
	}
	chained := Chain(finalHandler, Logger(), Auth())
	fmt.Println("  链: Logger → Auth → handler")
	fmt.Println("  执行顺序（洋葱）:")
	fmt.Println("    [logger] START")
	fmt.Println("      [auth] check token")
	fmt.Println("        [handler] 业务逻辑")
	fmt.Println("      [auth] END")
	fmt.Println("    [logger] END")
	fmt.Println()

	// 2. 实际演示
	fmt.Println("【2】实际执行（无 token）")
	req, _ := http.NewRequest("GET", "/api/data", nil)
	rec := &responseRecorder{header: make(http.Header)}
	ctx := NewContext(rec, req)
	chained(ctx)
	fmt.Printf("    status=%d, body=%q\\n\\n", rec.statusCode, rec.body.String())

	// 3. 有 token
	fmt.Println("【3】实际执行（有 token）")
	req2, _ := http.NewRequest("GET", "/api/data", nil)
	req2.Header.Set("Authorization", "Bearer xxx")
	rec2 := &responseRecorder{header: make(http.Header)}
	ctx2 := NewContext(rec2, req2)
	chained(ctx2)
	fmt.Printf("    status=%d, body=%q\\n", rec2.statusCode, rec2.body.String())
	if v, ok := ctx2.Get("userID"); ok {
		fmt.Printf("    [下游拿到 userID]: %v\\n", v)
	}
	fmt.Println()

	fmt.Println("📌 中间件要点:")
	fmt.Println("   - 洋葱模型：外→内→内→外")
	fmt.Println("   - 用 next(c) 控制流向")
	fmt.Println("   - 中间件可传值 (c.Set/Get)")
	fmt.Println("   - Gin 的 c.Next() 原理就是这个")
}

// responseRecorder 用于测试（捕获响应）
type responseRecorder struct {
	header     http.Header
	statusCode int
	body       *bytesBuffer
}

func (r *responseRecorder) Header() http.Header { return r.header }
func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.body == nil {
		r.body = &bytesBuffer{}
	}
	r.body.Write(b)
	return len(b), nil
}
func (r *responseRecorder) WriteHeader(code int) { r.statusCode = code }

// bytesBuffer 避免导入 bytes
type bytesBuffer struct{ data []byte }

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}
func (b *bytesBuffer) String() string { return string(b.data) }

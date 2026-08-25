package main

import (
	"fmt"
	"net/http"
)

// Engine 类似 Gin 的 gin.Engine
type Engine struct {
	*Router
	middlewares []HandlerFunc
}

func New() *Engine {
	return &Engine{Router: NewRouter()}
}

// Use 注册全局中间件（类似 Gin 的 r.Use()）
func (e *Engine) Use(mw ...HandlerFunc) {
	e.middlewares = append(e.middlewares, mw...)
}

// runHandler 走中间件链
func (e *Engine) runHandler(c *Context) {
	// 中间件包住最终的 handler（这里简化为直接调用）
	for i, mw := range e.middlewares {
		_ = i
		c.handler = append(c.handler, mw)
	}
}

// DemoServer 演示完整 server 启动
func DemoServer() {
	fmt.Println("=== 完整 server 启动 ===")
	fmt.Println()

	e := New()

	// 全局中间件
	e.Use(func(c *Context) {
		fmt.Printf("    [middleware] %s %s\\n", c.Method, c.Path)
	})

	// 路由
	e.GET("/", func(c *Context) {
		c.String(200, "hello world")
	})
	e.GET("/health", func(c *Context) {
		c.JSON(200, map[string]string{"status": "ok"})
	})

	fmt.Println("Gin 实际用法:")
	fmt.Println("  r := gin.Default() // 含 Logger + Recovery 中间件")
	fmt.Println("  r.GET(\"/\", handler)")
	fmt.Println("  r.Run(\":8080\")")
	fmt.Println()

	fmt.Println("我们的简化版:")
	fmt.Println("  e := New()")
	fmt.Println("  e.Use(middleware...)")
	fmt.Println("  e.GET(path, handler)")
	fmt.Println("  http.ListenAndServe(\":8080\", e)")
	fmt.Println()

	fmt.Println("📌 Gin vs 标准库:")
	fmt.Println("  - 标准库: HandlerFunc + http.Handle (繁琐)")
	fmt.Println("  - Gin:     Engine + Route Group + Middleware")
	fmt.Println("  - 中间间链: c.Next() 实现洋葱调用")
	fmt.Println("  - 性能:   Gin 用 radix tree, 比 map 快很多")

	// 实际启动一个短命 server 测试一下
	fmt.Println("\\n实际启动测试:")
	fmt.Println("  启动 :18080, 访问 / 和 /health")
	go func() {
		_ = http.ListenAndServe(":18080", e)
	}()

	// 真实请求
	// 注意：sandbox 可能没有网络访问，这里只演示启动逻辑
	fmt.Println("  (服务器已启动，请用 curl 测试，或跳过此 demo)")
}

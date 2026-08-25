package main

import (
	"fmt"
	"net/http"
)

// Router 简化版 Gin Router（用 map 模拟，真正的 Gin 用 radix tree）
type Router struct {
	routes map[string]map[string]HandlerFunc // path -> method -> handler
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]HandlerFunc),
	}
}

// Handle 注册路由（类似 Gin 的 GET/POST/PUT/DELETE）
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]HandlerFunc)
	}
	r.routes[path][method] = handler
}

func (r *Router) GET(path string, handler HandlerFunc)    { r.Handle("GET", path, handler) }
func (r *Router) POST(path string, handler HandlerFunc)   { r.Handle("POST", path, handler) }
func (r *Router) PUT(path string, handler HandlerFunc)    { r.Handle("PUT", path, handler) }
func (r *Router) DELETE(path string, handler HandlerFunc) { r.Handle("DELETE", path, handler) }

// ServeHTTP 让 Router 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	if handlers, ok := r.routes[path]; ok {
		if handler, ok := handlers[method]; ok {
			ctx := NewContext(w, req)
			handler(ctx)
			return
		}
	}
	// 404
	ctx := NewContext(w, req)
	ctx.String(404, "404 page not found")
}

// ListRoutes 打印所有注册的路由
func (r *Router) ListRoutes() {
	fmt.Println("注册的路由:")
	for path, handlers := range r.routes {
		for method := range handlers {
			fmt.Printf("  %-6s %s\n", method, path)
		}
	}
}

// DemoRouter 演示路由注册
func DemoRouter() {
	fmt.Println("=== 简单路由注册 ===")
	fmt.Println()

	r := NewRouter()
	r.GET("/hello", func(c *Context) {
		c.JSON(200, map[string]string{"msg": "hello"})
	})
	r.GET("/users", func(c *Context) {
		c.String(200, "users list")
	})
	r.POST("/users", func(c *Context) {
		c.String(201, "user created")
	})

	r.ListRoutes()
	fmt.Println()
	fmt.Println("Gin 用 radix tree 实现，性能比 map 更好")
	fmt.Println("但 API 一样: r.GET(path, handler)")
}

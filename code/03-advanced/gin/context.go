package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HandlerFunc 简化版 Gin 的 handler 类型
type HandlerFunc func(*Context)

// Context 简化版 Gin 的 Context
// Gin 的 Context 实际有 100+ 方法，这里只演示核心功能
type Context struct {
	Writer  http.ResponseWriter
	Req     *http.Request
	Path    string
	Method  string
	Status  int
	Keys    map[string]any // 用于中间件传值
	handler []HandlerFunc  // 中间件链
	index   int           // 当前执行到第几个 handler
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := &Context{
		Writer: w,
		Req:    r,
		Status: 200,
		Keys:   make(map[string]any),
		index:  -1,
	}
	if r != nil {
		c.Path = r.URL.Path
		c.Method = r.Method
	}
	return c
}

// String 写字符串响应
func (c *Context) String(code int, msg string) {
	c.Status = code
	c.Writer.WriteHeader(code)
	fmt.Fprint(c.Writer, msg)
}

// JSON 写 JSON 响应
func (c *Context) JSON(code int, data any) error {
	c.Status = code
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(code)
	return json.NewEncoder(c.Writer).Encode(data)
}

// Set 设置键值（中间件之间传值）
func (c *Context) Set(key string, value any) {
	c.Keys[key] = value
}

// Get 获取键值
func (c *Context) Get(key string) (any, bool) {
	v, ok := c.Keys[key]
	return v, ok
}

// Param 获取 URL 参数（路由中的 :id 等）
func (c *Context) Param(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Query 获取 query string
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Next 继续执行中间件链（洋葱模型核心）
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handler) {
		if c.handler[c.index] != nil {
			c.handler[c.index](c)
		}
		c.index++
	}
}

// DemoContext 演示 Context 用法
func DemoContext() {
	fmt.Println("=== Context 传值 ===")
	fmt.Println()

	// 1. 基本响应
	fmt.Println("【1】基本响应 (String / JSON)")
	ctx := NewContext(nil, nil)
	ctx.String(200, "hello")
	fmt.Printf("  Status = %d\\n", ctx.Status)

	ctx2 := NewContext(nil, nil)
	_ = ctx2.JSON(201, map[string]string{"id": "1"})
	fmt.Printf("  JSON 已编码\\n\\n", )

	// 2. 传值
	fmt.Println("【2】中间件传值 (Set/Get)")
	ctx3 := NewContext(nil, nil)
	ctx3.Set("userID", 42)
	if v, ok := ctx3.Get("userID"); ok {
		fmt.Printf("  userID = %v (类型 %T)\\n", v, v)
	}
	fmt.Println()

	// 3. URL 参数
	fmt.Println("【3】URL 参数 (Query / Param)")
	fmt.Println("  在 Gin 里 r.GET('/users/:id', ...) 然后 c.Param('id')")
	fmt.Println("  在本 demo: c.Query('id') 取 query string 的 id")
	fmt.Println()

	fmt.Println("📌 Context 是 Gin 的精髓:")
	fmt.Println("   - 封装了 request/response")
	fmt.Println("   - 提供 params/keys/c.Next() 等便捷方法")
	fmt.Println("   - 复用 c（不每次创建），性能高")
}

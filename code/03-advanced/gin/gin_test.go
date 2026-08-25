package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRouter 验证路由匹配
func TestRouter(t *testing.T) {
	r := NewRouter()
	r.GET("/hello", func(c *Context) {
		c.String(200, "hi")
	})

	// 测试匹配
	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "hi" {
		t.Errorf("expected hi, got %q", w.Body.String())
	}

	// 测试不匹配的 404
	req = httptest.NewRequest("GET", "/notfound", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// TestContextSetGet 验证 Context 传值
func TestContextSetGet(t *testing.T) {
	c := NewContext(nil, nil)
	c.Set("user", "alice")
	if v, ok := c.Get("user"); !ok || v != "alice" {
		t.Errorf("expected alice, got %v ok=%v", v, ok)
	}
}

// TestContextJSON 验证 JSON 响应
func TestContextJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c := NewContext(w, httptest.NewRequest("GET", "/", nil))
	err := c.JSON(200, map[string]string{"status": "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected json content type, got %s", w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Errorf("expected ok in body, got %s", w.Body.String())
	}
}

// TestMiddlewareChain 验证中间件链
func TestMiddlewareChain(t *testing.T) {
	var order []string
	finalHandler := func(c *Context) {
		order = append(order, "handler")
	}
	mw1 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			order = append(order, "mw1-before")
			next(c)
			order = append(order, "mw1-after")
		}
	}
	mw2 := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) {
			order = append(order, "mw2-before")
			next(c)
			order = append(order, "mw2-after")
		}
	}
	chained := Chain(finalHandler, mw1, mw2)
	c := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	chained(c)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Errorf("at %d: expected %s, got %s", i, expected[i], order[i])
		}
	}
}

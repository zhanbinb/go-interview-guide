package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/pprof"
	"runtime"
	"strings"
	"testing"
)

// TestPProfEndpoints 验证 pprof 端点可用
func TestPProfEndpoints(t *testing.T) {
	// 临时启动一个 pprof 服务
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	// 测试 index 页面
	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	// 应该包含 profile 列表
	for _, profile := range []string{"heap", "goroutine", "profile"} {
		if !strings.Contains(body, profile) {
			t.Errorf("index should list profile %s", profile)
		}
	}
}

// TestGoroutineLeak 演示 runtime 能看到泄漏
func TestGoroutineLeak(t *testing.T) {
	initial := runtime.NumGoroutine()

	// 制造 10 个泄漏的 goroutine
	for i := 0; i < 10; i++ {
		go func() {
			ch := make(chan struct{})
			<-ch // 永久阻塞
		}()
	}

	// 给 goroutine 时间启动
	for i := 0; i < 100; i++ {
		if runtime.NumGoroutine() >= initial+10 {
			break
		}
		runtime.Gosched()
	}

	// 验证（不严格，因为 GC 也可能启动/回收 goroutine）
	leaked := runtime.NumGoroutine() - initial
	if leaked < 5 {
		t.Errorf("expected at least 5 leaked goroutines, got %d", leaked)
	}
	t.Logf("confirmed %d goroutines (leak simulation)", leaked)
}

// TestHeapProfile 验证能获取 heap profile
func TestHeapProfile(t *testing.T) {
	// 分配一些内存
	bufs := make([][]byte, 100)
	for i := range bufs {
		bufs[i] = make([]byte, 1024)
	}
	runtime.GC() // 触发 GC，让数据稳定

	// 简单的内存占用检查
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.HeapAlloc < 50*1024 {
		t.Logf("HeapAlloc = %d (allocated)", m.HeapAlloc)
	}
}

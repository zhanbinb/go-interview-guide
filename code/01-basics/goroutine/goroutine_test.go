package main

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestLoopVarIsolated 验证 Go 1.22+ 循环变量在 goroutine 中独立
func TestLoopVarIsolated(t *testing.T) {
	var seen []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			seen = append(seen, i)
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(seen) != 5 {
		t.Fatalf("expected 5 values, got %d", len(seen))
	}

	// 验证每个值都出现且只出现一次
	counts := make(map[int]int)
	for _, v := range seen {
		counts[v]++
	}
	for i := 0; i < 5; i++ {
		if counts[i] != 1 {
			t.Errorf("expected i=%d to appear exactly once, got %d times", i, counts[i])
		}
	}
}

// TestGoroutineNoLeak 验证正常 goroutine 创建后能正确退出
func TestGoroutineNoLeak(t *testing.T) {
	initial := runtime.NumGoroutine()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}
	wg.Wait()

	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	if got := runtime.NumGoroutine(); got > initial+5 {
		t.Errorf("goroutine leak: initial=%d, after=%d", initial, got)
	}
}

// TestStackGrowth 验证 goroutine 栈可以动态增长
func TestStackGrowth(t *testing.T) {
	done := make(chan int, 1)

	go func() {
		// 深度递归，每次占用栈空间
		var deep func(n int) int
		deep = func(n int) int {
			// 故意占栈空间（编译器无法优化掉）
			buf := make([]byte, 256)
			_ = buf
			if n == 0 {
				return 0
			}
			return deep(n-1) + 1
		}
		done <- deep(50)
	}()

	select {
	case depth := <-done:
		if depth != 50 {
			t.Errorf("expected recursion depth 50, got %d", depth)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("stack growth test timed out")
	}
}


// BenchmarkGoroutineCreate 测试 goroutine 创建开销
func BenchmarkGoroutineCreate(b *testing.B) {
	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
		}()
	}
	wg.Wait()
}

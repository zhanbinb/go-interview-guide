package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGoConcurrency 演示 Go 并发的轻量
// 1000 个 goroutine 几乎瞬间创建完
func TestGoConcurrency(t *testing.T) {
	const N = 1000
	var wg sync.WaitGroup
	var counter int64
	start := time.Now()
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	if counter != N {
		t.Errorf("expected %d, got %d", N, counter)
	}
	t.Logf("创建 %d 个 goroutine 用时: %v", N, elapsed)
	// Go: 1000 个 goroutine 几乎瞬间 (< 10ms)
	// Java thread: 1000 个需要 1+ 秒，OS 资源紧张
	// Python thread: 受 GIL 限制，且开销大
}

// BenchmarkGoGoroutine vs 其他语言
func BenchmarkGoGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ch := make(chan struct{})
		go func() {
			close(ch)
		}()
		<-ch
	}
}

// BenchmarkGoChannel 测 channel 通信
func BenchmarkGoChannel(b *testing.B) {
	ch := make(chan int, 100)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
	}()
	for range ch {
	}
}

// BenchmarkGoSync 测 sync.Mutex
func BenchmarkGoSync(b *testing.B) {
	var mu sync.Mutex
	counter := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		counter++
		mu.Unlock()
	}
}

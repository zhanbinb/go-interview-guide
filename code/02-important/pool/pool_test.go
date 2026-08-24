package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestPoolGetPut 验证 Pool Get/Put 基本功能
func TestPoolGetPut(t *testing.T) {
	pool := &sync.Pool{
		New: func() any { return "new" },
	}
	v := pool.Get()
	if v != "new" {
		t.Errorf("expected new, got %v", v)
	}
	pool.Put("cached")
	v = pool.Get()
	if v != "cached" {
		t.Errorf("expected cached, got %v", v)
	}
}

// TestPoolConcurrent 验证并发安全
func TestPoolConcurrent(t *testing.T) {
	pool := &sync.Pool{
		New: func() any { return &struct{ n int }{} },
	}
	const N = 1000
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			x := pool.Get().(*struct{ n int })
			atomic.AddInt64(&counter, 1)
			pool.Put(x)
		}()
	}
	wg.Wait()
	if counter != int64(N) {
		t.Errorf("expected %d, got %d", N, counter)
	}
}

// TestWorkerPool 验证 Worker Pool 模式
func TestWorkerPool(t *testing.T) {
	const (
		workers = 3
		jobs    = 10
	)
	ch := make(chan int, jobs)
	var wg sync.WaitGroup
	var processed int64
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range ch {
				atomic.AddInt64(&processed, 1)
				_ = n
			}
		}()
	}
	for i := 0; i < jobs; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()
	if processed != int64(jobs) {
		t.Errorf("expected %d processed, got %d", jobs, processed)
	}
}

// TestSyncOnce 验证 sync.Once 只执行一次
func TestSyncOnce(t *testing.T) {
	var once sync.Once
	count := 0
	for i := 0; i < 100; i++ {
		once.Do(func() { count++ })
	}
	if count != 1 {
		t.Errorf("once.Do should run once, got %d", count)
	}
}

func BenchmarkSyncPool(b *testing.B) {
	pool := &sync.Pool{New: func() any { return make([]byte, 1024) }}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get().([]byte)
		pool.Put(buf)
	}
}

func BenchmarkDirectAlloc(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

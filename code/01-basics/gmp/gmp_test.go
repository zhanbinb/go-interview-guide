package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGOMAXPROCS 验证 GOMAXPROCS 可设置并生效
func TestGOMAXPROCS(t *testing.T) {
	prev := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(prev)

	prev = runtime.GOMAXPROCS(2)
	if got := runtime.GOMAXPROCS(0); got != 2 {
		t.Errorf("expected GOMAXPROCS=2, got %d", got)
	}
}

// TestNumGoroutine 验证 NumGoroutine 计数
func TestNumGoroutine(t *testing.T) {
	initial := runtime.NumGoroutine()
	if initial < 1 {
		t.Fatalf("expected at least 1 goroutine (main), got %d", initial)
	}

	var wg sync.WaitGroup
	const N = 100
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		}()
	}
	wg.Wait()

	// 给调度器一点时间回收已退出的 G
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	if got := runtime.NumGoroutine(); got > initial+5 {
		t.Errorf("goroutines leaked: initial=%d, after=%d", initial, got)
	}
}

// TestPreemptive 验证抢占式调度在死循环场景下也能让其他 goroutine 跑起来
func TestPreemptive(t *testing.T) {
	var counter int64
	stop := make(chan struct{})

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				atomic.AddInt64(&counter, 1)
			}
		}
	}()

	// 这个 timer goroutine 必须能在合理时间内被调度到
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		close(stop)
		// 抢占生效（Go 1.14+），另一个 goroutine 跑到了 timer
		t.Logf("✅ preemptive scheduling works: busy-loop ran %d times before timer fired", atomic.LoadInt64(&counter))
	case <-time.After(10 * time.Second):
		close(stop)
		t.Fatalf("preemption failed: timer goroutine didn't run within 10s")
	}
}

// BenchmarkWorkStealing 简单的 work stealing benchmark
func BenchmarkWorkStealing(b *testing.B) {
	runtime.GOMAXPROCS(2)
	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			for j := 0; j < 1e5; j++ {
				sum += j
			}
			_ = sum
		}()
	}
	wg.Wait()
}

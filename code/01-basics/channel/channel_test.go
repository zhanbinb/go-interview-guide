package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestUnbufferedChannelSync 验证 unbuffered channel 同步性
func TestUnbufferedChannelSync(t *testing.T) {
	ch := make(chan int)
	go func() {
		ch <- 42
	}()
	v := <-ch
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

// TestBufferedChannelAsync 验证 buffered channel 异步性（不阻塞直到满）
func TestBufferedChannelAsync(t *testing.T) {
	ch := make(chan int, 3)
	for i := 0; i < 3; i++ {
		ch <- i
	}
	if len(ch) != 3 {
		t.Errorf("expected len=3, got %d", len(ch))
	}
	for i := 0; i < 3; i++ {
		v := <-ch
		if v != i {
			t.Errorf("expected %d, got %d", i, v)
		}
	}
}

// TestCloseAndRead 验证 close 后读不阻塞，返回零值 + ok=false
func TestCloseAndRead(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)

	v, ok := <-ch
	if v != 1 || !ok {
		t.Errorf("expected (1,true), got (%d,%v)", v, ok)
	}
	v, ok = <-ch
	if v != 2 || !ok {
		t.Errorf("expected (2,true), got (%d,%v)", v, ok)
	}
	v, ok = <-ch
	if v != 0 || ok {
		t.Errorf("expected (0,false), got (%d,%v)", v, ok)
	}
}

// TestPanicSendToClosed 验证向 closed channel 发送会 panic
func TestPanicSendToClosed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from send to closed channel")
		} else {
			t.Logf("got expected panic: %v", r)
		}
	}()
	ch := make(chan int)
	close(ch)
	ch <- 1
}

// TestPanicCloseClosed 验证重复 close 会 panic
func TestPanicCloseClosed(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from close of closed channel")
		}
	}()
	ch := make(chan int)
	close(ch)
	close(ch)
}

// TestSelectTimeout 验证 select + time.After 超时
func TestSelectTimeout(t *testing.T) {
	ch := make(chan int)
	got := ""
	select {
	case <-ch:
		got = "received"
	case <-time.After(50 * time.Millisecond):
		got = "timeout"
	}
	if got != "timeout" {
		t.Errorf("expected timeout, got %q", got)
	}
}

// TestSelectDefault 验证 default 实现非阻塞
func TestSelectDefault(t *testing.T) {
	ch := make(chan int)
	executed := "default"
	select {
	case v := <-ch:
		executed = "received"
		_ = v
	default:
		executed = "default"
	}
	if executed != "default" {
		t.Errorf("expected default branch, got %q", executed)
	}
}

// TestNilChannelInSelect 验证 nil channel 在 select 中被跳过
func TestNilChannelInSelect(t *testing.T) {
	var nilCh chan int
	executed := false
	select {
	case nilCh <- 42:
		executed = true // 永远不会执行
	default:
	}
	if executed {
		t.Error("nil channel should never be ready")
	}
}

// TestSemaphorePattern 验证 cap=1 channel 当信号量
func TestSemaphorePattern(t *testing.T) {
	sem := make(chan struct{}, 1)
	var concurrent int32
	var maxConcurrent int32
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			cur := atomic.AddInt32(&concurrent, 1)
			for {
				old := atomic.LoadInt32(&maxConcurrent)
				if cur <= old {
					break
				}
				if atomic.CompareAndSwapInt32(&maxConcurrent, old, cur) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&concurrent, -1)
		}()
	}
	wg.Wait()

	if maxConcurrent > 1 {
		t.Errorf("semaphore failed: max concurrent = %d (should be 1)", maxConcurrent)
	}
}

// TestChannelFanOut 验证 Fan-out 模式
func TestChannelFanOut(t *testing.T) {
	jobs := make(chan int, 100)
	var wg sync.WaitGroup
	var processed int32

	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				atomic.AddInt32(&processed, 1)
			}
		}()
	}

	for i := 0; i < 30; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	if processed != 30 {
		t.Errorf("expected 30 processed, got %d", processed)
	}
}

// TestPipeline 验证 pipeline 模式
func TestPipeline(t *testing.T) {
	stage1 := make(chan int)
	stage2 := make(chan int)

	go func() {
		defer close(stage1)
		for i := 1; i <= 5; i++ {
			stage1 <- i
		}
	}()
	go func() {
		defer close(stage2)
		for v := range stage1 {
			stage2 <- v * 2
		}
	}()

	var results []int
	for v := range stage2 {
		results = append(results, v)
	}

	expected := []int{2, 4, 6, 8, 10}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	for i, v := range results {
		if v != expected[i] {
			t.Errorf("at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

// TestNoLeakAfterUse 验证正常使用后 goroutine 能正确回收
func TestNoLeakAfterUse(t *testing.T) {
	initial := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		ch := make(chan int, 1)
		ch <- i
		<-ch
	}
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	if got := runtime.NumGoroutine(); got > initial+5 {
		t.Errorf("goroutine leak: initial=%d, after=%d", initial, got)
	}
}

// BenchmarkUnbufferedChan 测试 unbuffered channel 通信开销
func BenchmarkUnbufferedChan(b *testing.B) {
	ch := make(chan int)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
	}()
	for range ch {
	}
}

// BenchmarkBufferedChan 测试 buffered channel 通信开销
func BenchmarkBufferedChan(b *testing.B) {
	ch := make(chan int, 1024)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
	}()
	for range ch {
	}
}

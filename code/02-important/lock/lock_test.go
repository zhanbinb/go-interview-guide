package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestMutex 验证 Mutex 基本功能
func TestMutex(t *testing.T) {
	var mu sync.Mutex
	var counter int
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	if counter != 100 {
		t.Errorf("expected 100, got %d", counter)
	}
}

// TestTryLock 验证 TryLock 行为
func TestTryLock(t *testing.T) {
	var mu sync.Mutex
	if !mu.TryLock() {
		t.Error("first TryLock should succeed")
	}
	if mu.TryLock() {
		t.Error("second TryLock should fail when locked")
	}
	mu.Unlock()
	if !mu.TryLock() {
		t.Error("TryLock should succeed after Unlock")
	}
	mu.Unlock()
}

// TestRWMutex 验证 RWMutex 读并发
func TestRWMutex(t *testing.T) {
	var rw sync.RWMutex
	var counter int
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw.RLock()
			_ = counter
			rw.RUnlock()
		}()
	}
	wg.Wait()
}

// TestAtomicAdd 验证原子加法
func TestAtomicAdd(t *testing.T) {
	var cnt int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&cnt, 1)
		}()
	}
	wg.Wait()
	if cnt != 1000 {
		t.Errorf("expected 1000, got %d", cnt)
	}
}

// TestCAS 验证 CAS
func TestCAS(t *testing.T) {
	var value int32 = 10
	if !atomic.CompareAndSwapInt32(&value, 10, 20) {
		t.Error("CAS should succeed")
	}
	if value != 20 {
		t.Errorf("expected 20, got %d", value)
	}
	if atomic.CompareAndSwapInt32(&value, 10, 30) {
		t.Error("CAS should fail (value is not 10)")
	}
}

// TestOnce 验证 sync.Once
func TestOnce(t *testing.T) {
	var once sync.Once
	count := 0
	for i := 0; i < 100; i++ {
		once.Do(func() { count++ })
	}
	if count != 1 {
		t.Errorf("once.Do should run exactly once, got %d", count)
	}
}

// BenchmarkAtomic 对比 atomic
func BenchmarkAtomic(b *testing.B) {
	var cnt int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&cnt, 1)
	}
}

// BenchmarkMutex 对比 mutex
func BenchmarkMutex(b *testing.B) {
	var mu sync.Mutex
	var cnt int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		cnt++
		mu.Unlock()
	}
}

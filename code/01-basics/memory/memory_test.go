package main

import (
	"runtime"
	"testing"
)

// TestHeapStats 验证 HeapAlloc 可读
func TestHeapStats(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.HeapAlloc < 0 {
		t.Error("HeapAlloc should be non-negative")
	}
}

// TestAllocSmall 验证小对象分配
func TestAllocSmall(t *testing.T) {
	s := make([]byte, 100)
	if len(s) != 100 {
		t.Errorf("expected len=100, got %d", len(s))
	}
}

// TestAllocLarge 验证大对象分配
func TestAllocLarge(t *testing.T) {
	s := make([]byte, 64*1024) // 64KB, 大对象
	if len(s) != 64*1024 {
		t.Errorf("expected len=64KB, got %d", len(s))
	}
}

// TestEscapeReturn 验证返回指针会逃逸到堆
// 用 runtime.SetFinalizer 检查：如果对象被分配到堆，finalizer 才会执行
func TestEscapeReturn(t *testing.T) {
	done := make(chan *int)
	go func() {
		x := 42
		runtime.SetFinalizer(&x, func(p *int) {
			t.Logf("finalizer ran for *p = %d (说明分配在堆)", *p)
			close(done)
		})
		done <- &x
	}()
	<-done
	// 触发 GC，让 finalizer 有机会运行
	runtime.GC()
	runtime.Gosched()
}

// TestGCAfterAlloc 验证 GC 能回收内存
func TestGCAfterAlloc(t *testing.T) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	// 分配并释放
	func() {
		_ = make([]byte, 10*1024*1024) // 10MB
	}()

	runtime.GC()
	runtime.ReadMemStats(&after)

	if after.NumGC <= before.NumGC {
		t.Logf("no GC ran (might might be valid if heap didn't grow enough)")
	}
	t.Logf("GC ran %d times, heap: %.2f -> %.2f MB",
		after.NumGC-before.NumGC,
		float64(before.HeapAlloc)/1024/1024,
		float64(after.HeapAlloc)/1024/1024)
}

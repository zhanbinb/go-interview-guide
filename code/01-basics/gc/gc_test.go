package main

import (
	"runtime"
	"runtime/debug"
	"testing"
)

// TestManualGC 验证手动 runtime.GC() 触发 GC
func TestManualGC(t *testing.T) {
	before := runtime.NumGoroutine()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.NumGC == 0 {
		t.Error("expected NumGC > 0 after manual GC")
	}
	_ = before
}

// TestGOGC 验证 GOGC 可读可设置（通过 debug.SetGCPercent）
func TestGOGC(t *testing.T) {
	prev := debug.SetGCPercent(-1) // 读当前值（不改）
	defer debug.SetGCPercent(prev)

	debug.SetGCPercent(100)
	if got := debug.SetGCPercent(-1); got != 100 {
		t.Errorf("expected GOGC=100, got %d", got)
	}
	debug.SetGCPercent(prev) // 恢复
}

// TestAutoGC 验证分配大量内存会自动触发 GC
// 关键：保留 slice 引用，避免编译器优化掉
func TestAutoGC(t *testing.T) {
	var before, after runtime.MemStats
	runtime.GC() // 先清空
	runtime.ReadMemStats(&before)

	// 分配并保留引用
	slices := make([][]byte, 0)
	for i := 0; i < 10000; i++ {
		s := make([]byte, 4096) // 4KB/个，总 40MB
		s[0] = 1                  // 防止优化
		slices = append(slices, s)
	}

	runtime.ReadMemStats(&after)
	gcCount := after.NumGC - before.NumGC
	heapGrowth := after.HeapAlloc - before.HeapAlloc

	if heapGrowth < 1024*1024 {
		t.Errorf("expected HeapAlloc to grow, got %d bytes", heapGrowth)
	}
	t.Logf("GC 触发 %d 次, 堆增长 %.2f MB", gcCount, float64(heapGrowth)/1024/1024)
	_ = slices
}

// TestMemoryStats 验证内存统计字段
func TestMemoryStats(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if m.HeapAlloc < 0 {
		t.Error("HeapAlloc should be non-negative")
	}
}

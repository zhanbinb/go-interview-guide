package main

import (
	"fmt"
	"runtime"
)

// DemoAllocate 演示内存分配层级
//
// ============================================================================
// 分配层级（runtime/malloc.go）：
//
//   ┌─────────────────────────────────────────┐
//   │         mheap (全局，页分配)             │  ← mmap 向 OS 申请
//   │  ┌───────────────────────────────────┐  │
//   │  │       mcentral (按 size class)     │  │  ← 全局，但需加锁
//   │  │  ┌─────────────────────────────┐  │  │
//   │  │  │  mcache (每个 P 一个，无锁)    │  │  │  ← 最快路径
//   │  │  │  span 数组（按 size class）│  │  │
//   │  │  └─────────────────────────────┘  │  │
//   │  └───────────────────────────────────┘  │
//   └─────────────────────────────────────────┘
//
// 对象大小分类：
//   - 微对象 (<16B)：用 mcache 的 tiny allocator
//   - 小对象 (16B ~ 32KB)：按 size class 从 mcache 拿 span
//   - 大对象 (≥32KB)：直接走 mheap，按页（8KB）对齐
// ============================================================================
func DemoAllocate() {
	fmt.Println("=== 内存分配层级 ===")
	fmt.Println()
	fmt.Println("分配流程（按对象大小）:")
	fmt.Println("  微对象 <16B    → mcache 的 tiny allocator（极快）")
	fmt.Println("  小对象 16B-32KB → mcache → mcentral → mheap")
	fmt.Println("  大对象 ≥32KB    → 直接 mheap，按 8KB 页对齐")
	fmt.Println()

	// 演示：不同大小对象
	fmt.Println("【实验】不同大小对象的分配")

	var m1, m2, m3 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 小对象：10000 个 100-byte slice
	small := make([][]byte, 0, 10000)
	for i := 0; i < 10000; i++ {
		s := make([]byte, 100)
		small = append(small, s)
	}
	runtime.ReadMemStats(&m2)

	// 大对象：10 个 64KB
	big := make([][]byte, 0, 10)
	for i := 0; i < 10; i++ {
		b := make([]byte, 64*1024)
		big = append(big, b)
	}
	runtime.ReadMemStats(&m3)

	fmt.Printf("  小对象 (10000 × 100B):  HeapAlloc 增加 %.2f KB\n",
		float64(m2.HeapAlloc-m1.HeapAlloc)/1024)
	fmt.Printf("  大对象 (10 × 64KB):      HeapAlloc 增加 %.2f MB\n",
		float64(m3.HeapAlloc-m2.HeapAlloc)/1024/1024)

	_ = small
	_ = big
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - 分配层级：mcache (P 私有) → mcentral (全局) → mheap (OS)")
	fmt.Println("   - 小对象走 size class 复用 span，无碎片")
	fmt.Println("   - 大对象直接按页分配，不复用 span")
	fmt.Println("   - P 私有 mcache 无锁，是 Go 高并发分配的关键")
}

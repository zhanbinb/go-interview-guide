package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

// DemoSizeClass 演示大对象 vs 小对象
//
// ============================================================================
// 对象大小分类（runtime/sizeclasses.go）：
//
//   微对象 <16B      → mcache tiny allocator（无 size class 概念）
//   小对象 16B-32KB   → 67 种 size class，按大小分配 span
//   大对象 ≥32KB     → 直接 mheap，按页对齐（不通过 size class）
//
// 为什么区分？
//   - 小对象复用 span，零碎片，分配快
//   - 大对象按页对齐，避免浪费空间
// ============================================================================
func DemoSizeClass() {
	fmt.Println("=== 大对象 vs 小对象 ===")
	fmt.Println()

	// 实验 1：验证 small < 32KB → 走 size class
	fmt.Println("【实验 1】小对象 (<32KB) → 走 mcache size class")
	showClass := func(size int) {
		var s []byte
		if size <= 16 {
			fmt.Printf("  %d B: 微对象（tiny allocator）\n", size)
		} else if size < 32*1024 {
			fmt.Printf("  %d B: 小对象（size class）\n", size)
		} else {
			fmt.Printf("  %d KB: 大对象（mheap 直接分配）\n", size/1024)
		}
		_ = s
	}
	for _, n := range []int{8, 16, 100, 1024, 10*1024, 100*1024} {
		showClass(n)
	}
	fmt.Println()

	// 实验 2：演示为什么小对象多了会有 GC 压力
	fmt.Println("【实验 2】GC 扫描成本 vs 对象数量")
	fmt.Println("  假设分配 10000 个 100B 对象 vs 1 个 1MB 对象")
	fmt.Println("  GC 扫描 10000 个对象 > 扫描 1 个大对象")
	fmt.Println("  即使总内存一样")
	fmt.Println()

	// 实验 3：栈 vs 堆大小限制
	fmt.Println("【实验 3】栈大小限制")
	var stackStat runtime.MemStats
	runtime.ReadMemStats(&stackStat)
	fmt.Printf("  默认 goroutine 栈: 2KB 起步, 最大 1GB\n")
	fmt.Printf("  当前线程栈（系统）: %d B (×8 = ~%d KB)\n",
		unsafe.Sizeof(stackStat), unsafe.Sizeof(stackStat)*8)
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - 小对象 <32KB 用 size class（67 种）")
	fmt.Println("   - 大对象 ≥32KB 走 mheap，按页（8KB）对齐")
	fmt.Println("   - 小对象多 → GC 扫描慢（数量问题，不是大小问题）")
	fmt.Println("   - 可以用 sync.Pool 复用小对象（降低 GC 压力）")
}

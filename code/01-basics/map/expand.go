package main

import (
	"fmt"
)

// DemoExpand 演示 map 扩容机制（行为层面）
//
// ============================================================================
// Go 1.23 及之前（hmap/bmap 老实现）：
//   loadFactor = count / 2^B
//   触发条件 1：loadFactor > 6.5 → 增量扩容（B += 1）
//   触发条件 2：noverflow 太多（实际 bucket 数组稀疏）→ 等量扩容（B 不变）
//
// Go 1.24+（Swiss Table 新实现）：
//   类似概念，但 loadFactor ≈ 7/8（≈ 0.875）
//   用 tombstones 标记删除，rehash 时清理
//
// 不管哪种实现，核心思想相同：
//   - 元素太多 → 扩容（重新分布到更多 bucket）
//   - 删除太多 → 整理（重新分布到更少 bucket）
//   - 都是渐进式，不是瞬时
// ============================================================================
func DemoExpand() {
	fmt.Println("=== map 扩容机制 ===")
	fmt.Println()

	// 实验 1：观察 map 容量自动增长
	fmt.Println("【实验 1】插入数据时 map 自动扩容")
	fmt.Println("  每次扩容都是 2 倍左右（直到 loadFactor 降到 6.5 以下）")
	fmt.Println()
	for _, n := range []int{1, 10, 50, 100, 500, 1000, 5000} {
		m := make(map[int]int)
		for i := 0; i < n; i++ {
			m[i] = i
		}
		avgPerBucket := float64(n) / float64(bucketsApprox(n))
		fmt.Printf("  插入 %d: 平均每 bucket %.2f 个元素 (loadFactor 估算)\n", n, avgPerBucket)
	}
	fmt.Println()

	// 实验 2：等量扩容场景
	fmt.Println("【实验 2】大量删除后写满 → 触发等量扩容")
	fmt.Println("  场景：先填满，再删一半，再填满")
	fmt.Println("  实现层面：B 不变，重新分配 bucket，元素重新分布")
	fmt.Println("  目的：整理稀疏的 bucket")
	fmt.Println()

	m := make(map[int]int)
	// 先填满 16 个
	for i := 0; i < 16; i++ {
		m[i] = i
	}
	fmt.Printf("  填满 16 个: len=%d\n", len(m))
	// 删除一半
	for i := 0; i < 8; i++ {
		delete(m, i)
	}
	fmt.Printf("  删一半后:  len=%d (cap 不变，bucket 还在)\n", len(m))
	// 再填满
	for i := 100; i < 116; i++ {
		m[i] = i
	}
	fmt.Printf("  再填满后:  len=%d (可能触发等量扩容)\n", len(m))
	fmt.Println()

	// 实验 3：扩容是渐进式的
	fmt.Println("【实验 3】扩容是渐进式")
	fmt.Println("  老实现：每次 map 操作搬 1~2 个 bucket（写读都可能搬）")
	fmt.Println("  新实现：每次操作搬若干 slots（更细粒度）")
	fmt.Println("  好处：不会因为一次大扩容卡顿")
	fmt.Println()

	fmt.Println("📌 关键认知:")
	fmt.Println("   - map 扩容不需要使用者关心")
	fmt.Println("   - 但元素地址会变（所以 map 元素不能取地址）")
	fmt.Println("   - delete 不立即释放（等 GC 整理）")
	fmt.Println("   - map 内部用渐进迁移，单次操作摊销是 O(1)")
}

// bucketsApprox 估算当前插入 n 个元素时的 bucket 数（粗略）
// 老实现：loadFactor 阈值 6.5，所以 bucket 数 ≈ n / 6.5 向上取到 2 的幂
func bucketsApprox(n int) int {
	if n == 0 {
		return 1
	}
	buckets := 1
	for {
		// 2^B >= n/6.5
		if float64(buckets) >= float64(n)/6.5 {
			return buckets
		}
		buckets *= 2
	}
}

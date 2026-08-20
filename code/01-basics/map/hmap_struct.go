package main

import (
	"fmt"
	"runtime"
)

// DemoHmapStruct 演示 map 内部结构（行为层面）
//
// ============================================================================
// Go 1.24+ 默认用 Swiss Table 实现（不再用 hmap/bmap）
// Go 1.23 及之前用老实现（hmap + bmap + 链表式溢出 bucket）
//
// 老实现（hmap）大致结构（仅供了解，Go 1.25 已不适用）：
//   type hmap struct {
//       count     int               // 元素数（len() 读这个）
//       B         uint8             // log_2(buckets 数)，bucket 数 = 2^B
//       flags     uint8             // iterator / hashWriting 等标志
//       noverflow uint16            // 溢出 bucket 数量
//       hash0     uint32            // 随机哈希种子（防 HashDoS）
//       buckets    unsafe.Pointer   // bucket 数组
//       oldbuckets unsafe.Pointer   // 扩容时的旧 bucket
//       ...
//   }
//   type bmap struct {
//       tophash [8]uint8            // 哈希高 8 位（快速比较）
//       keys    [8]K
//       values  [8]V
//       overflow *bmap              // 溢出 bucket 链
//   }
//
// Swiss Table 实现（Go 1.24+ 默认）：
//   - 开放寻址（不用链地址）
//   - groups of 8 slots
//   - 每个 slot 有 control byte（标记空/满/deleted）
//   - 用 SIMD 加速探测
//
// 面试重点：
//   - 老实现：哈希冲突 → 链表 → 扩容 → 负载因子 6.5
//   - 新实现：开放寻址 + control byte + 性能更好
//   - 两种实现对外接口完全一样
// ============================================================================
func DemoHmapStruct() {
	fmt.Println("=== map 内部结构（行为层面）===")
	fmt.Println()

	// 实验 1：演示 map 容量是 2 的幂
	fmt.Println("【实验 1】bucket 数永远是 2 的幂")
	fmt.Println("  插入数据后, 内部 bucket 数 = 2^B")
	fmt.Println("  B 会自动调整以保持负载因子 (老实现: 6.5)")
	fmt.Println()

	m := make(map[int]int)
	fmt.Printf("  初始 make(map[int]int): len=%d\n", len(m))

	// 插入越来越多数据，观察 map 自动扩容
	for _, n := range []int{1, 10, 100, 1000, 10000} {
		m2 := make(map[int]int)
		for i := 0; i < n; i++ {
			m2[i] = i
		}
		fmt.Printf("  插入 %5d 个: len=%d\n", n, len(m2))
	}
	_ = m
	fmt.Println()

	// 实验 2：演示 map 的并发安全检测（runtime 检查，不是 map 内部加锁）
	fmt.Println("【实验 2】map 并发写检测")
	fmt.Println("  Go runtime 在 map 操作时会检查 flags & hashWriting")
	fmt.Println("  如果有并发写, 触发 fatal error: concurrent map writes")
	fmt.Println("  这是主动检测, 不是 map 内部有锁")
	fmt.Println()

	// 实验 3：Go 版本与 map 实现
	fmt.Println("【实验 3】Go 版本与 map 实现")
	fmt.Printf("  当前 Go 版本: %s\n", runtime.Version())
	fmt.Println()
	fmt.Println("  实现演进:")
	fmt.Println("    Go 1.23 及之前: 老 hmap/bmap（链表式溢出）")
	fmt.Println("    Go 1.24+:        Swiss Table 默认（开放寻址）")
	fmt.Println("    区别: 性能、内存布局不同，但 API 完全一样")
	fmt.Println()

	// 实验 4：行为层面验证关键特性
	fmt.Println("【实验 4】行为层面验证关键特性")

	// key 必须可比较
	type Point struct{ X, Y int }
	m1 := map[Point]string{{1, 2}: "origin"}
	fmt.Printf("  struct 作 key: %v (字段都可比较 ✓)\n", m1)

	// 遍历无序
	order1 := map[string]int{"a": 1, "b": 2, "c": 3}
	var keys1 []string
	for k := range order1 {
		keys1 = append(keys1, k)
	}
	keys2 := []string{}
	for k := range order1 {
		keys2 = append(keys2, k)
	}
	fmt.Printf("  第 1 次遍历: %v\n", keys1)
	fmt.Printf("  第 2 次遍历: %v (顺序可能不同)\n", keys2)
	fmt.Println()

	fmt.Println("📌 关键认知:")
	fmt.Println("   - 内部结构是黑盒（Go 1.24+ Swiss Table，layout 不稳定）")
	fmt.Println("   - 不要依赖 unsafe 窥探（不同版本可能完全不一样）")
	fmt.Println("   - 关注行为：len()、遍历无序、key 可比较、并发安全")
	fmt.Println("   - 面试时讲清特性（哈希冲突 / 扩容 / 并发检测）即可")
}

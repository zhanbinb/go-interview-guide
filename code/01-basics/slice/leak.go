package main

import (
	"fmt"
	"runtime"
)

// DemoLeak 演示大 slice 截取小 slice 导致的内存泄漏
//
// 经典场景：
//   读一个大文件到 buf[:n]，n 很小
//   buf 的 cap 是整个文件大小（比如 1GB）
//   即使只用 buf[:10]，整个 1GB 也常驻内存（GC 不会回收）
//
// 修复：append 到一个新的 slice 复制一份
func DemoLeak() {
	fmt.Println("=== slice 内存泄漏 ===")
	fmt.Println()

	reportMem := func(label string) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("  %s: HeapAlloc=%.2f MB, Sys=%.2f MB\n",
			label, float64(m.HeapAlloc)/1024/1024, float64(m.Sys)/1024/1024)
	}

	// 模拟分配 100MB 数据
	const size = 100 * 1024 * 1024 // 100MB
	reportMem("分配前")
	big := make([]byte, size)
	for i := range big {
		big[i] = byte(i % 256)
	}
	reportMem("分配 100MB 后")
	runtime.GC()
	reportMem("GC 后")
	fmt.Println()

	// 场景 1：错误做法：截取小 slice
	fmt.Println("【场景 1】错误：截取小 slice")
	small1 := big[:10] // small1 共享底层 100MB
	fmt.Printf("  small1: len=%d, cap=%d\n", len(small1), cap(small1))

	big = nil // 看似释放了 big
	runtime.GC()
	reportMem("big=nil + GC 后")
	fmt.Printf("  但 small1 还引用底层数组，100MB 没释放！\n\n")

	// 重新分配
	big = make([]byte, size)
	for i := range big {
		big[i] = byte(i % 256)
	}

	// 场景 2：正确做法 1：用 copy 复制一份
	fmt.Println("【场景 2】正确做法 1：用 copy 复制一份")
	small2 := make([]byte, 10)
	copy(small2, big[:10])
	big = nil
	runtime.GC()
	reportMem("copy + big=nil + GC 后")
	fmt.Printf("  small2: len=%d, cap=%d（独立的底层数组，10MB）\n\n", len(small2), cap(small2))

	// 重新分配
	big = make([]byte, size)
	for i := range big {
		big[i] = byte(i % 256)
	}

	// 场景 3：正确做法 2：append 到新 slice
	fmt.Println("【场景 3】正确做法 2：append 到 nil slice")
	small3 := append([]byte(nil), big[:10]...)
	big = nil
	runtime.GC()
	reportMem("append+big=nil+GC 后")
	fmt.Printf("  small3: len=%d, cap=%d（独立的底层数组）\n\n", len(small3), cap(small3))

	// 场景 4：strings 包也会复制
	fmt.Println("【场景 4】strings 包用 []byte 构造 string 是会复制底层数组的")
	big = make([]byte, size)
	for i := range big {
		big[i] = byte(i % 256)
	}
	_ = string(big[:10]) // strings.Builder 或 string([]byte) 会复制
	big = nil
	runtime.GC()
	reportMem("string(big[:10]) + big=nil + GC 后")
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - 截取 slice 不复制底层数组")
	fmt.Println("   - 即使 len 很小，只要 cap 还在引用，整个底层数组都常驻")
	fmt.Println("   - 但 Go scavenger 会通过 madvise 把页还给 OS，所以 Sys 也不一定高")
	fmt.Println("   - 正确处理：copy() 或 append([]byte{}, src[:n]...)")
	fmt.Println("   - 实际项目：处理大文件/大消息时务必注意")
}

package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// DemoStackGrowth 演示 goroutine 栈的动态增长
//
// ============================================================================
// 关键事实：
//   - goroutine 初始栈仅 2KB（vs OS 线程 8MB，可创建百万级 goroutine）
//   - 栈按需增长（拷贝 + 重定向 stack pointer）
//   - 最大默认 1GB（可通过 GOMAXSTACKSIZE 或 debug.SetMaxStack 调整）
//   - 运行时维护 stack pool 复用退出 goroutine 的栈内存
//
// ============================================================================
func DemoStackGrowth() {
	fmt.Println("=== Goroutine 栈增长演示 ===")
	fmt.Println()

	// ---------- 实验 1：抓 goroutine 自己的栈信息 ----------
	fmt.Println("实验 1：runtime.Stack() 看 goroutine 自己的栈")
	stackCh := make(chan string, 1)
	go func() {
		// 故意占用一点栈空间
		buf := make([]byte, 512)
		_ = buf

		stack := make([]byte, 4096)
		n := runtime.Stack(stack, false)
		stackCh <- string(stack[:n])
	}()
	stackInfo := <-stackCh
	lines := strings.Split(stackInfo, "\n")
	fmt.Println("Goroutine 栈片段（前 6 行）:")
	for i := 0; i < 6 && i < len(lines); i++ {
		fmt.Printf("  %s\n", lines[i])
	}
	fmt.Println()

	// ---------- 实验 2：1000 个 goroutine 的内存占用 ----------
	fmt.Println("实验 2：1000 个 goroutine 的内存开销")
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	fmt.Printf("初始 HeapAlloc: %.2f KB\n", float64(m1.HeapAlloc)/1024)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// 每个 goroutine 用一点栈空间
			buf := make([]byte, 1024)
			_ = buf
			_ = id
		}(i)
	}
	wg.Wait()

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)
	fmt.Printf("1000 个 goroutine 退出后 HeapAlloc: %.2f KB\n", float64(m2.HeapAlloc)/1024)
	fmt.Printf("  增加: %.2f KB（每个 goroutine 约 %.2f KB）\n",
		float64(m2.HeapAlloc-m1.HeapAlloc)/1024,
		float64(m2.HeapAlloc-m1.HeapAlloc)/1024/1000)
	fmt.Println()

	// ---------- 实验 3：栈动态增长（深度递归） ----------
	fmt.Println("实验 3：栈动态增长（深度递归）")
	depth := 0
	var deepRecur func(n int) int
	deepRecur = func(n int) int {
		depth++
		buf := make([]byte, 256) // 故意占栈空间，逼编译器无法优化
		_ = buf
		if n == 0 {
			return 0
		}
		return deepRecur(n-1) + 1
	}
	result := deepRecur(100)
	fmt.Printf("递归深度: %d（每次栈至少扩 256 字节）\n", result)
	fmt.Printf("  实际调用次数: %d\n", depth)
	fmt.Println()

	// ---------- 实验 4：触发 stack growth panic ----------
	fmt.Println("实验 4：栈大小限制（如果取消注释会 panic: stack overflow）")
	fmt.Println("  默认最大 1GB（debug.SetMaxStack 可调）")
	fmt.Println("  var deepFunc func(n int) int")
	fmt.Println("  deepFunc = func(n int) int {")
	fmt.Println("      buf := make([]byte, 1024*1024) // 每次 1MB")
	fmt.Println("      _ = buf")
	fmt.Println("      if n == 0 { return 0 }")
	fmt.Println("      return deepFunc(n-1) + 1")
	fmt.Println("  }")
	fmt.Println("  // deepFunc(2000) 会 panic: runtime: goroutine stack exceeds 1000000000-byte limit")

	fmt.Println()
	fmt.Println("📌 关键 takeaway:")
	fmt.Println("   - 初始栈 2KB，可以轻松创建百万级 goroutine")
	fmt.Println("   - 栈按需动态增长，编译器插入 stack growth probe 检查")
	fmt.Println("   - 默认上限 1GB（避免无限递归吃光内存）")
}

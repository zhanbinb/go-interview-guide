package main

import (
	"fmt"
	"runtime"
	"time"
)

// DemoLeak 演示内存泄漏的 4 种常见场景
//
// ============================================================================
// 什么是内存泄漏？
//   内存被分配后，永远无法被 GC 回收
//
// 4 种常见场景：
//   1. goroutine 泄漏（永不退出）
//   2. 全局 slice/map 持续增长（没删除）
//   3. 循环引用（Go GC 能处理一般循环，但 channel 闭包等仍可能）
//   4. 未关闭的 timer / cgo 内存
//
// 排查工具：
//   - pprof: go tool pprof http://host/debug/pprof/heap
//   - go test -memprofile
//   - runtime.MemStats: HeapAlloc / HeapInuse / HeapReleased
// ============================================================================
func DemoLeak() {
	fmt.Println("=== 内存泄漏 4 种场景 ===")
	fmt.Println()

	report := func(label string) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("  %s: HeapAlloc=%.2f MB, NumGoroutine=%d\n",
			label, float64(m.HeapAlloc)/1024/1024, runtime.NumGoroutine())
	}

	report("初始")

	// 泄漏 1：goroutine 永久阻塞
	fmt.Println("\n【泄漏 1】goroutine 永久阻塞（用 select{} 代替具体场景）")
	for i := 0; i < 1000; i++ {
		go func() {
			ch := make(chan int)
			_ = ch
			select {} // 永久阻塞
		}()
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 1 后")
	runtime.GC()
	report("GC 后")

	// 泄漏 2：全局 map 持续增长
	fmt.Println("\n【泄漏 2】全局 map 持续增长（忘记删除）")
	globalMap := make(map[int][]byte)
	for i := 0; i < 10000; i++ {
		globalMap[i] = make([]byte, 1024) // 1KB/个
	}
	report("泄漏 2 后")
	runtime.GC()
	report("GC 后 (map 仍持有引用, 不会被回收)")

	fmt.Println()
	fmt.Println("📌 排查命令:")
	fmt.Println("   # 1. runtime stats")
	fmt.Println("   go tool pprof http://localhost:6060/debug/pprof/heap")
	fmt.Println("   # 2. 在测试里检查")
	fmt.Println("   go test -memprofile=mem.out")
	fmt.Println("   go tool pprof mem.out")
	fmt.Println()
	fmt.Println("📌 实战经验:")
	fmt.Println("   - goroutine 泄漏最常见：用 runtime.NumGoroutine() 监控")
	fmt.Println("   - 全局容器增长：用 weak reference 或定期清理")
	fmt.Println("   - cgo 内存：runtime 不能追踪，必须手动 free")
	fmt.Println("   - 排查黄金法则：先 pprof heap top，再看具体分配代码")
}

package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// DemoLeak 演示 channel 资源泄漏的 4 种常见场景
//
// ============================================================================
// 什么是 channel 泄漏？
//   goroutine 因为 channel 操作永久阻塞，无法退出 → goroutine 泄漏 → 内存泄漏
//
// 4 种典型场景：
//   1. 发送方阻塞：goroutine 发送数据到无缓冲 channel，但无 receiver
//   2. 接收方阻塞：goroutine 从 channel 接收，但 sender 早早 return
//   3. buffered channel 持续生产，无人消费
//   4. goroutine 等永远不会被关闭/取消的 channel
// ============================================================================
func DemoLeak() {
	fmt.Println("=== Channel 泄漏演示 ===")
	fmt.Println()

	report := func(label string) {
		fmt.Printf("  %s: NumGoroutine = %d\n", label, runtime.NumGoroutine())
	}
	report("初始")

	// ---------- 泄漏 1：发送方阻塞 ----------
	fmt.Println("\n【泄漏 1】向无缓冲 channel 发送，但无 receiver")
	unbuf := make(chan int)
	for i := 0; i < 3; i++ {
		go func(id int) {
			unbuf <- id // 永久阻塞
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 1 后")

	// ---------- 泄漏 2：接收方阻塞 ----------
	fmt.Println("\n【泄漏 2】从 channel 接收，但 sender 早早 return")
	oneShot := make(chan int)
	go func() {
		// sender 只发一次就退出，但 channel 不关闭
		oneShot <- 42
		fmt.Println("    sender 已 return")
	}()
	go func() {
		v := <-oneShot
		fmt.Printf("    receiver 收到: %d\n", v)
		// 之后 receiver 等下一个值，但永远等不到
		v2 := <-oneShot
		fmt.Printf("    receiver 又收到: %d (不会发生)\n", v2)
	}()
	time.Sleep(50 * time.Millisecond)
	report("泄漏 2 后")

	// ---------- 泄漏 3：buffered channel 持续生产，无人消费 ----------
	fmt.Println("\n【泄漏 3】buffered channel 持续生产，无人消费")
	buf := make(chan int, 100) // 缓冲 100
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				buf <- id*10 + j // 没人消费
			}
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 3 后")
	fmt.Printf("    buf 现状: len=%d, cap=%d\n", len(buf), cap(buf))

	// ---------- 泄漏 4：永远不取消的 context ----------
	fmt.Println("\n【泄漏 4】等永远不会 cancel 的 context")
	for i := 0; i < 3; i++ {
		go func(id int) {
			ctx, _ := context.WithCancel(context.Background()) // 故意不 cancel
			<-ctx.Done()                                       // 永久阻塞
			_ = id
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 4 后")

	fmt.Println()
	fmt.Println("⚠️  准备触发 fatal error: all goroutines are asleep...")
	fmt.Println("   （主 goroutine 阻塞后，runtime 检测所有 goroutine 都卡住）")

	time.Sleep(500 * time.Millisecond)
	select {} // 主 goroutine 也阻塞
}

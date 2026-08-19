package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// DemoLeak 演示 goroutine 泄漏的 4 种常见模式
//
// ============================================================================
// 什么是泄漏？
//   goroutine 永不退出，且无法被 GC 回收 → 内存泄漏
//
// 4 种最常见的泄漏模式：
//   1. channel 收发永远没人对端
//   2. context 没 cancel，子 goroutine 永久等 ctx.Done()
//   3. WaitGroup.Add 永远没 Done
//   4. 死循环 + 没有退出条件
//
// 怎么检测？
//   - runtime.NumGoroutine() 看趋势
//   - go test + uber-go/goleak 库
//   - pprof goroutine profile：go tool pprof http://.../debug/pprof/goroutine
// ============================================================================
func DemoLeak() {
	fmt.Println("=== Goroutine 泄漏演示 ===")
	fmt.Println()

	report := func(label string) {
		fmt.Printf("  %s: NumGoroutine = %d\n", label, runtime.NumGoroutine())
	}

	report("初始")

	// ---------- 泄漏 1：channel 没人接收 ----------
	fmt.Println("\n【泄漏 1】向无缓冲 channel 发送数据，但没人接收")
	leakCh := make(chan int)
	for i := 0; i < 3; i++ {
		go func(id int) {
			fmt.Printf("  goroutine %d 准备发送\n", id)
			leakCh <- id // 永久阻塞
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 1 后")

	// ---------- 泄漏 2：context 没取消 ----------
	fmt.Println("\n【泄漏 2】context 没 cancel，子 goroutine 等 ctx.Done()")
	for i := 0; i < 3; i++ {
		go func(id int) {
			ctx, _ := context.WithCancel(context.Background())
			_ = ctx
			<-make(chan struct{}) // 永久阻塞
			_ = id
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 2 后")

	// ---------- 泄漏 3：WaitGroup Add 没 Done ----------
	fmt.Println("\n【泄漏 3】WaitGroup.Add(1) 但 goroutine 永远不 Done")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			// 永远不会调用 wg.Done()
			select {} // 死循环
			_ = id
		}(i)
	}
	// 注意：这里故意不调用 wg.Wait()
	_ = wg
	time.Sleep(100 * time.Millisecond)
	report("泄漏 3 后")

	// ---------- 泄漏 4：死循环无退出条件 ----------
	fmt.Println("\n【泄漏 4】死循环 + 没退出条件")
	for i := 0; i < 3; i++ {
		go func(id int) {
			for {
				// 没有 break / return / ctx.Done 检查
			}
			_ = id
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 4 后")

	fmt.Println()
	fmt.Printf("📌 启动前 +12 个泄漏 goroutine，NumGoroutine 从初始涨到当前\n")
	fmt.Println()
	fmt.Println("⚠️  准备触发 fatal error: all goroutines are asleep...")
	fmt.Println("   （主 goroutine 也会阻塞，让 runtime 检测到所有 goroutine 都卡住）")
	fmt.Println()

	time.Sleep(500 * time.Millisecond)
	select {} // 故意阻塞，触发 runtime 检测
}

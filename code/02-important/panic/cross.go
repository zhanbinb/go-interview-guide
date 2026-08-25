package main

import (
	"fmt"
	"time"
)

// DemoCross 演示跨 goroutine panic 隔离
//
// ============================================================================
// 关键认知：
//   - panic 在哪个 goroutine 触发，就影响那个 goroutine
//   - 子 goroutine panic 不会自动传播到主 goroutine
//   - 主 goroutine panic（无 recover）→ 整个进程退出
//   - 子 goroutine panic（无 recover）→ 该 goroutine 崩溃，进程也退出
//
// 所以：每个 goroutine 都需要独立的 defer recover
// 推荐用 errgroup 库处理 goroutine 错误
// ============================================================================
func DemoCross() {
	fmt.Println("=== 跨 goroutine panic 隔离 ===")
	fmt.Println()

	// 实验 1：子 goroutine panic，主 goroutine 不影响
	fmt.Println("【实验 1】子 goroutine panic（无 recover）")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  子 goroutine recover: %v\\n", r)
			}
		}()
		panic("💥 子 goroutine panic")
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("  主 goroutine 没受影响 ✨")
	fmt.Println()

	// 实验 2：子 goroutine 没 recover，演示 goroutine 死亡
	fmt.Println("【实验 2】多个 goroutine，其中一个 panic")
	for i := 0; i < 3; i++ {
		go func(id int) {
			// 不 recover（演示用，生产代码必须 recover）
			if id == 2 {
				panic(fmt.Sprintf("💥 worker %d panic", id))
			}
			fmt.Printf("    worker %d 完成\\n", id)
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println("  panic 的 goroutine 死亡，其他不受影响（但 panic 输出会刷屏）")
	fmt.Println()

	// 实验 3：主 goroutine panic → 整个进程退出
	fmt.Println("【实验 3】主 goroutine panic（无 recover）→ 整个进程退出")
	fmt.Println("  defer func() { recover() }()")
	fmt.Println("  panic(\"主 goroutine\")  // 没 defer recover，进程退出")
	fmt.Println()

	// 实验 4：所有 goroutine 都需要独立 recover
	fmt.Println("【实验 4】实践模式：每个 goroutine 都 recover")
	for i := 0; i < 3; i++ {
		go worker(i)
	}
	time.Sleep(100 * time.Millisecond)
	fmt.Println()

	fmt.Println("📌 实战模式:")
	fmt.Println("   1. 每个 goroutine 都用 defer recover")
	fmt.Println("   2. 推荐用 errgroup.WithContext (golang.org/x/sync/errgroup)")
	fmt.Println("   3. 业务 goroutine: recover + 打印日志 + 上报 metric")
	fmt.Println("   4. 关键路径 goroutine: recover + re-panic + 终止进程")
}

// worker 安全的 worker 模板
func worker(id int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    worker %d recover: %v\\n", id, r)
		}
	}()

	// 模拟工作
	fmt.Printf("    worker %d 工作中...\\n", id)
	time.Sleep(20 * time.Millisecond)

	// 模拟 panic
	if id == 2 {
		panic(fmt.Sprintf("worker %d 业务错误", id))
	}
	fmt.Printf("    worker %d 完成\\n", id)
}

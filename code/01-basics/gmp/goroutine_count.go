package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// DemoGoroutineCount 演示 NumGoroutine 的变化
//
// 关键结论：
//   - NumGoroutine() 返回当前存在的 goroutine 数量
//   - 包括：主 goroutine、运行中、阻塞中、即将退出未回收的所有 G
//   - Go 1.22+ 循环变量语义变了（每轮独立），不需要临时变量技巧
//   - goroutine 泄漏 = 永不退出的 goroutine（如下面的 select{} 场景）
//   - 当所有 goroutine 都阻塞时，runtime 会报 "fatal error: all goroutines are asleep"
func DemoGoroutineCount() {
	fmt.Printf("初始 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println()

	// 场景 1：启动 1000 个会立即退出的 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_ = id
		}(i)
	}
	wg.Wait()

	// 启动后立即退出的 G 会被运行时回收
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("启动 1000 个立即退出的 goroutine 后: %d（已回收）\n", runtime.NumGoroutine())
	fmt.Println()

	// 场景 2：启动 1000 个永久阻塞的 goroutine（演示泄漏 + runtime 检测）
	fmt.Println(">> 创建 1000 个永久阻塞 goroutine（select{}）...")
	for i := 0; i < 1000; i++ {
		go func() {
			select {} // 永久阻塞
		}()
	}

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("创建后 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println()

	// 故意让主 goroutine 也阻塞 1 秒，给用户时间看输出
	// 之后 runtime 会检测到 "all goroutines are asleep"，自动报错退出
	// （这本身就是 goroutine 泄漏的最好证据）
	fmt.Println("⚠️  准备触发 goroutine 泄漏检测：所有 goroutine 都阻塞时 runtime 会 fatal error")
	fmt.Println("📌 实际项目中的 goroutine 泄漏常见场景:")
	fmt.Println("   - channel 收发没人接收/发送")
	fmt.Println("   - context 没取消，子 goroutine 永远不退出")
	fmt.Println("   - sync.WaitGroup.Add 在 goroutine 内调用，Done 没调用到")
	fmt.Println("   - 死循环里没有退出条件")
	fmt.Println()
	fmt.Println("等待 1 秒后会触发 fatal error（这是预期行为，不是 bug）...")

	time.Sleep(1 * time.Second)
	select {} // 让主 goroutine 也阻塞，触发 runtime 检测
}

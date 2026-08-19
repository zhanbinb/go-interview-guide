package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// DemoWorkStealing 演示 Work Stealing 机制
//
// 关键结论：
//   - 每个 P 有自己的本地 runqueue（LRQ），无锁，最快
//   - P 的 LRQ 满了，新 G 才入全局 runqueue（GRQ）
//   - 当某个 P 的 LRQ 空了，它会去别的 P 偷一半 G（work stealing）
//   - 每 61 次调度会检查一次 GRQ，防止全局队列被饿死
//
// 面试问法："调度器的设计策略？"
func DemoWorkStealing() {
	fmt.Println("设置 GOMAXPROCS=2，制造负载不均，便于观察 work stealing")
	runtime.GOMAXPROCS(2)
	fmt.Println()

	var wg sync.WaitGroup
	start := time.Now()

	// 100 个 CPU 密集任务
	// 初始分配可能集中到一个 P 的 LRQ，另一个 P 应该有 stealing 发生
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sum := 0
			for j := 0; j < 1e7; j++ {
				sum += j
			}
			_ = sum
			_ = id
		}(i)
	}
	wg.Wait()

	fmt.Printf("100 个 CPU 密集任务完成，用时 %v\n", time.Since(start))
	fmt.Println()
	fmt.Println("📌 用 GODEBUG=schedtrace=1000,scheddetail=1 跑本 demo，观察:")
	fmt.Println("   - P0/P1 的 LRQ 长度变化")
	fmt.Println("   - 空闲 P 从其他 P 偷任务的时刻")
	fmt.Println()
	fmt.Println("📌 进阶：还可以观察 gomaxprocs=1 时的串行行为作为对比")
}

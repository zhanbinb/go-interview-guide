package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoMutex 演示 Mutex 两种模式
//
// ============================================================================
// sync.Mutex (Go 1.9+) 两种模式：
//
// 1. normal 模式（默认）：
//    - 新来的 goroutine 和等待者公平竞争
//    - 等待者可被新来者"抢"（CAS 自旋几次）
//    - 性能高，但可能让等待者饥饿
//
// 2. starvation 模式：
//    - goroutine 等待 >1ms 后进入此模式
//    - 严格 FIFO（不再让新来者抢）
//    - 保证公平，但性能略低
// ============================================================================
func DemoMutex() {
	fmt.Println("=== Mutex 两种模式 ===")
	fmt.Println()

	// 实验 1：基本用法
	fmt.Println("【实验 1】基本用法")
	var mu sync.Mutex
	counter := 0
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  1000 个 goroutine 加锁后 counter = %d (期望 1000)\n\n", counter)

	// 实验 2：观察锁竞争（高并发争抢）
	fmt.Println("【实验 2】高并发争抢（看自旋效果）")
	var mu2 sync.Mutex
	var contended int
	start := time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mu2.Lock()
				contended++
				mu2.Unlock()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("  10000 次加锁: %v\n", elapsed)
	fmt.Printf("  平均每次锁操作: %v\n\n", elapsed/10000)

	// 实验 3：TryLock (Go 1.18+)
	fmt.Println("【实验 3】TryLock (非阻塞尝试)")
	var mu3 sync.Mutex
	if mu3.TryLock() {
		fmt.Println("  TryLock 成功")
		mu3.Unlock()
	}
	mu3.Lock()
	if !mu3.TryLock() {
		fmt.Println("  锁已被持, TryLock 失败")
	}
	mu3.Unlock()
	fmt.Println()

	fmt.Println("📌 Mutex 要点:")
	fmt.Println("   - Go 1.9+ 两种模式自动切换（用户无感）")
	fmt.Println("   - normal 模式：性能优先，可能饥饿")
	fmt.Println("   - starvation 模式：1ms 后切换，保证公平")
	fmt.Println("   - 不要递归加锁（死锁）")
}

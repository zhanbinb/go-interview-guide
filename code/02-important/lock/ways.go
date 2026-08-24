package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// DemoWays 对比 5 种同步共享变量的方式
//
// ============================================================================
// 同步共享变量，Go 给你的选择：
//
//   1. sync.Mutex         通用，写独占
//   2. sync.RWMutex       读多写少
//   3. atomic             简单类型（性能最好）
//   4. channel            通过通信共享内存（推荐）
//   5. sync.Once          一次性初始化
//
// 选择指南：
//   - 简单状态 → atomic
//   - 复杂结构 → Mutex/RWMutex
//   - 涉及 goroutine 协调 → channel
//   - 单例/初始化 → sync.Once
// ============================================================================
func DemoWays() {
	fmt.Println("=== 5 种同步方式对比 ===")
	fmt.Println()

	// 1. sync.Mutex
	fmt.Println("【1】sync.Mutex")
	var mu sync.Mutex
	var muCount int
	mu.Lock()
	muCount++
	mu.Unlock()
	fmt.Printf("  muCount = %d (写独占)\n", muCount)

	// 2. sync.RWMutex
	fmt.Println("【2】sync.RWMutex")
	var rw sync.RWMutex
	rw.RLock()
	rw.RUnlock()
	fmt.Println("  RLock/RUnlock 读并发 (实现略)")

	// 3. atomic
	fmt.Println("【3】atomic")
	var atomicCnt int64
	atomic.AddInt64(&atomicCnt, 1)
	fmt.Printf("  atomicCnt = %d (无锁)\n", atomic.LoadInt64(&atomicCnt))

	// 4. channel
	fmt.Println("【4】channel")
	ch := make(chan int, 1)
	ch <- 42
	v := <-ch
	fmt.Printf("  ch = %d (通过通信共享内存)\n", v)

	// 5. sync.Once
	fmt.Println("【5】sync.Once")
	var once sync.Once
	init := func() {
		fmt.Println("    只执行一次 (即使多次调用)")
	}
	for i := 0; i < 3; i++ {
		once.Do(init)
	}
	fmt.Println()

	fmt.Println("📌 选择决策树:")
	fmt.Println("  简单类型状态/计数器 → atomic")
	fmt.Println("  复杂结构读写互斥 → Mutex / RWMutex")
	fmt.Println("  goroutine 间数据传递 → channel")
	fmt.Println("  延迟初始化 / 单例 → sync.Once")
	fmt.Println("  读多写少 + 复杂结构 → RWMutex")
	fmt.Println()
	fmt.Println("📌 Go 哲学:")
	fmt.Println("  \"不要通过共享内存来通信,\"")
	fmt.Println("   \"通过通信来共享内存\"")
	fmt.Println("  (实际: 看场景，都用)")
}

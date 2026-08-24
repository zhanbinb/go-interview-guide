package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoRWMutex 演示 RWMutex 与 Mutex 对比
//
// ============================================================================
// sync.RWMutex vs sync.Mutex：
//
//   RWMutex：读锁可并发，写锁独占
//   Mutex：  任何锁都是独占
//
//   - 读多写少：用 RWMutex 性能更好
//   - 写多读多：用 Mutex 更简单
//
// 注意：不能递归加读锁/写锁（会死锁）
// ============================================================================
func DemoRWMutex() {
	fmt.Println("=== RWMutex vs Mutex ===")
	fmt.Println()

	// 实验 1：RWMutex 读并发
	fmt.Println("【实验 1】RWMutex：100 个读并发")
	rw := sync.RWMutex{}
	data := 0
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw.RLock()
			_ = data
			time.Sleep(10 * time.Millisecond) // 模拟读耗时
			rw.RUnlock()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("  100 个并发读: %v (因为读可并发，实际几乎同时完成)\n\n", elapsed)

	// 实验 2：Mutex 读并发（对比）
	fmt.Println("【实验 2】Mutex：100 个读串行")
	mu := sync.Mutex{}
	start = time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			_ = data
			time.Sleep(10 * time.Millisecond)
			mu.Unlock()
		}()
	}
	wg.Wait()
	elapsed = time.Since(start)
	fmt.Printf("  100 个串行读: %v (因为 Mutex 独占，串行)\n\n", elapsed)

	// 实验 3：写锁独占
	fmt.Println("【实验 3】写锁独占：写时不能读")
	rw2 := sync.RWMutex{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		rw2.Lock()
		fmt.Println("  写锁获取，开始写 (300ms)")
		time.Sleep(300 * time.Millisecond)
		rw2.Unlock()
	}()
	time.Sleep(50 * time.Millisecond) // 等写锁获取

	// 这时候读应该阻塞
	readCh := make(chan struct{})
	go func() {
		rw2.RLock()
		close(readCh)
		rw2.RUnlock()
	}()
	select {
	case <-readCh:
		fmt.Println("  ❌ 读立即成功（不应该）")
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  ✅ 读被阻塞（写锁独占期间）")
	}
	wg.Wait()
	fmt.Println()

	fmt.Println("📌 选择建议:")
	fmt.Println("   - 读占 90%+：用 RWMutex 收益大")
	fmt.Println("   - 写占 50%+：用 Mutex 反而更快（避免 RWMutex 内部复杂度）")
	fmt.Println("   - 简单场景：默认 Mutex")
}

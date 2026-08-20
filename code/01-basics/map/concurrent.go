package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoConcurrent 演示 map 并发访问的 fatal error 和 3 种解决方案
//
// 原生 map 不是并发安全的：
//   并发写（含一个写 + 一个读）会触发 fatal error: concurrent map writes
//   这是运行时主动检测的，触发后整个程序崩溃
//
// 3 种解决方案：
//   1. sync.Mutex + map       通用，写频繁也可以
//   2. sync.RWMutex + map     读多写少
//   3. sync.Map               key 集合稳定 + 读多写少（专用优化）
func DemoConcurrent() {
	fmt.Println("=== map 并发安全 ===")
	fmt.Println()

	// ---------- 1. 演示 fatal error ----------
	fmt.Println("【实验 1】原生 map 并发写 → fatal error")
	fmt.Println("  下面的代码如果取消注释会让程序崩溃:")
	fmt.Println("    m := map[int]int{}")
	fmt.Println("    go func() { for i := 0; ; i++ { m[i] = i } }()")
	fmt.Println("    go func() { for i := 0; ; i++ { m[i] = i } }()")
	fmt.Println("  → fatal error: concurrent map writes")
	fmt.Println()

	// ---------- 2. 方案 1：sync.Mutex + map ----------
	fmt.Println("【实验 2】方案 1：sync.Mutex + map")
	type MutexMap struct {
		mu sync.Mutex
		m  map[string]int
	}
	mm := &MutexMap{m: make(map[string]int)}
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mm.mu.Lock()
			mm.m[fmt.Sprintf("k%d", id)] = id
			mm.mu.Unlock()
		}(i)
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mm.mu.Lock()
			_ = mm.m[fmt.Sprintf("k%d", id)]
			mm.mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Printf("  100 写 + 100 读 用时 %v, len=%d (并发安全)\n", time.Since(start), len(mm.m))
	fmt.Println()

	// ---------- 3. 方案 2：sync.RWMutex + map ----------
	fmt.Println("【实验 3】方案 2：sync.RWMutex + map")
	type RWMap struct {
		mu sync.RWMutex
		m  map[string]int
	}
	rwm := &RWMap{m: make(map[string]int)}
	start = time.Now()
	// 100 写 + 1000 读（读多写少）
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rwm.mu.Lock()
			rwm.m[fmt.Sprintf("k%d", id)] = id
			rwm.mu.Unlock()
		}(i)
	}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rwm.mu.RLock()
			_ = rwm.m["k50"]
			rwm.mu.RUnlock()
		}()
	}
	wg.Wait()
	fmt.Printf("  100 写 + 1000 读 用时 %v, len=%d (并发安全)\n", time.Since(start), len(rwm.m))
	fmt.Println()

	// ---------- 4. 方案 3：sync.Map ----------
	fmt.Println("【实验 4】方案 3：sync.Map (内置并发安全)")
	var sm sync.Map
	start = time.Now()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sm.Store(fmt.Sprintf("k%d", id), id)
		}(i)
	}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = sm.Load("k50")
		}()
	}
	wg.Wait()
	count := 0
	sm.Range(func(_, _ any) bool { count++; return true })
	fmt.Printf("  100 写 + 1000 读 用时 %v, count=%d (并发安全)\n", time.Since(start), count)
	fmt.Println()

	fmt.Println("📌 方案选择:")
	fmt.Println("   - sync.Mutex + map:     通用，写多读多都行")
	fmt.Println("   - sync.RWMutex + map:   读多写少，性能更好")
	fmt.Println("   - sync.Map:             key 集合稳定 + 读极多写少（专用）")
	fmt.Println("   - sync.Map 不是万能的: key 不稳定时 dirty map 升级代价高")
}

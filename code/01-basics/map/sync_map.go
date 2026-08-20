package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoSyncMap 对比 sync.Map 和 RWMutex + map
//
// sync.Map 内部结构（实现两层 map）：
//   - read map（只读，无锁）：atomic.Value 保护
//   - dirty map（写入）：需要 mu 锁
//   - 读时优先查 read，没找到才查 dirty
//   - 写时只写 dirty
//   - 当 dirty 被"提升"为 read 时，需要双倍内存
//
// 适用场景：
//   ✅ key 集合稳定（写少 + key 不变）
//   ✅ 读极多写少（10:1 以上）
//   ❌ key 集合频繁变化（dirty map 升级成本高）
func DemoSyncMap() {
	fmt.Println("=== sync.Map vs RWMutex + map ===")
	fmt.Println()

	// 实验 1：读极多写少场景（sync.Map 应该更快）
	fmt.Println("【实验 1】读极多写少 (1000:1) - sync.Map 优势场景")
	benchmark := func(name string, fn func()) {
		start := time.Now()
		fn()
		fmt.Printf("  %s: %v\n", name, time.Since(start))
	}

	var sm sync.Map
	rwm := struct {
		sync.RWMutex
		m map[int]int
	}{m: make(map[int]int)}

	benchmark("sync.Map (1000读 + 1写)", func() {
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, _ = sm.Load(id % 10)
			}(i)
		}
		sm.Store(0, 999)
		wg.Wait()
	})

	benchmark("RWMutex+map (1000读 + 1写)", func() {
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				rwm.RLock()
				_ = rwm.m[id%10]
				rwm.RUnlock()
			}(i)
		}
		rwm.Lock()
		rwm.m[0] = 999
		rwm.Unlock()
		wg.Wait()
	})
	fmt.Println()

	// 实验 2：key 频繁变化场景（sync.Map 劣势）
	fmt.Println("【实验 2】key 频繁变化 (100 写 + 100 读, 每次 key 不同)")
	benchmark("sync.Map (key 全不同)", func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				sm.Store(id, id*10)
			}(i)
		}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, _ = sm.Load(id)
			}(i)
		}
		wg.Wait()
	})

	benchmark("Mutex+map (key 全不同)", func() {
		var wg sync.WaitGroup
		var mu sync.Mutex
		m := map[int]int{}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.Lock()
				m[id] = id * 10
				mu.Unlock()
			}(i)
		}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				mu.Lock()
				_ = m[id]
				mu.Unlock()
			}(i)
		}
		wg.Wait()
	})
	fmt.Println()

	// 实验 3：sync.Map API 对比
	fmt.Println("【实验 3】sync.Map API")
	sm.Store("k1", 1)
	sm.Store("k2", 2)
	if v, ok := sm.Load("k1"); ok {
		fmt.Printf("  Load(\"k1\"): v=%d, ok=%v\n", v, ok)
	}
	sm.Range(func(k, v any) bool {
		fmt.Printf("  Range: k=%v, v=%v\n", k, v)
		return true
	})
	sm.Delete("k2")
	if _, ok := sm.Load("k2"); !ok {
		fmt.Printf("  Delete(\"k2\") 后 Load: ok=%v\n", ok)
	}
	fmt.Println()

	fmt.Println("📌 sync.Map 使用建议:")
	fmt.Println("   - ✅ 适用: key 稳定 + 读极多写少（cache、配置中心）")
	fmt.Println("   - ❌ 不适用: key 集合频繁变化（普通业务场景）")
	fmt.Println("   - ❌ 不适用: 类型不一致（需要 any，类型断言成本）")
	fmt.Println("   - 通用场景: sync.RWMutex + map 更稳")
}

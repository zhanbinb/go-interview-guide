package main

import (
	"fmt"
	
	"runtime"
	"runtime/debug"
)

// DemoTrigger 演示 GC 触发时机
//
// GC 会在以下任一条件满足时触发：
//   1. 堆内存相比上次 GC 翻了 GOGC 倍（默认 100，即翻倍）
//   2. 2 分钟内没 GC（forcegc 守护协程）
//   3. 手动调用 runtime.GC()
//   4. Go 1.21+: 达到 GOMEMLIMIT 软限制
//
// 面试问法："GC 什么时候触发？"
func DemoTrigger() {
	fmt.Println("=== GC 触发时机 ===")

	// 实验 1：手动触发 GC
	fmt.Println("【实验 1】手动 runtime.GC()")
	var m1, m2 runtime.MemStats
	runtime.GC() // 先 GC 一次，得到 baseline
	runtime.ReadMemStats(&m1)
	fmt.Printf("  GC 后 NumGC=%d, HeapAlloc=%.2f MB\n",
		m1.NumGC, float64(m1.HeapAlloc)/1024/1024)

	// 分配一些内存
	for i := 0; i < 10000; i++ {
		_ = make([]byte, 1024)
	}
	runtime.ReadMemStats(&m2)
	fmt.Printf("  分配后 NumGC=%d, HeapAlloc=%.2f MB\n",
		m2.NumGC, float64(m2.HeapAlloc)/1024/1024)
	fmt.Println("  （可能自动触发了 GC，因为堆增长超过 GOGC 阈值）")
	fmt.Println()

	// 实验 2：查看 GC 配置
	fmt.Println("【实验 2】查看 GC 配置")
	gogc := debug.SetGCPercent(-1) // 读当前值（不改）
	debug.SetGCPercent(gogc)
	fmt.Printf("  GOGC=%d (堆翻倍阈值，100 表示翻倍触发)\n", gogc)
	// GOMEMLIMIT 是进程启动时设置的，只能读不能改
	// 没有运行时 API，只能从环境变量推算
	fmt.Println("  GOMEMLIMIT=启动时设置（用 GOMEMLIMIT=100MiB ./demo 查看）")
	fmt.Println()

	// 实验 3：自动 GC 演示
	fmt.Println("【实验 3】触发自动 GC")
	fmt.Println("  分配大量内存，看 NumGC 自动增加:")
	start := runtime.NumGoroutine()
	_ = start
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 100000; i++ {
		_ = make([]byte, 1024)
	}
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	fmt.Printf("  分配 100MB 后 NumGC: %d → %d\n", before.NumGC, after.NumGC)
	fmt.Printf("  GC 次数增加了 %d 次（自动触发的）\n", after.NumGC-before.NumGC)
}

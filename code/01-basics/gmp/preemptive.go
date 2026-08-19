package main

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// DemoPreemptive 演示 Go 1.14+ 抢占式调度
//
// 关键结论：
//   - Go 1.14 之前：协作式调度（依赖函数调用触发，安全点）
//     一个死循环 goroutine 会一直占着 M，其他 goroutine 永远跑不到
//   - Go 1.14+  ：基于信号的抢占式调度
//     sysmon 线程（独立于 P）检测到 G 占用 > 10ms，向 M 发送 SIGURG
//     M 收到信号后，在安全点插入抢占标记，让 G 调度出去
//
// 面试问法："抢占式调度是如何抢占的？"
func DemoPreemptive() {
	fmt.Printf("Go runtime version: %s\n", runtime.Version())
	fmt.Println()

	// 场景：1 个死循环 goroutine + 1 个定时器 goroutine
	// Go 1.14+ 环境下，后者的定时器到点后必然会被调度
	var counter int64
	done := make(chan struct{})

	// 死循环 goroutine（无函数调用，纯算术）
	go func() {
		for {
			atomic.AddInt64(&counter, 1)
		}
		//fmt.Println("死循环 goroutine 结束")
	}()

	// 另一个 goroutine，应该在 ~500ms 后被调度到
	go func() {
		time.Sleep(500 * time.Millisecond)
		close(done)
		fmt.Println("定时器 goroutine 结束")
	}()

	start := time.Now()
	<-done
	elapsed := time.Since(start)

	fmt.Printf("另一个 goroutine 实际被调度的耗时: %v\n", elapsed)
	fmt.Printf("期间死循环执行次数: %d\n", atomic.LoadInt64(&counter))
	fmt.Printf("平均执行速率: %.0f 次/ms\n", float64(counter)/float64(elapsed.Milliseconds()))
	fmt.Println()

	if elapsed > 5*time.Second {
		fmt.Println("⚠️  等待超过 5s，抢占失效！你的 Go 版本应该是 <1.14")
	} else {
		fmt.Println("✅  ~500ms 说明抢占式调度正常工作（Go 1.14+）")
		fmt.Println("   原理：sysmon 线程每 10ms 检测一次，向占用的 M 发送 SIGURG")
		fmt.Println("   M 在安全点（函数调用、栈增长等）检测到抢占标记，让出 CPU")
	}
}

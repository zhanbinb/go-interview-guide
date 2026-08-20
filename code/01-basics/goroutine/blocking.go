package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DemoBlocking 演示 goroutine 的各种阻塞场景
//
// ============================================================================
// 关键认知：
//
//	"goroutine 阻塞" ≠ "M 阻塞" ≠ "程序阻塞"
//
// 阻塞分类：
//
//  1. 用户态阻塞（runtime 调度器处理）：
//     - channel 收发（无缓冲 / 缓冲空 / 缓冲满）
//     - sync.Mutex / RWMutex
//     - select 等 channel 就绪
//     - WaitGroup.Wait
//     → 阻塞时 M 让出 P，runtime 调度别的 G 执行
//
//  2. 内核态阻塞（syscall）：
//     - 文件 I/O（os.Read 等）
//     - time.Sleep
//     - C 系统调用
//     → M 进入内核，新 M 接管 P；netpoller 处理网络 I/O
//
//  3. 死循环（Go 1.14 之前会卡死所有 G，Go 1.14+ 被抢占）
//
// ============================================================================
func DemoBlocking() {
	fmt.Println("=== Goroutine 阻塞场景演示 ===")
	fmt.Println()
	fmt.Println("📌 关键原则：goroutine 阻塞时，runtime 会让 P 执行别的 goroutine，")
	fmt.Println("   所以单个 goroutine 阻塞 ≠ 程序阻塞。")

	// ---------- 场景 1：channel 收发阻塞 ----------
	fmt.Println("【场景 1】channel 收发阻塞（无缓冲必须配对）")
	ch := make(chan int)
	start := time.Now()
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch <- 42
	}()
	v := <-ch
	fmt.Printf("  收到: %d (耗时 %v)\n\n", v, time.Since(start))

	// ---------- 场景 2：Mutex 阻塞 ----------
	fmt.Println("【场景 2】Mutex 阻塞（goroutine 等锁时让出 P）")
	var mu sync.Mutex
	mu.Lock()
	go func() {
		time.Sleep(200 * time.Millisecond)
		mu.Unlock()
	}()
	start = time.Now()
	mu.Lock()
	mu.Unlock()
	fmt.Printf("  Mutex 已释放 (等待 %v)\n\n", time.Since(start))

	// ---------- 场景 3：select 阻塞 ----------
	fmt.Println("【场景 3】select 阻塞（等多个 channel 中任意一个就绪）")
	done := make(chan struct{})
	go func() {
		time.Sleep(250 * time.Millisecond)
		close(done)
	}()
	start = time.Now()
	select {
	case <-done:
		fmt.Printf("  done 已关闭 (耗时 %v)\n\n", time.Since(start))
	case <-time.After(5 * time.Second):
		fmt.Println("  timeout!")
	}

	// ---------- 场景 4：syscall 阻塞 ----------
	fmt.Println("【场景 4】time.Sleep syscall 阻塞（M 进内核，netpoller 处理）")
	start = time.Now()
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("  sleep 完成 (耗时 %v)\n\n", time.Since(start))

	// ---------- 场景 5：context 等待 ----------
	fmt.Println("【场景 5】context.Done() 阻塞（goroutine 等 context 取消）")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start = time.Now()
	<-ctx.Done()
	fmt.Printf("  context 超时 (耗时 %v)\n\n", time.Since(start))

	fmt.Println("✅ 5 种阻塞场景全部演示完毕")
	fmt.Println()
	fmt.Println("📌 面试陷阱题：")
	fmt.Println("   Q: time.Sleep 会阻塞 goroutine 还是 M？")
	fmt.Println("   A: goroutine 阻塞进入等待队列，同时 M 也会被绑定（如果时间短）")
	fmt.Println("      runtime 会根据情况决定：是否要把 M 释放给别的 P")
	fmt.Println("   Q: 网络 I/O 阻塞会卡死整个程序吗？")
	fmt.Println("   A: 不会，netpoller 会把网络 fd 交给 epoll/kqueue，goroutine 等待但不占 M")
}

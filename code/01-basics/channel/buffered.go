package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoBuffered 演示 buffered vs unbuffered vs cap=1 信号量模式
//
// ============================================================================
// 关键认知：
//   unbuffered (cap=0): 同步语义 — sender 阻塞直到有 receiver
//   buffered  (cap=N):  异步语义 — 缓冲未满时 sender 不阻塞
//   cap=1 特殊用法:     可当"信号量"（一次只允许一个 goroutine 持有）
// ============================================================================
func DemoBuffered() {
	fmt.Println("=== buffered vs unbuffered channel ===")
	fmt.Println()

	// ---------- 1. unbuffered ----------
	fmt.Println("【实验 1】unbuffered channel = 同步握手")
	unbuf := make(chan int)
	start := time.Now()

	go func() {
		time.Sleep(200 * time.Millisecond) // 模拟准备工作
		fmt.Println("  [sender] 准备发送")
		unbuf <- 42
		fmt.Println("  [sender] 已发送（receiver 接收到才返回）")
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("  [receiver] 准备接收")
	v := <-unbuf
	fmt.Printf("  [receiver] 收到: %d (sender 的 \"已发送\" 也立即出现)\n", v)
	fmt.Printf("  总耗时: %v\n\n", time.Since(start))

	// ---------- 2. buffered ----------
	fmt.Println("【实验 2】buffered channel = 异步队列（缓冲未满时不阻塞）")
	buf := make(chan int, 3)
	start = time.Now()

	go func() {
		for i := 1; i <= 3; i++ {
			fmt.Printf("  [sender] 发送 %d (剩余容量 %d)\n", i, cap(buf)-len(buf))
			buf <- i
		}
		fmt.Println("  [sender] 缓冲已满，下次发送会阻塞")
	}()

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("  主 goroutine 立刻读 1 个: %d (剩余 %d)\n", <-buf, len(buf))
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("  再读 1 个: %d (剩余 %d)\n", <-buf, len(buf))
	fmt.Printf("  再读 1 个: %d (剩余 %d)\n\n", <-buf, len(buf))
	_ = start

	// ---------- 3. cap=1 信号量模式 ----------
	fmt.Println("【实验 3】cap=1 channel = 信号量（限流 / 互斥）")
	sem := make(chan struct{}, 1) // 容量为 1 的 buffered channel
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{} // 获取信号量
			fmt.Printf("  [worker %d] 获取信号量，开始工作\n", id)
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("  [worker %d] 工作完成，释放\n", id)
			<-sem // 释放信号量
		}(i)
	}
	wg.Wait()
	fmt.Println("  5 个 worker 串行执行（同一时刻只有 1 个持有信号量）")
	fmt.Println()

	// ---------- 4. 对比总结 ----------
	fmt.Println("【实验 4】对比 unbuffered / cap=1 / cap=N 的 send 行为")
	for _, cap := range []int{0, 1, 3} {
		ch := make(chan int, cap)
		start = time.Now()
		sendDone := make(chan struct{})
		go func() {
			ch <- 1
			close(sendDone)
		}()
		select {
		case <-sendDone:
			fmt.Printf("  cap=%d: 立即发送成功（耗时 %v）\n", cap, time.Since(start))
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("  cap=%d: 100ms 内未发送（说明阻塞）\n", cap)
		}
	}
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - unbuffered: 必须有 receiver 才算完成一次通信")
	fmt.Println("   - buffered:   缓冲是队列，满了才阻塞")
	fmt.Println("   - cap=1:      经典信号量模式（一次只允许一个持有）")
	fmt.Println("   - 选择哪种：决定权在你，需要同步还是解耦生产/消费速度")
}

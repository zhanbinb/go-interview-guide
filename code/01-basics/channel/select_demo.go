package main

import (
	"fmt"
	"time"
)

// DemoSelect 演示 select 的各种用法
//
// ============================================================================
// select 关键特性：
//   1. 等待多个 channel 中任意一个就绪
//   2. 多个 case 就绪时，随机选一个执行
//   3. default case 没就绪时立即执行（实现非阻塞 select）
//   4. time.After(N) 可做超时控制
//   5. nil channel 在 select 中永远不 ready（被跳过）← 经典技巧
// ============================================================================
func DemoSelect() {
	fmt.Println("=== select 多路复用 ===")
	fmt.Println()

	// ---------- 1. 等待多个 channel ----------
	fmt.Println("【实验 1】等两个 channel 中任意一个就绪")
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(300 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "two"
	}()

	select {
	case msg := <-ch1:
		fmt.Printf("  收到 ch1: %q\n", msg)
	case msg := <-ch2:
		fmt.Printf("  收到 ch2: %q (先就绪的)\n", msg)
	}
	// 实际会收到 ch2，因为它先就绪 (200ms < 300ms)
	fmt.Println()

	// ---------- 2. 多 case + 超时 ----------
	fmt.Println("【实验 2】select + time.After 超时")
	ch := make(chan string)
	go func() {
		time.Sleep(1 * time.Second) // 故意延迟
		ch <- "slow"
	}()

	select {
	case msg := <-ch:
		fmt.Printf("  收到: %q\n", msg)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("  ⏰ 500ms 超时（slow 还没就绪）")
	}
	fmt.Println()

	// ---------- 3. default（非阻塞）----------
	fmt.Println("【实验 3】default 实现非阻塞收发")
	emptyCh := make(chan int)
	select {
	case v := <-emptyCh:
		fmt.Printf("  收到: %d\n", v)
	default:
		fmt.Println("  default 触发（emptyCh 没数据，立刻走 default）")
	}
	fmt.Println()

	// ---------- 4. nil channel 在 select 中跳过 ----------
	fmt.Println("【实验 4】nil channel 在 select 中被跳过（经典技巧）")
	fmt.Println("  场景：有一个 quit channel，要根据 quit 是否传值来切换两种行为")
	_ = "quit pattern (commented out)" // unused for demo, kept for reference
	// 模拟：根据 quit 的值，发送不同的消息到 ch1 或 ch2
	ch1 = make(chan string)
	ch2 = make(chan string)
	var ch1Active, ch2Active = true, true

	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(50 * time.Millisecond)
			// 通过 select 发送
			select {
			case ch1 <- fmt.Sprintf("msg-%d", i):
				if !ch1Active {
					// 不会发生
				}
			case ch2 <- fmt.Sprintf("msg-%d", i):
				if !ch2Active {
					// 不会发生
				}
			}
		}
	}()

	// 接收端：根据需要禁用某个 case
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(80 * time.Millisecond)
			select {
			case msg := <-ch1:
				fmt.Printf("    <-ch1: %s\n", msg)
			case msg := <-ch2:
				fmt.Printf("    <-ch2: %s\n", msg)
			}
			if i == 1 {
				// 模拟 "禁用 ch1" —— 把 ch1 设为 nil
				fmt.Println("    [禁用 ch1 case]")
				_ = ch1Active
				ch1 = nil
			}
		}
	}()

	time.Sleep(400 * time.Millisecond)
	fmt.Println()

	// ---------- 5. 多个 case 同时就绪时随机选择 ----------
	fmt.Println("【实验 5】多个 case 同时就绪时随机选择（防止饥饿）")
	// 关键：cap=2 才能保证循环里 push 不阻塞（push 一个，select 读一个）
	chA := make(chan struct{}, 2)
	chB := make(chan struct{}, 2)
	// 预先各放 1 个，让 select 一上来就两个都就绪
	chA <- struct{}{}
	chB <- struct{}{}

	countA, countB := 0, 0
	for i := 0; i < 100; i++ {
		select {
		case <-chA:
			countA++
			chA <- struct{}{} // 消费后再补一个
		case <-chB:
			countB++
			chB <- struct{}{}
		}
	}
	fmt.Printf("  100 次里 chA=%d, chB=%d（接近 50:50，证明随机选择）\n\n", countA, countB)

	fmt.Println("📌 面试要点:")
	fmt.Println("   - select 等第一个就绪的 case；都就绪则随机选")
	fmt.Println("   - default + select 实现非阻塞收发")
	fmt.Println("   - nil channel 在 select 中被跳过（动态启用/禁用 case 的核心）")
	fmt.Println("   - 常见模式：select { case <-ctx.Done(): ... }")
}

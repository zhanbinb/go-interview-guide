package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func main() {
	//demoStates()
	DemoLeak()
}

func demoStates() {
	fmt.Println("【演示 channel 状态】")
	var nilCh chan int
	fmt.Printf("  类型: %T, 值: %v, 是 nil: %v\n", nilCh, nilCh, nilCh == nil)
	fmt.Println("  <-nilCh: 永久阻塞（演示：goroutine 里读 200ms 超时）")
	tryRead(nilCh, 200*time.Millisecond, "nil")

	fmt.Println("  nilCh<-: 永久阻塞（演示：goroutine 里写 200ms 超时）")
	tryWrite(nilCh, 42, 200*time.Millisecond, "nil")

	fmt.Println("  close(nilCh): panic (close of nil channel)")
	tryClose(nilCh, "nil")
	fmt.Println()

}

func tryRead(ch chan int, timeout time.Duration, label string) {
	done := make(chan struct{})
	var got int
	var ok bool
	go func() {
		got, ok = <-ch
		close(done)
	}()
	select {
	case <-done:
		fmt.Printf("    读到: %d, ok=%v\n", got, ok)
	case <-time.After(timeout):
		fmt.Printf("    ⏰ %v 内没读到（说明 %s 阻塞）\n", timeout, label)
	}
}

// tryWrite 尝试向 ch 写，超时则放弃
// 返回：是否 panic（true 表示 panic 了）
func tryWrite(ch chan int, val int, timeout time.Duration, label string) {
	if ch == nil {
		fmt.Println("    ⏰ nil channel 永久阻塞")
		return
	}
	type result struct {
		panicked bool
		panicMsg any
	}
	done := make(chan result, 1)
	go func() {
		var r result
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					r.panicked = true
					r.panicMsg = rec
				}
			}()
			ch <- val
		}()
		done <- r
	}()
	select {
	case r := <-done:
		if r.panicked {
			fmt.Printf("    💥 PANIC: %v\n", r.panicMsg)
		} else {
			fmt.Printf("    ✅ 写入成功\n")
		}
	case <-time.After(timeout):
		fmt.Printf("    ⏰ %v 内没写入（说明 %s 阻塞）\n", timeout, label)
	}
}

// tryClose 尝试关闭 ch，捕获 panic
func tryClose(ch chan int, label string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    💥 PANIC: %v\n", r)
		}
	}()
	close(ch)
	fmt.Printf("    ✅ 关闭成功\n")
}
func DemoLeak() {
	fmt.Println("=== Channel 泄漏演示 ===")
	fmt.Println()

	report := func(label string) {
		fmt.Printf("  %s: NumGoroutine = %d\n", label, runtime.NumGoroutine())
	}
	report("初始")

	// ---------- 泄漏 1：发送方阻塞 ----------
	fmt.Println("\n【泄漏 1】向无缓冲 channel 发送，但无 receiver")
	unbuf := make(chan int)
	for i := 0; i < 3; i++ {
		go func(id int) {
			unbuf <- id // 永久阻塞
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 1 后")

	// ---------- 泄漏 2：接收方阻塞 ----------
	fmt.Println("\n【泄漏 2】从 channel 接收，但 sender 早早 return")
	oneShot := make(chan int)
	go func() {
		// sender 只发一次就退出，但 channel 不关闭
		oneShot <- 42
		fmt.Println("    sender 已 return")
	}()
	go func() {
		v := <-oneShot
		fmt.Printf("    receiver 收到: %d\n", v)
		// 之后 receiver 等下一个值，但永远等不到
		v2 := <-oneShot
		fmt.Printf("    receiver 又收到: %d (不会发生)\n", v2)
	}()
	time.Sleep(50 * time.Millisecond)
	report("泄漏 2 后")

	// ---------- 泄漏 3：buffered channel 持续生产，无人消费 ----------
	fmt.Println("\n【泄漏 3】buffered channel 持续生产，无人消费")
	buf := make(chan int, 100) // 缓冲 100
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				buf <- id*10 + j // 没人消费
			}
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 3 后")
	fmt.Printf("    buf 现状: len=%d, cap=%d\n", len(buf), cap(buf))

	// ---------- 泄漏 4：永远不取消的 context ----------
	fmt.Println("\n【泄漏 4】等永远不会 cancel 的 context")
	for i := 0; i < 3; i++ {
		go func(id int) {
			ctx, _ := context.WithCancel(context.Background()) // 故意不 cancel
			<-ctx.Done()                                       // 永久阻塞
			_ = id
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	report("泄漏 4 后")

	fmt.Println()
	fmt.Println("⚠️  准备触发 fatal error: all goroutines are asleep...")
	fmt.Println("   （主 goroutine 阻塞后，runtime 检测所有 goroutine 都卡住）")

	time.Sleep(500 * time.Millisecond)
	select {} // 主 goroutine 也阻塞
}

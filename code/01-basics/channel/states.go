package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoStates 演示 channel 的 3 状态 × 3 操作行为矩阵
//
// 状态：
//   nil channel     — var ch chan int
//   open channel    — make(chan int) 或 make(chan int, N)
//   closed channel  — close(ch) 之后
//
// 操作：读 (<-ch) / 写 (ch<-) / close(ch)
//
// ============================================================================
//                 │  nil channel  │  open channel   │  closed channel
// ────────────────┼───────────────┼─────────────────┼──────────────────
// <-ch (读)       │ 永久阻塞       │ 阻塞/拿到值      │ 立即返回零值, ok=false
// ch<- (写)       │ 永久阻塞       │ 阻塞/写入成功    │ PANIC send on closed
// close(ch)       │ PANIC close   │ 关闭成功         │ PANIC close of closed
//                  │   of nil     │                  │
// ============================================================================
func DemoStates() {
	fmt.Println("=== Channel 3 状态 × 3 操作行为矩阵 ===")
	fmt.Println()
	fmt.Println("📌 速记（15字口诀）：")
	fmt.Println("   nil 全阻塞，关闭写 panic，关闭可读零值")
	fmt.Println()

	// ---------- nil channel ----------
	fmt.Println("【nil channel】var ch chan int（零值）")
	var nilCh chan int
	fmt.Printf("  类型: %T, 值: %v, 是 nil: %v\n", nilCh, nilCh, nilCh == nil)
	fmt.Println("  <-nilCh: 永久阻塞（演示：goroutine 里读 200ms 超时）")
	tryRead(nilCh, 200*time.Millisecond, "nil")
	fmt.Println("  nilCh<-: 永久阻塞（演示：goroutine 里写 200ms 超时）")
	tryWrite(nilCh, 42, 200*time.Millisecond, "nil")
	fmt.Println("  close(nilCh): panic (close of nil channel)")
	tryClose(nilCh, "nil")
	fmt.Println()

	// ---------- open channel (unbuffered) ----------
	fmt.Println("【open unbuffered】make(chan int)")
	openCh := make(chan int)
	fmt.Printf("  类型: %T, 容量: %d, 长度: %d\n", openCh, cap(openCh), len(openCh))
	fmt.Println("  openCh<-: 阻塞直到有 receiver（演示：异步发送 + 超时）")
	tryWrite(openCh, 42, 200*time.Millisecond, "open-unbuffered")
	fmt.Println("  close(openCh): OK")
	tryClose(openCh, "open-unbuffered")
	fmt.Println()

	// ---------- closed channel ----------
	fmt.Println("【closed channel】close(ch) 之后")
	closedCh := make(chan int)
	close(closedCh)
	fmt.Println("  <-closedCh: 不阻塞，返回 (0, false)")
	v, ok := <-closedCh
	fmt.Printf("    读到: v=%d, ok=%v\n", v, ok)
	fmt.Println("  closedCh<-: 立即 panic (send on closed channel)")
	tryWrite(closedCh, 42, 200*time.Millisecond, "closed")
	fmt.Println("  close(closedCh): panic (close of closed channel)")
	tryClose(closedCh, "closed")
	fmt.Println()

	// ---------- buffered open channel ----------
	fmt.Println("【open buffered】make(chan int, 3) 写满后行为")
	bufCh := make(chan int, 3)
	bufCh <- 1
	bufCh <- 2
	bufCh <- 3
	fmt.Printf("  写满 3 个后: len=%d, cap=%d\n", len(bufCh), cap(bufCh))
	fmt.Println("  bufCh<-: 阻塞（缓冲满）")
	tryWrite(bufCh, 4, 200*time.Millisecond, "open-buffered-full")
	fmt.Println("  <-bufCh: 立即拿到")
	fmt.Printf("    收到: %d (剩余 len=%d)\n", <-bufCh, len(bufCh))
	fmt.Println()

	// ---------- nil channel 妙用 ----------
	fmt.Println("【nil channel 妙用】nilCh <- 42 在 select 中被跳过")
	var nilCh2 chan int
	select {
	case nilCh2 <- 42:
		fmt.Println("  case nilCh2 <- 42: 触发（永远不可能）")
	default:
		fmt.Println("  default: 触发（因为 nil channel 永远不 ready）")
	}
	fmt.Println()

	fmt.Println("✅ 行为矩阵演示完毕。详细总结见 README.md")
}

// tryRead 尝试从 ch 读，超时则放弃（不阻塞主流程）
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

// suppress unused warning
var _ = sync.WaitGroup{}

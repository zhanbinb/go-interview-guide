package main

import (
	"context"
	"fmt"
	"time"
)

// DemoPatterns 演示 Context 的 4 种实战模式
//
// 1. 超时控制（HTTP / RPC）
// 2. goroutine 协调（用 select 等 ctx.Done()）
// 3. 级联取消（父 ctx 取消 → 子 ctx 自动取消）
// 4. 跨 API 边界传值（traceID、userID）
func DemoPatterns() {
	fmt.Println("=== 4 种实战模式 ===")
	fmt.Println()

	// ---------- 模式 1：HTTP 请求超时 ----------
	fmt.Println("【模式 1】HTTP 请求超时")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel1()
	select {
	case <-time.After(50 * time.Millisecond):
		fmt.Println("  ✓ 请求完成 (50ms)")
	case <-ctx1.Done():
		fmt.Printf("  ✗ 超时: %v\n", ctx1.Err())
	}
	fmt.Println()

	// ---------- 模式 2：goroutine 协调 ----------
	fmt.Println("【模式 2】goroutine 协调（用 select 等 ctx.Done()）")
	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		for i := 0; ; i++ {
			select {
			case <-ctx2.Done():
				fmt.Println("  worker 收到取消信号, 退出")
				return
			case <-time.After(50 * time.Millisecond):
				fmt.Printf("  worker working... %d\n", i)
			}
		}
	}()
	time.Sleep(200 * time.Millisecond)
	cancel2() // 取消
	fmt.Println()

	// ---------- 模式 3：级联取消 ----------
	fmt.Println("【模式 3】级联取消（父 ctx 取消 → 子 ctx 自动取消）")
	parent, pcancel := context.WithCancel(context.Background())
	child, ccancel := context.WithCancel(parent)
	defer pcancel()
	defer ccancel()
	go func() {
		<-child.Done()
		fmt.Printf("  子 ctx 取消了: %v\n", child.Err())
	}()
	time.Sleep(50 * time.Millisecond)
	pcancel() // 取消父 ctx
	time.Sleep(100 * time.Millisecond) // 等待 goroutine 打印
	fmt.Println()

	// ---------- 模式 4：传值 ----------
	fmt.Println("【模式 4】跨 API 边界传值 (traceID)")
	root := context.Background()
	ctx4 := context.WithValue(root, "traceID", "abc-def-789")
	handleRequest(ctx4, "/users")
	fmt.Println()

	fmt.Println("📌 最佳实践:")
	fmt.Println("   - HTTP/RPC 调用必须传 ctx（WithTimeout）")
	fmt.Println("   - defer cancel() 释放资源")
	fmt.Println("   - 业务函数第一参数接收 ctx")
	fmt.Println("   - WithValue 只传 traceID / userID 这类请求范围数据")
}

func handleRequest(ctx context.Context, path string) {
	traceID := ctx.Value("traceID")
	fmt.Printf("  handling %s, traceID=%v\n", path, traceID)
}

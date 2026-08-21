package main

import (
	"context"
	"fmt"
	"time"
)

// DemoStruct 演示 Context 接口和 4 个方法
//
// Context 接口（context/context.go）：
//   type Context interface {
//       Deadline() (deadline time.Time, ok bool)
//       Done() <-chan struct{}
//       Err() error
//       Value(key any) any
//   }
//
// 面试问法："Context 接口有哪些方法？"
func DemoStruct() {
	fmt.Println("=== Context 接口 ===")
	fmt.Println()

	// 用 WithTimeout 创建一个有 deadline 的 ctx
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 1. Deadline()
	deadline, ok := ctx.Deadline()
	fmt.Println("【方法 1】 Deadline() 返回截止时间")
	fmt.Printf("  deadline=%v, ok=%v (ok 表示 ctx 有截止时间)\n", deadline, ok)
	fmt.Println()

	// 2. Done()
	done := ctx.Done()
	fmt.Println("【方法 2】 Done() <-chan struct{} (取消信号)")
	fmt.Printf("  done channel type: %T, value: %v\n", done, done)
	fmt.Println("  → channel 关闭（或收到值）就表示 ctx 被取消")
	fmt.Println()

	// 3. Err() — 在 ctx 没取消时返回 nil
	fmt.Println("【方法 3】 Err() error (取消原因)")
	fmt.Printf("  当前 err: %v (ctx 未取消 → nil)\n", ctx.Err())
	fmt.Println()

	// 4. Value()
	fmt.Println("【方法 4】 Value(key any) any (存/取值)")
	ctx2 := context.WithValue(ctx, "userID", 42)
	fmt.Println("  ctx2.Value(userID) =", ctx2.Value("userID"))
	fmt.Println("  ctx2.Value(missing) =", ctx2.Value("missing"), "(nil)")
	fmt.Println()

	fmt.Println("📌 4 个方法的用途:")
	fmt.Println("   Deadline: 判断 ctx 是否有超时")
	fmt.Println("   Done:     select 监听取消")
	fmt.Println("   Err:      判断取消原因（canceled / deadline exceeded）")
	fmt.Println("   Value:    跨 API 边界传少量数据（不要滥用）")
}

package main

import (
	"context"
	"fmt"
	"time"
)

// DemoFunctions 演示 6 个 Context 创建函数
//
// ============================================================================
// 创建函数（context 包）：
//   1. Background()             → 根 ctx，永远不取消
//   2. TODO()                   → 占位 ctx，不知道用啥时
//   3. WithCancel(parent)       → 手动 cancel
//   4. WithTimeout(parent, d)   → 超时自动 cancel
//   5. WithDeadline(parent, t)  → 到时间自动 cancel（= WithTimeout 到时间点）
//   6. WithValue(parent, k, v)  → 存数据
//
// 面试问法："Context 怎么用？怎么取消？"
func DemoFunctions() {
	fmt.Println("=== 6 个 Context 创建函数 ===")
	fmt.Println()

	// 1. Background
	ctx0 := context.Background()
	fmt.Println("【1】 Background()")
	fmt.Printf("  ctx0 type: %T, Err: %v (永不取消)\n", ctx0, ctx0.Err())
	fmt.Println()

	// 2. TODO
	ctx1 := context.TODO()
	fmt.Println("【2】 TODO()")
	fmt.Printf("  ctx1 type: %T (跟 Background 一样，但表示 \"还不确定用啥\")\n", ctx1)
	fmt.Println()

	// 3. WithCancel
	fmt.Println("【3】 WithCancel(parent)")
	parent := context.Background()
	ctx2, cancel2 := context.WithCancel(parent)
	fmt.Printf("  cancel type: %T (返回 cancel func)\n", cancel2)
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel2() // 手动取消
	}()
	<-ctx2.Done()
	fmt.Printf("  ctx2.Err() = %v (取消原因)\n", ctx2.Err())
	fmt.Println()

	// 4. WithTimeout
	fmt.Println("【4】 WithTimeout(parent, d) — 推荐用这个做超时")
	ctx3, cancel3 := context.WithTimeout(parent, 300*time.Millisecond)
	defer cancel3()
	select {
	case <-time.After(500 * time.Millisecond):
		fmt.Println("  work 超时了")
	case <-ctx3.Done():
		fmt.Printf("  ctx3 超时取消, err = %v, deadline = %v\n", ctx3.Err(), time.Until(time.Now()))
	}
	fmt.Println()

	// 5. WithDeadline
	fmt.Println("【5】 WithDeadline(parent, t)")
	deadline := time.Now().Add(200 * time.Millisecond)
	ctx4, cancel4 := context.WithDeadline(parent, deadline)
	defer cancel4()
	<-ctx4.Done()
	fmt.Printf("  ctx4.Err() = %v (到时间取消)\n", ctx4.Err())
	fmt.Println()

	// 6. WithValue
	fmt.Println("【6】 WithValue(parent, k, v)")
	ctx5 := context.WithValue(parent, "traceID", "abc-123")
	ctx5 = context.WithValue(ctx5, "userID", 42)
	fmt.Printf("  traceID = %v\n", ctx5.Value("traceID"))
	fmt.Printf("  userID  = %v\n", ctx5.Value("userID"))
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - Background / TODO 是 ctx 的\"根\"")
	fmt.Println("   - WithCancel/Timeout/WithDeadline 都返回 cancel func")
	fmt.Println("   - defer cancel() 是好习惯（避免资源泄漏）")
	fmt.Println("   - WithValue 别滥用（只用于请求范围的数据）")
}

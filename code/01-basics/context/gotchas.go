package main

import (
	"context"
	"fmt"
	"time"
)

// DemoGotchas 演示 Context 的 3 个常见陷阱
//
// ============================================================================
// 1. 忘记 defer cancel() → ctx 泄漏（虽然 ctx 本身不占很多资源，但关联的 goroutine 可能）
// 2. WithValue 用字符串作 key（应该用自定义类型，避免冲突）
// 3. ctx 传给结构体（应该作为函数参数显式传递）
// ============================================================================
func DemoGotchas() {
	fmt.Println("=== 3 个常见陷阱 ===")
	fmt.Println()

	// ---------- 陷阱 1：忘记 cancel ----------
	fmt.Println("【陷阱 1】忘记 defer cancel()")
	_, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// 忘记 defer cancel()
	fmt.Println("  ctx 在 3 秒后才被 GC（虽然不严重，但 goroutine 计时器会泄漏）")
	_ = cancel
	fmt.Println("  ✓ 正确: defer cancel() 立即释放")
	fmt.Println()

	// ---------- 陷阱 2：WithValue 用 string 作 key ----------
	fmt.Println("【陷阱 2】WithValue 用 string 作 key")
	ctx := context.WithValue(context.Background(), "traceID", "abc") // ⚠️ 字符串 key
	// 任何包都可以用 "traceID" 作 key，可能冲突
	fmt.Println("  用 string key: 风险大，容易冲突")
	fmt.Println("  ✓ 正确: 自定义类型作 key")
	type traceIDKey struct{} // 私有类型
	ctx2 := context.WithValue(context.Background(), traceIDKey{}, "abc")
	fmt.Printf("    ctx2 = %v\n", ctx2)
	_ = ctx
	fmt.Println()

	// ---------- 陷阱 3：ctx 放进结构体 ----------
	fmt.Println("【陷阱 3】把 ctx 放进结构体")
	type BadStruct struct {
		ctx context.Context // ⚠️ 不要这样
		data string
	}
	_ = BadStruct{}
	fmt.Println("  ctx 不应该在 struct 里")
	fmt.Println("  ✓ 正确: ctx 作为函数第一参数")
	type GoodHandler struct {
		data string
	}
	handle := GoodHandler{data: "x"}
	processRequest(context.Background(), handle)
	fmt.Println()

	fmt.Println("📌 经验法则:")
	fmt.Println("   1. defer cancel()（除非故意不取消）")
	fmt.Println("   2. WithValue key 用自定义类型")
	fmt.Println("   3. ctx 永远作为函数参数，不进结构体")
}

func processRequest(ctx context.Context, h struct{ data string }) {
	fmt.Printf("  processRequest: ctx type=%T, data = %v\n", ctx, h.data)
}

package main

import "fmt"

// DemoEscape 演示 5 种逃逸场景
//
// 逃逸分析（escape analysis）是编译器在编译期判断对象放栈还是堆
// 栈分配：函数返回时自动回收，无 GC 压力
// 堆分配：需要 GC 跟踪
//
// 看逃逸分析的命令：
//   go build -gcflags="-m"       # -m 打印分析
//   go build -gcflags="-m -m"    # 更详细
//
// 5 种典型逃逸场景：

// 场景 1：返回局部变量指针 → 逃逸
func escapeReturnPtr() *int {
	x := 42
	return &x // x 必须逃逸到堆（外部要引用）
}

// 场景 2：闭包捕获外部变量 → 逃逸
func escapeClosure() func() int {
	y := 100
	return func() int { return y } // y 被闭包捕获，必须逃逸
}

// 场景 3：发送到 channel（指针/interface）→ 逃逸
func escapeChannel() {
	ch := make(chan *int, 1)
	x := 1
	ch <- &x // 发送到 channel 后可能被其他 goroutine 接收，必须堆分配
	close(ch)
}

// 场景 4：slice 包含指针 → slice 元素逃逸
func escapeSlice() {
	s := make([]*int, 10)
	for i := 0; i < 10; i++ {
		x := i
		s[i] = &x // 局部变量的地址被放进 slice，必须堆分配
	}
	_ = s
}

// 场景 5：大对象 / 切片扩容后超过栈空间 → 逃逸
func escapeLarge() {
	// 局部变量太大（或编译器判定为太大）→ 逃逸到堆
	s := make([]byte, 1024*1024) // 1MB
	s[0] = 1
	_ = s
}

func DemoEscape() {
	fmt.Println("=== 5 种逃逸场景 ===")
	fmt.Println()

	fmt.Println("【场景 1】返回局部变量指针")
	_ = escapeReturnPtr()
	fmt.Println("  → x 逃逸到堆（外部还要用）")
	fmt.Println()

	fmt.Println("【场景 2】闭包捕获外部变量")
	_ = escapeClosure()
	fmt.Println("  → y 逃逸到堆（闭包要用）")
	fmt.Println()

	fmt.Println("【场景 3】发送到 channel")
	escapeChannel()
	fmt.Println("  → x 逃逸到堆（可能被其他 goroutine 接收）")
	fmt.Println()

	fmt.Println("【场景 4】slice 包含指针")
	escapeSlice()
	fmt.Println("  → s[i] = &x 导致 x 逃逸到堆")
	fmt.Println()

	fmt.Println("【场景 5】大对象")
	escapeLarge()
	fmt.Println("  → 1MB 大小可能逃逸（栈默认 8KB 太小）")
	fmt.Println()

	fmt.Println("📌 自己验证：")
	fmt.Println("   go build -gcflags=\"-m\" code/01-basics/memory/escape.go")
	fmt.Println()
	fmt.Println("📌 面试要点:")
	fmt.Println("   - 编译器做逃逸分析（不需要运行时）")
	fmt.Println("   - 栈分配：函数返回即回收，无 GC 压力 → 性能好")
	fmt.Println("   - 堆分配：需要 GC → 有开销")
	fmt.Println("   - 经验：能用栈就尽量用栈（避免不必要的指针返回）")
}

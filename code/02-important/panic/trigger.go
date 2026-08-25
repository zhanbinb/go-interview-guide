package main

import "fmt"

// DemoTrigger 演示 6 种 panic 触发场景
func DemoTrigger() {
	fmt.Println("=== 6 种 panic 触发场景 ===")
	fmt.Println()

	// 场景 1：数组越界
	fmt.Println("【场景 1】数组越界（运行时检测）")
	tryRun("数组越界", func() {
		s := []int{1, 2, 3}
		_ = s[10]
	})
	fmt.Println()

	// 场景 2：nil 指针解引用
	fmt.Println("【场景 2】nil 指针解引用")
	tryRun("nil 指针", func() {
		var p *int = nil
		_ = *p
	})
	fmt.Println()

	// 场景 3：类型断言失败（不带 ok）
	fmt.Println("【场景 3】类型断言失败（不带 ok 形式）")
	tryRun("类型断言", func() {
		var i interface{} = "hello"
		_ = i.(int)
	})
	fmt.Println()

	// 场景 4：关闭已关闭的 channel
	fmt.Println("【场景 4】关闭已关闭的 channel")
	tryRun("close closed", func() {
		ch := make(chan int)
		close(ch)
		close(ch) // panic
	})
	fmt.Println()

	// 场景 5：向已关闭的 channel 发送
	fmt.Println("【场景 5】向已关闭的 channel 发送")
	tryRun("send closed", func() {
		ch := make(chan int)
		close(ch)
		ch <- 1 // panic
	})
	fmt.Println()

	// 场景 6：除数为 0（编译期就报错，不是 panic）
	fmt.Println("【场景 6】除数为 0（编译期就报错，不会 panic）")
	fmt.Println("  var x int = 1 / 0  // 编译错误")
	fmt.Println()

	fmt.Println("📌 注意：map 并发读写会触发 fatal error（不是普通 panic）")
	fmt.Println("   goroutine 是 fatal: concurrent map writes")
	fmt.Println("   → 不可 recover，整个进程退出")
}

// tryRun 执行 fn，捕获 panic
func tryRun(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  💥 %s: panic 捕获 = %v\n", label, r)
		}
	}()
	fn()
}

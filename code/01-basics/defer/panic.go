package main

import "fmt"

// DemoPanicRecover 演示 defer + recover 捕获 panic
//
// 关键点：
//   - recover 必须在 defer 里调用才有效
//   - recover 只能捕获当前 goroutine 的 panic
//   - recover 后程序继续执行（不会终止）
//
// 面试问法："defer recover 怎么用？"
func DemoPanicRecover() {
	fmt.Println("=== defer + recover ===")

	// 安全的调用：包一层
	safeCall := func(name string, fn func()) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  [recover] %s: 捕获 panic: %v\n", name, r)
			}
		}()
		fmt.Printf("  调用 %s...\n", name)
		fn()
		fmt.Printf("  %s 正常结束\n", name)
	}

	// 场景 1：正常函数（无 panic）
	safeCall("normal", func() {})

	// 场景 2：会 panic 的函数
	safeCall("crash", func() {
		panic("💥 出错了！")
	})

	// 场景 3：recover 后程序继续
	fmt.Println("\n  程序继续执行，没崩 ✨")
}

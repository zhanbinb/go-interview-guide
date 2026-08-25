package main

import "fmt"

// DemoRecover 演示 recover 的用法
//
// ============================================================================
// recover 用法：
//
//   1. 基本用法：捕获 panic，程序继续
//   2. re-panic：恢复后重新 panic（保留堆栈，让上层处理）
//   3. 库代码：用 named return + defer recover 包装错误
//
// 关键规则：
//   - recover 必须在 defer 里才生效
//   - recover 只能捕获当前 goroutine 的 panic
//   - recover 后函数返回零值（正常流程）
// ============================================================================
func DemoRecover() {
	fmt.Println("=== recover 用法 ===")
	fmt.Println()

	// 1. 基本用法
	fmt.Println("【用法 1】基本用法：defer recover")
	fmt.Println("  调用前: 程序会崩溃")
	safeCall("normal", func() {
		fmt.Println("    normal 函数执行中")
	})
	safeCall("crash", func() {
		panic("💥 出错了！")
	})
	fmt.Println("  调用后: 程序继续 ✨")
	fmt.Println()

	// 2. 包装错误
	fmt.Println("【用法 2】包装错误（库代码模式）")
	doWork := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("panic: %v", r)
				}
			}
		}()
		// 模拟业务
		panic("biz error")
	}
	err := doWork()
	fmt.Printf("  err = %v (从 panic 转成 error 返回)\\n", err)
	fmt.Println()

	// 3. re-panic 模式
	fmt.Println("【用法 3】re-panic（库代码 + 上层处理）")
	wrapper := func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("    [inner] 第一次 recover: %v\\n", r)
				panic(r) // 重新 panic，让外层处理
			}
		}()
		panic("💥 inner panic")
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("    [outer] 最终 recover: %v\\n", r)
			}
		}()
		wrapper()
	}()
	fmt.Println()

	// 4. 区分错误类型
	fmt.Println("【用法 4】区分错误类型（用类型 switch）")
	handle := func(fn func()) {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case string:
					fmt.Printf("    string panic: %q\\n", e)
				case int:
					fmt.Printf("    int panic: %d\\n", e)
				case error:
					fmt.Printf("    error panic: %v\\n", e)
				default:
					fmt.Printf("    unknown panic: %v\\n", e)
				}
			}
		}()
		fn()
	}
	handle(func() { panic("字符串错误") })
	handle(func() { panic(42) })
	handle(func() { panic(fmt.Errorf("error 错误")) })
	fmt.Println()

	fmt.Println("📌 经验法则:")
	fmt.Println("   - 业务代码：可以用 recover，但要打印日志")
	fmt.Println("   - 库代码：慎重 recover（可能吞掉 bug）")
	fmt.Println("   - 跨 goroutine：每个 goroutine 需要独立 recover")
}

// safeCall 用 defer recover 包装函数调用
func safeCall(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  [safeCall] %s 捕获: %v\\n", name, r)
		}
	}()
	fmt.Printf("  调用 %s...\\n", name)
	fn()
	fmt.Printf("  %s 正常结束\\n", name)
}

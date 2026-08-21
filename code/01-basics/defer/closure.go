package main

import "fmt"

// DemoClosure 演示 defer 的闭包陷阱
//
// 两种写法：
//   defer func(x int) { ... }(x)   // 参数立即求值（拷贝当时的值）
//   defer func() { ... }()         // 闭包延迟求值（看到最终值）
//
// 面试问法："defer 的参数什么时候求值？"
func DemoClosure() {
	fmt.Println("=== 闭包参数陷阱 ===")

	// 实验 1：参数立即求值
	fmt.Println("【实验 1】defer f(x)：参数立即求值")
	x := 1
	defer fmt.Println("  defer 1:", x) // 拷贝 x=1
	x = 999
	fmt.Println("  x 改为 999")

	// 实验 2：闭包延迟求值
	fmt.Println("\n【实验 2】defer func(){...}()：闭包延迟求值")
	y := 1
	defer func() { fmt.Println("  defer 2:", y) }() // 闭包，延迟读 y
	y = 999
	fmt.Println("  y 改为 999")

	// 实验 3：循环里的闭包陷阱（经典！）
	fmt.Println("\n【实验 3】循环里 defer 的闭包陷阱")
	for i := 0; i < 3; i++ {
		defer func() { fmt.Printf("  v = %d\n", i) }()
	}
	// Go 1.22+：i 每次循环独立，所以会输出 2, 1, 0（LIFO）
	// Go 1.21-：所有闭包共享同一 i，会输出 3, 3, 3
	// 你跑的是 Go 1.25，所以应该是 2, 1, 0
}

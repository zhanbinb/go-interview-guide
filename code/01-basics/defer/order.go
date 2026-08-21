package main

import "fmt"

// DemoOrder 演示 defer 的 LIFO 执行顺序
//
// 多个 defer 像栈一样：后注册的先执行
// 面试问法："defer 的执行顺序？"
func DemoOrder() {
	fmt.Println("=== LIFO 顺序 ===")
	defer fmt.Println("defer 1 (最先注册, 最后执行)")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3 (最后注册, 最先执行)")
	fmt.Println("函数体执行完，即将 return")
}

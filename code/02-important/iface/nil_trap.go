package main

import (
	"fmt"
)

// DemoNilTrap 演示 nil 接口陷阱
//
// ============================================================================
// 经典陷阱：接口变量的 nil 是分两层的
//   - 整个接口变量是 nil：i == nil → true（没有任何动态类型和动态值）
//   - 接口有动态类型，但动态值是 nil：i == nil → false（有 *int 类型）
//
// 这导致：
//   - 函数返回 error 时，如果返回的是 (err *MyError)(nil)，调用方判断 err == nil 会出错
//   - map 里查不存在的 key，返回 (zero, false) 不是 nil
// ============================================================================
func DemoNilTrap() {
	fmt.Println("=== nil 接口陷阱 ===")
	fmt.Println()

	// 实验 1：完全 nil 的接口
	fmt.Println("【实验 1】完全 nil 的接口")
	var i1 interface{}
	fmt.Printf("  i1 = nil: %v (没有动态类型也没动态值)\n", i1 == nil)
	fmt.Println()

	// 实验 2：有动态类型，动态值是 nil
	fmt.Println("【实验 2】有动态类型 (*int)，动态值是 nil")
	var p *int = nil
	var i2 interface{} = p
	fmt.Printf("  i2 = nil: %v ⚠️ (false!)\n", i2 == nil)
	fmt.Printf("  i2 类型: %T, 动态值: %v\n", i2, i2)
	fmt.Println()

	// 实验 3：经典错误示例
	fmt.Println("【实验 3】error 接口的常见错误")
	var err error
	fmt.Printf("  默认 err: %v, == nil: %v\n", err, err == nil)

	// 模拟一个返回 nil error 的函数
	err = returnNil()
	fmt.Printf("  returnNil() 返回 err: %v, == nil: %v ⚠️\n", err, err == nil)
	if err != nil {
		fmt.Printf("    实际是类型 %T，值 %v\n", err, err)
	}
	fmt.Println()

	// 实验 4：interface 转具体类型时也要小心
	fmt.Println("【实验 4】从接口转回具体类型")
	var i3 interface{} = (*int)(nil)
	if v, ok := i3.(*int); ok {
		fmt.Printf("  转换成功: v = %v, v == nil: %v ⚠️\n", v, v == nil)
	}
	fmt.Println()

	fmt.Println("📌 防御性写法:")
	fmt.Println("   if err != nil { ... }") // 跟 nil 比较
	fmt.Println("   vs")
	fmt.Println("   var concrete *MyError")
	fmt.Println("   if errors.As(err, &concrete) { ... }") // Go 1.13+ 推荐")
}

// returnNil 模拟一个返回 nil error 但类型不是 nil 的函数
type myErr struct{}

func (myErr) Error() string { return "" }

func returnNil() error {
	var e *myErr = nil
	return e // // 实际返回的是 (*myErr, nil)，不是 nil
}

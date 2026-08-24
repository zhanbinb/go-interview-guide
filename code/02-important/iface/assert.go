package main

import "fmt"

// DemoAssert 演示类型断言
//
// ============================================================================
// 类型断言的语法：
//   1. v := x.(T)             // 不安全，失败 panic
//   2. v, ok := x.(T)         // 安全，ok=false 表示失败
//   3. switch x.(type) { ... } // 类型 switch（只能用在 switch 里）
//
// 类型断言 vs 类型转换：
//   - x.(T): 运行期，x 必须是 interface
//   - T(x):  编译期，类型必须兼容（int↔int64, []byte↔string）
// ============================================================================
func DemoAssert() {
	fmt.Println("=== 类型断言 ===")
	fmt.Println()

	// 实验 1：基础断言（带 ok）
	fmt.Println("【实验 1】基础类型断言 (v, ok := x.(T))")
	var i interface{} = "hello"
	if v, ok := i.(string); ok {
		fmt.Printf("  i 是 string, 值 = %q\n", v)
	} else {
		fmt.Printf("  i 不是 string\n")
	}
	if v, ok := i.(int); ok {
		fmt.Printf("  i 是 int, 值 = %d\n", v)
	} else {
		fmt.Printf("  i 不是 int (ok=%v)\n", ok)
	}
	fmt.Println()

	// 实验 2：不带 ok 的断言（会 panic）
	fmt.Println("【实验 2】不带 ok 的断言 (失败会 panic)")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  💥 panic: %v\n", r)
			}
		}()
		var j interface{} = "hello"
		_ = j.(int) // 失败，panic
	}()
	fmt.Println()

	// 实验 3：类型 switch
	fmt.Println("【实验 3】类型 switch (一次判断多种类型)")
	check := func(i interface{}) string {
		switch v := i.(type) {
		case int:
			return fmt.Sprintf("int: %d", v)
		case string:
			return fmt.Sprintf("string: %q", v)
		case bool:
			return fmt.Sprintf("bool: %v", v)
		case nil:
			return "nil"
		default:
			return fmt.Sprintf("unknown: %T", v)
		}
	}
	for _, v := range []interface{}{42, "hi", true, nil, 3.14} {
		fmt.Printf("  %v → %s\n", v, check(v))
	}
	fmt.Println()

	// 实验 4：类型转换（编译期）
	fmt.Println("【实验 4】类型转换 T(x) - 编译期")
	var a int = 42
	var b int64 = int64(a)   // int → int64（编译期 OK）
	fmt.Printf("  int(%d) → int64 = %d\n", a, b)
	var c []byte = []byte("hello")
	var s string = string(c)  // []byte → string
	fmt.Printf("  []byte → string = %q\n", s)
	fmt.Println()

	fmt.Println("📌 关键区别:")
	fmt.Println("   - x.(T):   运行期，x 必须 interface，失败 panic")
	fmt.Println("   - T(x):    编译期，类型必须兼容")
	fmt.Println("   - 实战总是用 v, ok 形式，不带 ok 的容易炸")
}

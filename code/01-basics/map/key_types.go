package main

import "fmt"

// DemoKeyTypes 演示 map 的 key 类型限制
//
// 规则：key 必须可用 == 比较
// 可比较：bool, 数值, string, pointer, channel, interface, 数组（元素可比较）, struct（字段可比较）
// 不可比较：slice, map, func（含这些字段的 struct 也不行）
func DemoKeyTypes() {
	fmt.Println("=== map key 类型限制 ===")
	fmt.Println()

	// ✅ 可以做 key 的类型
	fmt.Println("【可以做 key 的类型】:")
	m1 := map[int]string{1: "a"} // int
	m2 := map[string]int{"x": 1} // string
	m3 := map[bool]int{true: 1}  // bool
	type Point struct{ X, Y int }
	m4 := map[Point]string{{1, 2}: "origin"} // struct (字段都可比较)
	m5 := map[[2]int]string{{1, 2}: "2d"}    // 数组（元素可比较）
	m6 := map[chan int]string{}              // channel
	m7 := map[*int]string{}                  // pointer

	fmt.Printf("  m1[int]      = %v\n", m1)
	fmt.Printf("  m2[string]   = %v\n", m2)
	fmt.Printf("  m3[bool]     = %v\n", m3)
	fmt.Printf("  m4[Point]    = %v\n", m4)
	fmt.Printf("  m5[[2]int]   = %v\n", m5)
	fmt.Printf("  m6[chan int] (空) = %v\n", m6)
	fmt.Printf("  m7[*int]     (空) = %v\n", m7)
	fmt.Println()

	// ❌ 不能做 key 的类型（注释掉，下面展示编译错误信息）
	fmt.Println("【不能做 key 的类型】:")
	fmt.Println("  ❌ slice:      map[[]int]string{}  // compile error")
	fmt.Println("  ❌ map:        map[map[int]int]string{} // compile error")
	fmt.Println("  ❌ func:       map[func()]string{}  // compile error")
	fmt.Println("  ❌ 含 slice/map/func 字段的 struct:")
	fmt.Println("    type Bad struct{ s []int }")
	fmt.Println("    map[Bad]string{}  // compile error")
	fmt.Println()

	// 验证 slice 不可做 key
	fmt.Println("【验证】下面的代码如果取消注释会编译报错:")
	fmt.Println("  var s []int")
	fmt.Println("  m := map[[]int]string{s: \"x\"}  // invalid: []int is not comparable")
	fmt.Println()

	// 验证 interface 做 key
	fmt.Println("【验证】interface 做 key 比较的是动态类型 + 值:")
	var i1 interface{} = "hello"
	var i2 interface{} = "hello"
	var i3 interface{} = 42
	var i4 interface{} = false
	_ = i4
	m8 := map[interface{}]string{i1: "1", i2: "2", i3: "3", i4: "4"}
	fmt.Printf("  m[interface{}] = %v\n", m8)
	fmt.Printf("  m[\"hello\"] = %q, m[42] = %q, m[false] = %q (动态类型不同不冲突)\n",
		m8[i1], m8[i3], m8[i4])
	fmt.Println()

	fmt.Println("📌 关键点:")
	fmt.Println("   - key 必须能用 == 比较")
	fmt.Println("   - 编译期就检查（不会运行时才报错）")
	fmt.Println("   - 数组/含可比较字段的 struct 可以做 key")
	fmt.Println("   - slice/map/func 及其容器都不行")
}

package main

import "fmt"

// DemoReturnValue 演示 defer 能否修改返回值
//
// 关键规则：
//   - 匿名返回值：return 时把值赋给一个临时变量，defer 改的是临时变量，外部看不到
//   - 命名返回值：return 时把值赋给命名的变量，defer 改的就是这个变量，外部能看到
//
// 面试问法："defer 能不能修改返回值？"
func DemoReturnValue() {
	fmt.Println("=== 命名 vs 匿名返回值 ===")

	// 匿名返回值：defer 改不了
	anon := func() int {
		ret := 1
		defer func() { ret = 999 }() // 改的是局部 ret
		return ret // return value = 1（拷贝）
	}
	fmt.Printf("匿名返回值: %d (defer 改的是副本)\n", anon())

	// 命名返回值：defer 能改
	named := func() (ret int) {
		ret = 1
		defer func() { ret = 999 }() // 改的是命名返回值 ret
		return // 等价 return ret
	}
	fmt.Printf("命名返回值: %d (defer 改的是原变量)\n", named())

	// 返回指针：能改
	ptr := func() *int {
		ret := 1
		defer func() { ret = 999 }()
		return &ret
	}
	fmt.Printf("返回指针:   %d (defer 改的是同一份)\n", *ptr())
}

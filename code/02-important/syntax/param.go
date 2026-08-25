package main

import "fmt"

type BigStruct struct {
	Data [1024]byte // 1KB
	Name string
}

// DemoParam 演示函数参数传值还是传指针
//
// 关键认知：
//   - Go 全是值传递（即使传指针，也是复制指针变量）
//   - 大结构体传指针（避免复制）
//   - 切片/map/channel 传引用（header 是值，但底层共享）
func DemoParam() {
	fmt.Println("=== 函数参数传递 ===")
	fmt.Println()

	// 1. 传值（复制整个结构体）
	fmt.Println("【实验 1】传值：函数内修改不影响外部")
	b := BigStruct{Name: "original"}
	modifyByValue(b)
	fmt.Printf("  函数返回后 b.Name = %q (没变)\\n\\n", b.Name)

	// 2. 传指针（共享同一个结构体）
	fmt.Println("【实验 2】传指针：函数内修改影响外部")
	b2 := BigStruct{Name: "original"}
	modifyByPointer(&b2)
	fmt.Printf("  函数返回后 b2.Name = %q (变了)\\n\\n", b2.Name)

	// 3. slice 传引用（header 是值，底层共享）
	fmt.Println("【实验 3】slice 传引用")
	s := []int{1, 2, 3}
	modifySlice(s)
	fmt.Printf("  函数返回后 s = %v (底层共享)\\n\\n", s)

	// 4. map 传引用
	fmt.Println("【实验 4】map 传引用")
	m := map[string]int{"a": 1}
	modifyMap(m)
	fmt.Printf("  函数返回后 m = %v (修改生效)\\n\\n", m)

	fmt.Println("📌 原则:")
	fmt.Println("   - 小结构体（< 几十字节）：传值")
	fmt.Println("   - 大结构体 / 需要修改：传指针")
	fmt.Println("   - slice/map/channel：传引用（直接传就行）")
	fmt.Println("   - 多个返回值 + error：经典 Go 模式")
}

// modifyByValue 传值 - 函数内修改不影响外部
func modifyByValue(b BigStruct) {
	b.Name = "modified"
	_ = b
}

// modifyByPointer 传指针 - 函数内修改影响外部
func modifyByPointer(b *BigStruct) {
	b.Name = "modified"
}

// modifySlice 修改元素 - 影响外部（slice 底层共享）
func modifySlice(s []int) {
	s[0] = 999
}

// modifyMap 修改元素 - 影响外部（map 底层共享）
func modifyMap(m map[string]int) {
	m["a"] = 999
	m["b"] = 2
}

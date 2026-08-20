package main

import "fmt"

// DemoPassParam 演示 slice 作为函数参数传递
//
// 关键认知：
//   Go 的参数传递是值传递，slice 也是值传递（slice header）
//   不过 slice header 包含一个 ptr，ptr 指向的底层数组是共享的
//   所以：
//     - 函数内修改 s[i] 会影响外部（共享底层数组）
//     - 函数内 append 不触发扩容不会影响外部
//     - 函数内 append 触发扩容后修改不影响外部（新建了底层数组）
func DemoPassParam() {
	fmt.Println("=== slice 作为参数传递 ===")
	fmt.Println()

	// 场景 1：函数内修改元素 → 影响外部
	fmt.Println("【场景 1】函数内修改 s[i] → 影响外部（共享底层数组）")
	s := []int{1, 2, 3, 4, 5}
	fmt.Printf("  调用前: %v\n", s)
	modifyElement(s, 0, 999)
	fmt.Printf("  调用后: %v (s[0] 被改了)\n", s)
	fmt.Println()

	// 场景 2：函数内 append 没扩容 → 改影响外部
	fmt.Println("【场景 2】函数内 append 没扩容 → 改影响外部")
	s = make([]int, 3, 10) // len=3, cap=10, append 不扩容
	s[0], s[1], s[2] = 1, 2, 3
	fmt.Printf("  调用前: %v (len=%d cap=%d)\n", s, len(s), cap(s))
	appendInPlace(s, 100)
	fmt.Printf("  调用后: %v (cap 够，append 在原数组写)\n", s)
	fmt.Println()

	// 场景 3：函数内 append 触发扩容 → 改不影响外部
	fmt.Println("【场景 3】函数内 append 触发扩容 → 改不影响外部")
	s = []int{1, 2, 3} // cap=3
	fmt.Printf("  调用前: %v (len=%d cap=%d)\n", s, len(s), cap(s))
	appendGrow(s, 100)
	fmt.Printf("  调用后: %v (append 扩容了，原 slice 不变)\n", s)
	fmt.Println()

	// 场景 4：返回新 slice
	fmt.Println("【场景 4】返回新 slice → 外部能拿到")
	s = []int{1, 2, 3}
	fmt.Printf("  调用前: %v\n", s)
	s2 := appendAfter(s, 100)
	fmt.Printf("  返回值: %v (扩容后的新 slice)\n", s2)
	fmt.Println()

	fmt.Println("📌 关键结论:")
	fmt.Println("   - slice header 是值拷贝（3 个字段）")
	fmt.Println("   - 但 ptr 指向的底层数组共享")
	fmt.Println("   - 所以函数内 s[i] = x 会影响外部")
	fmt.Println("   - 但函数内 s = append(s, x) 不一定影响外部（看是否扩容）")
	fmt.Println("   - 要让函数 append 生效，必须返回新 slice")
}

// modifyElement 修改 s[i] 的值
func modifyElement(s []int, i int, v int) {
	s[i] = v
}

// appendInPlace append 一个元素，不扩容（假设 cap 够）
func appendInPlace(s []int, v int) {
	// 调用方保证 len < cap，所以 append 不会扩容
	s = append(s, v)
	s[0] = 999 // 验证外部看到变化
}

// appendGrow append 一个元素，触发扩容
func appendGrow(s []int, v int) {
	s = append(s, v) // cap 不够，扩容
	s[0] = 999 // 这次改的是新数组
}

// appendAfter append 并返回新 slice
func appendAfter(s []int, v int) []int {
	return append(s, v)
}

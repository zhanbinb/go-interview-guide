package main

import (
	"fmt"
	"unsafe"
)

// DemoShare 演示截取 slice 共享底层数组
//
// 关键认知：
//   s2 := s1[1:3] 是 O(1) 操作
//   只是改了 ptr/len/cap，不复制元素
//   所以 s1 和 s2 共享底层数组
//   改 s2[0] 会影响 s1[1]
func DemoShare() {
	fmt.Println("=== slice 共享底层数组 ===")
	fmt.Println()

	// 实验 1：截取不复制
	fmt.Println("【实验 1】s2 := s1[1:3] 是否复制了元素？")
	s1 := []int{10, 20, 30, 40, 50}
	s2 := s1[1:3]
	fmt.Printf("  s1 = %v, len=%d cap=%d\n", s1, len(s1), cap(s1))
	fmt.Printf("  s2 = %v, len=%d cap=%d (cap 继承 s1)\n", s2, len(s2), cap(s2))
	fmt.Printf("  s1 和 s2 共享底层数组: &s1[1] == &s2[0]: %v\n",
		uintptr(unsafe.Pointer(&s1[1])) == uintptr(unsafe.Pointer(&s2[0])))
	fmt.Println()

	// 实验 2：改 s2 影响 s1
	fmt.Println("【实验 2】改 s2[0] 影响 s1[1]")
	s1 = []int{10, 20, 30, 40, 50}
	s2 = s1[1:3]
	fmt.Printf("  修改前: s1=%v, s2=%v\n", s1, s2)
	s2[0] = 999
	fmt.Printf("  s2[0]=999 后: s1=%v, s2=%v (s1[1] 同步变了)\n", s1, s2)
	fmt.Println()

	// 实验 3：append 超过 cap 会切断共享
	fmt.Println("【实验 3】s2 append 超过 cap → 切断共享（新建底层数组）")
	s1 = []int{10, 20, 30, 40, 50}
	s2 = s1[1:3] // s2 cap=4（继承 s1 从 index 1 到末尾）
	fmt.Printf("  s1=%v, s2 cap=%d\n", s1, cap(s2))
	// cap=4，可以 append 两个不触发扩容
	s2 = append(s2, 60)
	fmt.Printf("  append 后 s2=%v: s1=%v (s1 还受影响)\n", s2, s1)
	// 再 append 就扩容了
	s2 = append(s2, 70)
	fmt.Printf("  再 append s2=%v: s1=%v (s1 不再受影响, s2 切断了)\n", s2, s1)
	fmt.Println()

	// 实验 4：3 层截取
	fmt.Println("【实验 4】多层截取")
	s1 = []int{1, 2, 3, 4, 5}
	s2 = s1[1:4] // [2, 3, 4]
	s3 := s2[1:3] // [3, 4]
	fmt.Printf("  s1=%v s2=%v s3=%v\n", s1, s2, s3)
	fmt.Printf("  &s1[1]==&s2[0]: %v\n", &s1[1] == &s2[0])
	fmt.Printf("  &s2[1]==&s3[0]: %v\n", &s2[1] == &s3[0])
	fmt.Printf("  &s1[2]==&s3[0]: %v (深度共享)\n", &s1[2] == &s3[0])
	fmt.Println()

	// 实验 5：从数组截取
	fmt.Println("【实验 5】从数组截取 slice")
	arr := [5]int{1, 2, 3, 4, 5}
	s4 := arr[1:3]
	fmt.Printf("  arr=%v\n", arr)
	fmt.Printf("  s4=%v (从 arr 截取)\n", s4)
	s4[0] = 999
	fmt.Printf("  s4[0]=999 后: arr=%v (数组也被改)\n", arr)
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - 截取 slice 是 O(1)，不复制元素")
	fmt.Println("   - 共享底层数组：改一个影响另一个")
	fmt.Println("   - append 超过 cap 会切断共享")
	fmt.Println("   - 想要完全独立的 slice：copy copy() 或 append 到 nil")
}

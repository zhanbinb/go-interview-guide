package main

import (
	"fmt"
	"unsafe"
)

// SliceHeader slice 的底层结构（runtime/slice.go）
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

// DemoStruct 演示 slice 三元组 (ptr, len, cap)
//
// slice 本质是这样一个 struct：
//   type SliceHeader struct {
//       Data uintptr  // 指向底层数组
//       Len  int      // 长度（<= Cap）
//       Cap  int      // 容量（从 Data 位置到底层数组末尾）
//   }
func DemoStruct() {
	s := make([]int, 3, 5)
	fmt.Printf("s := make([]int, 3, 5)\n")
	fmt.Printf("  len = %d\n", len(s))
	fmt.Printf("  cap = %d\n", cap(s))
	fmt.Printf("  unsafe.Sizeof(s) = %d 字节（SliceHeader 的大小）\n", unsafe.Sizeof(s))
	fmt.Println()

	// 把 slice 强转为 SliceHeader 看内部
	hdr := (*SliceHeader)(unsafe.Pointer(&s))
	fmt.Printf("SliceHeader:\n")
	fmt.Printf("  Data = 0x%x (底层数组地址)\n", hdr.Data)
	fmt.Printf("  Len  = %d\n", hdr.Len)
	fmt.Printf("  Cap  = %d\n", hdr.Cap)
	fmt.Println()

	// 验证：s[i] 地址 = Data + i*sizeof(elem)
	fmt.Println("验证：&s[i] = Data + i*sizeof(int)")
	for i := 0; i < len(s); i++ {
		addr := uintptr(unsafe.Pointer(&s[i]))
		expected := hdr.Data + uintptr(i)*unsafe.Sizeof(s[0])
		fmt.Printf("  s[%d] 地址: 0x%x, 计算值: 0x%x, 一致: %v\n",
			i, addr, expected, addr == expected)
	}
	fmt.Println()

	// append 演示：填满 + 扩容
	fmt.Println("append 触发扩容:")
	fmt.Printf("  初始: len=%d, cap=%d, ptr=0x%x\n", len(s), cap(s), hdr.Data)
	for i := 0; i < 10; i++ {
		s = append(s, i)
		fmt.Printf("  append %d: len=%d, cap=%d, ptr=0x%x\n",
			i, len(s), cap(s), (*SliceHeader)(unsafe.Pointer(&s)).Data)
	}
	fmt.Println()
	fmt.Println("📌 关键观察:")
	fmt.Println("   - len 满了但 cap 没满时，ptr 不变（不扩容）")
	fmt.Println("   - len 超过 cap 时，ptr 改变（新建底层数组 + 拷贝）")
}

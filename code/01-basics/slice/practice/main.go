package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("【演示 slice 状态】")
	//DemoStruct()
	DemoExpand()
}

type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

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

func DemoExpand() {
	fmt.Println("=== slice 扩容规则 ===")
	fmt.Println()

	// 实验 1：从 cap=0 开始，看每一轮 append 后 cap 的变化
	fmt.Println("【实验 1】从 cap=0 开始扩容:")
	s := []int{}
	prevCap := cap(s)
	for i := 0; i < 20; i++ {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("  append 触发扩容: cap %d → %d\n", prevCap, cap(s))
			prevCap = cap(s)
		}
	}
	fmt.Println()

	// 实验 2：跳到 cap 256 附近，看阶梯增长（让 len=cap 强制扩容）
	fmt.Println("【实验 2】cap ≥ 256 时的阶梯增长:")
	s = make([]int, 256, 256) // len=cap=256, append 一个必扩容
	prevCap = cap(s)
	for i := 0; i < 10; i++ {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("  cap %d → %d (增长 %.2fx)\n",
				prevCap, cap(s), float64(cap(s))/float64(prevCap))
			prevCap = cap(s)
		}
	}
	fmt.Println()

	// 实验 3：直观看到 cap 没满时不扩容（&s[0] 不变）
	fmt.Println("【实验 3】验证：cap 没满时 append 不扩容，&s[0] 不变")
	s = make([]int, 1, 3) // len=1, cap=3，可以 append 2 个不扩容
	fmt.Printf("  初始: len=%d cap=%d, &s[0]=%p\n", len(s), cap(s), &s[0])
	for i := 0; i < 5; i++ {
		old := &s[0]
		s = append(s, i)
		fmt.Printf("  append %d: len=%d cap=%d &s[0]=%p (相同: %v)\n",
			i, len(s), cap(s), &s[0], &s[0] == old)
	}
	fmt.Println()

	// 实验 4：扩展示意图
	fmt.Println("【实验 4】扩容原理示意:")
	fmt.Println("  旧底层数组: [a b c]         cap=3, 满了")
	fmt.Println("  append(d):")
	fmt.Println("    → 算 newcap (cap < 256, newcap = 3*2 = 6)")
	fmt.Println("    → 申请新数组 [    ] (cap=6)")
	fmt.Println("    → 拷贝 a b c 到新数组")
	fmt.Println("    → 写入 d 到 [d]")
	fmt.Println("    → s.ptr 指向新数组")
	fmt.Println()

	// 实验 5：内存对齐的影响（让 len=cap 强制扩容）
	fmt.Println("【实验 5】内存对齐的实际影响:")
	for _, oldCap := range []int{3, 5, 7, 9, 11, 100, 200} {
		s := make([]int, oldCap, oldCap) // len=cap, append 1必扩容
		s = append(s, 0)
		fmt.Printf("  oldCap=%-3d → newCap=%-3d (理论翻倍 = %d)\n",
			oldCap, cap(s), oldCap*2)
	}

	fmt.Println()
	fmt.Println("📌 面试要点:")
	fmt.Println("   - cap < 256：翻倍 (×2)")
	fmt.Println("   - cap ≥ 256：阶梯 ((cap + 3*256) / 4 ≈ ×1.25)")
	fmt.Println("   - 最后按 size class 对齐到 8/16/32/48/64...")
	fmt.Println("   - 扩容代价：分配新数组 + memmove 拷贝数据")
	fmt.Println("   - 预分配 cap 是个优化习惯：make([]int, 0, n)")
}

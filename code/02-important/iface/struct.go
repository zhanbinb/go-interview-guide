package main

import (
	"fmt"
	"unsafe"
)

// iface 实际是 16 字节：两个指针（tab/_type + data）
// 完整定义（runtime/runtime2.go）：
//
//   type iface struct {
//       tab  *itab          // 带方法接口
//       data unsafe.Pointer // 数据指针
//   }
//
//   type eface struct {
//       _type *_type        // 空接口
//       data  unsafe.Pointer
//   }
//
//   type itab = struct {
//       inter  *interfacetype
//       _type  *_type
//       hash   uint32
//       fun    [1]uintptr  // 方法地址数组
//   }

// ifaceRaw 接口变量的原始 16 字节
type ifaceRaw struct {
	Word1 uintptr // tab (iface) 或 _type (eface)
	Word2 uintptr // data
}

// ifacePtr 把任意 interface 转成 *ifaceRaw
func ifacePtr(i any) *ifaceRaw {
	return (*ifaceRaw)(unsafe.Pointer(&i))
}

// DemoStruct 演示 iface/eface 内部结构
func DemoStruct() {
	fmt.Println("=== iface / eface 内部结构 ===")
	fmt.Println()

	// 实验 1：空接口 eface（任何类型）
	fmt.Println("【实验 1】空接口 (any) → 16 字节")
	var i1 any = 42
	raw1 := ifacePtr(i1)
	fmt.Printf("  i1 = 42 (int)\n")
	fmt.Printf("  word1 (类型元数据): 0x%x\n", raw1.Word1)
	fmt.Printf("  word2 (数据指针):   0x%x\n", raw1.Word2)
	fmt.Printf("  sizeof(any): %d 字节 (8+8)\n\n", unsafe.Sizeof(i1))

	// 实验 2：空接口存字符串
	fmt.Println("【实验 2】空接口存 string")
	var i2 any = "hello"
	raw2 := ifacePtr(i2)
	fmt.Printf("  i2 = %q\n", i2)
	fmt.Printf("  word1 (string 类型): 0x%x\n", raw2.Word1)
	fmt.Printf("  word2 (string header): 0x%x\n", raw2.Word2)
	fmt.Println()

	// 实验 3：带方法的接口 iface（复用 polymorphism.go 的类型）
	fmt.Println("【实验 3】带方法的接口 (Speaker) → iface 结构")
	var sp Speaker = Dog{Name: "旺财"}
	raw3 := ifacePtr(sp)
	fmt.Printf("  sp = Speaker(Dog{Name: 旺财})\n")
	fmt.Printf("  word1 (itab):       0x%x (含方法表)\n", raw3.Word1)
	fmt.Printf("  word2 (Dog 数据):   0x%x\n", raw3.Word2)
	fmt.Printf("  sizeof(Speaker): %d 字节\n", unsafe.Sizeof(Speaker(nil)))
	fmt.Println()

	// 实验 4：nil 接口的两个状态
	fmt.Println("【实验 4】nil 接口的两种状态")
	var inil any
	rawNil := ifacePtr(inil)
	fmt.Printf("  var inil any → word1=0x%x, word2=0x%x (完全 nil)\n", rawNil.Word1, rawNil.Word2)

	var p *int
	var ityped any = p
	rawTyped := ifacePtr(ityped)
	fmt.Printf("  var p *int; ityped=nil p → word1=0x%x, word2=0x%x (有类型, 值 nil)\n",
		rawTyped.Word1, rawTyped.Word2)
	fmt.Println()

	// 实验 5：unsafe.Sizeof
	fmt.Println("【实验 5】所有接口变量都是 16 字节")
	fmt.Printf("  sizeof(any)       = %d\n", unsafe.Sizeof(any(0)))
	fmt.Printf("  sizeof(Speaker)   = %d\n", unsafe.Sizeof(Speaker(nil)))
	fmt.Printf("  sizeof(error)     = %d\n", unsafe.Sizeof(error(nil)))
	fmt.Println()

	fmt.Println("📌 关键认知:")
	fmt.Println("   - 接口变量就是两个指针 (16 字节)")
	fmt.Println("   - eface: (type, data) → 装箱任意类型")
	fmt.Println("   - iface: (itab, data) → 带方法类型有额外方法表")
	fmt.Println("   - 调用接口方法 = itab.fun[i] → 函数指针调用")
	fmt.Println()
	fmt.Println("⚠️ unsafe 代码只演示指针位置，不深入解引用")
	fmt.Println("   (runtime itab/_type 内部字段太多，没必要窥探)")
}

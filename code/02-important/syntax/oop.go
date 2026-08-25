package main

import "fmt"

// Animal 动物（基类）
type Animal struct {
	name string // 小写：包内可见
}

// Name 公开方法（首字母大写）
func (a *Animal) Name() string { return a.name }

// Speak 公开方法
func (a *Animal) Speak() string { return "..." }

// SetName 私有方法（首字母小写）
func (a *Animal) setName(n string) { a.name = n }

// Dog 继承 Animal（结构体组合 = has-a）
type Dog struct {
	*Animal      // 嵌入指针，自动获得 Animal 的方法
	bark string
}

// Speak 方法覆盖（多态）
func (d *Dog) Speak() string {
	return d.name + ": 汪！"
}

// DemoOOP 演示 Go 的 OOP 三件套
func DemoOOP() {
	fmt.Println("=== Go 的 OOP 三件套 ===")
	fmt.Println()

	// 1. 封装
	fmt.Println("【1】封装：首字母大小写控制可见性")
	d := &Dog{Animal: &Animal{name: "旺财"}}
	fmt.Printf("  d.Name() = %q (公开方法)\\n", d.Name())
	// d.setName("x") // 编译错，setName 是小写，包外不可访问
	fmt.Println("  小写方法只在包内可见（封装）")
	fmt.Println()

	// 2. 继承（结构体组合 = has-a）
	fmt.Println("【2】继承：结构体组合（不是 is-a，是 has-a）")
	d2 := &Dog{
		Animal: &Animal{name: "小黑"},
		bark:   "Woof!",
	}
	fmt.Printf("  d2.Name() = %q (来自 Animal)\\n", d2.Name())
	fmt.Printf("  d2.Speak() = %q (Dog 自己实现)\\n", d2.Speak())
	fmt.Println()

	// 3. 多态
	fmt.Println("【3】多态：interface 隐式实现")
	type Speaker interface{ Speak() string }
	var s Speaker = d2
	fmt.Printf("  Speaker 接口接受 Dog: %q\\n", s.Speak())
	fmt.Println()

	// 4. Go vs Java 核心区别
	fmt.Println("【4】Go vs Java 核心区别")
	fmt.Println("  - 无 class，只有 struct")
	fmt.Println("  - 无 extends，组合代替继承")
	fmt.Println("  - 无 implements 关键字（鸭子类型）")
	fmt.Println("  - 无 public/private 关键字（首字母大小写）")
	fmt.Println("  - 没有构造函数（用 NewXXX 函数）")
	fmt.Println("  - 没有 try-catch（panic + recover）")
	fmt.Println("  - 没有泛型类（Go 1.18+ 有泛型函数）")
	fmt.Println()

	fmt.Println("📌 设计哲学:")
	fmt.Println("   \"少即是多\"")
	fmt.Println("   Go 故意去掉 OOP 里的复杂特性")
	fmt.Println("   让你用简单组合解决问题")
}

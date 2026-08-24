package main

import "fmt"

// DemoPolymorphism 演示接口多态
//
// Go 的多态：鸭子类型
//   - 不需要显式 implements
//   - 只要实现了接口的所有方法，就"是"这个接口
//   - 函数参数用 interface 接受，可以接受任何实现
//
// 这是 Go 灵活性的核心，也是面试常考点
type Speaker interface {
	Speak() string
}

type Eater interface {
	Eat(food string)
}

type Dog struct{ Name string }

func (d Dog) Speak() string { return d.Name + ": 汪！" }
func (d Dog) Eat(food string) {
	fmt.Printf("  Dog %s 吃 %s\n", d.Name, food)
}

type Cat struct{ Name string }

func (c Cat) Speak() string { return c.Name + ": 喵~" }
func (c Cat) Eat(food string) {
	fmt.Printf("  Cat %s 吃 %s\n", c.Name, food)
}

func DemoPolymorphism() {
	fmt.Println("=== 接口多态 ===")
	fmt.Println()

	// 多态调用
	animals := []Speaker{
		Dog{Name: "旺财"},
		Cat{Name: "小花"},
		Dog{Name: "小黑"},
	}
	fmt.Println("【多态演示】所有动物说话:")
	for _, a := range animals {
		fmt.Println("  " + a.Speak())
	}
	fmt.Println()

	// 接口组合
	fmt.Println("【接口组合】Dog 同时是 Speaker 和 Eater")
	d := Dog{Name: "旺财"}
	var sp Speaker = d
	var e Eater = d
	fmt.Printf("  Speaker: %s\n", sp.Speak())
	e.Eat("骨头")
	fmt.Println()

	// 接口作为参数
	fmt.Println("【实战模式】func feed(e Eater, food string)")
	feed := func(e Eater, food string) {
		e.Eat(food)
	}
	feed(Dog{Name: "旺财"}, "肉")
	feed(Cat{Name: "小花"}, "鱼")
	fmt.Println()

	fmt.Println("📌 多态要点:")
	fmt.Println("   - 不需要 implements 关键字（鸭子类型）")
	fmt.Println("   - 值接收者 vs 指针接收者会影响接口实现")
	fmt.Println("   - 接口可以组合（接口嵌入）")
	fmt.Println("   - 函数参数用 interface 接受 → 接受任何实现")
	fmt.Println()
	fmt.Println("⚠️ 常见坑:")
	fmt.Println("   - 值接收者：T 和 *T 都实现了接口")
	fmt.Println("   - 指针接收者：只有 *T 实现了接口（T 不行）")
}

// 基础语法 + 面向对象演示入口
package main

import (
	"fmt"
	"os"
)

type demo struct {
	name string
	desc string
	fn   func()
}

var demos = []demo{
	{"string", "单/双/反引号", DemoString},
	{"rune", "rune/byte + uint 溢出", DemoRuneByte},
	{"param", "函数参数传值还是传指针", DemoParam},
	{"oop", "封装/组合/多态", DemoOOP},
}

func main() {
	if len(os.Args) < 2 {
		printMenu()
		return
	}
	for _, d := range demos {
		if d.name == os.Args[1] {
			fmt.Printf(">>> %s\n\n", d.desc)
			d.fn()
			return
		}
	}
	fmt.Fprintf(os.Stderr, "未找到: %s\n", os.Args[1])
	os.Exit(1)
}

func printMenu() {
	fmt.Println("基础语法 + 面向对象演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
}

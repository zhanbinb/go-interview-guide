// 接口演示入口
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
	{"struct", "iface / eface 内部结构（unsafe）", DemoStruct},
	{"nil", "nil 接口经典陷阱", DemoNilTrap},
	{"assert", "类型断言 + switch", DemoAssert},
	{"poly", "多态实战", DemoPolymorphism},
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
	fmt.Println("接口演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
}

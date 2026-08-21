// Defer 演示入口
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
	{"order", "LIFO 执行顺序", DemoOrder},
	{"return", "命名 vs 匿名返回值", DemoReturnValue},
	{"closure", "闭包参数陷阱", DemoClosure},
	{"panic", "defer + recover", DemoPanicRecover},
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
	fmt.Println("Defer 演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-8s  %s\n", d.name, d.desc)
	}
}

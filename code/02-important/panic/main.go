// panic/recover 演示入口
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
	{"trigger", "6 种 panic 触发场景", DemoTrigger},
	{"recover", "recover 用法（基本 + re-panic）", DemoRecover},
	{"cross", "跨 goroutine panic 隔离", DemoCross},
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
	fmt.Println("panic/recover 演示")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
}

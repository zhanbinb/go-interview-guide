// 内存分配 + 逃逸分析演示入口
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
	{"allocate", "分配层级 + span class", DemoAllocate},
	{"escape", "5 种逃逸场景", DemoEscape},
	{"leak", "内存泄漏 4 种场景", DemoLeak},
	{"size", "大对象小对象", DemoSizeClass},
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
	fmt.Println("内存分配演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
	fmt.Println("\n提示：跑 escape 后，单独用 go build -gcflags=\"-m\" escape.go 看编译期分析")
}

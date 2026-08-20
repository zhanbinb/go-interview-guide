// Goroutine 演示入口
//
// 运行方式:
//
//	go run .              # 列出所有 demo
//	go run . <demo-name>  # 运行指定 demo
package main

import (
	"fmt"
	"os"
	"strings"
)

type demo struct {
	name string
	desc string
	fn   func()
}

var demos = []demo{
	{"loopvar", "Go 1.22+ 循环变量语义变更（§2 传参陷阱）", DemoLoopVar},
	{"stack", "Goroutine 栈增长（§2 什么是 goroutine）", DemoStackGrowth},
	{"blocking", "5 种阻塞场景（§2 什么情况下阻塞）", DemoBlocking},
	{"leak", "Goroutine 泄漏 4 种模式（§2 传参/泄漏）", DemoLeak},
}

func main() {
	if len(os.Args) < 2 {
		printMenu()
		return
	}

	name := os.Args[1]
	for _, d := range demos {
		if d.name == name {
			fmt.Printf(">>> 运行 demo: %s\n", d.desc)
			fmt.Println(strings.Repeat("-", 60))
			d.fn()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "未找到 demo: %s\n可用: ", name)
	for i, d := range demos {
		if i > 0 {
			fmt.Fprint(os.Stderr, ", ")
		}
		fmt.Fprint(os.Stderr, d.name)
	}
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func printMenu() {
	fmt.Println("Goroutine 演示")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("可用 demo:")
	for _, d := range demos {
		fmt.Printf("  %-12s  %s\n", d.name, d.desc)
	}
	fmt.Println()
	fmt.Println("推荐先看:")
	fmt.Println("  go run . loopvar    # Go 1.22+ 循环变量变更（经典陷阱）")
	fmt.Println("  go run . blocking   # 5 种阻塞模式")
	fmt.Println("  go run . leak       # 4 种泄漏模式（会触发 fatal error）")
}

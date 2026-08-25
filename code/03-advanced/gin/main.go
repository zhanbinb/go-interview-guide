// Gin 风格 demo（标准库实现）
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
	{"router", "简单路由注册", DemoRouter},
	{"middleware", "中间件洋葱模型", DemoMiddleware},
	{"context", "Context 传值", DemoContext},
	{"server", "完整 server 启动", DemoServer},
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
	fmt.Println("Gin 风格演示（标准库实现）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
	fmt.Println("\nGin 真实使用: go get github.com/gin-gonic/gin")
	fmt.Println("本 demo 用标准库演示 Gin 的设计原理")
}

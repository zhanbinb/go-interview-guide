// Channel 演示入口
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
	{"states", "3 状态 × 3 操作行为矩阵（§3 nil/closed/open）", DemoStates},
	{"hchan", "hchan 结构 + unsafe 窥探（§3 底层原理）", DemoHchanStruct},
	{"buffered", "buffered vs unbuffered（§3 有/无缓冲）", DemoBuffered},
	{"select", "select 多路复用 + nil 跳过（§3 select）", DemoSelect},
	{"patterns", "5 种常见模式（§3 使用场景）", DemoPatterns},
	{"leak", "channel 泄漏 4 种场景（§3 资源泄漏）", DemoLeak},
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
	fmt.Println("Channel 演示（Go 面试最重要的题）")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("可用 demo:")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
	fmt.Println()
	fmt.Println("推荐顺序:")
	fmt.Println("  1. states   — 行为矩阵（必看）")
	fmt.Println("  2. hchan    — 底层结构")
	fmt.Println("  3. buffered — buffered vs unbuffered")
	fmt.Println("  4. select   — 多路复用")
	fmt.Println("  5. patterns — 实战模式")
	fmt.Println("  6. leak     — 泄漏场景（会触发 fatal error）")
}

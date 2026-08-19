// GMP 调度模型演示入口
//
// 运行方式:
//   go run .                              # 列出所有 demo
//   go run . <demo-name>                  # 运行指定 demo
//   GODEBUG=schedtrace=1000 go run . ...  # 观察调度 trace
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
	{"gomaxprocs", "GOMAXPROCS 与 NumCPU 的关系（§1 P 的个数）", DemoGOMAXPROCS},
	{"preemptive", "Go 1.14+ 抢占式调度演示（§1 抢占式调度）", DemoPreemptive},
	{"work-steal", "Work Stealing 演示（§1 调度策略）", DemoWorkStealing},
	{"count", "NumGoroutine 变化观察（§1 什么是 goroutine）", DemoGoroutineCount},
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
	fmt.Println("GMP 调度模型演示")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("可用 demo:")
	for _, d := range demos {
		fmt.Printf("  %-15s  %s\n", d.name, d.desc)
	}
	fmt.Println()
	fmt.Println("推荐配合 GODEBUG=schedtrace=1000 观察调度:")
	fmt.Println("  GODEBUG=schedtrace=1000,scheddetail=1 go run . preemptive")
	fmt.Println()
	fmt.Println("或直接指定 demo:")
	fmt.Println("  go run . preemptive")
}

// 锁与原子演示入口
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
	{"mutex", "Mutex 两种模式", DemoMutex},
	{"rwmutex", "RWMutex vs Mutex", DemoRWMutex},
	{"atomic", "atomic 包", DemoAtomic},
	{"ways", "5 种同步方式对比", DemoWays},
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
	fmt.Println("锁与原子演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
}

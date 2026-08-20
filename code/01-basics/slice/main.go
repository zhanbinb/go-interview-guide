// Slice 演示入口
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
	{"struct", "slice 三元组 (ptr, len, cap)", DemoStruct},
	{"expand", "扩容规则（Go 1.18+）", DemoExpand},
	{"param", "函数参数传递（值拷贝 vs 共享）", DemoPassParam},
	{"share", "共享底层数组的副作用", DemoShare},
	{"leak", "大 slice 截取 → 内存泄漏", DemoLeak},
}

func main() {
	if len(os.Args) < 2 {
		printMenu()
		return
	}
	for _, d := range demos {
		if d.name == os.Args[1] {
			fmt.Printf(">>> %s\n", d.desc)
			fmt.Println(strings.Repeat("-", 60))
			d.fn()
			return
		}
	}
	fmt.Fprintf(os.Stderr, "未找到 demo: %s\n", os.Args[1])
	os.Exit(1)
}

func printMenu() {
	fmt.Println("Slice 演示")
	fmt.Println(strings.Repeat("=", 60))
	for _, d := range demos {
		fmt.Printf("  %-8s  %s\n", d.name, d.desc)
	}
	fmt.Println("\n推荐顺序: struct → expand → param → share → leak")
}

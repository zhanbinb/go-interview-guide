// Map 演示入口
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
	{"key", "key 类型限制（可比较类型）", DemoKeyTypes},
	{"nil", "nil map vs 空 map", DemoNilEmpty},
	{"concurrent", "并发写 fatal error + 3 种解法", DemoConcurrent},
	{"hmap", "hmap 结构窥探（unsafe）", DemoHmapStruct},
	{"expand", "扩容机制（增量 + 等量）", DemoExpand},
	{"sync", "sync.Map vs Mutex+map 对比", DemoSyncMap},
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
	fmt.Println("Map 演示")
	fmt.Println(strings.Repeat("=", 60))
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
	fmt.Println("\n推荐顺序: key → nil → concurrent → hmap → expand → sync")
}

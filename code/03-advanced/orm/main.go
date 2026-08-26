// ORM 演示入口
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
	{"model", "模型定义（struct + tag）", DemoModel},
	{"orm", "CRUD 操作", DemoORM},
	{"hooks", "钩子机制", DemoHooks},
	{"tx", "事务", DemoTx},
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
	fmt.Println("ORM 演示（精简版）")
	for _, d := range demos {
		fmt.Printf("  %-10s  %s\n", d.name, d.desc)
	}
	fmt.Println("\n实际项目用 GORM:")
	fmt.Println("  go get gorm.io/gorm")
}

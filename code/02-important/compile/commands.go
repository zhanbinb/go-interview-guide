package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// DemoCommands 演示 go build/install/run 区别
//
// ============================================================================
// 三个常用命令的区别：
//
//   go build  : 编译，但不安装
//              输出: 当前目录的二进制文件（package main 时）
//              默认编译整个包，但保留中间结果在 build cache
//
//   go install: 编译 + 安装
//              输出: 放到 GOPATH/bin
//              适合: 工具安装（go install golang.org/x/tools/...）
//
//   go run    : 编译 + 立即运行
//              临时二进制在临时目录
//              适合: 开发调试（每次重新编译）
// ============================================================================
func DemoCommands() {
	fmt.Println("=== go build / install / run 区别 ===")
	fmt.Println()

	// 实验 1：go build（编译到当前目录）
	fmt.Println("【实验 1】go build -o /tmp/demo-build")
	cmd := exec.Command("go", "build", "-o", "/tmp/demo-build")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("  错误: %v\n", err)
	} else {
		fmt.Println("  ✓ 编译成功")
		if len(out) > 0 {
			fmt.Printf("  输出: %s\n", strings.TrimSpace(string(out)))
		}
	}
	fmt.Printf("  生成二进制: /tmp/demo-build (用户决定输出位置)\n\n")

	// 实验 2：go install（编译 + 安装到 GOPATH/bin）
	fmt.Println("【实验 2】go install（演示用，不真安装避免污染）")
	fmt.Println("  命令: go install . (会把二进制装到 GOPATH/bin)")
	fmt.Println("  实际生产用法:")
	fmt.Println("    go install golang.org/x/tools/cmd/godoc@latest")
	fmt.Println("    → 在 GOPATH/bin 生成 godoc 可执行文件")

	// 实验 3：go run（编译 + 立即运行）
	fmt.Println("【实验 3】go run . （用于开发）")
	fmt.Println("  命令: go run . gopath")
	fmt.Println("  临时编译在 $TMPDIR/go-build.../")
	fmt.Println("  不会污染当前目录")
	fmt.Println()

	// 实验 4：build cache 验证
	fmt.Println("【实验 4】build cache（重复构建会复用）")
	out1, _ := exec.Command("go", "build", "-a", "-o", "/tmp/demo-a", ".").CombinedOutput()
	out2, _ := exec.Command("go", "build", "-o", "/tmp/demo-b", ".").CombinedOutput()
	fmt.Printf("  第一次（强制重编）: %s\n", trimOrEmpty(out1))
	fmt.Printf("  第二次（用 cache）:  %s\n", trimOrEmpty(out2))
	fmt.Println("  → 第二次明显更快（用了 cache）")

	fmt.Println("📌 命令选择:")
	fmt.Println("   - 开发调试 → go run")
	fmt.Println("   - CI/部署 → go build -o binary")
	fmt.Println("   - 安装工具 → go install")
	fmt.Println("   - 跨平台编译 → GOOS=linux GOARCH=amd64 go build")
}

func trimOrEmpty(s []byte) string {
	t := strings.TrimSpace(string(s))
	if t == "" {
		return "(无)"
	}
	return t
}

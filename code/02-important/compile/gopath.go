package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DemoGOPATH 演示 GOROOT / GOPATH / go.mod 的关系
//
// ============================================================================
// 三个核心路径：
//
//   GOROOT  Go 安装路径（自带的 src/pkg）
//            例: /usr/local/go
//            内容: 标准库源码 + 编译器
//
//   GOPATH  旧版工作目录（pre-modules 时代唯一方案）
//            作用: 存放 go install 安装的二进制
//            现在的项目不再依赖 GOPATH 路径
//
//   go.mod  Go 1.11+ 引入的模块系统
//            作用: 项目级依赖管理
//            推荐: 所有新项目都用 go.mod
// ============================================================================
func DemoGOPATH() {
	fmt.Println("=== GOROOT / GOPATH / go.mod ===")
	fmt.Println()

	// 实验 1：查看 GOROOT
	fmt.Println("【实验 1】GOROOT（Go 安装路径）")
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		out, _ := exec.Command("go", "env", "GOROOT").Output()
		goroot = strings.TrimSpace(string(out))
	}
	fmt.Printf("  GOROOT = %s\n", goroot)
	fmt.Println("  → 编译器、汇编器、链接器、标准库源码都在这")
	fmt.Println()

	// 实验 2：查看 GOPATH
	fmt.Println("【实验 2】GOPATH（旧版工作目录）")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		out, _ := exec.Command("go", "env", "GOPATH").Output()
		gopath = strings.TrimSpace(string(out))
	}
	fmt.Printf("  GOPATH = %s\n", gopath)
	if gopath != "" {
		binPath := filepath.Join(gopath, "bin")
		fmt.Printf("  GOPATH/bin = %s\n", binPath)
		fmt.Println("  → go install 把二进制安装到这里")
	}
	fmt.Println()

	// 实验 3：查看 go env 关键变量
	fmt.Println("【实验 3】go env 关键变量")
	out, _ := exec.Command("go", "env").Output()
	envLines := strings.Split(string(out), "\n")
	for _, line := range envLines {
		if line == "" {
			continue
		}
		// 只显示关键变量
		key := strings.SplitN(line, "=", 2)[0]
		if key == "GOOS" || key == "GOARCH" ||
			key == "GO111MODULE" || key == "GOMODCACHE" ||
			key == "GOPROXY" || key == "CGO_ENABLED" {
			fmt.Printf("  %s\n", line)
		}
	}
	fmt.Println()

	// 实验 4：go.mod
	fmt.Println("【实验 4】go.mod（项目级依赖管理）")
	out, _ = exec.Command("cat", "go.mod").Output()
	fmt.Printf("  当前 go.mod:\n")
	for _, line := range strings.Split(string(out), "\n") {
		if line != "" {
			fmt.Printf("    %s\n", line)
		}
	}
	fmt.Println()

	fmt.Println("📌 实战建议:")
	fmt.Println("   - GOROOT: 一般不修改，使用默认")
	fmt.Println("   - GOPATH: 现在只是 go install 的目标")
	fmt.Println("   - go.mod: 现代 Go 项目必备")
	fmt.Println("   - GO111MODULE=on: 强制使用 modules（默认）")
}

package main

import (
	"fmt"
	"runtime"
)

// DemoProcess 演示编译链接过程 + 程序启动过程
//
// ============================================================================
// 编译链接流程（cmd/compile）：
//
//   1. 词法分析 → Token
//   2. 语法分析 → AST
//   3. 类型检查 → 类型化 AST
//   4. 中间代码生成 → SSA IR
//   5. 优化（内联、死代码消除等）
//   6. 机器码生成 → .o 文件
//   7. 链接（cmd/link）→ 单一可执行二进制
//
// 程序启动流程（runtime）：
//   1. OS 加载 ELF/Mach-O 二进制
//   2. 跳到入口点（汇编 _rt0_amd64_linux 等）
//   3. 初始化 runtime（g0、m0）
//   4. 初始化 GC、栈分配器、调度器
//   5. 启动第一个 goroutine 运行用户 main()
//   6. main() 返回 → runtime 清理退出
// ============================================================================
func DemoProcess() {
	fmt.Println("=== 编译链接 + 启动过程 ===")
	fmt.Println()

	// 实验 1：查看编译过程
	fmt.Println("【实验 1】编译过程示意")
	fmt.Println("  源码: fmt.Println(\"hello\")")
	fmt.Println()
	fmt.Println("  1. 词法分析  →  [fmt, ., Println, (, \"hello\", ), ;]")
	fmt.Println("  2. 语法分析  →  AST (表达式语句)")
	fmt.Println("  3. 类型检查  →  类型化 AST")
	fmt.Println("  4. SSA 中间码 →  Println.Call(hello)")
	fmt.Println("  5. 优化      →  内联 fmt.Println")
	fmt.Println("  6. 机器码    →  amd64 指令序列")
	fmt.Println("  7. 链接      →  包含 runtime + fmt + 你代码的单一二进制")
	fmt.Println()

	// 实验 2：程序启动流程
	fmt.Println("【实验 2】程序启动流程")
	fmt.Println("  1. OS loader 加载 ELF/Mach-O 二进制")
	fmt.Println("  2. 跳到入口: runtime/rt0_xxx.s (汇编)")
	fmt.Println("  3. 初始化 g0 (初始 goroutine 栈)")
	fmt.Println("  4. 初始化 m0 (主 OS 线程)")
	fmt.Println("  5. 初始化 runtime: GC / 调度器 / P")
	fmt.Println("  6. 启动 user main goroutine")
	fmt.Println("  7. main() 返回 → runtime.exit()")
	fmt.Println()

	// 实验 3：runtime 自检
	fmt.Println("【实验 3】查看当前 runtime 信息")
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  GOOS:       %s\n", runtime.GOOS)
	fmt.Printf("  GOARCH:     %s\n", runtime.GOARCH)
	fmt.Printf("  NumCPU:     %d\n", runtime.NumCPU())
	fmt.Printf("  NumGoroutine: %d\n", runtime.NumGoroutine())
	fmt.Println()

	fmt.Println("📌 关键源码:")
	fmt.Println("   - 编译入口: src/cmd/compile/internal/gc")
	fmt.Println("   - 链接入口: src/cmd/link/internal/ld")
	fmt.Println("   - 运行时:   src/runtime")
	fmt.Println("   - 汇编入口: src/runtime/rt0_xxx.s")
	fmt.Println()
	fmt.Println("📌 编译优化:")
	fmt.Println("   - 内联 (inlining)")
	fmt.Println("   - 死代码消除")
	fmt.Println("   - 逃逸分析（决定栈/堆）")
	fmt.Println("   - SSA 优化（constant folding, dead code 等）")
}

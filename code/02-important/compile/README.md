# Go 编译演示（精简版）

> 对应复习指南：§15 Go 编译与 go tool

## 面试只需了解 4 个点

1. GoRoot vs GoPath vs Go modules
2. 编译链接过程（源码 -> 编译 -> 链接 -> 二进制）
3. go build / go install / go run 区别
4. Go 程序启动过程（runtime 初始化）

## 文件清单

- main.go: 菜单
- gopath.go: GOROOT / GOPATH / go.mod 关系
- commands.go: go build/install/run 区别
- process.go: 编译链接过程示意
- compile_test.go: 测试

## 跑

    make run TOPIC=compile
    make run TOPIC=compile DEMO=path
    make run TOPIC=compile DEMO=cmd
    make run TOPIC=compile DEMO=process
    make test TOPIC=compile

## 面试速记

GOROOT / GOPATH / go.mod:
  - GOROOT:  Go 安装目录（自带标准库）
  - GOPATH:  旧版依赖模式的工作目录（现在放 go install 的二进制）
  - go.mod:  Go 1.11+ 推荐的项目依赖管理

编译链接：
  词法/语法分析 → AST → 类型检查 → SSA 中间码 → 机器码 → 链接

go build/install/run:
  - go build: 编译但不安装（默认输出到当前目录）
  - go install: 编译 + 安装到 GOPATH/bin
  - go run: 编译 + 立即运行（开发用）

启动过程：
  - 操作系统加载二进制
  - 跳到 runtime.rt0_go（汇编入口）
  - 初始化 runtime（goroutine 栈、GC、P）
  - 执行用户 main 函数
  - 程序退出，runtime 清理


# Goroutine 演示

> 📖 对应复习指南：[指南 §2 Goroutine](../../../docs/Go-Interview-Study-Guide.md#2-goroutine)

## 🎯 学习目标

通过亲手验证，掌握以下 goroutine 核心特性：

1. **轻量**：初始栈 2KB，按需动态增长
2. **Go 1.22+ 循环变量**：每次迭代独立变量，不再共享
3. **多种阻塞方式**：channel、锁、syscall、select 都会让 G 阻塞（但 M 不一定阻塞）
4. **goroutine 泄漏**：常见 4 种模式 + 检测方法

## 📁 文件清单

| 文件 | 演示内容 | 对应面试题 |
|------|---------|------------|
| `main.go` | 入口 + demo 菜单 | - |
| `loopvar.go` | Go 1.22+ 循环变量语义变更 | "for range 时地址会变化么？" |
| `stack_growth.go` | 初始栈 2KB + 动态增长 | "goroutine 跟线程的区别？" |
| `blocking.go` | 各种阻塞场景 | "goroutine 什么情况下会阻塞？" |
| `leak.go` | goroutine 泄漏 4 种常见模式 | "goroutine 创建时传参注意点？" |
| `goroutine_test.go` | 单元测试 + 泄漏检测 | - |

## 🚀 怎么跑

### Makefile（推荐）

```bash
# 从项目根目录
make run TOPIC=goroutine                       # 菜单
make run TOPIC=goroutine DEMO=loopvar          # 循环变量语义
make run TOPIC=goroutine DEMO=stack            # 栈增长
make run TOPIC=goroutine DEMO=blocking         # 阻塞场景
make run TOPIC=goroutine DEMO=leak             # 泄漏演示
make test TOPIC=goroutine                      # 跑测试
```

### 直接 go run

```bash
cd code/01-basics/goroutine
go run .
go run . loopvar
```

## 🧪 实验详解

### 实验 1：循环变量语义（必看）

```bash
make run TOPIC=goroutine DEMO=loopvar
```

**背景**：Go 1.22 之前，`for i := 0; i < N; i++` 的 `i` 在**整个循环里共享一个地址**，所有 goroutine 闭包捕获的都是同一个变量**，最终几乎都读到 `N`。老代码靠"作为参数传入"绕过：

```go
for i := 0; i < 3; i++ {
    go func(i int) { fmt.Println(i) }(i)  // 老写法：必须传参
}
```

**Go 1.22 起**：每次迭代的 `i` 都是独立变量，闭包捕获自动正确：

```go
for i := 0; i < 3; i++ {
    go func() { fmt.Println(i) }()  // 现在也安全了
}
```

本 demo 会同时跑两种写法（slice / for-range），对比看结果。

### 实验 2：栈增长

```bash
make run TOPIC=goroutine DEMO=stack
```

观察：
- 1000 个 goroutine 创建前后 `HeapAlloc` 几乎不变（栈复用）
- 用 `runtime.Stack()` 抓取 goroutine 栈，看 `goroutine N [running]:` 格式
- 关键事实：**goroutine 初始栈仅 2KB**，最大可到 1GB（`GOMAXSTACKSIZE` 控制）

### 实验 3：阻塞场景

```bash
make run TOPIC=goroutine DEMO=blocking
```

依次演示 5 种阻塞：
1. **channel 收发阻塞**（无缓冲 channel 必须配对）
2. **Mutex 阻塞**（goroutine 等锁时会让出 P）
3. **select 阻塞**（多 channel 等其中一个就绪）
4. **time.Sleep / syscall 阻塞**（特殊：会让 M 进内核）
5. **死循环**（Go 1.14+ 也会被抢占）

### 实验 4：goroutine 泄漏

```bash
make run TOPIC=goroutine DEMO=leak
```

故意制造 4 种泄漏，看 `runtime.NumGoroutine()` 增长：

1. **channel 没人接收**
2. **context 没取消，子 goroutine 永远等 `ctx.Done()`**
3. **WaitGroup.Add 但没 Done**
4. **死循环 + 无退出条件**

主 goroutine 最后也 `select{}` 阻塞 → 触发 runtime "all goroutines asleep" → fatal error。

## 🔧 检测 goroutine 泄漏

实际项目里可以用 [`uber-go/goleak`](https://github.com/uber-go/goleak) 在测试里检测：

```go
func TestNoLeak(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ... 你的代码
}
```

## 📌 面试速记卡

```
goroutine = 用户态协程（G），初始栈 2KB，最大 1GB
go func() 创建 G，调度由 runtime 负责
Go 1.22+ 循环变量每次独立（老版本需作为参数传入）
阻塞不阻塞 M：channel/锁/select → G 阻塞但 M 可执行别的
            syscall（网络/文件 I/O）→ M 进内核，netpoller 会处理
泄漏检测：runtime.NumGoroutine()、goleak、pprof goroutine profile
```

# panic / recover 演示（精简版）

> 对应复习指南：§14 panic/recover

## 面试只需掌握 4 个点

1. panic 触发场景：数组越界、nil 指针、类型断言失败、close closed channel
2. recover 必须在 defer 里，且只能捕获当前 goroutine
3. 跨 goroutine panic 隔离：一个 goroutine panic 不影响其他
4. re-panic 模式：recover 后重新 panic（库代码用）

## 文件清单

- main.go: 菜单
- trigger.go: 6 种 panic 触发场景
- recover.go: recover 用法（基本 + re-panic）
- cross.go: 跨 goroutine panic 隔离
- panic_test.go: 测试

## 跑

    make run TOPIC=panic
    make run TOPIC=panic DEMO=trigger
    make run TOPIC=panic DEMO=recover
    make run TOPIC=panic DEMO=cross
    make test TOPIC=panic

## 面试速记

panic 触发场景：
  - 数组越界（运行时检测）
  - nil 指针解引用
  - 类型断言失败（不带 , ok 形式）
  - 关闭已关闭的 channel
  - 向已关闭的 channel 发送
  - map 并发读写（fatal error，不可 recover）

recover 要点：
  - 必须在 defer 里调用
  - 只能恢复当前 goroutine 的 panic
  - 返回 panic 的值
  - 库代码要慎重 recover（可能吞掉 panic）
  - recover 后可重新 panic 保留堆栈

跨 goroutine：
  - 每个 goroutine 需要独立的 defer recover
  - 子 goroutine panic 不影响主 goroutine
  - errgroup 库处理这个场景

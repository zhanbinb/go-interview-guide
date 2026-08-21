# Context 演示（精简版）

> 📖 对应复习指南：[指南 §9 Context](../../../docs/Go-Interview-Study-Guide.md#9-context)

## 🎯 面试只需掌握 4 个点

1. **Context 接口**：4 个方法（Deadline / Done / Err / Value）
2. **创建**：Background / TODO / WithCancel / WithTimeout / WithDeadline / WithValue
3. **级联取消**：父 ctx 取消 → 子 ctx 自动取消
4. **Value 传递**：用 WithValue 存请求范围数据（traceID、userID）

## 📁 文件清单（每个文件 < 50 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `struct.go` | Context 接口 + 4 个方法 |
| `functions.go` | 6 个创建函数演示 |
| `patterns.go` | 4 种实战模式 |
| `gotchas.go` | 3 个常见陷阱 |
| `context_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=context
make run TOPIC=context DEMO=struct
make run TOPIC=context DEMO=functions
make run TOPIC=context DEMO=patterns
make run TOPIC=context DEMO=gotchas
make test TOPIC=context
```

## 📌 面试速记

```
Context 接口（4 个方法）：
  Deadline() (time.Time, bool)   // 返回截止时间
  Done() <-chan struct{}         // 返回取消信号 channel
  Err() error                     // 返回取消原因
  Value(key any) any              // 返回 key 对应的 value

创建（6 个）：
  Background()    根 ctx（永不取消）
  TODO()          占位（不知道用啥时）
  WithCancel(parent)    手动 cancel
  WithTimeout(parent, d) 超时自动 cancel
  WithDeadline(parent, t) 到时间自动 cancel
  WithValue(parent, k, v) 存数据

使用规范：
  - 第一参数：ctx context.Context
  - 不要放在结构体里，显式传
  - 不要传传可选参数（应该用函数参数）
  - defer cancel() 避免泄漏
```

## 🔑 实战模式

```go
// 1. HTTP 请求超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

// 2. goroutine 协调
ctx, cancel := context.WithCancel(context.Background())
go func() {
    select {
    case <-ctx.Done():
        return  // 收到取消信号
    case <-work:
        // 处理
    }
}()

// 3. 传值（traceID）
ctx := context.WithValue(parent, "trace_id", traceID)

// 4. 超时 + 级联取消
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()
// 子 goroutine 也会被取消
```

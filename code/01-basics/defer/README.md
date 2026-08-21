# Defer 演示（精简版）

> 📖 对应复习指南：[指南 §6 Defer](../../../docs/Go-Interview-Study-Guide.md#6-defer)

## 🎯 面试只需掌握 4 个点

1. **执行顺序**：LIFO（后进先出，像栈）
2. **执行时机**：在 `return` 之后（具体是 return value 赋值之后）
3. **能否修改返回值**：命名返回值 → ✅；匿名返回值 → ❌
4. **闭包陷阱**：参数立即求值 vs 闭包延迟求值

加 1 个常用模式：
- **`defer + recover`**：捕获 panic（库代码慎用）

## 📁 文件清单（每个文件 < 40 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `order.go` | LIFO 顺序 |
| `return.go` | 匿名 vs 命名返回值 |
| `closure.go` | 闭包参数陷阱 |
| `panic.go` | defer + recover |
| `defer_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=defer                       # 菜单
make run TOPIC=defer DEMO=order
make run TOPIC=defer DEMO=return
make run TOPIC=defer DEMO=closure
make run TOPIC=defer DEMO=panic
make test TOPIC=defer
```

## 📌 面试速记

```
执行顺序：LIFO（多个 defer 反序执行）
执行时机：return value 赋值之后，函数真正返回之前
返回值：
  - func f() int       { defer ...; return 1 }   defer 改不了
  - func f() (r int)   { defer ...; return 1 }   defer 能改 r
闭包：
  - defer f(x)         立即求值，x 是当时的值
  - defer func() { ... }()   延迟求值，能看到最终值
recover：必须在 defer 里调用，且只能恢复当前 goroutine 的 panic
```

# GC 演示（精简版）

> 📖 对应复习指南：[指南 §7 GC](../../../docs/Go-Interview-Study-Guide.md#7-gc-三色标记--混合写屏障)

## 🎯 面试只需了解 4 个点

1. **触发时机**：堆翻倍 / 2 分钟 / 手动 `runtime.GC()` / 内存上限
2. **三色标记**：白/灰/黑 + 流程（为什么需要它）
3. **写屏障**：用户程序和 GC 并发时，怎么保证不漏标对象
4. **混合屏障**（Go 1.8+）：同时用插入 + 删除屏障，几乎消除了 STW

## 📁 文件清单（每个文件 < 40 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `trigger.go` | GC 触发时机 |
| `tricolor.go` | 三色标记示意 |
| `barrier.go` | 写屏障是什么、为什么需要 |
| `gc_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=gc
make run TOPIC=gc DEMO=trigger
make run TOPIC=gc DEMO=tricolor
make run TOPIC=gc DEMO=barrier
make test TOPIC=gc
```

## 📌 面试速记

```
GC 算法：Go 用的是 **并发三色标记 + 混合写屏障**

触发时机（满足任一）：
  - 堆内存相比上次 GC 翻了 GOGC 倍（默认 100，即翻倍）
  - 2 分钟内没 GC
  - 手动 runtime.GC()
  - Go 1.21+: 达到 GOMEMLIMIT（软限制）

三色标记：
  - 白：未扫描
  - 灰：已标记，子引用未扫描
  - 黑：已标记，子引用已扫描
  → 从根出发扫描，最后白色对象被回收

为什么需要写屏障：
  GC 和用户程序并发执行时，用户程序可能：
  - A. 把白色对象赋值给黑色对象（漏标 → 错误回收）
  - B. 把黑色对象指向的白色对象断开（漏标 → 错误回收）
  写屏障就是在读写时插入一段代码，通知 GC 处理。

混合写屏障（Go 1.8+）：
  - 插入屏障：被引用时立即标灰（Dijkstra）
  - 删除屏障：引用断开时标灰（Yuasa）
  - 优势：消除 STW 重新标记阶段
```

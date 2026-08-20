# Channel 演示

> 📖 对应复习指南：[指南 §3 Channel](../../../docs/Go-Interview-Study-Guide.md#3-channel)

## 🎯 Channel — Go 面试最重要的一题

Channel 是 Go 实现 "**通过通信共享内存**"（vs Java 通过共享内存通信）的核心原语，也是面试必问之必问。

## 📚 速记：3 状态 × 3 操作的行为矩阵（"15字口诀"）

| 操作 | nil channel | closed channel | open channel（有数据） | open channel（空/满） |
|------|:---:|:---:|:---:|:---:|
| **读 <-ch** | 永久阻塞 | 返回零值 + `ok=false` | 读到值 | 阻塞（空） |
| **写 ch<-** | 永久阻塞 | **panic: send on closed** | 阻塞（满） | 写入成功 |
| **close(ch)** | **panic: close of nil** | **panic: close of closed** | 可关闭 | 可关闭 |

## 🎯 学习目标

通过亲手验证，掌握：

1. **3 状态 × 3 操作**：上面那张表
2. **hchan 数据结构**：runtime 内部的 channel 实现
3. **buffered vs unbuffered**：同步 vs 异步语义
4. **select**：多 channel 等待 + nil channel 跳过 + timeout
5. **5 种常见模式**：producer-consumer / fan-in / fan-out / pipeline / quit
6. **channel 泄漏**：4 种典型场景

## 📁 文件清单

| 文件 | 演示内容 | 对应面试题 |
|------|---------|------------|
| `main.go` | 入口 + 菜单 | - |
| `states.go` | 3 状态 × 3 操作行为矩阵 | "nil/closed/open channel 读写关闭行为？" |
| `hchan_struct.go` | hchan 结构 + unsafe 窥探内部字段 | "channel 底层实现原理？" |
| `buffered.go` | buffered vs unbuffered vs cap=1 信号量 | "有缓存和无缓存的区别？" |
| `select_demo.go` | select 各种 case + nil channel 跳过 | "select 底层数据结构" |
| `patterns.go` | 5 种常见模式 | "channel 主要使用场景？" |
| `leak.go` | channel 泄漏 4 种场景 | "channel 资源泄漏" |
| `channel_test.go` | 单元测试 + benchmark | - |

## 🚀 怎么跑

```bash
# 从项目根目录
make run TOPIC=channel                       # 菜单
make run TOPIC=channel DEMO=states          # 行为矩阵（必看！）
make run TOPIC=channel DEMO=hchan           # hchan 结构
make run TOPIC=channel DEMO=buffered        # buffered vs unbuffered
make run TOPIC=channel DEMO=select          # select
make run TOPIC=channel DEMO=patterns        # 5 种模式
make run TOPIC=channel DEMO=leak            # 泄漏（会触发 fatal error）
make test TOPIC=channel                     # 测试
make bench TOPIC=channel                    # benchmark
```

## 🧪 实验详解

### 实验 1：行为矩阵（必看）

```bash
make run TOPIC=channel DEMO=states
```

每一行演示一种 "状态 × 操作" 组合。所有"永久阻塞"的场景用 goroutine + timeout 模拟，不会真的卡住。

重点观察：
- **nil channel 的所有操作都阻塞**（这个特性有用：用在 select 里跳过该 case）
- **写 closed channel 立即 panic**
- **重复 close 立即 panic**
- **从 closed channel 读不会 panic**，返回零值（int→0, ptr→nil, bool→false）

### 实验 2：hchan 结构

```bash
make run TOPIC=channel DEMO=hchan
```

用 `unsafe.Pointer` 窥探 `hchan` 内部字段：
- `qcount` / `dataqsiz` / `buf` / `closed` / `sendx` / `recvx`
- 验证 buffered vs unbuffered 的 `dataqsiz` 差异
- 验证 close 后 `closed` 字段变成 1

⚠️ unsafe 代码依赖 Go runtime 内部布局，**不同版本可能不同**，demo 里做了版本提示。

### 实验 3：buffered vs unbuffered

```bash
make run TOPIC=channel DEMO=buffered
```

对比 3 种情况：
- `make(chan int)` — unbuffered（同步）
- `make(chan int, 1)` — buffered cap=1（可当信号量）
- `make(chan int, 3)` — buffered cap=3（队列）

演示发送/接收完成时间，理解异步/同步语义。

### 实验 4：select

```bash
make run TOPIC=channel DEMO=select
```

4 个小场景：
1. select 等第一个 ready case
2. select + default（非阻塞）
3. select + time.After（超时）
4. **nil channel 在 select 中被跳过**（经典用法：用 nil 关闭 case）

### 实验 5：patterns

```bash
make run TOPIC=channel DEMO=patterns
```

5 种经典模式：
1. **Producer-Consumer**：1 producer + 1 consumer
2. **Fan-out**：1 producer → N consumer（负载分摊）
3. **Fan-in**：N producer → 1 consumer
4. **Pipeline**：阶段式数据流（带 cancellation）
5. **Quit 信号**：`done <-chan struct{}` 协调退出

### 实验 6：leak

```bash
make run TOPIC=channel DEMO=leak
```

4 种泄漏场景：
1. 向无缓冲 channel 发送，没 receiver
2. 从 channel 接收，但 sender 早早 return
3. buffered channel 持续生产，没人消费
4. goroutine 等待永远不会被关闭的 channel

最后主 goroutine 也 `select{}` → 触发 fatal error "all goroutines asleep"。

## 📌 面试速记卡

```
channel = Go 通信原语（hchan 结构 + 环形队列 + sendq/recvq + mutex）
底层 hchan 包含：buf (环形队列) + sendx/recvx (索引) + sendq/recvq (等待队列)
线程安全：hchan 内部有 mutex 保护
unbuffered：同步（sender 等 receiver）
buffered  ：异步（缓冲未满时 sender 不阻塞）
关闭原则：谁创建谁关闭；多个 receiver 时用 done 协调
nil channel 妙用：select 中用作"禁用 case"
资源泄漏：永久阻塞的收发、忘记关闭 channel、context 没取消
```

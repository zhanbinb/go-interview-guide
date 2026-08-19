# GMP 调度模型 Demo

> 📖 对应复习指南：[指南 §1 GMP 调度模型](../../../docs/Go-Interview-Study-Guide.md#1-gmp-调度模型)

## 🎯 学习目标

亲手验证 GMP 调度模型的几个关键结论：

1. **GOMAXPROCS = P 的个数**（默认 = NumCPU，可手动修改）
2. **M 的个数 ≥ P**（syscall / 阻塞时会临时创建新 M）
3. **Go 1.14+ 抢占式调度**（死循环也会被强制让出 CPU）
4. **Work Stealing**：空闲 P 会从其他 P 的 LRQ 偷一半 G
5. **Goroutine 栈动态增长**：从 2KB 起步，按需扩展到 GB 级

## 📁 文件清单

| 文件 | 演示内容 | 对应面试题 |
|------|---------|------------|
| `main.go` | 入口 + demo 菜单 | - |
| `gomaxprocs.go` | GOMAXPROCS / NumCPU 关系 | "G、P、M 的个数问题？" |
| `preemptive.go` | Go 1.14+ 抢占式调度 | "抢占式调度是如何抢占的？" |
| `work_stealing.go` | Work Stealing 观察 | "调度器的设计策略？" |
| `goroutine_count.go` | NumGoroutine 计数 | "什么是 goroutine？" |
| `gmp_test.go` | 单元测试 + benchmark | - |

## 🚀 怎么跑

### 启动菜单

```bash
cd code/01-basics/gmp
go run .
```

### 指定 demo

```bash
go run . preemptive      # 抢占式调度
go run . work-steal      # Work Stealing
go run . gomaxprocs      # P 的个数
go run . count           # NumGoroutine 变化
```

### 观察调度 trace（重点！）

```bash
GODEBUG=schedtrace=1000 go run .
GODEBUG=schedtrace=1000,scheddetail=1 go run .
```

输出示例：

```
SCHED 0ms: gomaxprocs=8 idleprocs=6 threads=4 spinning=0 ...
SCHED 1ms: gomaxprocs=8 idleprocs=2 threads=8 spinning=2 ...
```

字段含义：
- `gomaxprocs` — P 的数量
- `idleprocs` — 空闲的 P 数量
- `threads` — M 的总数量（**注意可能 > gomaxprocs**）
- `spinning` — 自旋等待的 M 数量

## 🧪 实验步骤

### 实验 1：验证 GOMAXPROCS

```bash
go run . gomaxprocs
```

观察：
- NumCPU = 8，GOMAXPROCS = 8（默认）
- 修改 `gomaxprocs.go` 里的 `runtime.GOMAXPROCS(2)`，再次运行

### 实验 2：验证抢占式调度

```bash
go run . preemptive
```

预期：另一个 goroutine 大约 500ms 后被调度执行。
- 如果耗时 ~500ms → Go 1.14+ 抢占式调度生效 ✅
- 如果耗时 >5s → Go 版本 < 1.14（协作式调度）⚠️

### 实验 3：观察 Work Stealing

```bash
GODEBUG=schedtrace=1000,scheddetail=1 go run . work-steal
```

观察两个 P 的 LRQ 长度变化：空 P 会从满 P 偷走一半 G。

### 实验 4：观察 M > P

```bash
GODEBUG=schedtrace=100 go run . gomaxprocs
```

当 goroutine 进入 syscall（如 `time.Sleep`）时，会创建新的 M 处理其他 P 的任务，threads 数量可能 > gomaxprocs。

## 📌 面试速记

详见复习指南 §1。核心结论：

```
P 的个数 = GOMAXPROCS（默认 NumCPU）
M 的个数 ≥ GOMAXPROCS（syscall 时可临时多）
go 关键字 → G 优先入 LRQ，LRQ 满则入 GRQ
空闲 P → 每 61 次调度查 GRQ，否则 work-steal
sysmon → 10ms 检测强制抢占（Go 1.14+）
```

## 🛠 Makefile 一键运行（推荐）

项目根目录的 Makefile 已经封装好 GOCACHE 等环境变量：

```bash
# 从项目根目录
make run DEMO=preemptive     # 跑指定 demo
make run DEMO=work-steal
make run DEMO=gomaxprocs
make run DEMO=count
make test                    # 跑测试
make bench                   # 跑 benchmark
make trace DEMO=preemptive   # 跑 demo + 抓调度 trace
make help                    # 查看所有命令
```

也可以直接用 `go run`（需要本地 Go 缓存可写）：

```bash
cd code/01-basics/gmp
go run . preemptive
```

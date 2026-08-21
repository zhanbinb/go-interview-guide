# Go 面试复习指南

> 📚 基于 [mao888/golang-guide](https://github.com/mao888/golang-guide/blob/main/golang/go-Interview/GOALNG_INTERVIEW_COLLECTION.md) 整理
>
> 每个主题都有可运行的 demo + 测试 + 面试速记，应付 Go 后端面试。

## 🎯 项目特点

- **9 大核心主题**：GMP / Goroutine / Channel / Slice / Map / Defer / GC / Memory / Context
- **每个主题独立**：可单独跑、单独测、单独提交
- **代码优先**：每个 demo 都可直接运行看输出，比背概念更扎实
- **面试导向**：每个 demo 最后都给了"面试速记卡"
- **可一键运行**：`make` 命令封装好 GOCACHE 等环境变量

## 🚀 快速开始

```bash
# 列出所有 topic
make list

# 跑某个 topic 的菜单
make run TOPIC=gmp
make run TOPIC=channel
make run TOPIC=map

# 跑指定 demo
make run TOPIC=gmp DEMO=preemptive
make run TOPIC=channel DEMO=states

# 跑测试 / benchmark
make test TOPIC=gmp
make bench TOPIC=channel
```

## 📁 项目结构

```
go-interview-guide/
├── README.md              ← 本文件
├── Makefile               ← 一键运行（封装 GOCACHE）
├── go.mod                 ← Go 1.25.5 module
│
├── docs/
│   └── Go-Interview-Study-Guide.md   ← 整理好的面试指南（956 行）
│                                  每个小节都有原文锚点 + 推荐阅读 + 源码位置
│
└── code/01-basics/
    ├── gmp/                ← GMP 调度模型（🔥必问）
    ├── goroutine/          ← Goroutine（🔥必问）
    ├── channel/            ← Channel（🔥必问，含 hchan unsafe 窥探）
    ├── slice/              ← Slice（🔥必问）
    ├── map/                ← Map（🔥必问，Swiss Table 适配）
    ├── defer/              ← Defer（⭐重点）
    ├── gc/                 ← GC 三色标记 + 混合写屏障（🔥必问）
    ├── memory/             ← 内存分配 + 逃逸分析（🔥必问）
    └── context/            ← Context（⭐重点）
```

## 📊 各主题详情

| Topic | 必问度 | demo 数 | 测试数 | 关键内容 |
|:------|:------:|:------:|:------:|---------|
| **GMP** | 🔥 90%+ | 4 | 3 | 调度模型、抢占、work stealing |
| **Goroutine** | 🔥 90%+ | 4 | 3 | 循环变量、栈增长、阻塞、泄漏 |
| **Channel** | 🔥 95%+ | 6 | 12 | 行为矩阵、hchan 结构、5 种模式 |
| **Slice** | 🔥 90%+ | 5 | 10 | 三元组、扩容、参数传递、共享 |
| **Map** | 🔥 90%+ | 6 | 9 | key 类型、并发安全、Swiss Table |
| **Defer** | ⭐ 70% | 4 | 6 | LIFO、返回值、闭包、recover |
| **GC** | 🔥 90%+ | 3 | 4 | 触发时机、三色标记、写屏障 |
| **Memory** | 🔥 90%+ | 4 | 5 | 分配层级、逃逸分析、5 种场景 |
| **Context** | ⭐ 70% | 4 | 6 | 4 方法、6 创建函数、级联取消 |

## 🎯 4 周复习路线

| 周 | 主题 | 目标 |
|:-:|------|------|
| **Week 1** | GMP + Goroutine + Channel | 能讲清调度机制，能手写 channel 用法 |
| **Week 2** | Slice + Map + Defer | 能画出底层数据结构图，能讲清扩容规则 |
| **Week 3** | GC + Memory + Context | 能讲清三色标记、混合写屏障，能用 Context |
| **Week 4** | 全部 demo + 模拟面试 | 能结合项目经验回答，能熟练 pprof 排查 |

## 📌 几个关键数据点（面试引用）

| 数据 | 出处 | 含义 |
|------|------|------|
| `~177ns` | goroutine demo | goroutine 创建开销 |
| `114ns / 29ns` | channel demo | unbuffered vs buffered channel 通信开销 |
| `476ns / 3854ns` | slice demo | 预分配 cap 比动态扩容快 8 倍 |
| `5.2ns / 13.2ns` | map demo | 普通 map 读比 sync.Map 还快 |
| `1.8+` | GC demo | 混合写屏障版本号 |

## 🔍 看 demo 源码

每个 demo 都有详细的注释和**面试要点**总结。直接看 `.go` 文件即可：

```bash
# GMP demo 的关键源码
cat code/01-basics/gmp/main.go
cat code/01-basics/gmp/preemptive.go

# Channel 行为矩阵（必看！）
cat code/01-basics/channel/states.go
```

## 🎁 配套资源

- **[`docs/Go-Interview-Study-Guide.md`](docs/Go-Interview-Study-Guide.md)**：整理好的面试指南（956 行）
  - 每个主题都有原文锚点链接 + 推荐阅读 + 源码位置 + 易错提醒
- **原文**：[mao888/golang-guide - GOALNG_INTERVIEW_COLLECTION.md](https://github.com/mao888/golang-guide/blob/main/golang/go-Interview/GOALNG_INTERVIEW_COLLECTION.md)

## 📝 实践建议

1. **每个 demo 亲手跑一遍**：看输出比背答案强 10 倍
2. **修改 demo 代码**：试着改变参数看输出变化
3. **结合项目经验**：每个 demo 想想在自己的项目里遇到过没
4. **动手写一遍**：在 `code/<topic>/practice/` 目录里练习（约定）

## 🌲 Git 历史

所有 demo 都已提交并推送到 GitHub：

- 仓库：https://github.com/zhanbinb/go-interview-guide
- 最新 commit：见 `git log --oneline`

## 🤝 贡献

发现错误或有改进建议？直接修改对应的 `.go` 文件，提交并推送即可。

---

> **面试不是只背答案，更重要的是理解 + 项目串联。** 每个知识点都要能"结合项目讲 5 分钟"。祝面试顺利！🚀

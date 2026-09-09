# Redis 面试复习指南

> 目标：精简主流，应付面试
> 风格：每个章节一个 README + demo.sh (redis-cli) + demo.go (Go 实战)

> 🎯 **[面试速查：10 大题组 + 标准答案 + 必背金句 →](./INTERVIEW-QA.md)** ← 面试前 30 分钟必看

---

## 📁 目录结构

```
code/05-redis/
├── README.md                      ← 本文件（总览大纲）
├── INTERVIEW-QA.md               ← 🎯 面试速查 10 大题组（推荐先看）
│
├── 01-datatype/                  ← 5 大基础 + 3 大特殊数据类型
├── 02-datastructure/             ← 底层数据结构（SDS/跳表/quicklist/dict）
├── 03-persistence/               ← RDB + AOF + 混合持久化
├── 04-high-availability/         ← 主从 + Sentinel + Cluster
├── 05-cache-design/              ← 缓存三大问题（雪崩/穿透/击穿）
├── 06-transaction-lua/            ← 事务 + Lua + Pipeline
└── 07-memory/                    ← 内存管理 + 淘汰策略
```

每个章节固定结构：
- `README.md` — 面试速记 + 核心要点 + 时间复杂度
- `demo.sh`  — 可执行 redis-cli 命令（最常用）
- `demo.go`  — Go 代码示例（实战）

---

## 🎯 章节优先级

### 🔥 必问核心
1. **01-datatype** — 数据类型 + 应用场景
2. **02-datastructure** — 底层数据结构（超高频）
3. **03-persistence** — 持久化（必问）
4. **04-high-availability** — 主从 + Sentinel + Cluster
5. **05-cache-design** — 缓存三大问题（实战）

### ⭐ 重点掌握
6. **06-transaction-lua** — 事务 + Lua 原子性

### 💡 选学
7. **07-memory** — 内存管理 + LRU/LFU

---

## 🎯 面试速查入口

| 文档 | 用途 | 推荐 |
|------|------|------|
| **[INTERVIEW-QA.md](./INTERVIEW-QA.md)** | 10 大题组 + 答案 + 40 个金句 | 🔥 面试前必看 |
| [01-datatype/README.md](./01-datatype/README.md) | 5+3 数据类型详解 | ⭐ 必看 |
| [02-datastructure/README.md](./02-datastructure/README.md) | SDS/跳表/quicklist/dict | ⭐ 必看 |
| [03-persistence/README.md](./03-persistence/README.md) | RDB + AOF + 混合 | ⭐ 必看 |
| [05-cache-design/README.md](./05-cache-design/README.md) | 三大问题 + 分布式锁 | 🔥 必看 |
| [07-memory/README.md](./07-memory/README.md) | 过期策略 + 8 种淘汰 | ⭐ 必看 |

> 💡 **使用建议**：先打开 INTERVIEW-QA.md 看题 → 不熟的回查对应章节 README → 跑对应章节的 demo.sh 验证

---

## 🚀 跑法（两种模式）

### A. redis-cli 命令（最快）
```bash
cd code/05-redis/01-datatype
redis-cli < demo.sh
```

### B. Go 代码示例（实战）
```bash
brew install redis && redis-server  # 先起 Redis
cd code/05-redis/01-datatype
go mod init demo
go get github.com/redis/go-redis/v9
go run demo.go
```

---

## 📋 面试速查（完整版见 [INTERVIEW-QA.md](./INTERVIEW-QA.md)）

> 这里只放 5 个**最常被问**的，**完整 10 大题组 40+ 追问 + 答案**在 INTERVIEW-QA.md。

1. **Redis 为什么快？** — 内存 + 单线程 + IO 多路复用 + 数据结构
2. **ZSet 为什么用跳表不用红黑树？** — 简单 + 范围查询 O(log N) + 局部锁
3. **缓存三大问题（穿透/击穿/雪崩）？** — 布隆+空值 / 互斥锁 / 过期加随机
4. **Redis 分布式锁怎么实现？** — `SET NX PX` + UUID + Lua + 看门狗
5. **Redis 怎么保证高可用？** — 主从 + Sentinel + Cluster（16384 槽位）

> 📌 完整题组：性能 / 数据类型 / 持久化 / 过期淘汰 / 三大问题 / 一致性 / 分布式锁 / 主从哨兵 / Cluster / 性能调优

---

## 🎯 5 类典型场景速查

| 场景 | 数据类型 + 命令 |
|------|---------------|
| **缓存** | `String SET key + EXPIRE 3600` |
| **计数器** | `String INCR key` |
| **分布式锁** | `SETNX lock + EXPIRE` |
| **排行榜** | `ZADD rank score member + ZREVRANGE 0 9` |
| **去重/共同好友** | `Set SINTER / SUNION / SDIFF` |
| **限流（滑动窗口）** | `ZADD + ZREMRANGEBYSCORE + ZCARD` |
| **延迟队列** | `ZADD score=时间戳，循环 ZRANGEBYSCORE` |
| **UV 统计** | `HyperLogLog PFADD/PFCOUNT` |
| **附近的人** | `GEOADD + GEORADIUS` |
| **签到/日活** | `Bitmap SETBIT/BITCOUNT` |

---

## 📊 速查表

```
| 类型      | 命令前缀 | 底层       | 时间复杂度 | 典型场景 |
|----------|---------|-----------|-----------|----------|
| String   | SET/GET  | SDS        | O(1)      | 缓存、计数器、锁 |
| List     | LPUSH    | quicklist  | 头尾 O(1)  | 队列、最新列表 |
| Hash     | HSET     | dict       | O(1)      | 对象存储 |
| Set      | SADD     | dict       | O(1)      | 去重、共同好友 |
| ZSet     | ZADD     | skiplist+dict | O(log N) | 排行榜 |
| Bitmap   | SETBIT   | string     | O(1)      | 签到、日活 |
| HyperLogLog| PFADD  | 概率       | O(1)      | UV 估算 |
| GEO      | GEOADD   | zset       | O(log N)  | 附近的人 |

Redis 为什么快：内存 + 单线程 + IO 多路复用 + 高效数据结构
持久化：RDB 快照 + AOF 日志 + 4.0 混合
高可用：主从 + Sentinel + Cluster（16384 槽位）
```

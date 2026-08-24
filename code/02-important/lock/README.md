# 锁与原子操作演示（精简版）

> 📖 对应复习指南：[指南 §12 锁与原子操作](../../../docs/Go-Interview-Study-Guide.md#12-锁与原子操作)

## 🎯 面试只需掌握 4 个点

1. **同步共享变量的方式**：Mutex / RWMutex / atomic / channel
2. **Mutex 两种模式**：normal（可自旋抢））vs starvation（1ms 后严格 FIFO）
3. **atomic 底层**：CPU 原子指令（CAS / LOCK CMPXCHG）
4. **自旋锁**：适合短临界区，Go 的 Mutex 内部就是自旋+阻塞混合

## 📁 文件清单（每个文件 < 50 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `mutex.go` | Mutex 两种模式 + 锁竞争 |
| `rwmutex.go` | RWMutex vs Mutex（读多写少场景对比） |
| `atomic.go` | atomic 包（Add/CAS/Load/Store/Swap） |
| `ways.go` | 5 种同步方式对比 |
| `lock_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=lock
make run TOPIC=lock DEMO=mutex
make run TOPIC=lock DEMO=rwmutex
make run TOPIC=lock DEMO=atomic
make run TOPIC=lock DEMO=ways
make test TOPIC=lock
```

## 📌 面试速记

```
sync.Mutex：
  - 两种模式：normal（默认）vs starvation（等待 >1ms 切换）
  - normal：自旋 + 抢等待者（性能高，但可能饥饿）
  - starvation：严格 FIFO（公平，但性能略低）
  - Go 1.9+ 实现，runtime 文档：

sync.RWMutex：
  - 读锁可并发（RLock），写锁独占（Lock）
  - 读多写少才用 RWMutex
  - 不能递归加锁（死锁）

atomic 包：
  - Add / Sub     原子加减
  - Load / Store  原子读/写
  - CAS           CompareAndSwap（乐观锁）
  - Swap          原子交换
  - 底层：CPU LOCK CMPXCHG 指令（x86）

5 种同步方式对比：
  1. sync.Mutex       通用
  2. sync.RWMutex     读多写少
  3. atomic.Xxx       简单类型（int/pointer）
  4. channel          通信共享内存
  5. sync.Once        一次性初始化

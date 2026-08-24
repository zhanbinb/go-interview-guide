# 并发模式演示（精简版）

> 对应复习指南：§13 并发模式

## 面试只需掌握 3 个点

1. sync.Pool：复用对象，降低 GC 压力
2. Worker Pool：固定数量 worker 处理任务
3. sync.Map vs Mutex+map（map demo 已覆盖）

## 文件清单

| 文件 | 内容 |
|------|------|
| main.go | 菜单 |
| pool.go | sync.Pool 基本用法 + 性能对比 |
| worker_pool.go | Worker Pool 模式实现 |
| fanout.go | 高级模式 |
| pool_test.go | 测试 + benchmark |

## 跑

    make run TOPIC=pool
    make run TOPIC=pool DEMO=pool
    make run TOPIC=pool DEMO=worker
    make run TOPIC=pool DEMO=fanout
    make test TOPIC=pool
    make bench TOPIC=pool

## 面试速记

sync.Pool:
  - 复用临时对象，减少 GC 压力
  - Get 获取对象，Put 归还
  - 池中的对象随时可能被 GC 回收
  - 每个 P 独立本地池

Worker Pool:
  - 固定数量 worker 处理任务
  - 任务用 channel 传递
  - WaitGroup 等所有 worker 完成

sync.Map:
  - 读多写少 + key 集合稳定时用
  - 内部 read/dirty 两层 map

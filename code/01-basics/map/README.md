# Map 演示

> 📖 对应复习指南：[指南 §5 Map 相关](../../../docs/Go-Interview-Study-Guide.md#5-map)

## 🎯 学习目标

1. **key 类型限制**：哪些能做 key，哪些不能
2. **nil map vs 空 map**：行为差异
3. **并发安全**：并发写 fatal error + 3 种解决方案
4. **hmap 数据结构**：count / B / buckets / overflow
5. **扩容机制**：增量扩容 vs 等量扩容
6. **sync.Map vs RWMutex+map**：各自适用场景

## 📁 文件清单

| 文件 | 演示内容 | 对应面试题 |
|------|---------|------------|
| `main.go` | 入口 + 菜单 | - |
| `key_types.go` | key 类型限制 | "什么类型可以作为 map 的 key" |
| `nil_empty.go` | nil map vs 空 map | "nil map 和空 map 有何不同" |
| `concurrent.go` | 并发写触发 fatal error + 3 种解法对比 | "怎么处理并发访问" |
| `hmap_struct.go` | hmap 字段窥探（unsafe） | "map 的数据结构是什么" |
| `expand.go` | 增量扩容 vs 等量扩容 | "是怎么实现扩容" |
| `sync_map.go` | sync.Map vs Mutex+map vs RWMutex+map | "sync.Map 的锁机制" |
| `map_test.go` | 测试 + benchmark | - |

## 🚀 怎么跑

```bash
make run TOPIC=map                       # 菜单
make run TOPIC=map DEMO=key              # key 类型
make run TOPIC=map DEMO=nil              # nil vs empty
make run TOPIC=map DEMO=concurrent       # 并发（会触发 fatal）
make run TOPIC=map DEMO=hmap             # hchan 结构
make run TOPIC=map DEMO=expand           # 扩容
make run TOPIC=map DEMO=sync             # sync.Map 对比
make test TOPIC=map                      # 测试
```

## 📌 面试速记卡

```
key 类型：必须可比较（==），slice/map/func 不行
nil map：可读（零值）+ len()，但写 panic
map 遍历：无序（每次顺序不同）
并发写：触发 fatal error "concurrent map writes"
map 删除：标记 deleted，不立即释放内存

hmap = (count, B, noverflow, hash0, buckets, oldbuckets, nevacuate)
B = log_2(buckets 数)，cap = 2^B
bmap = [8]tophash + [8]key + [8]value + overflow

扩容条件：loadFactor > 6.5 或 overflow 太多
增量扩容：B += 1，每次操作搬 1~2 bucket
等量扩容：B 不变，整理（解决大量删除）

并发方案对比：
- sync.Mutex + map       通用，但读写互斥
- sync.RWMutex + map     读多写少更优
- sync.Map               key 稳定 + 读写比例极端（10:1+）

时间复杂度：平均 O(1)，最坏 O(n)（全冲突）
不能对 map 元素取地址（&m["k"] 编译错误）
```

## 🧪 实验亮点

### 实验 1：key 类型

```go
m1 := map[[2]int]string{}        // ✅ 数组可比较
m2 := map[string]int{}           // ✅ string
// m3 := map[[]int]string{}      // ❌ slice 不可比较（编译错）
```

### 实验 2：nil vs 空 map

```go
var m map[string]int  // nil
m["k"] = 1           // panic: assignment to entry in nil map
m = make(map[string]int)
m["k"] = 1           // OK

len(m)               // 两个都可以
delete(m, "k")       // 两个都可以（即使 nil）
```

### 实验 3：并发写

```go
go func() { m["k"] = 1 }()  // fatal error: concurrent map writes
go func() { m["k"] = 2 }()  // ❌ 原生 map 不是并发安全
```

### 实验 4：hmap 窥探

观察 `B` (log_2 buckets 数)、`count` (元素数)、`noverflow` (溢出 bucket 数) 在插入/删除时的变化。

### 实验 5：扩容

写满 6.5 倍 B 后，B += 1（增量扩容）。删大量后再写满相同 B 时，触发等量扩容。

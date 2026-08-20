# Slice 演示

> 📖 对应复习指南：[指南 §4 Slice](../../../docs/Go-Interview-Study-Guide.md#4-slice)

## 🎯 学习目标

Slice 是 Go 面试 🔥 必问题。本 demo 围绕 **3 个最高频考点**展开：

1. **slice 三元组** `(ptr, len, cap)` 是什么
2. **扩容规则** 怎么算（Go 1.18+ 内存对齐）
3. **共享底层数组** 带来的副作用

## 📁 文件清单

| 文件 | 演示内容 | 对应面试题 |
|------|---------|------------|
| `main.go` | 入口 + 菜单 | - |
| `struct.go` | slice 三元组 (ptr, len, cap) | "slice 底层数据结构" |
| `expand.go` | 扩容规则（1.18+ 内存对齐） | "新旧扩容策略" |
| `pass_param.go` | 函数参数传递（值拷贝 vs 共享底层） | "数组 slice 作为参数区别" |
| `share.go` | 共享底层数组的副作用 | "截取 slice 的成本" |
| `leak.go` | 大 slice 截取小 slice 导致内存泄漏 | "内存泄漏排查" |
| `slice_test.go` | 测试 + benchmark | - |

## 🚀 怎么跑

```bash
make run TOPIC=slice                       # 菜单
make run TOPIC=slice DEMO=struct           # slice 三元组
make run TOPIC=slice DEMO=expand           # 扩容
make run TOPIC=slice DEMO=param            # 参数传递
make run TOPIC=slice DEMO=share            # 共享底层
make run TOPIC=slice DEMO=leak             # 内存泄漏
make test TOPIC=slice                      # 测试
```

## 🧪 实验速览

### 实验 1：slice 三元组（核心）

```bash
make run TOPIC=slice DEMO=struct
```

打印 `s := make([]int, 3, 5)` 后：
- `ptr`（指向底层数组的指针）
- `len = 3`
- `cap = 5`
- 验证 `append` 超 cap 时会触发扩容

### 实验 2：扩容规则

```bash
make run TOPIC=slice DEMO=expand
```

观察：
- cap < 256 时：翻倍
- cap ≥ 256 时：阶梯式 (`(newcap + 3*256) / 4`) + 内存对齐
- 打印扩容前后 cap，看变化曲线

### 实验 3：参数传递

```bash
make run TOPIC=slice DEMO=param
```

`append` 是否影响外部？看代码就懂：
- 触发扩容 → 不影响原 slice（新建底层数组）
- 没触发扩容 → 影响（修改共享底层数组）

### 实验 4：共享底层数组

```bash
make run TOPIC=slice DEMO=share
```

`s2 := s1[1:3]` 不复制底层数组 → 改 s2[0] 会影响 s1[1]

### 实验 5：内存泄漏

```bash
make run TOPIC=slice DEMO=leak
```

经典坑：从大文件读 1GB 到 `buf[:10]`，**整个 1GB 都不会被 GC**（因为 buf 还引用着底层数组）。正确做法是 `s2 := append([]byte{}, buf[:10]...)` 复制一份。

## 📌 面试速记卡

```
slice = (ptr, len, cap) 三元组（本质是 struct）
数组  = 值类型，长度固定；slice = 引用类型，长度可变
append 可能扩容：cap 不够时新建底层数组 + 拷贝 + ptr 改变
扩容规则：
  - cap < 256：newcap = cap * 2
  - cap ≥ 256：newcap = (cap + 3*256) / 4（接近 1.25x）
  - 最后按内存对齐（mallocgc 的 size class）取整
参数传递：slice header 是值拷贝，但底层数组共享
截取 slice：O(1)，不复制底层数组
泄漏陷阱：大 slice 截取小 slice → 整个底层数组常驻
```

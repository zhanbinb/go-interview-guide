# 内存分配 + 逃逸分析（精简版）

> 📖 对应复习指南：[指南 §8 内存相关](../../../docs/Go-Interview-Study-Guide.md#8-内存分配--逃逸分析)

## 🎯 面试只需了解 4 个点

1. **分配层级**：mcache → mcentral → mheap（多级缓存，按大小分类）
2. **逃逸分析**：编译器判断对象放栈还是堆（`go build -gcflags="-m"`）
3. **5 种逃逸场景**：返回指针、闭包、interface、栈空间不够、大对象
4. **大对象小对象**：小对象 < 32KB 用 mcache，大对象直接走 mheap

## 📁 文件清单（每个文件 < 50 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `allocate.go` | 分配层级 + span class |
| `escape.go` | 5 种逃逸场景 + 编译期分析 |
| `leak.go` | 内存泄漏 4 种场景 |
| `size.go` | 大对象小对象 |
| `memory_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=memory
make run TOPIC=memory DEMO=allocate
make run TOPIC=memory DEMO=escape
make run TOPIC=memory DEMO=leak
make run TOPIC=memory DEMO=size

# 看逃逸分析（编译期）
go build -gcflags="-m" code/01-basics/memory/escape.go
```

## 📌 面试速记

```
分配层级（runtime/malloc.go）：
  mcache (P 私有，无锁) → mcentral (全局，需锁) → mheap (页分配)
  
  小对象 (<32KB): P 的 mcache 拿，没有就从 mcentral 要
  大对象 (≥32KB): 直接走 mheap，按页对齐

逃逸分析：
  go build -gcflags="-m"   # -m 打印逃逸分析结果
  -m -m                   更详细

5 种典型逃逸场景：
  1. 函数返回局部变量指针
  2. 闭包捕获外部变量
  3. 发送到 channel（指针/interface）
  4. slice/map/interface 包含指针
  5. 大对象（>栈空间）或切片扩容后太大
```

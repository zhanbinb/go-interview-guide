# 接口体系演示（精简版）

> 📖 对应复习指南：[指南 §11 接口](../../../docs/Go-Interview-Study-Guide.md#11-接口体系)

## 🎯 面试只需掌握 5 个点

1. **iface vs eface**：带方法的接口 vs 空接口 `any`/`interface{}`
2. **nil 接口陷阱**：接口变量的 nil 分两层（动态类型 + 动态值）
3. **类型断言**：`x.(T)` 的两种用法（普通 + switch）
4. **类型转换 vs 类型断言**：编译期 vs 运行期
5. **多态**：鸭子类型，函数参数接受接口

## 📁 文件清单（每个文件 < 50 行）

| 文件 | 内容 |
|------|------|
| `main.go` | 菜单 |
| `struct.go` | iface / eface 内部结构（unsafe 窥探） |
| `nil_trap.go` | 经典 nil 接口陷阱 |
| `assert.go` | 类型断言 + 类型 switch |
| `polymorphism.go` | 多态实战 |
| `iface_test.go` | 测试 |

## 🚀 跑

```bash
make run TOPIC=iface
make run TOPIC=iface DEMO=struct
make run TOPIC=iface DEMO=nil
make run TOPIC=iface DEMO=assert
make run TOPIC=iface DEMO=poly
make test TOPIC=iface
```

## 📌 面试速记

```
接口底层两个结构体（runtime/runtime2.go）：
  iface: 带方法的接口
    type iface struct {
        tab  *itab          // 方法表 + 类型信息
        data unsafe.Pointer // 动态值指针
    }

  eface: 空接口 interface{} / any
    type eface struct {
        _type *_type       // 类型信息
        data  unsafe.Pointer
    }

  itab:
    type itab struct {
        inter  *interfacetype
        _type  *_type
        link   *itab
        hash   uint32
        fun    [1]uintptr  // 方法表（变长）
    }

经典陷阱：nil 接口 vs nil 动态值
  var p *int = nil
  var i interface{} = p
  i == nil      → false（i 有动态类型 *int）
  i.(*int) == nil → true（动态值是 nil）

类型断言：
  v, ok := x.(T)    // 安全，ok=false 表示失败
  switch x.(type) {  // 在 switch 里用

类型转换 vs 类型断言：
  T(x)     编译期，类型必须"兼容"（如 int↔int64）
  x.(T)    运行期，x 必须是 interface

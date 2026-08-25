# 基础语法 + 面向对象演示（精简版）

> 对应复习指南：§16 基础语法细节 + §17 面向对象

## 面试只需了解 8 个点

§16 基础语法：
1. 单/双/反引号区别
2. rune vs byte 类型
3. uint 溢出
4. 函数参数传值还是传指针

§17 面向对象：
5. 封装（首字母大小写）
6. 继承（结构体组合 = has-a）
7. 多态（interface 隐式实现 = 鸭子类型）
8. Go 和 Java 的核心区别

## 文件清单

- main.go: 菜单
- string.go: 单/双/反引号
- rune_byte.go: rune vs byte + uint 溢出
- param.go: 函数参数传值还是传指针
- oop.go: 封装/组合/多态
- syntax_test.go: 测试

## 面试速记

单/双/反引号：
  - 单引号 a 是 rune 字面量（int32）
  - 双引号 abc 是 string（解析转义）
  - 反引号 abc 是 raw string（不解析转义）

rune vs byte：
  - byte = uint8 (1字节，ASCII)
  - rune = int32 (4字节，Unicode 码点)

uint 溢出：
  - uint8(255) + 1 = 0（回卷）
  - 边界判断要小心

函数参数：
  - Go 全是值传递
  - 大结构体传指针省内存
  - 切片/map/channel 本身是引用（header 拷贝，但底层共享）

OOP 三件套：
  - 封装:首字母大小写控制可见性
  - 继承:结构体组合（嵌入）= has-a，不是 is-a
  - 多态:interface 隐式实现

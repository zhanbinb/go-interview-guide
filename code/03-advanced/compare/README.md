# Go 与其他语言对比

> 对应复习指南：§21 Go 与其他语言对比

## 面试只需了解 3 个对比

1. Go vs Java: 部署简单、并发轻量、适合云原生
2. Go vs Python: 静态类型、速度快 10-100 倍、适合服务开发
3. Go vs Node.js: 性能更强、类型安全、适合高并发

## 文件清单

- main.go: 菜单
- vs_java.go: Go vs Java 对比
- vs_python.go: Go vs Python 对比
- vs_node.go: Go vs Node.js 对比
- compare_test.go: 性能对比 benchmark（演示 Go 性能优势）

## 跑

    make run TOPIC=compare
    make run TOPIC=compare DEMO=java
    make run TOPIC=compare DEMO=python
    make run TOPIC=compare DEMO=node
    make bench TOPIC=compare

## 面试速记

Go vs Java:
  - 部署：Go 单二进制 / Java 需 JVM
  - 并发：goroutine (M) / thread (OS)
  - 类型：Go 接口隐式 / Java 显示 implements
  - 生态：Java 成熟 / Go 云原生强
  - 性能：Go 启动快、内存低 / Java 优化空间大

Go vs Python:
  - 类型：静态 / 动态
  - 性能：Go 快 10-100 倍
  - 并发：goroutine / asyncio
  - 用途：Go 服务 / Python 数据/脚本

Go vs Node.js:
  - 性能：Go 快得多（CPU 密集）
  - 类型：Go 静态 / Node 动态（TS 弥补）
  - 生态：Node npm 更丰富
  - 用途：Go 后端 / Node I/O 密集 + 前端全栈


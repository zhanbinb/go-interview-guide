# Gin 框架演示（精简版）

> 对应复习指南：§18 框架

## 说明

sandbox 环境无法联网下载 Gin 包，所以本 demo 用 Go 标准库 `net/http`
实现 Gin 的核心设计模式。你可以学到 Gin 的原理：
- radix tree 路由（用 prefix tree 模拟）
- 中间件洋葱模型
- HandlerFunc 模式
- Context 传递

## 面试只需了解 4 个点

1. Gin = 高性能 Go web 框架（httprouter 改造的 radix tree）
2. 中间件洋葱模型（c.Next() 控制流程）
3. Gin.Context 封装了 request/response
4. gin.Default() 自动加 Logger + Recovery 中间件

## 文件清单

- main.go: 菜单
- router.go: 简单路由注册（用 map 模拟）
- middleware.go: 中间件洋葱模型
- context.go: 自定义 Context 传递值
- server.go: 完整 server 启动 demo
- gin_test.go: 测试

## 跑

    make run TOPIC=gin
    make run TOPIC=gin DEMO=router
    make run TOPIC=gin DEMO=middleware
    make run TOPIC=gin DEMO=context
    make run TOPIC=gin DEMO=server
    make test TOPIC=gin

## 面试速记

Gin 核心:
  - Router: radix tree (前缀树) 路由匹配
  - HandlerFunc: func(c *Context)
  - 中间件: chain of responsibility
  - Context: 封装 *http.Request, ResponseWriter, params, keys

Gin vs net/http:
  - net/http: 标准库，繁琐
  - Gin: 封装好，性能高（httprouter）
  - 其他选择: echo / chi / iris

 
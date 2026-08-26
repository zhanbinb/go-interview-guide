# pprof 性能排查演示

> 对应复习指南：§20 性能排查

## 面试只需了解 4 个点

1. pprof 是 Go runtime 自带的性能分析工具（net/http/pprof）
2. CPU profile: 找 CPU 热点
3. Heap profile: 找内存分配
4. Goroutine profile: 找 goroutine 泄漏

## 文件清单

- main.go: 启动带 pprof 端点的 HTTP 服务
- leak.go: 模拟 goroutine 泄漏
- pprof_test.go: 测试（不用真起服务）

## 跑

实际起服务：
    make run TOPIC=pprof
    # 另一个终端：
    curl http://localhost:6060/debug/pprof/heap > heap.out
    go tool pprof heap.out
    # 在 pprof 交互模式：(pprof) top10

## 面试速记

pprof 4 类 profile：
  - CPU:     CPU 占用热点（找慢函数）
  - HEAP:    内存分配（找分配最多的代码）
  - GOROUTINE: goroutine 数量（找泄漏）
  - ALLOC:   内存分配次数

排查流程：
  1. CPU 高 → CPU profile → top10
  2. 内存涨 → HEAP profile → top10
  3. goroutine 涨 → GOROUTINE profile
  4. trace 看调度延迟

命令行：
  go tool pprof http://host/debug/pprof/heap
  go tool pprof http://host/debug/pprof/profile?seconds=30
  go test -bench=. -benchmem

可视化：
  go tool pprof -http=:8080 heap.out
  # 浏览器看火焰图


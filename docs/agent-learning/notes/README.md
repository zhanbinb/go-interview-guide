# 学习笔记目录

> 按主题组织的深度笔记，配合代码一起看。

## 📚 笔记清单

| # | 笔记 | 对应学习路线 | 关键代码 | 状态 |
|---|------|------------|---------|------|
| 1 | [refactoring.md](./refactoring.md) | Step 5 → Step 7 重构 | 整个项目 | ✅ |
| 2 | [context-management.md](./context-management.md) | Step 6 · Context Management | [`memory/context.go`](../../../code/06-agent/go-agent-demo/memory/context.go) | ✅ |
| 3 | [agent-orchestration.md](./agent-orchestration.md) | Step 7 · Agent 编排模式 | [`workflow/order.go`](../../../code/06-agent/go-agent-demo/workflow/order.go) | ✅ |
| 4 | [memory-preview.md](./memory-preview.md) | Step 8 · Memory（预览）| [`memory/memory.go`](../../../code/06-agent/go-agent-demo/memory/memory.go) | ⬜ |

## 🗺️ 推荐阅读顺序

### 第一次学（按顺序）

1. [refactoring.md](./refactoring.md) — 理解为什么模块这样切分
2. [context-management.md](./context-management.md) — Step 6 核心概念
3. [agent-orchestration.md](./agent-orchestration.md) — Step 7 三种模式

### 复习时（按主题跳）

- 面试被问 Context Management → 看 [context-management.md](./context-management.md)
- 面试被问 Workflow vs Agent → 看 [agent-orchestration.md](./agent-orchestration.md)
- 准备开始 Step 8 → 看 [memory-preview.md](./memory-preview.md)

## 📝 每篇笔记的结构

每篇笔记都包含：
- 🎯 核心问题（一句话讲清学这个干嘛）
- 🧠 关键概念（区别 / 对比 / 类比）
- 💻 代码解读（对应到实际文件 + 行数）
- ⚠️ 已知问题 / 局限
- 🔮 下一步 / 进阶方向
- ✅ 自检问题

## 🔗 相关资源

- [学习路线总览](../README.md)
- [Demo 代码 + README](../../../code/06-agent/go-agent-demo/README.md)

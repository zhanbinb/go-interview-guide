# MySQL 一条 SQL 的执行流程（基础架构）

> 面试**入门必问**，中等频率
> 重点：6 大组件 + 查询流程 + 更新流程的差异
> SQL Server 也有类似流程，但组件划分不同

---

## 🔥 必问核心

### 1. MySQL 基础架构图（必背）

```
┌─────────────────────────────────────────────────┐
│                  Client (连接池)                  │
└────────────────────┬────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────┐
│  Server 层（所有存储引擎共用）                      │
│  ┌────────┐  ┌─────────┐  ┌─────────┐  ┌────────┐ │
│  │连接器  │→│ 查询缓存 │→│ 分析器  │→│ 优化器 │→│ 执行器 │ │
│  └────────┘  └─────────┘  └────┬────┘  └────────┘  │
└───────────────────────────────┼────────────────────┘
                                │ 词法/语法分析
                                ↓
┌───────────────────────────────▼────────────────────┐
│  存储引擎层（可插拔）                              │
│  InnoDB / MyISAM / Memory                          │
└─────────────────────────────────────────────────┘
```

### 2. 6 大组件详解（按执行顺序）

| 组件 | 作用 | 关键点 |
|------|------|--------|
| **1. 连接器** | 管理连接 + 鉴权 | `mysql -u root -p` 走这里 |
| **2. 查询缓存** | 缓存 SELECT 结果（**8.0 已删**）| 表改了缓存就失效，弊大于利 |
| **3. 分析器** | 词法 + 语法 + 语义分析 | `SELECT * FRMO t` 报错 "syntax error" |
| **4. 优化器** | 生成最优执行计划 | 选哪个索引、JOIN 顺序等 |
| **5. 执行器** | 执行计划 + 调用存储引擎 | 调 `engine.query()` |
| **6. 存储引擎** | 真正存/取数据 | InnoDB / MyISAM / Memory |

### 3. 一条 SELECT 的执行流程（重点！）

```sql
SELECT * FROM t_user WHERE name = 'Alice' AND age > 20 LIMIT 10;
```

```
① 连接器（Connector）
   └─ 验证用户名/密码（不匹配 → ERROR 1045）
   └─ 从权限表读权限（之后所有判断都用这份快照）
   └─ 查询缓存里有没有（8.0 已删，8.0 这步不存在）

② 分析器（Parser）
   └─ 词法分析：识别 SELECT、FROM、WHERE、LIMIT 等关键字
   └─ 语法分析：检查 SQL 语法（错就报 You have an error）
   └─ 生成解析树：
       SELECT
         └─ * FROM t_user
         └─ WHERE name = 'Alice' AND age > 20
         └─ LIMIT 10

③ 优化器（Optimizer）
   └─ 选择索引：name_idx 还是 age_idx 还是全表扫？
   └─ 决定 JOIN 顺序（多表时）
   └─ 评估成本（Cost-based Optimizer）
   └─ 输出 执行计划（EXPLAIN 看的就是这个）

④ 执行器（Executor）
   └─ 按执行计划调存储引擎：
       for i in 0..10:
         engine.query('t_user', name='Alice' AND age > 20, LIMIT 10)
   └─ 过滤 + 排序 + LIMIT 在 Server 层做

⑤ 存储引擎（InnoDB）
   └─ 真正读数据页 → 返回给 Server 层 → 返回给客户端
```

### 4. 一条 UPDATE 的执行流程（重点！与 SELECT 差异）

```sql
UPDATE t_user SET age = 30 WHERE name = 'Alice';
```

```
①②③ 与 SELECT 相同（连接 → 解析 → 优化 → 执行计划）

④ 执行器（关键差异）
   └─ 调存储引擎前：先问 Buffer Pool 有没有这行（缓存命中？）
   └─ 没命中：从磁盘读数据页到 Buffer Pool
   └─ 修改前：写 **undo log**（用于回滚 + MVCC）
   └─ 修改 Buffer Pool 中的页（内存里）
   └─ 写 **redo log（prepare）**
   └─ 写 **binlog**
   └─ 写 **redo log（commit）**
   └─ 标记 redo log 为 committed
   └─ 返回给客户端"成功"
   （数据页异步刷盘）

⑤ 存储引擎
   └─ 只修改 Buffer Pool，磁盘 I/O 延后
```

**关键差异**：
- SELECT 走 **快照读**（MVCC）
- UPDATE 走 **当前读**（最新已提交版本）
- UPDATE 比 SELECT 多了 **3 个日志**（undo / redo / binlog）

### 5. SQL Server 对比（你专享）

| 组件 | SQL Server | MySQL |
|------|-----------|-------|
| 协议层 | TDS（SQL Server 自有）| MySQL Wire Protocol |
| 连接管理 | 线程池（默认无连接池）| 连接器（默认 151 线程池）|
| 查询缓存 | ❌ 没有（早期有，被废弃）| ⚠️ 8.0 已删（建议关闭）|
| 解析器 | T-SQL Parser | Parser |
| 优化器 | Cost-Based + Rule-Based | Cost-Based |
| 执行器 | 执行计划 | 执行器 |
| 存储引擎 | 只有一种（InnoDB-like）| **可插拔**（InnoDB / MyISAM / Memory）|

**最关键差异**：MySQL 存储引擎**可插拔**，SQL Server 只有一种。

### 6. 查询缓存为什么被删？（必考历史）

**MySQL 8.0 之前**有查询缓存，8.0 完全移除：
```sql
-- 8.0 之前
SHOW VARIABLES LIKE 'query_cache_type';
-- 0: 关  1: 全部缓存  2: 只缓存 SELECT SQL_CACHE

-- 8.0 起这个变量都没了
```

**为什么删**：
- 缓存命中要求**SQL 完全相同**（连大小写都得一致）
- 表任何数据变更 → **整个表所有缓存失效**
- 高并发写入下命中率极低（通常 < 5%）
- 维护缓存本身有性能开销

**生产建议**：
- 8.0+ 不用考虑查询缓存
- 老版本（5.7）建议关闭（`query_cache_type=0`）
- 需要缓存用 **Redis** 或 **应用层缓存**

---

## ⭐ 重点掌握

### 7. 长连接 vs 短连接

| 类型 | 含义 | 场景 |
|------|------|------|
| **短连接** | 每次查询都新建连接 | 简单，但慢 |
| **长连接** | 复用连接（默认）| 高性能，但占内存 |

**坑**：长连接会累积临时内存（`wait_timeout` 默认 8 小时后才断开）：
```sql
SHOW VARIABLES LIKE 'wait_timeout';
SET GLOBAL wait_timeout = 28800;  -- 8 小时
```

**应用层建议**：连接池（Java HikariCP / Go database/sql）控制空闲连接

### 8. Buffer Pool 是什么（高频）？

**定义**：MySQL 内存里的数据缓存区（默认 128MB）

**作用**：
- 缓存磁盘读出的数据页
- 写入先写内存（Buffer Pool），再异步刷盘
- 大幅减少磁盘 I/O（内存比磁盘快 1000 倍）

**核心指标**：
- `SHOW ENGINE INNODB STATUS\G` 看 Buffer Pool hit rate
- 命中率 > 99% 才正常（生产）

### 9. SQL 执行的"权限检查"在哪个阶段？

**重要：先查权限，再查数据！**

```
连接器：验证用户名密码
  ↓
权限表里查这个用户有没有这个库的权限
  ↓
权限快照（在内存里缓存，后续用这个快照）
  ↓
后续所有 SQL 都不再查权限表（即使管理员改了权限也没用）
  ↓
要重新连接才能用新权限
```

### 10. MySQL 服务端 vs 客户端解析差异

**MySQL 服务端**：
- 解析 SQL → 生成执行计划 → 执行
- 每次都要走一遍这个流程

**客户端**：
- 一些 ORM/驱动会做**预编译**（Prepared Statement）：
  ```java
  PreparedStatement ps = conn.prepareStatement("SELECT * FROM t WHERE id = ?");
  ps.setInt(1, 100);
  ```
- 只解析一次 SQL，多次执行只传参数
- **好处**：防 SQL 注入 + 减少服务端解析开销

---

## 💡 选学

### 11. 词法/语法分析是怎么工作的？

**词法分析**：把 SQL 字符串切成 token
```
SELECT * FROM t WHERE id = 1
   ↓
[SELECT] [*] [FROM] [t] [WHERE] [id] [=] [1]
```

**语法分析**：用 BNF 语法规则检查 token 顺序，生成**解析树**（AST）

### 12. 优化器内部有什么？

**Rule-Based Optimizer (RBO)**：基于规则（如"小表驱动大表"）
**Cost-Based Optimizer (CBO)**：基于成本统计（默认 8.0+）

**MySQL 8.0 引入了 hash join**（之前只有 nested loop）

### 13. 一条 SQL 在 InnoDB 引擎层的完整流程

```
Server 层优化器（决定用哪个索引）
    ↓
InnoDB 引擎 API：handler::ha_index_read()
    ↓
Buffer Pool 查缓存
    ↓
未命中：去磁盘读数据页 → 加载到 Buffer Pool
    ↓
返回数据 → InnoDB → Server 层 → 客户端
```

---

## 🎯 面试最常被问的 5 个问题

1. **MySQL 一条 SQL 怎么执行的？**
   - 答：连接器（鉴权）→ 分析器（词法+语法）→ 优化器（生成执行计划）→ 执行器（调存储引擎）→ 存储引擎（读数据）

2. **SELECT 和 UPDATE 的执行流程区别？**
   - 答：SELECT 走快照读（MVCC）；UPDATE 走当前读 + 写 3 个日志（undo/redo/binlog）+ 标记 Buffer Pool 脏页

3. **MySQL 查询缓存为什么被删了？**
   - 答：缓存命中率极低（<5%），表更新就全失效，维护成本高。生产用 Redis

4. **MySQL 和 SQL Server 架构最关键区别？**
   - 答：MySQL 存储引擎可插拔（InnoDB/MyISAM/Memory），SQL Server 只有一种

5. **Buffer Pool 是什么？默认多大？**
   - 答：MySQL 内存数据缓存，默认 128MB，生产建议设为机器内存的 60-70%

---

## 📋 速查表

```
| 组件       | 作用                 | 8.0 变化 |
|------------|---------------------|---------|
| 连接器     | 鉴权 + 连接管理      | -       |
| 查询缓存   | 缓存 SELECT 结果     | 已删除  |
| 分析器     | 词法+语法+语义       | -       |
| 优化器     | 生成最优执行计划     | + hash join |
| 执行器     | 调存储引擎 API       | -       |
| 存储引擎   | InnoDB（默认）/ 其他 | -       |

SELECT 路径：连接器 → 优化器 → 存储引擎（快照读）
UPDATE 路径：连接器 → 优化器 → undo log → Buffer Pool → redo log (prepare) → binlog → redo log (commit)
```

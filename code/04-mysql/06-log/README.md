# MySQL 日志（binlog + redo log + undo log + 两阶段提交）

> 面试**必问**章节
> 重点：三种日志的作用 + 两阶段提交保证一致性
> 这些日志是事务、复制、性能调优的基础

---

## 🔥 必问核心

### 1. MySQL 三种核心日志（必背）

| 日志 | 层级 | 内容 | 作用 | 关键点 |
|------|------|------|------|--------|
| **redo log** | InnoDB | 物理日志（页修改）| **crash-safe / 持久性 D** | 写 WAL，循环写 |
| **binlog** | Server | 逻辑日志（SQL 行）| **主从复制 / 数据恢复** | 追加写，ROW 模式 |
| **undo log** | InnoDB | 逻辑日志（反向操作）| **回滚 + MVCC** | 事务开始前生成 |

**记忆口诀**：
- **redo** = 重新做 = 恢复数据（crash-safe）
- **undo** = 反着做 = 回滚或 MVCC 找历史版本
- **binlog** = bin = 二进制 = 给从库用的

### 2. redo log（重做日志）

**作用**：保证**持久性 D**（事务提交后数据不丢）

**核心机制：WAL（Write-Ahead Logging）**
- 事务提交前，先写 redo log，再写数据
- 内存里的脏页还没刷盘，redo log 已经持久化了
- 即使 MySQL 崩溃，重启后用 redo log 自动恢复

**写入流程**：
```
SQL 更新 → InnoDB
  ├─ 写 undo log（用于回滚）
  ├─ 更新 Buffer Pool（内存）
  ├─ 写 redo log（顺序追加到 log buffer）
  └─ 事务提交 → redo log 刷盘（fsync）
     数据页 → 异步刷盘
```

**redo log 是物理日志**：记录的是"在某个页的某个位置做了什么修改"，不是 SQL

**redo log 是循环写**：
```
┌── active (当前写) ──┐
│   ↓                 │
│   写入位置          │
└────────────────────┘
  ↓ write pos
  ↑ check point
```

- write pos：当前记录位置
- check point：已刷盘的位置
- write pos 追上 check point 时，必须停下来刷盘（等 checkpoint 推进）

**面试常问：为什么 redo log 比直接刷数据快？**
- 答：顺序写（O(1) 复杂度）vs 随机写（要找数据页在哪）

### 3. binlog（二进制日志）

**作用**：**主从复制** + **数据恢复**（Point-in-time Recovery）

**Server 层日志**（所有存储引擎共用）

**三种格式**（面试必问）：
| 格式 | 含义 | 优缺点 |
|------|------|--------|
| **STATEMENT** | 记录 SQL 语句 | 简单，但 `now()` 这类函数主从不一致 |
| **ROW** | 记录每行数据变化 | 准确，文件大（默认推荐）|
| **MIXED** | 混合模式 | MySQL 自动选择 |

**查看当前格式**：
```sql
SHOW VARIABLES LIKE 'binlog_format';
```

**生产建议**：**用 ROW 格式**（8.0 默认）

**binlog vs redo log 关键区别**：

| 维度 | redo log | binlog |
|------|----------|--------|
| 层级 | InnoDB 引擎层 | Server 层 |
| 内容 | 物理日志（页修改）| 逻辑日志（SQL/行）|
| 用途 | crash-safe | 主从复制 / 数据恢复 |
| 写入 | 循环写 | 追加写（不覆盖）|
| 大小 | 固定（4 个文件各 1GB）| 无限（按时间切分）|

### 4. undo log（回滚日志）

**作用**：
1. **回滚**事务（rollback）
2. **MVCC** 快照读（找到历史版本）

**机制**：
- 事务开始时，记录要修改的数据的**原始值**
- 比如 UPDATE t SET age=30 WHERE id=1，会先记下 `id=1, age=25`（原值）
- ROLLBACK 时把 age 改回 25
- MVCC 快照读时通过 DB_ROLL_PTR 找到历史版本

**特点**：
- 逻辑日志（反向 SQL）
- 事务提交后**不立即删除**（MVCC 还要用）
- 由 purge 线程异步清理（当没有事务需要这个版本时）

### 5. 两阶段提交（必问难点！）

**问题**：redo log 和 binlog 是**两个独立的写操作**。如何保证它们**要么都写，要么都不写**？

**经典场景**（崩溃发生在两个写中间）：
```
1. 写 redo log（prepare 阶段）
2. 写 binlog
3. 写 redo log（commit 阶段）
```

**两阶段提交流程**：
```
START TRANSACTION
  ↓
写 undo log（用于回滚和 MVCC）
  ↓
更新 Buffer Pool（内存）
  ↓
写 redo log（标记为 prepare 状态）
  ↓
写 binlog
  ↓
写 redo log（标记为 commit 状态）
  ↓
COMMIT
```

**崩溃恢复规则**：
- redo log 是 commit 状态 → 提交
- redo log 是 prepare + binlog 完整 → 提交
- redo log 是 prepare + binlog 不完整 → 回滚

**面试必问：为什么需要两阶段？**
- 答：因为 redo log 和 binlog 是两个独立的写
- 简单先后写无法保证一致性
- 两阶段确保两者**要么都成功，要么都不影响**

### 6. SQL Server vs MySQL 日志对比

| 日志 | SQL Server | MySQL |
|------|-----------|-------|
| 事务日志 | `transaction log`（ldf 文件）| `redo log`（物理）+ `undo log` |
| 复制日志 | `transaction log` 兼任 | `binlog` 独立 |
| 数据恢复 | 完整恢复模式 + 简单恢复模式 | `binlog` + 全量备份 |
| 写入顺序 | 先写日志再写数据（WAL）| 同（左） |
| 复制模式 | 发布/订阅（推/拉）| 主从 binlog 拉取 |

---

## ⭐ 重点掌握

### 7. binlog 写入机制

**关键参数**：
```sql
SHOW VARIABLES LIKE 'sync_binlog';   -- 刷盘策略
SHOW VARIABLES LIKE 'binlog_format';
SHOW MASTER STATUS;                    -- 当前 binlog 位置
```

**sync_binlog 三种策略**：
- `0`：依赖 OS 刷盘（性能最好，可能丢数据）
- `1`：每次事务提交都 fsync（最安全，性能差）— **推荐**
- `N`：每 N 个事务刷一次（折中）

### 8. binlog vs redo log 写入顺序

**重点**：两阶段提交是 InnoDB 引擎的实现细节，应用层用 `START TRANSACTION ... COMMIT` 就好

| 顺序 | 内容 |
|------|------|
| 1 | 写 undo log |
| 2 | 更新内存数据 |
| 3 | **写 redo log（prepare）** ← 第一次 |
| 4 | **写 binlog** |
| 5 | **写 redo log（commit）** ← 第二次 |
| 6 | COMMIT 返回给客户端 |

### 9. 主从复制与 binlog

```
   Master                   Slave
   ┌─────┐                ┌──────┐
   │ binlog │  ─── 传 ───→ │ IO Thread │
   └─────┘                │ 写入 relay log │
                          │  └─→ SQL Thread │
                          │      重放 SQL  │
                          └──────┘
```

**异步复制**（默认）：master 写完 binlog 就返回，slave 异步拉取
- 问题：master 崩溃时，slave 数据可能落后（数据丢失）

**半同步复制**（MySQL 5.7+）：
- master 等待至少一个 slave 收到 binlog 后才返回
- 解决数据丢失问题，但性能有损

### 10. 事务提交后数据真的写盘了吗？

**不一定**！
- 事务提交 = redo log 已刷盘（持久性 D 满足）
- 但**数据页**可能还在内存（Buffer Pool）
- 数据库崩溃后重启会从 redo log 自动恢复

```sql
-- 查看 Buffer Pool 大小
SHOW VARIABLES LIKE 'innodb_buffer_pool_size';
-- 查看脏页
SHOW ENGINE INNODB STATUS;
```

### 11. 三种日志协作

```
SQL UPDATE t SET age = 30 WHERE id = 1
        ↓
┌─ undo log: 记录 (id=1, age=25) ← 用于回滚/MVCC
│
├─ Buffer Pool: 内存里 id=1 改成 age=30
│
└─ redo log: 记录 "id=1 改成 age=30" ← 用于崩溃恢复

事务 COMMIT
  ↓
  写 binlog（逻辑：UPDATE ... id=1）
  ↓
  redo log 标记为 committed
```

---

## 💡 选学

### 12. MySQL 还有哪些日志（不深究）

| 日志 | 作用 |
|------|------|
| 慢查询日志 | 记录慢 SQL（`slow_query_log`）|
| 错误日志 | MySQL 启动/运行错误 |
| 中继日志 | 主从复制时 slave 暂存（relay log）|
| binlog 索引文件 | 记录所有 binlog 文件列表 |
| DDL 日志 | 记录 DDL 操作（metadata log）|

**面试不会深问这些，知道存在即可**

### 13. WAL（Write-Ahead Logging）思想

不只是 MySQL，**所有现代数据库都用 WAL**：
- PostgreSQL
- Oracle（redo log）
- SQL Server（transaction log）
- RocksDB / LevelDB

**核心思想**：先写日志再写数据，保证崩溃后能恢复

---

## 🎯 面试最常被问的 5 个问题

1. **MySQL 有哪些日志？redo/undo/binlog 区别？**
   - 答：redo（物理，crash-safe）、undo（回滚+MVCC）、binlog（逻辑，主从复制）

2. **什么是 WAL？为什么需要先写日志？**
   - 答：Write-Ahead Logging。先写日志再写数据，崩溃后用日志恢复，避免数据丢失

3. **两阶段提交是什么？为什么需要？**
   - 答：redo log 和 binlog 独立，两阶段（prepare + commit）保证两者原子性。崩溃恢复规则：commit 状态 → 提交；prepare + binlog 完整 → 提交；prepare + binlog 不完整 → 回滚

4. **binlog 有几种格式？生产用哪个？**
   - 答：STATEMENT（SQL）、ROW（行数据）、MIXED（混合）。生产用 ROW（准确，8.0 默认）

5. **redolog 和 binlog 区别？**
   - 答：redolog 是 InnoDB 物理日志（循环写、crash-safe），binlog 是 Server 逻辑日志（追加写、主从复制）

---

## 📋 速查表

```
| 日志       | 层级   | 内容    | 用途          | 写入   |
|------------|--------|---------|---------------|--------|
| redo log   | InnoDB | 物理    | crash-safe    | 循环   |
| binlog     | Server | 逻辑    | 主从复制      | 追加   |
| undo log   | InnoDB | 逻辑    | 回滚 + MVCC   | 追加   |

两阶段提交：redo (prepare) → binlog → redo (commit)
WAL：先写日志再写数据
binlog 推荐 ROW 格式
sync_binlog=1 最安全
```

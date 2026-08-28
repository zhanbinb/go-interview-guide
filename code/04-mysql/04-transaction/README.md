# MySQL 事务（ACID + 隔离级别 + MVCC）

> 面试**必问**章节
> 重点：4 隔离级别、MVCC 实现、幻读与 Next-Key Lock
> 与 SQL Server 强相关：默认隔离级别不同

---

## 🔥 必问核心

### 1. 事务的 ACID 四大特性

| 特性 | 含义 | MySQL 实现 |
|------|------|------------|
| **A**tomicity 原子性 | 全部成功或全部失败 | **undo log**（回滚用）|
| **C**onsistency 一致性 | 数据从一个一致状态到另一个 | AID 三者共同保证 |
| **I**solation 隔离性 | 并发事务互不干扰 | **锁 + MVCC** |
| **D**urability 持久性 | 提交后永久保存 | **redo log** |

**面试口诀**：AID 是手段，C 是目的

### 2. 4 个隔离级别（最常考！）

| 隔离级别 | 脏读 | 不可重复读 | 幻读 | MySQL 默认 | SQL Server 默认 |
|---------|-----|----------|------|-----------|----------------|
| READ UNCOMMITTED | ✅ | ✅ | ✅ | 支持 | 支持 |
| READ COMMITTED (RC) | ❌ | ✅ | ✅ | 支持 | **✅ 默认** |
| REPEATABLE READ (RR) | ❌ | ❌ | ✅（InnoDB 用 Next-Key Lock 防住）| **✅ MySQL 默认** | 支持 |
| SERIALIZABLE | ❌ | ❌ | ❌ | 支持 | 支持 |
| SNAPSHOT | ❌ | ❌ | ❌ | ❌（用 MVCC 替代）| 支持（2016+）|

**关键差异**：
- **MySQL InnoDB 默认 RR**（不是 RC！）
- **SQL Server 默认 RC**
- MySQL 在 RR 级别通过 **Next-Key Lock** 解决了幻读（SQL Server 需要 SERIALIZABLE 或 SNAPSHOT）

### 3. 三大问题（必背）

| 问题 | 含义 | 举例 |
|------|------|------|
| **脏读** Dirty Read | 读到别的事务**未提交**的数据 | A 改 → B 读 → A 回滚 → B 读到错的 |
| **不可重复读** | 同一行**两次读**结果不同 | A 读 → B 改 → A 再读（同一行变了）|
| **幻读** Phantom | 同一范围**两次读**行数不同 | A 读范围 → B 插入新行 → A 再读（多了一行）|

**关键区分**：
- 不可重复读：**单行数据**变了
- 幻读：**范围查询**行数变了

### 4. MVCC 多版本并发控制（核心机制）

**MySQL InnoDB 的 RR 级别实现快照读的关键**：

每行数据有 2 个隐藏列：
- `DB_TRX_ID`：最近修改的事务 ID
- `DB_ROLL_PTR`：回滚指针（指向 undo log）

**快照读（普通 SELECT）**：
- 读的是**事务开始时的快照版本**
- 不会看到别的事务的修改（即使已提交）
- 通过 undo log 链 + DB_TRX_ID 对比实现

**当前读（INSERT/UPDATE/DELETE/SELECT FOR UPDATE）**：
- 读的是**最新已提交版本**
- 通过行锁保证

**MVCC 流程**（快照读）：
1. 事务开始时，记录当前活跃事务 ID 列表
2. 读一行时，对比行的 DB_TRX_ID：
   - 如果 TRX_ID 在活跃列表里 → 顺着 undo log 找更早版本
   - 如果 TRX_ID 不在活跃列表里 → 读这个版本（说明已提交）
3. 反复 2 直到找到**事务开始前已提交**的版本

### 5. 当前读 vs 快照读（面试必问）

| 操作 | 类型 | 例子 |
|------|------|------|
| 普通 SELECT | 快照读 | `SELECT * FROM t` |
| SELECT FOR UPDATE | 当前读 | `SELECT * FROM t FOR UPDATE` |
| UPDATE | 当前读 | `UPDATE t SET ...` |
| DELETE | 当前读 | `DELETE FROM t WHERE ...` |
| INSERT | 当前读 | `INSERT INTO t VALUES (...)` |

**重要**：在 RR 级别，**快照读不阻塞、当前读阻塞**

### 6. Next-Key Lock（防幻读的核心！）

**问题**：普通行锁只能锁单行，**不能阻止其他事务在范围内插入新行**（幻读）

**Next-Key Lock = 记录锁 + 间隙锁（Gap Lock）**

```sql
-- 假设有 id: 1, 5, 10, 15

SELECT * FROM t WHERE id BETWEEN 5 AND 10 FOR UPDATE;
-- InnoDB 实际锁住：
-- - 记录锁: id=5, id=10
-- - 间隙锁: (-∞, 5), (5, 10), (10, 15)
-- → 其他事务不能在 1-5、5-10、10-15 范围插入
```

**这就是 MySQL RR 级别防幻读的秘诀！**

### 7. 事务实际操作（生产经验）

```sql
-- 开启事务
START TRANSACTION;
-- 或
BEGIN;

-- 业务操作
INSERT INTO orders ...;
UPDATE inventory ...;

-- 提交或回滚
COMMIT;
-- 或
ROLLBACK;
```

**坑**：
- DDL 语句（CREATE/DROP/ALTER）**会隐式提交**事务
- 用 `set autocommit=0` 关闭自动提交（不推荐，影响连接池）
- 建议用 `START TRANSACTION` 显式开

### 8. 大事务的问题（高频坑！）

**大事务的危害**：
- 锁太多行，并发度低
- undo log 太大，回滚慢
- 主从延迟大（binlog 要等事务完成）
- 连接占用时间长

**解决方案**：
- 拆成小事务
- 用 `SELECT ... FOR UPDATE LIMIT n` 限制行数
- 异步处理大任务（用消息队列）

---

## ⭐ 重点掌握

### 9. 4 隔离级别实现原理对比

| 隔离级别 | 实现 | 锁 |
|---------|------|------|
| RU | 无锁 | 无 |
| RC | 每条 SELECT 重新生成快照 | 普通行锁 |
| RR | 事务开始时生成快照，整个事务用同一个 | **Next-Key Lock** |
| Serializable | 全用锁（没有快照读）| 全用共享锁 |

### 10. 幻读问题详解

**MySQL RR 级别**：
- 快照读：通过 MVCC + 一致性读解决（不阻塞）
- 当前读：通过 Next-Key Lock 解决（阻塞其他事务插入）

**SQL Server 默认 RC**：
- 没有 MVCC 快照读，每次读都是最新已提交
- 所以 RC 会有不可重复读 + 幻读
- 想要 RR 隔离级别要显式设置

### 11. 间隙锁的副作用

```sql
-- RR 级别
BEGIN;
SELECT * FROM t WHERE id > 100 FOR UPDATE;  -- 加间隙锁 (-∞, ∞)
-- 其他事务无法插入任何新行（！）
-- 后果：insert 会被阻塞直到超时
```

**生产经验**：
- 大事务 + FOR UPDATE 容易锁全表
- 建议用 `WHERE id > ? AND id < ?` 限制范围
- 或者用 `SKIP LOCKED`（MySQL 8.0+）

### 12. 查看当前隔离级别

```sql
-- 全局
SELECT @@global.transaction_isolation;
-- 当前会话
SELECT @@session.transaction_isolation;
-- 设置
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
```

### 13. SQL Server vs MySQL 事务对比（你专享）

| 特性 | SQL Server | MySQL |
|------|-----------|-------|
| 默认隔离 | **RC** | **RR** |
| 快照读 | 需要 SNAPSHOT 隔离（额外开销）| RR 级别默认就有 |
| 幻读解决 | 需要 SERIALIZABLE 或 SNAPSHOT | RR 级别 + Next-Key Lock |
| 锁粒度 | RID | 主键 |
| 嵌套事务 | SAVE TRANSACTION | SAVEPOINT |
| 分布式事务 | MSDTC | XA / Seata |

---

## 💡 选学（高级）

### 14. 当前读幻读案例

```sql
-- Session A
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
START TRANSACTION;
SELECT COUNT(*) FROM t_user WHERE age > 20;  -- 快照读，结果: 100

-- Session B（同时）
INSERT INTO t_user (age) VALUES (30);  -- 插入成功
COMMIT;

-- Session A 继续
SELECT COUNT(*) FROM t_user WHERE age > 20;  -- 仍是 100（快照读）

UPDATE t_user SET name = 'x' WHERE age > 20;  -- 当前读！
SELECT COUNT(*) FROM t_user WHERE age > 20 WHERE name = 'x';  -- 101（当前读看到新行）
-- 同一事务内出现幻读！RR 不是完全防幻读，只是快照读防
```

### 15. 分布式事务

- **2PC**：MySQL XA 事务，性能差，不推荐
- **TCC**：Try-Confirm-Cancel，业务侵入大
- **Seata**：阿里开源，主流方案
- **本地消息表**：最常用，异步最终一致
- **Seata AT 模式**：无侵入，自动回滚

---

## 🎯 面试最常被问的 5 个问题

1. **MySQL 默认隔离级别是什么？和 SQL Server 有什么区别？**
   - 答：MySQL RR，SQL Server RC。MySQL 在 RR 级别用 Next-Key Lock 防幻读

2. **什么是 MVCC？**
   - 答：每行有 DB_TRX_ID 和 DB_ROLL_PTR，快照读通过对比事务 ID + undo log 找到事务开始时的版本

3. **当前读和快照读的区别？**
   - 答：普通 SELECT 是快照读（不加锁，RR 下读事务开始时版本）；FOR UPDATE/UPDATE/DELETE 是当前读（加锁，读最新）

4. **幻读是什么？怎么解决？**
   - 答：同一范围两次读行数不同。MySQL RR 用 Next-Key Lock 防当前读的幻读；快照读天然防（用 MVCC）

5. **ACID 怎么实现？**
   - 答：A 原子性 = undo log；C 一致性 = AID 共同保证；I 隔离性 = 锁 + MVCC；D 持久性 = redo log

---

## 📋 速查表

```
| 概念       | 关键点 |
|------------|--------|
| 隔离默认    | MySQL RR，SQL Server RC |
| 幻读解决    | MySQL RR + Next-Key Lock |
| MVCC 字段  | DB_TRX_ID + DB_ROLL_PTR |
| 快照读     | 普通 SELECT（不阻塞）|
| 当前读     | FOR UPDATE / UPDATE（阻塞）|
| 原子性     | undo log |
| 持久性     | redo log |
| RR 锁     | 记录锁 + 间隙锁（Next-Key）|
```

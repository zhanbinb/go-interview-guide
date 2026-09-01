# MySQL 锁机制（行锁 + Next-Key Lock + 死锁）

> 面试**必问**章节，跟事务强相关
> 重点：行锁类型、Next-Key Lock、意向锁、死锁排查
> 比 SQL Server 复杂（Next-Key Lock 是 SQL Server 没有的）

---

## 🔥 必问核心

### 1. 全局锁 / 表级锁 / 行级锁（按粒度分）

| 锁类型 | 粒度 | 冲突 | 性能 | 场景 |
|--------|------|------|------|------|
| 全局锁 | 整个 MySQL | 所有 DML/DDL 阻塞 | 极差 | 备份（`FLUSH TABLES WITH READ LOCK`）|
| 表级锁 | 整张表 | 其他事务的写阻塞 | 差 | MyISAM、METADATA LOCK |
| 行级锁 | 单行 | 只冲突同一行 | 好 | **InnoDB 默认** |

**MySQL InnoDB 默认用行锁**，但**行锁是基于索引实现的**（重要！）

### 2. 共享锁（S Lock）vs 排他锁（X Lock）

| 锁类型 | 读 | 写 |
|--------|-----|------|
| 共享锁 S | ✅ 兼容 | ❌ 冲突 |
| 排他锁 X | ❌ 冲突 | ❌ 冲突 |

**SQL 对应的锁**：
- `SELECT ... LOCK IN SHARE MODE` → 加 S 锁
- `SELECT ... FOR UPDATE` → 加 X 锁
- `UPDATE / DELETE / INSERT` → 自动加 X 锁

### 3. 意向锁（Intent Lock，必问！）

**为什么需要意向锁？**

```
事务 A：SELECT * FROM t WHERE id = 1 FOR UPDATE;  -- 加行级 X 锁

事务 B：ALTER TABLE t ADD COLUMN ...;  -- 需要表级 X 锁
```

**没有意向锁的问题**：事务 B 扫描整张表的所有行，才能知道有没有行锁。慢！

**有意向锁**：
- 事务 A 加行 X 锁时，**自动给表加一个 IX 锁**（意向排他锁）
- 事务 B 想加表 X 锁时，**先看表有没有 IX 锁**，有则等待
- 避免扫描所有行

| 锁类型 | 含义 | 兼容性 |
|--------|------|--------|
| IS (意向共享) | 事务打算在某些行加 S 锁 | 与 IX/X 冲突，与 IS/S 兼容 |
| IX (意向排他) | 事务打算在某些行加 X 锁 | 与 S/X/IS/IX 都冲突 |

### 4. 行锁算法：记录锁 / 间隙锁 / Next-Key Lock

**InnoDB 行锁的 3 种算法**（按范围分）：

| 算法 | 锁定范围 | 解决什么 |
|------|----------|----------|
| **记录锁 Record Lock** | 单个索引记录 | 单行更新冲突 |
| **间隙锁 Gap Lock** | 索引记录之间的间隙 | 阻止插入（防幻读）|
| **Next-Key Lock** | 记录 + 间隙（前开后闭）| RR 级别防幻读 |

**Next-Key Lock 详解**（MySQL 防幻读的核心）：

```sql
-- 假设有 id: 1, 5, 10, 15

SELECT * FROM t WHERE id BETWEEN 5 AND 10 FOR UPDATE;
-- 加的锁：
--   记录锁: id=5, id=10
--   间隙锁: (-∞, 5), (5, 10), (10, 15)
-- 其他事务不能插入 id=3, 6, 7, 8, 9, 11, 12, 13, 14
```

**Next-Key Lock = 记录锁 + 它前面的间隙锁**

### 5. SQL Server vs MySQL 锁对比

| 维度 | SQL Server | MySQL |
|------|-----------|-------|
| 隔离默认 | RC | RR |
| 锁粒度（行） | RID | 主键 / 索引项 |
| 间隙锁 | ❌（用 SERIALIZABLE 模拟）| ✅ Next-Key Lock |
| 意向锁 | ✅ | ✅ |
| 乐观锁 | `ROWVERSION` 列 | 版本号字段 + `WHERE version = ?` |
| 死锁检测 | 自动回滚成本低的事务 | 自动回滚（一样）|

**最关键差异**：MySQL RR 级别用 Next-Key Lock 防幻读（SQL Server 做不到）

---

## ⭐ 重点掌握

### 6. 行锁的"陷阱"：基于索引！

**重要**：**InnoDB 行锁是加在索引上的，不是数据行上**

```sql
-- id 是主键（有索引）
SELECT * FROM t WHERE id = 1 FOR UPDATE;  -- 行锁，加在 id=1

-- name 无索引
SELECT * FROM t WHERE name = 'Alice' FOR UPDATE;
-- 走全表扫描 → 锁定所有行（可能锁全表！）
```

**死循环坑**：
```sql
-- 即使你只查 1 行，但 WHERE 条件没用上索引 → 锁全表
UPDATE t SET name = 'x' WHERE name = 'Alice';
-- 如果 name 无索引，会锁住所有行！
```

**解决**：所有 WHERE 条件都加索引

### 7. 乐观锁 vs 悲观锁

| 类型 | 实现 | 适用场景 |
|------|------|----------|
| **悲观锁** | `SELECT FOR UPDATE` 加锁 | 写多读少、冲突多 |
| **乐观锁** | 版本号 + CAS | 读多写少、冲突少 |

**乐观锁 MySQL 实现**（无内置，靠应用层）：
```sql
-- 1. 表加 version 列
ALTER TABLE t ADD COLUMN version INT NOT NULL DEFAULT 0;

-- 2. 更新时带 version 条件
UPDATE t SET name = 'new', version = version + 1
WHERE id = 1 AND version = 0;
-- 如果影响行数 = 0，说明版本冲突

-- 3. 失败时重试
```

### 8. 死锁排查（必问实战）

**死锁经典场景**：

```
Session A:                Session B:
BEGIN;                    BEGIN;
UPDATE t SET ... WHERE id=1;  -- 锁 id=1
                          UPDATE t SET ... WHERE id=2;  -- 锁 id=2
UPDATE t SET ... WHERE id=2;  -- 等 id=2 的锁
                          UPDATE t SET ... WHERE id=1;  -- 等 id=1 的锁
                                                -- 死锁！
```

MySQL 自动检测死锁，回滚**成本较小的事务**。

**排查命令**：
```sql
-- 查看最近一次死锁信息
SHOW ENGINE INNODB STATUS\G

-- 找 LATEST DETECTED DEADLOCK 段
-- 里面有死锁的 SQL、事务信息
```

**预防死锁**：
1. 保持事务小
2. 固定顺序访问资源（都先 id=1 再 id=2）
3. 加合适的索引（减少锁范围）
4. 降低隔离级别（如用 RC 而非 RR）
5. 加锁超时自动回滚（`innodb_lock_wait_timeout`）

### 9. 死锁 vs 锁等待

| 现象 | 原因 | 默认行为 |
|------|------|---------|
| 锁等待 | 事务 A 持锁，事务 B 等 | 等待 `innodb_lock_wait_timeout`（默认 50s）|
| 死锁 | 事务 A 和 B 互相等待 | MySQL 自动检测，回滚一个事务 |

**锁等待超时配置**：
```sql
SHOW VARIABLES LIKE 'innodb_lock_wait_timeout';
SET GLOBAL innodb_lock_wait_timeout = 10;  -- 10 秒超时
```

### 10. MySQL 8.0 新增：SKIP LOCKED 和 NOWAIT

**SKIP LOCKED**（跳过已锁定的行）：
```sql
-- 队列场景：worker 抢任务，跳过被其他 worker 锁定的
SELECT * FROM t_task WHERE status = 'pending'
  ORDER BY id LIMIT 10
  FOR UPDATE SKIP LOCKED;
```

**NOWAIT**（不等锁直接报错）：
```sql
SELECT * FROM t WHERE id = 1 FOR UPDATE NOWAIT;
-- 如果被锁，立即报错（不等待）
-- ERROR 3572: Statement aborted because lock(s) could not be acquired immediately
```

### 11. 自增锁（自增 ID 竞争）

**问题**：并发 INSERT 怎么保证自增 ID 唯一？

**MySQL 解决方案**：自增锁 + 轻量级锁

```sql
SHOW VARIABLES LIKE 'innodb_autoinc_lock_mode';
-- 0: 传统（每次都锁，串行）
-- 1: 连续（默认，批量插入不锁）
-- 2: 混合（推荐，简单 INSERT 锁少，批量 INSERT 锁多）
```

**生产建议**：用默认 1

---

## 💡 选学

### 12. 元数据锁（Metadata Lock, MDL）

**MySQL 5.5 引入**，自动加锁保护表结构不被并发修改。

```sql
-- Session A: 事务里查 t 表
START TRANSACTION;
SELECT * FROM t;  -- 自动加 MDL 读锁

-- Session B: 想改 t 表结构
ALTER TABLE t ADD COLUMN ...;  -- 等待 MDL 读锁释放！
-- 结果：所有 DDL 在事务执行时会被阻塞！
```

**坑**：长事务 + 频繁查询 → 后续所有 ALTER TABLE 都会排队

**解决**：监控长事务，DDL 放在低峰期

### 13. 锁等待的 5 个诊断 SQL

```sql
-- 1. 当前所有事务
SELECT * FROM information_schema.INNODB_TRX\G

-- 2. 锁等待关系
SELECT 
    r.trx_id AS waiting_trx,
    r.trx_mysql_thread_id AS waiting_thread,
    b.trx_id AS blocking_trx,
    b.trx_mysql_thread_id AS blocking_thread
FROM information_schema.INNODB_LOCK_WAITS w
JOIN information_schema.INNODB_TRX r ON w.requesting_trx_id = r.trx_id
JOIN information_schema.INNODB_TRX b ON w.blocking_trx_id = b.trx_id;

-- 3. 当前锁信息
SELECT * FROM performance_schema.data_locks LIMIT 10\G

-- 4. 当前运行的 SQL
SELECT * FROM information_schema.PROCESSLIST WHERE INFO IS NOT NULL;

-- 5. 死锁日志
SHOW ENGINE INNODB STATUS\G
```

### 14. MySQL 8.0 性能元数据锁改进

8.0 用 `NOWAIT` / `SKIP LOCKED` 让 DDL 也能跳过锁：
```sql
ALTER TABLE t ADD COLUMN x INT, ALGORITHM=INSTANT, LOCK=NONE;
-- ALGORITHM=INSTANT: 不需要 rebuild 表（瞬时加列）
-- LOCK=NONE: 不锁表（在线 DDL）
```

---

## 🎯 面试最常被问的 5 个问题

1. **MySQL 锁有哪几种？**
   - 全局锁 / 表级锁 / 行级锁；行锁又分记录锁/间隙锁/Next-Key Lock

2. **什么是 Next-Key Lock？解决什么问题？**
   - 记录锁 + 间隙锁（前开后闭），RR 级别防幻读

3. **为什么需要意向锁？**
   - 避免扫描所有行判断表级锁冲突；加行锁时自动加表级意向锁

4. **什么是死锁？怎么排查？怎么避免？**
   - 互相等待对方持有的锁。SHOW ENGINE INNODB STATUS 看死锁日志。预防：固定访问顺序、加索引、小事务

5. **乐观锁 vs 悲观锁？MySQL 怎么实现？**
   - 悲观用 FOR UPDATE；乐观用 version 列 + CAS。无内置乐观锁，靠应用层实现

---

## 📋 速查表

```
| 锁             | 粒度    | 解决什么 |
|----------------|---------|----------|
| 全局锁         | 整个 DB | 备份     |
| 表级锁         | 整表    | DDL     |
| 记录锁         | 单行    | 单行更新 |
| 间隙锁         | 索引间隙 | 防插入   |
| Next-Key Lock  | 记录+间隙 | RR 防幻读 |
| 意向锁 IS/IX   | 表级   | 加速DDL判断 |
| MDL            | 表结构  | 保护元数据 |
| 自增锁          | AUTO_INCREMENT | 唯一性 |

排查：SHOW ENGINE INNODB STATUS
预防：固定顺序 + 加索引 + 小事务
8.0 新增：SKIP LOCKED / NOWAIT
```

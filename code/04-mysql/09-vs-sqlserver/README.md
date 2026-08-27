# MySQL vs SQL Server 对比（SQL Server 用户专享）

> 你的背景：熟悉 SQL Server，准备面 MySQL 岗位
> 目标：把 SQL Server 知识迁移到 MySQL，理解核心差异
> 风格：精简 + 对比表 + 真实场景

---

## 🔥 必问核心（面试加分）

### 1. 事务隔离级别（最大差异！）

| 隔离级别 | SQL Server 默认 | MySQL InnoDB 默认 |
|---------|----------------|-------------------|
| READ UNCOMMITTED | 支持 | 支持 |
| READ COMMITTED  | ✅ **默认** | 支持 |
| REPEATABLE READ | 支持 | ✅ **默认** |
| SNAPSHOT        | 支持（需开启）| **不支持**（用 MVCC 替代）|
| SERIALIZABLE    | 支持 | 支持 |

**关键差异**：
- **SQL Server 默认 RC**（读已提交）
- **MySQL 默认 RR**（可重复读）
- 面试必问：**MySQL 为什么默认 RR？怎么避免幻读？**
  - 答：用 **Next-Key Lock**（记录锁 + 间隙锁）阻止范围插入

### 2. 自增主键

| 特性 | SQL Server | MySQL |
|------|-----------|-------|
| 语法 | `IDENTITY(1,1)` | `AUTO_INCREMENT` |
| 获取刚插入的 ID | `SCOPE_IDENTITY()` | `LAST_INSERT_ID()` |
| 起始值设置 | `IDENTITY(100,1)` | `AUTO_INCREMENT=100` |
| 步长 | `IDENTITY(1,2)`（步长2）| 只能 1（要步长用 sequence）|
| 雪花算法 | 不支持 | **不支持**（用 sequence 或应用层）|

**面试加分**：为什么 MySQL 官方建议用自增主键？
- 答：聚簇索引顺序插入，减少页分裂；UUID 随机写性能差

### 3. 字符串类型

| 场景 | SQL Server | MySQL |
|------|-----------|-------|
| 定长 | `CHAR(n)` | `CHAR(n)` |
| 变长 | `VARCHAR(n)` | `VARCHAR(n)` |
| 超大文本 | `VARCHAR(MAX)` | `TEXT/LONGTEXT`（类型不同）|
| Unicode | `NVARCHAR(n)` | `VARCHAR(n)` + utf8mb4 字符集 |
| 二进制 | `VARBINARY(MAX)` | `BLOB/LONGBLOB` |

**关键差异**：
- MySQL 的 `VARCHAR(n)` 中 **n 是字符数**（不是字节数）
- 但**实际占空间 = 字符数 × 最大字节宽度**（utf8mb4 = 4 字节/字符）
- 所以 `VARCHAR(255)` 在 utf8mb4 下最大占 255*4 = 1020 字节（超过 768 字节需要用 TEXT 或独立列）

### 4. LIMIT vs TOP/OFFSET

| 用法 | SQL Server | MySQL |
|------|-----------|-------|
| 取前 10 | `SELECT TOP 10 * FROM t` | `SELECT * FROM t LIMIT 10` |
| 跳过 20 取 10 | `OFFSET 20 ROWS FETCH NEXT 10 ROWS ONLY`（2012+）| `LIMIT 10 OFFSET 20` 或 `LIMIT 20, 10` |
| 取第 100-120 行 | `OFFSET 100 ROWS FETCH NEXT 20 ROWS ONLY` | `LIMIT 20 OFFSET 100` |

### 5. NULL 处理

| 函数 | SQL Server | MySQL |
|------|-----------|-------|
| 判断 NULL | `IS NULL` | `IS NULL` |
| 转 NULL | `ISNULL(col, 0)` | `IFNULL(col, 0)` 或 `COALESCE(col, 0)` |
| NULL 拼接 | `CONCAT(a,b)`（自动忽略 NULL）| `CONCAT(a,b)`（NULL 变 NULL）|
| NULL 比较 | `NULL = NULL` → NULL | `NULL = NULL` → NULL |

### 6. 日期时间

| 类型 | SQL Server | MySQL |
|------|-----------|-------|
| 日期时间 | `DATETIME`（精度 3.33ms）| `DATETIME`（精度 1s）|
| 高精度 | `DATETIME2(7)`（精度 100ns）| `DATETIME(6)`（精度 1μs）|
| 时间戳 | `TIMESTAMP`（8字节）| `TIMESTAMP`（4字节，2038年问题）|
| 当前时间 | `GETDATE()` | `NOW()` / `CURRENT_TIMESTAMP` |
| 日期加减 | `DATEADD(DAY, 1, col)` | `DATE_ADD(col, INTERVAL 1 DAY)` |
| 日期差 | `DATEDIFF(DAY, a, b)` | `DATEDIFF(a, b)`（参数顺序相反）|

**MySQL 的坑**：
- `TIMESTAMP` 受时区影响，`DATETIME` 不受影响
- 推荐用 `DATETIME` 存业务时间

### 7. 字符集（高频坑！）

| 字符集 | SQL Server | MySQL |
|--------|-----------|-------|
| 默认 | `nvarchar` 默认 UCS-2 | `utf8mb4`（MySQL 5.7+ 推荐，8.0 默认）|
| 历史坑 | 无 | 老 `utf8` 是 3 字节（不能存 emoji）|
| emoji | 直接 NVARCHAR | 必须用 `utf8mb4` |

**MySQL 迁移必做**：
```sql
-- 建库时指定
CREATE DATABASE mydb DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- 表/列也可以单独指定
ALTER TABLE t CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 8. 临时表

| 特性 | SQL Server | MySQL |
|------|-----------|-------|
| 语法 | `SELECT * INTO #temp FROM t` | `CREATE TEMPORARY TABLE temp_t AS SELECT * FROM t` |
| 作用域 | 会话（连接）级别 | 会话（连接）级别 |
| 全局临时表 | `##global_temp`（所有会话可见）| ❌ 不支持 |
| 自动清理 | 会话结束 | 会话结束（连接断开）|

**MySQL 临时表注意**：
- 不能用事务回滚
- 索引只能用 InnoDB（不能 FULLTEXT）

---

## ⭐ 重点掌握

### 9. 索引差异

| 特性 | SQL Server | MySQL InnoDB |
|------|-----------|--------------|
| 聚簇索引 | ✅（数据按聚簇键排序）| ✅（默认按主键聚簇）|
| 覆盖索引 | ✅ | ✅ |
| 包含列索引 | ✅ `INCLUDE` 子句 | ❌（用联合索引替代）|
| 过滤索引 | ✅ `WHERE` 子句 | ❌（8.0 部分支持）|
| 列存储索引 | ✅（2012+）| ❌ |
| 全文索引 | ✅ | ✅（InnoDB 5.6+ 支持）|
| 索引提示 | `WITH (INDEX(idx))` | `USE INDEX / FORCE INDEX` |

**关键差异**：
- MySQL InnoDB **表数据本身就是按主键聚簇存储**（SQL Server 是聚簇索引指向数据行）
- 这意味着 MySQL 主键选择更重要

### 10. 自增溢出

| DB | 上限 | 解决方案 |
|----|------|----------|
| SQL Server `INT` | 2,147,483,647（约 21 亿）| 改 `BIGINT` |
| MySQL `INT UNSIGNED` | 4,294,967,295（约 42 亿）| 改 `BIGINT UNSIGNED` |
| MySQL `INT`（有符号）| 同 SQL Server 21 亿 | 改 `BIGINT` |

### 11. JSON 支持

| 特性 | SQL Server 2016+ | MySQL 5.7+ |
|------|-----------------|-------------|
| JSON 类型 | `NVARCHAR(MAX)` + `ISJSON` / `JSON_VALUE` | 原生 `JSON` 类型 |
| 查询 | `JSON_VALUE(col, '$.key')` | `col->>'$.key'` / `JSON_EXTRACT` |
| 索引 | 计算列 + 索引 | `CAST(... AS CHAR) + 索引` |
| 性能 | 较好 | 一般（大 JSON 慢）|

### 12. 分页查询（深分页问题）

| 写法 | SQL Server | MySQL |
|------|-----------|-------|
| 标准 | `OFFSET 100 ROWS FETCH NEXT 20 ROWS ONLY` | `LIMIT 20 OFFSET 100` |
| **深分页问题** | OFFSET 越大越慢 | **更明显**（OFFSET 100万 扫描 100万 + 20 行）|
| 优化（子查询）| `WHERE id > (SELECT id FROM t ORDER BY id OFFSET 100) FETCH NEXT 20` | `WHERE id > ? LIMIT 20` |

**深分页解决方案**：
```sql
-- 错误（深分页慢）
SELECT * FROM t ORDER BY id LIMIT 10 OFFSET 1000000;

-- 正确（用上次查询的最大 ID 游标分页）
SELECT * FROM t WHERE id > 1000000 ORDER BY id LIMIT 10;
```

---

## 💡 易踩的坑（生产经验）

### 13. 不要在生产用 `SELECT *`

| DB | 行为 |
|----|------|
| SQL Server | 会返回所有列，包括将来加的（可能导致代码兼容性）|
| MySQL | 同上，**但** TEXT/BLOB 字段可能截断，且查询更慢 |

### 14. 大表 `COUNT(*)` 性能

| DB | 行为 |
|----|------|
| SQL Server | 全表扫描（可加索引优化）|
| MySQL InnoDB | **不缓存行数**，每次都全表扫描（`COUNT(*) WHERE` 也一样）|
| MySQL MyISAM | **缓存行数**（快）但不支持事务 |

**MySQL 优化**：
- 维护计数表：`UPDATE count_t SET cnt = (SELECT COUNT(*) FROM big_t)`
- 或用 `information_schema.tables` 的 `TABLE_ROWS`（不准确）

### 15. 字符串拼接

| DB | 语法 |
|----|------|
| SQL Server | `SELECT 'a' + 'b'`（注意：`+` NULL 变 NULL）|
| MySQL | `SELECT CONCAT('a', 'b')` |
| SQL Server 多行 | `SELECT STRING_AGG(col, ',') FROM t` |
| MySQL 多行 | `SELECT GROUP_CONCAT(col SEPARATOR ',') FROM t` |

### 16. 字符串比较

| 行为 | SQL Server | MySQL |
|------|-----------|-------|
| 默认排序规则 | `SQL_Latin1_General_CP1_CI_AS` | `utf8mb4_0900_ai_ci`（8.0 默认）|
| 大小写敏感 | `WHERE col = 'ABC'` 默认不敏感 | 同上（8.0 默认不敏感）|
| 尾部空格 | `WHERE col = 'abc'` 不匹配 'abc ' | **不忽略尾部空格**（可能漏匹配）|

### 17. 布尔值

| DB | 存储 | 字面量 |
|----|------|--------|
| SQL Server | `BIT` | `1` / `0` / `TRUE` / `FALSE` |
| MySQL | `TINYINT(1)` 或 `BOOLEAN`（别名）| `1` / `0`（`TRUE` 是 `1` 的别名）|

---

## 🎯 面试最常被问的 5 个对比

1. **隔离级别默认不同**（RC vs RR）— 答案：MySQL 用 Next-Key Lock 避免幻读
2. **自增主键** — MySQL 推荐自增 UUID 性能差
3. **varchar vs varchar(MAX)** — MySQL 用 TEXT
4. **深分页** — OFFSET 越大越慢，用游标分页
5. **utf8 vs utf8mb4** — 老 utf8 不能存 emoji

---

## 📋 速查表（打印贴墙上）

```
| 概念     | SQL Server        | MySQL                |
|----------|-------------------|----------------------|
| 隔离默认 | RC                | RR                   |
| 自增     | IDENTITY(1,1)     | AUTO_INCREMENT       |
| 顶部     | TOP 10            | LIMIT 10              |
| 跳过     | OFFSET 20 FETCH 10| LIMIT 10 OFFSET 20   |
| NULL处理 | ISNULL(a,b)       | IFNULL(a,b)          |
| 当前时间 | GETDATE()         | NOW()                 |
| 时间戳   | TIMESTAMP (8字节) | TIMESTAMP (4字节)    |
| 字符串大 | VARCHAR(MAX)      | TEXT/LONGTEXT         |
| Unicode  | NVARCHAR          | VARCHAR + utf8mb4     |
| 临时表   | #temp             | CREATE TEMP TABLE    |
| 字符集   | nvarchar(默认)    | utf8mb4(8.0默认)     |
| JSON     | NVARCHAR+函数     | JSON (原生)          |
| 分页     | OFFSET/FETCH      | LIMIT/OFFSET         |
| 行数     | 缓存(Heap)        | 不缓存(每次扫表)     |
| 布尔     | BIT               | TINYINT(1) / BOOLEAN |
| 字符串拼接| +                | CONCAT()             |
```

---

## 🚀 实践建议（迁移到 MySQL 时）

1. **建库必加**：`CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`
2. **主键推荐自增 INT UNSIGNED 或 BIGINT**
3. **深分页必须用游标**（不能 OFFSET 100万）
4. **大表 COUNT(*) 用计数表**（不能直接 COUNT）
5. **避免 SELECT \*\***（特别是含 TEXT 字段）
6. **时间用 DATETIME**（不要 TIMESTAMP）

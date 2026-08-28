# MySQL 索引（最高频考点）

> 这是 MySQL 面试**最高频**的章节，必问中的必问
> 涵盖：B+Tree、聚簇/非聚簇、联合索引、最左前缀、覆盖索引、索引下推、EXPLAIN、索引失效

---

## 🔥 必问核心

### 1. 为什么用索引能加快查询？

**没索引**：全表扫描（顺序读 N 行）
**有索引**：B+Tree 树查找（log(N) 复杂度）

举例：1000 万行的表
- 无索引：扫 1000 万行
- 有索引：树高 3-4 层，只需 3-4 次磁盘 IO

### 2. B+Tree 索引结构（必画图）

```
                [50]
              /     \
        [20, 30]   [70, 80]
        /  |  \    /  |  \
      1  20 30  50 70 80  100     ← 叶子节点（数据页）
      |  |  |   |  |  |   |       ← 同一层，链表连接
      ↓  ↓  ↓   ↓  ↓  ↓   ↓
     ... ...   ... ...   ...

特点：
- 叶子节点**包含所有数据**（聚簇索引）或**主键指针**（非聚簇）
- 叶子节点**有序链表**连接（范围查询快）
- 非叶子节点只存键值（树矮，IO 少）
```

**对比 B 树**：
- B+Tree：数据都在叶子（树更矮，范围扫快）
- B 树：数据在每层（树更高，范围扫要回溯）

### 3. 聚簇 vs 非聚簇索引（核心区分！）

| 类型 | MySQL InnoDB | SQL Server |
|------|--------------|-------------|
| 聚簇索引 | ✅ **表数据本身就是按主键聚簇存储** | ✅（聚簇键是数据行的物理顺序）|
| 非聚簇索引 | ✅ 二级索引叶子存**主键值**（不是指针）| ✅ 叶子存**RID**（行标识符）|

**MySQL 关键差异**：
- 聚簇索引**只有 1 个**（一般是主键）
- 二级索引查询需要**回表**（拿主键再去聚簇索引找完整数据）
- 这就是为什么**主键长度会影响所有二级索引大小**（主键越长，二级索引越大）

**SQL Server 关键差异**：
- 二级索引叶子存 RID（行号）
- 回表直接按 RID 找（一步）
- 聚簇索引键可改变（不像 MySQL 主键改不了）

### 4. 联合索引（复合索引）

**最左前缀原则**（必问）：

```sql
-- 创建联合索引
CREATE INDEX idx_user ON t_user (name, age, city);

-- 走索引：
WHERE name = 'Alice'              -- ✅ 用到 name
WHERE name = 'Alice' AND age = 25  -- ✅ 用到 name, age
WHERE name = 'Alice' AND age = 25 AND city = 'BJ'  -- ✅ 全部
WHERE age = 25                    -- ❌ 不走（跳过了 name）
WHERE city = 'BJ'                 -- ❌ 不走（跳过了 name）
WHERE name = 'Alice' AND city = 'BJ'  -- ⚠️ 只用 name（age 断了）
```

**口诀**：从最左开始，不能跳过中间列

### 5. 覆盖索引（避免回表）

**回表问题**：用非主键索引查时，先找到主键，再回聚簇索引查整行 → 两次 IO

**覆盖索引**：查询的列**全部在索引里**，不用回表

```sql
-- 索引: (name, age)
SELECT name, age FROM t_user WHERE name = 'Alice';  -- ✅ 覆盖
SELECT * FROM t_user WHERE name = 'Alice';          -- ❌ 需要回表
```

**面试必问**：怎么优化 `SELECT *`？
- 答：用覆盖索引；只 SELECT 需要的列

### 6. 索引下推（ICP，MySQL 5.6+）

**没有 ICP**：
```sql
-- 索引: (name, age)
SELECT * FROM t_user WHERE name = 'Alice' AND age > 20;
-- 1. 用 name 索引找到所有 name='Alice' 的主键
-- 2. 全部回表（10 行）
-- 3. 在内存里过滤 age > 20
```

**有 ICP**：
```sql
-- 1. 用 name 索引找到所有 name='Alice' 的主键
-- 2. 在索引层就过滤 age > 20（只回表 2 行）
-- 3. 节省 8 次回表
```

**如何确认 ICP 生效**：`EXPLAIN` 的 Extra 列出现 `Using index condition`

### 7. EXPLAIN 详解（必会读！）

```sql
EXPLAIN SELECT * FROM t_user WHERE name = 'Alice' AND age = 25;
```

| 字段 | 含义 | 关键值 |
|------|------|--------|
| `id` | SELECT 编号 | 越大越先执行 |
| `select_type` | 查询类型 | SIMPLE/PRIMARY/SUBQUERY |
| `table` | 表名 | - |
| `type` | **访问类型（性能关键）** | system > const > eq_ref > ref > range > index > **ALL** |
| `possible_keys` | 可能用到的索引 | - |
| `key` | **实际用到的索引** | NULL = 全表扫 |
| `key_len` | 索引使用长度（字节）| 越短越好 |
| `ref` | 索引哪一列被引用 | const / func |
| `rows` | **预估扫描行数** | 越少越好 |
| `Extra` | **额外信息（很重要）** | 见下表 |

**Extra 关键值**：
- `Using index`：覆盖索引（好）
- `Using where`：用 WHERE 过滤
- `Using index condition`：ICP 生效（好）
- `Using filesort`：需要额外排序（差）
- `Using temporary`：用了临时表（差）
- `Using join buffer`：join 缓冲（差）
- `NULL`：没用到索引（差）

### 8. 索引失效场景（10 种必记）

```sql
-- 假设索引: idx_name_age (name, age)

-- ❌ 1. 不满足最左前缀
WHERE age = 25

-- ❌ 2. 范围查询导致后续列失效
WHERE name = 'Alice' AND age > 20  -- age 走索引
WHERE name = 'Alice' AND age = 20 AND city = 'BJ'  -- 走 name, age;city 失效

-- ❌ 3. 在索引列上做运算/函数
WHERE SUBSTRING(name, 1, 3) = 'Ali'
WHERE name + 'X' = 'AliceX'  -- 改成 name = 'Alice'

-- ❌ 4. 隐式类型转换
WHERE name = 123  -- name 是 varchar,123 是 int → 索引失效
WHERE name = '123'  -- 正确

-- ❌ 5. 隐式字符集转换
-- 两个表字符集不同，join 时索引失效（5.7 以后优化）

-- ❌ 6. LIKE 以 % 开头
WHERE name LIKE '%Alice'  -- 失效
WHERE name LIKE 'Alice%'  -- 走索引
WHERE name LIKE '%Alice%' -- 失效

-- ❌ 7. OR 前后有非索引列
WHERE name = 'Alice' OR age = 25  -- age 无索引，全表扫

-- ❌ 8. NOT、!=、<> 可能失效
WHERE name != 'Alice'  -- MySQL 不走索引（SQL Server 看统计信息）

-- ❌ 9. IS NULL / IS NOT NULL
WHERE name IS NULL  -- 看基数，低基数时可能失效
WHERE name IS NOT NULL -- 同上

-- ❌ 10. IN 太多值
WHERE id IN (1,2,3,...,10000)  -- 可能走全表扫（超过 30%）
```

### 9. 为什么建议用自增主键？

- ✅ 顺序插入：每次都往最后追加，不移动已有数据
- ❌ UUID 随机：每次插入随机位置，频繁页分裂
- ✅ 自增主键小（4-8 字节），二级索引也小

```sql
-- 推荐
CREATE TABLE t (id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, ...);

-- 不推荐（除非有特殊需求）
CREATE TABLE t (id CHAR(36) PRIMARY KEY, ...);  -- 36字节,UUID
```

### 10. 索引设计原则（实战）

1. **不为频繁更新的列建索引**（更新维护成本高）
2. **不重复建索引**（如 (a) 和 (a, b) 重复）
3. **联合索引优于单列索引**（可覆盖更多查询）
4. **长字符串用前缀索引**：`ALTER TABLE t ADD INDEX idx (col(10))`
5. **基数（区分度）低的列不建索引**（如性别、状态）
6. **NULL 多的列慎建**（索引效率低）
7. **小表不建索引**（全表扫更快）

---

## ⭐ 重点掌握

### 11. 前缀索引（应对长字符串）

```sql
-- 问题：email 字段 VARCHAR(100)，建完整索引太大
-- 解决：前缀索引
CREATE INDEX idx_email ON t_user (email(10));

-- 取多少前缀？找区分度最高的点
SELECT 
    COUNT(DISTINCT LEFT(email, 5)) / COUNT(*) AS sel_5,
    COUNT(DISTINCT LEFT(email, 10)) / COUNT(*) AS sel_10,
    COUNT(DISTINCT LEFT(email, 15)) / COUNT(*) AS sel_15
FROM t_user;
-- 选区分度接近 1.0 的最小前缀长度
```

**缺点**：不能用 ORDER BY 和 GROUP BY（部分场景）

### 12. 索引合并（Index Merge）

MySQL 5.0+ 支持：
- `Using union(idx_a, idx_b)`：OR 条件用两个索引
- `Using intersect(idx_a, idx_b)`：AND 条件用两个索引

**少见且性能不一定好**，通常建议改写为联合索引。

### 13. COUNT(*) 用哪个索引？

- **有二级索引**：走二级索引（比全表扫小）
- **没有索引**：全表扫
- **小技巧**：用 `COUNT(1)` 跟 `COUNT(*)` 等价（8.0 优化了）

### 14. 索引下推 vs 覆盖索引 区别

| 概念 | 解决什么 |
|------|----------|
| 覆盖索引 | 不用回表（避免主键查找）|
| 索引下推（ICP）| 在索引层就过滤 WHERE 条件（减少回表次数）|

---

## 💡 选学（高级）

### 15. 自适应哈希索引（InnoDB）

- InnoDB 自动判断，热点页加载到内存时建哈希
- 用户不可控，只能关（`innodb_adaptive_hash_index = OFF`）
- 等值查询快（接近 O(1)）

### 16. 全文索引（FULLTEXT）

- 5.6+ InnoDB 支持
- 替代 LIKE '%keyword%'
- 中文需要 ngram 分词器（5.7+）

```sql
CREATE FULLTEXT INDEX ft_content ON t_article (content) WITH PARSER ngram;
SELECT * FROM t_article WHERE MATCH(content) AGAINST('MySQL' IN NATURAL LANGUAGE MODE);
```

### 17. 函数索引（MySQL 8.0+）

```sql
-- 对 email 的 lowercase 建索引
CREATE INDEX idx_email_lower ON t_user ((LOWER(email)));

SELECT * FROM t_user WHERE LOWER(email) = 'alice@example.com';
-- 现在走索引
```

---

## 🎯 面试最常被问的 5 个问题

1. **B+Tree 和 B 树有什么区别？为什么用 B+Tree？**
   - 答：数据全在叶子节点（树更矮），叶子链表连接（范围扫快）

2. **聚簇索引和非聚簇索引的区别？**
   - 答：聚簇索引的叶子就是数据行（数据按索引顺序存储）；非聚簇叶子存主键或 RID

3. **最左前缀原则是什么？**
   - 答：联合索引从最左列开始，不能跳过中间列

4. **什么是覆盖索引？怎么避免回表？**
   - 答：查询列全部在索引里；只 SELECT 需要的列

5. **索引什么时候失效？**
   - 答：函数/运算/类型转换/LIKE %开头/OR 等

---

## 📋 速查表

```
| 概念         | 关键点 |
|--------------|--------|
| 索引结构      | B+Tree（数据全在叶子 + 链表连接）|
| 聚簇索引      | 1 个（数据按主键聚簇）|
| 二级索引      | 叶子存主键（不是 RID）|
| 联合索引      | 最左前缀，不能跳列 |
| 覆盖索引      | 避免回表 |
| 索引下推      | 在索引层过滤（少回表）|
| 主键建议      | 自增 BIGINT（顺序写）|
| 失效场景      | 函数/类型转换/LIKE %/OR |
| EXPLAIN type  | system < const < eq_ref < ref < range < index < ALL |
| EXPLAIN Extra | Using index > Using index condition > Using where > NULL |
```

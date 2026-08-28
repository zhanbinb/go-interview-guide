-- ============================================================
-- MySQL 索引 demo
-- 涵盖：B+Tree、聚簇/非聚簇、联合索引、覆盖索引、ICP、失效场景
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- ============================================================

SELECT '=== MySQL 索引 Demo（高频考点）===' AS title;
SELECT VERSION() AS mysql_version;

-- 准备测试库
DROP DATABASE IF EXISTS demo_index;
CREATE DATABASE demo_index DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_index;

-- ============================================================
-- 建表 + 准备数据（10万行模拟）
-- ============================================================
SELECT '--- 建表 + 数据 ---' AS step;

CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    city VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 单列索引
CREATE INDEX idx_name ON t_user (name);
-- 联合索引
CREATE INDEX idx_name_age ON t_user (name, age);
-- 联合索引（三个字段）
CREATE INDEX idx_name_age_city ON t_user (name, age, city);

-- 批量插入 10 万行（用 stored procedure）
DELIMITER //
CREATE PROCEDURE insert_users()
BEGIN
    DECLARE i INT DEFAULT 0;
    WHILE i < 100000 DO
        INSERT INTO t_user (name, age, city, email) VALUES (
            CONCAT('user_', FLOOR(RAND() * 1000)),
            FLOOR(20 + RAND() * 40),
            ELT(FLOOR(1 + RAND() * 5), 'Beijing', 'Shanghai', 'Guangzhou', 'Shenzhen', 'Hangzhou'),
            CONCAT('user_', i, '@example.com')
        );
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;

CALL insert_users();
SELECT COUNT(*) AS total_rows FROM t_user;

-- ============================================================
-- Demo 1: EXPLAIN 基础
-- ============================================================
SELECT '--- Demo 1: EXPLAIN 看索引使用情况 ---' AS step;

-- 走主键
EXPLAIN SELECT * FROM t_user WHERE id = 100;

-- 走 idx_name
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500';

-- 走 idx_name_age（最左前缀）
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' AND age = 30;

-- ============================================================
-- Demo 2: 最左前缀验证
-- ============================================================
SELECT '--- Demo 2: 最左前缀原则 ---' AS step;

-- ✅ 用到 name
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500';
-- 注意 key 列 = idx_name

-- ❌ 跳过 name，只用 age（不走索引）
EXPLAIN SELECT * FROM t_user WHERE age = 30;
-- 注意 key 列 = NULL（全表扫）

-- ⚠️ 用到 name + age，city 失效（范围之后中断）
EXPLAIN SELECT * FROM t_user 
WHERE name = 'user_500' AND age > 25 AND city = 'Beijing';
-- 注意 key_len 只包含 name+age 的长度

-- ============================================================
-- Demo 3: 覆盖索引（避免回表）
-- ============================================================
SELECT '--- Demo 3: 覆盖索引 vs 回表 ---' AS step;

-- ❌ SELECT * 需要回表
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500';
-- Extra: NULL (没覆盖索引)

-- ✅ 只查索引包含的列（覆盖索引）
EXPLAIN SELECT name FROM t_user WHERE name = 'user_500';
-- Extra: Using index (覆盖索引)

-- ✅ 联合索引覆盖
EXPLAIN SELECT name, age FROM t_user WHERE name = 'user_500';
-- key = idx_name_age, Extra: Using index

-- ============================================================
-- Demo 4: 索引下推 (ICP)
-- ============================================================
SELECT '--- Demo 4: 索引下推 (ICP) ---' AS step;

-- 没有 ICP 的话，MySQL 5.6 之前会先回表再过滤
-- 有 ICP：在索引层就过滤 WHERE 条件

-- 关闭 ICP 试试
SET optimizer_switch = 'index_condition_pushdown=off';
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' AND age > 20;
-- 注意 Extra: Using where (没有 ICP)

-- 开启 ICP
SET optimizer_switch = 'index_condition_pushdown=on';
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' AND age > 20;
-- 注意 Extra: Using index condition (ICP 生效)

-- ============================================================
-- Demo 5: 索引失效场景（10 种）
-- ============================================================
SELECT '--- Demo 5: 10 种索引失效场景 ---' AS step;

-- ❌ 1. 不满足最左前缀
EXPLAIN SELECT * FROM t_user WHERE age = 30 \G
-- key = NULL

-- ❌ 2. 范围查询导致后续列失效
EXPLAIN SELECT * FROM t_user 
WHERE name = 'user_500' AND age > 20 AND city = 'Beijing' \G
-- key_len 只到 name+age

-- ❌ 3. 在索引列上做函数
EXPLAIN SELECT * FROM t_user WHERE UPPER(name) = 'USER_500' \G
-- key = NULL

-- ❌ 4. 在索引列上做运算
EXPLAIN SELECT * FROM t_user WHERE age + 1 = 30 \G
-- key = NULL

-- ❌ 5. 隐式类型转换（name 是 varchar，传 int）
EXPLAIN SELECT * FROM t_user WHERE name = 123 \G
-- key = NULL

-- ✅ 正确：传字符串
EXPLAIN SELECT * FROM t_user WHERE name = '123' \G
-- key = idx_name

-- ❌ 6. LIKE 以 % 开头
EXPLAIN SELECT * FROM t_user WHERE name LIKE '%500' \G
-- key = NULL

-- ✅ 正确：% 在后面
EXPLAIN SELECT * FROM t_user WHERE name LIKE 'user_5%' \G
-- key = idx_name

-- ❌ 7. OR 前后有非索引列
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' OR email = 'a@b.com' \G
-- key = NULL（email 无索引，全表扫）

-- ❌ 8. NOT、!= 失效（取决于基数）
EXPLAIN SELECT * FROM t_user WHERE name != 'user_500' \G
-- 通常 key = NULL

-- ❌ 9. IS NULL（看基数）
EXPLAIN SELECT * FROM t_user WHERE name IS NULL \G
-- 大多数 NULL 少时 key = NULL

-- ❌ 10. IN 太多值
EXPLAIN SELECT * FROM t_user WHERE id IN (1,2,3,4,5) \G
-- 少量走索引

-- ============================================================
-- Demo 6: 前缀索引
-- ============================================================
SELECT '--- Demo 6: 前缀索引（长字符串）---' AS step;

CREATE INDEX idx_email_prefix ON t_user (email(10));

-- 测试：email 完整长度 25-30，但索引只存前 10 字符
-- 验证前缀索引大小
SELECT 
    INDEX_NAME,
    STAT_VALUE * @@innodb_page_size / 1024 / 1024 AS index_size_mb
FROM mysql.innodb_index_stats 
WHERE TABLE_NAME = 't_user' AND STAT_NAME = 'size';

-- 查找最佳前缀长度
SELECT 
    COUNT(DISTINCT LEFT(email, 5)) / COUNT(*) AS distinct_5,
    COUNT(DISTINCT LEFT(email, 10)) / COUNT(*) AS distinct_10,
    COUNT(DISTINCT LEFT(email, 15)) / COUNT(*) AS distinct_15
FROM t_user;

-- ============================================================
-- Demo 7: 强制走索引（USE INDEX）
-- ============================================================
SELECT '--- Demo 7: 索引提示 ---' AS step;

-- 正常情况（可能选错索引）
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' OR age = 30 \G

-- USE INDEX：建议用哪个索引
EXPLAIN SELECT * FROM t_user USE INDEX (idx_name_age) 
WHERE name = 'user_500' OR age = 30 \G

-- FORCE INDEX：强制用某个索引
EXPLAIN SELECT * FROM t_user FORCE INDEX (idx_name) 
WHERE name = 'user_500' \G

-- IGNORE INDEX：忽略某个索引
EXPLAIN SELECT * FROM t_user IGNORE INDEX (idx_name_age) 
WHERE name = 'user_500' \G

-- ============================================================
-- Demo 8: 自增主键 vs UUID（页分裂）
-- ============================================================
SELECT '--- Demo 8: 主键选择 ---' AS step;

-- 推荐：自增主键
CREATE TABLE t_inc (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    data VARCHAR(100)
) ENGINE=InnoDB;

-- 不推荐：UUID 主键（36字节，且随机写入）
CREATE TABLE t_uuid (
    id CHAR(36) PRIMARY KEY,
    data VARCHAR(100)
) ENGINE=InnoDB;

-- 索引大小对比（自增主键 8 字节 vs UUID 36 字节）
-- 同样的 10 万行，二级索引里：
-- - t_inc 二级索引叶子存 8 字节主键
-- - t_uuid 二级索引叶子存 36 字节主键
-- 4.5 倍差距！

-- ============================================================
-- Demo 9: COUNT(*) 性能对比
-- ============================================================
SELECT '--- Demo 9: COUNT(*) 用哪个索引？---' AS step;

-- 没有 WHERE：全表扫
EXPLAIN SELECT COUNT(*) FROM t_user \G

-- 有 WHERE + 索引：用二级索引（小）
EXPLAIN SELECT COUNT(*) FROM t_user WHERE name = 'user_500' \G

-- 复杂查询：可能选错索引
EXPLAIN SELECT COUNT(*) FROM t_user WHERE city = 'Beijing' \G
-- city 无索引，全表扫

-- ============================================================
-- Demo 10: 排序（ORDER BY）走索引
-- ============================================================
SELECT '--- Demo 10: ORDER BY 走索引 ---' AS step;

-- ✅ 走索引排序（避免 filesort）
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' ORDER BY age \G
-- Extra: Using index condition; Using where

-- ❌ 不同列排序（filesort）
EXPLAIN SELECT * FROM t_user WHERE name = 'user_500' ORDER BY city \G
-- Extra: Using filesort

-- ============================================================
-- 清理
-- ============================================================
DROP PROCEDURE IF EXISTS insert_users;
DROP DATABASE IF EXISTS demo_index;
SELECT '--- Demo 结束 ---' AS end;

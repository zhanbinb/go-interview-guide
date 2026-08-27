-- ============================================================
-- MySQL vs SQL Server 对比 demo
-- ============================================================
-- 用法：
--   1. 启动 MySQL 8.0+（docker / 本地）
--   2. mysql -uroot -p < demo.sql
--   3. 看每个 section 的输出，对比 SQL Server 行为
-- ============================================================

SELECT '=== MySQL 8.0 vs SQL Server 对比 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- ============================================================
-- Demo 1: 字符集（最容易踩的坑）
-- ============================================================
SELECT '--- Demo 1: utf8mb4 vs utf8 ---' AS demo;

-- 查看当前字符集
SHOW VARIABLES LIKE 'character_set_server';
SHOW VARIABLES LIKE 'collation_server';

-- 推荐建库语句
CREATE DATABASE IF NOT EXISTS demo_vs CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_vs;

-- 演示 utf8mb4 能存 emoji
CREATE TABLE t_emoji (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO t_emoji (name) VALUES ('Hello'), ('😀'), ('你好'), ('🌍地球');
SELECT * FROM t_emoji;

-- ============================================================
-- Demo 2: 自增主键 + LAST_INSERT_ID
-- ============================================================
SELECT '--- Demo 2: AUTO_INCREMENT ---' AS demo;

CREATE TABLE t_user (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,  -- 重点: UNSIGNED (42亿)
    name VARCHAR(50)
) ENGINE=InnoDB;

INSERT INTO t_user (name) VALUES ('Alice'), ('Bob'), ('Charlie');
SELECT * FROM t_user;

-- 重点: LAST_INSERT_ID()（对应 SQL Server 的 SCOPE_IDENTITY()）
INSERT INTO t_user (name) VALUES ('David');
SELECT LAST_INSERT_ID() AS new_id;

-- 自增起始值设置（SQL Server: IDENTITY(100,1) → MySQL: AUTO_INCREMENT=100）
ALTER TABLE t_user AUTO_INCREMENT = 100;
INSERT INTO t_user (name) VALUES ('Eve');
SELECT * FROM t_user WHERE name = 'Eve';

-- ============================================================
-- Demo 3: LIMIT vs TOP/OFFSET
-- ============================================================
SELECT '--- Demo 3: LIMIT/OFFSET vs TOP ---' AS demo;

CREATE TABLE t_product (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50),
    price DECIMAL(10,2)
) ENGINE=InnoDB;

INSERT INTO t_product (name, price) VALUES
('Apple', 1.00), ('Banana', 2.00), ('Cherry', 3.00),
('Date', 4.00), ('Elderberry', 5.00), ('Fig', 6.00),
('Grape', 7.00), ('Honeydew', 8.00);

-- SQL Server: SELECT TOP 3 * FROM t_product
SELECT * FROM t_product ORDER BY id LIMIT 3;  -- MySQL 等价

-- SQL Server 2012+: OFFSET 2 ROWS FETCH NEXT 3 ROWS ONLY
SELECT * FROM t_product ORDER BY id LIMIT 3 OFFSET 2;  -- MySQL 等价
-- 或: SELECT * FROM t_product ORDER BY id LIMIT 2, 3;

-- ============================================================
-- Demo 4: IFNULL vs ISNULL
-- ============================================================
SELECT '--- Demo 4: IFNULL / ISNULL / COALESCE ---' AS demo;

SELECT
    IFNULL(NULL, 'default') AS ifnull_result,         -- MySQL 专属
    COALESCE(NULL, NULL, 'fallback') AS coalesce,    -- 标准 SQL（推荐）
    NULLIF('a', 'a') AS nullif,                     -- 两边相等返回 NULL
    NULLIF('a', 'b') AS not_null;                    -- 两边不等返回第一个

-- ============================================================
-- Demo 5: CONCAT vs + 字符串拼接
-- ============================================================
SELECT '--- Demo 5: 字符串拼接差异 ---' AS demo;

-- MySQL: 任何参数为 NULL → 结果 NULL
SELECT CONCAT('a', NULL, 'b') AS mysql_concat;   -- 'NULL'

-- MySQL: 用 CONCAT_WS 跳过 NULL
SELECT CONCAT_WS(',', 'a', NULL, 'b', NULL, 'c') AS concat_ws;  -- 'a,b,b,c'

-- 多行聚合（SQL Server: STRING_AGG；MySQL: GROUP_CONCAT）
SELECT GROUP_CONCAT(name ORDER BY id SEPARATOR '|') AS all_names
FROM t_product
WHERE id <= 5;

-- ============================================================
-- Demo 6: NOW / CURRENT_TIMESTAMP
-- ============================================================
SELECT '--- Demo 6: 时间函数 ---' AS demo;

SELECT
    NOW() AS now_func,                          -- 当前时间
    CURRENT_TIMESTAMP AS current_ts,           -- 同 NOW()
    CURDATE() AS curdate,                      -- 只日期
    CURTIME() AS curtime,                      -- 只时间
    DATE_ADD(NOW(), INTERVAL 1 DAY) AS tomorrow,  -- 日期加 1 天
    DATE_SUB(NOW(), INTERVAL 1 HOUR) AS hour_ago, -- 日期减 1 小时
    DATEDIFF('2026-12-31', NOW()) AS days_left,   -- 日期差（参数顺序：未来-现在）
    DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') AS formatted; -- 格式化

-- ============================================================
-- Demo 7: 临时表
-- ============================================================
SELECT '--- Demo 7: 临时表 ---' AS demo;

-- SQL Server: SELECT * INTO #temp_orders FROM orders
-- MySQL:
CREATE TEMPORARY TABLE temp_top_products AS
SELECT id, name, price
FROM t_product
WHERE price > 3.00
ORDER BY price DESC
LIMIT 3;

SELECT * FROM temp_top_products;

-- 自动清理（会话结束）
-- 查看临时表
SHOW TABLES;  -- 注意：临时表在 SHOW TABLES 里看不到

-- ============================================================
-- Demo 8: 数据库隔离级别验证
-- ============================================================
SELECT '--- Demo 8: 隔离级别默认是 RR（不是 RC）---' AS demo;

SELECT @@global.transaction_isolation AS global_isolation;
SELECT @@session.transaction_isolation AS session_isolation;

-- MySQL 默认是 REPEATABLE READ（不是 SQL Server 的 RC）
-- 重要含义：同一事务内多次 SELECT 看到的数据一致（即使别的事务已修改）

-- 验证：可以会话级改隔离级别
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
SELECT @@session.transaction_isolation AS now_isolation;

-- 恢复默认
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- ============================================================
-- Demo 9: JSON 类型
-- ============================================================
SELECT '--- Demo 9: JSON 类型（MySQL 5.7+ 原生支持）---' AS demo;

CREATE TABLE t_user_json (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50),
    profile JSON  -- MySQL 原生 JSON 类型（不是 NVARCHAR）
) ENGINE=InnoDB;

INSERT INTO t_user_json (name, profile) VALUES
('Alice', JSON_OBJECT('age', 25, 'city', 'Beijing', 'tags', JSON_ARRAY('admin', 'vip'))),
('Bob', JSON_OBJECT('age', 30, 'city', 'Shanghai'));

-- 提取 JSON 字段
SELECT
    name,
    profile->>'$.city' AS city,           -- ->> 返回文本
    profile->'$.age' AS age_int,         -- ->  返回 JSON
    JSON_EXTRACT(profile, '$.tags') AS tags
FROM t_user_json;

-- JSON 函数
SELECT
    name,
    JSON_LENGTH(profile) AS field_count,
    JSON_KEYS(profile) AS keys,
    JSON_CONTAINS(profile, '"Beijing"', '$.city') AS is_beijing
FROM t_user_json;

-- ============================================================
-- Demo 10: 深分页问题（用游标解决）
-- ============================================================
SELECT '--- Demo 10: 深分页 ---' AS demo;

-- ❌ 错误：OFFSET 越大越慢（要扫 OFFSET+N 行）
EXPLAIN SELECT * FROM t_product ORDER BY id LIMIT 10 OFFSET 1000000;

-- ✅ 正确：用游标分页（前提：id 连续）
EXPLAIN SELECT * FROM t_product WHERE id > 1000000 ORDER BY id LIMIT 10;
-- 这才是 MySQL 处理深分页的正确姿势

-- ============================================================
-- 清理
-- ============================================================
DROP DATABASE IF EXISTS demo_vs;
SELECT '--- Demo 结束 ---' AS end;

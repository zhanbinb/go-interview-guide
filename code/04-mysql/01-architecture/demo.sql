-- ============================================================
-- MySQL 一条 SQL 怎么执行 demo
-- 涵盖：连接器、分析器、优化器、执行器、Buffer Pool
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- ============================================================

SELECT '=== MySQL 一条 SQL 怎么执行 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- 准备测试库
DROP DATABASE IF EXISTS demo_arch;
CREATE DATABASE demo_arch DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_arch;

CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    city VARCHAR(50) NOT NULL,
    INDEX idx_age (age),
    INDEX idx_name_age (name, age)
) ENGINE=InnoDB;

INSERT INTO t_user (name, age, city) VALUES
('Alice', 25, 'Beijing'),
('Bob', 30, 'Shanghai'),
('Charlie', 35, 'Guangzhou'),
('David', 28, 'Shenzhen'),
('Eve', 32, 'Hangzhou');

-- ============================================================
-- Demo 1: 查看执行流程相关参数
-- ============================================================
SELECT '--- Demo 1: 查看关键参数 ---' AS step;

-- 连接相关
SHOW VARIABLES LIKE 'max_connections';         -- 最大连接数（默认 151）
SHOW VARIABLES LIKE 'wait_timeout';            -- 空闲连接超时（默认 28800s = 8h）

-- 优化器相关
SHOW VARIABLES LIKE 'optimizer_switch';        -- 优化器开关
SHOW VARIABLES LIKE 'optimizer_trace';         -- 优化器追踪（默认 disabled）

-- Buffer Pool 相关（InnoDB）
SHOW VARIABLES LIKE 'innodb_buffer_pool_size'; -- 默认 128M

-- 查询缓存（8.0 已删）
SHOW VARIABLES LIKE 'query_cache%';
-- 应该返回 Empty set

-- ============================================================
-- Demo 2: 用 EXPLAIN 看执行计划
-- ============================================================
SELECT '--- Demo 2: EXPLAIN 看执行计划（理解优化器输出）---' AS step;

-- 简单查询：走主键
EXPLAIN SELECT * FROM t_user WHERE id = 3\G

-- 走索引（idx_age）
EXPLAIN SELECT * FROM t_user WHERE age > 30\G

-- 走复合索引（idx_name_age）
EXPLAIN SELECT * FROM t_user WHERE name = 'Alice' AND age = 25\G

-- ============================================================
-- Demo 3: 用 optimizer_trace 看更详细的信息（需要开启）
-- ============================================================
SELECT '--- Demo 3: 优化器追踪（OPTIMIZER_TRACE）---' AS step;

SET optimizer_trace = 'enabled=on';
SELECT * FROM information_schema.optimizer_trace\G
SET optimizer_trace = 'enabled=off';

SELECT '  上面会显示优化器做的成本估算、JOIN 顺序等' AS note;

-- ============================================================
-- Demo 4: 模拟长连接效果
-- ============================================================
SELECT '--- Demo 4: 长连接效果 ---' AS step;

-- 查看当前连接
SELECT CONNECTION_ID() AS my_conn_id;
SHOW PROCESSLIST;

-- ============================================================
-- Demo 5: Buffer Pool 状态
-- ============================================================
SELECT '--- Demo 5: Buffer Pool 状态 ---' AS step;

SHOW ENGINE INNODB STATUS\G

-- 找 BUFFER POOL AND MEMORY 部分
-- 看 Buffer pool hit rate（应该 > 99%）

-- ============================================================
-- Demo 6: 模拟一次完整查询（观察各阶段）
-- ============================================================
SELECT '--- Demo 6: 模拟完整查询 ---' AS step;

-- 开启 general_log 看 SQL 执行记录（仅学习用，生产慎开）
SET GLOBAL general_log = 1;
SET GLOBAL log_output = 'TABLE';

-- 执行一条查询
SELECT * FROM t_user WHERE name = 'Alice' AND age > 20;

-- 查看刚执行的 SQL
SELECT event_time, argument
FROM mysql.general_log
WHERE event_time > NOW() - INTERVAL 1 MINUTE
  AND argument LIKE '%SELECT%t_user%'
ORDER BY event_time DESC LIMIT 5;

-- 关闭
SET GLOBAL general_log = 0;

-- ============================================================
-- Demo 7: 用 PreparedStatement（避免重复解析）
-- ============================================================
SELECT '--- Demo 7: PreparedStatement 演示 ---' AS step;

-- MySQL 命令行模拟 prepare
PREPARE stmt FROM 'SELECT * FROM t_user WHERE id = ?';
SET @id = 1;
EXECUTE stmt USING @id;

SET @id = 2;
EXECUTE stmt USING @id;

SET @id = 3;
EXECUTE stmt USING @id;

-- 释放
DEALLOCATE PREPARE stmt;

SELECT '  PreparedStatement 只解析一次，多次执行只传参数' AS note;

-- ============================================================
-- Demo 8: SHOW PROFILE 和 SHOW WARNINGS
-- ============================================================
SELECT '--- Demo 8: 查看 SQL 执行详细信息 ---' AS step;

-- 8.0 推荐用 performance_schema 替代 SHOW PROFILE
SELECT * FROM performance_schema.events_statements_history
WHERE thread_id = PS_CURRENT_THREAD_ID()
ORDER BY event_id DESC LIMIT 3\G

-- ============================================================
-- Demo 9: 一个 UPDATE 涉及的日志（演示用）
-- ============================================================
SELECT '--- Demo 9: UPDATE 涉及的 3 个日志（06-log 复习）---' AS step;

SELECT '  UPDATE 流程会写：' AS step;
SELECT '    1. undo log（用于回滚和 MVCC）' AS log;
SELECT '    2. redo log（prepare）（持久性 D）' AS log;
SELECT '    3. binlog（主从复制）' AS log;
SELECT '    4. redo log（commit）' AS log;
SELECT '  这是两阶段提交！' AS note;

-- ============================================================
-- Demo 10: SQL Server 对比（你专享）
-- ============================================================
SELECT '--- Demo 10: SQL Server vs MySQL 架构对比 ---' AS step;

SELECT '  MySQL 优势：存储引擎可插拔（InnoDB/MyISAM/Memory）' AS compare;
SELECT '  SQL Server 只有一种引擎（类 InnoDB）' AS compare;
SELECT '  查询缓存：MySQL 8.0 已删（弊大于利），SQL Server 早期有，现在没了' AS compare;
SELECT '  协议层：MySQL 有自己的协议，SQL Server 用 TDS' AS compare;

-- ============================================================
-- 清理
-- ============================================================
SET GLOBAL general_log = 0;
DROP DATABASE IF EXISTS demo_arch;
SELECT '--- Demo 结束 ---' AS end;

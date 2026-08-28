-- ============================================================
-- MySQL 日志 demo
-- 涵盖：redo log / binlog / undo log / 两阶段提交
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- ============================================================

SELECT '=== MySQL 日志 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- 准备测试库
DROP DATABASE IF EXISTS demo_log;
CREATE DATABASE demo_log DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_log;

CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL
) ENGINE=InnoDB;

INSERT INTO t_user (name, age) VALUES ('Alice', 25), ('Bob', 30);

-- ============================================================
-- Demo 1: 查看 binlog 配置
-- ============================================================
SELECT '--- Demo 1: binlog 配置 ---' AS step;

SHOW VARIABLES LIKE 'log_bin';              -- 是否开启 binlog
SHOW VARIABLES LIKE 'binlog_format';       -- ROW / STATEMENT / MIXED
SHOW VARIABLES LIKE 'sync_binlog';          -- 刷盘策略
SHOW VARIABLES LIKE 'binlog_row_image';     -- 8.0 新增：full / minimal / noblob
SHOW VARIABLES LIKE 'expire_logs_days';     -- binlog 保留天数

-- ============================================================
-- Demo 2: 查看 redo log 配置
-- ============================================================
SELECT '--- Demo 2: redo log 配置 ---' AS step;

SHOW VARIABLES LIKE 'innodb_log_file_size';     -- 单个文件大小
SHOW VARIABLES LIKE 'innodb_log_files_in_group'; -- 文件数量
SHOW VARIABLES LIKE 'innodb_log_buffer_size';   -- log buffer
SHOW VARIABLES LIKE 'innodb_flush_log_at_trx_commit'; -- 刷盘策略
-- 0: 每秒刷，1: 每次提交刷（默认，最安全），2: 每次提交写 OS 缓存

-- 查看 redo log 文件
SHOW VARIABLES LIKE 'innodb_log_group_home_dir';

-- ============================================================
-- Demo 3: undo log 配置
-- ============================================================
SELECT '--- Demo 3: undo log 配置 ---' AS step;

SHOW VARIABLES LIKE 'innodb_undo_tablespaces';   -- undo 表空间数量
SHOW VARIABLES LIKE 'innodb_undo_log_truncate';   -- 8.0 开启自动回收
SHOW VARIABLES LIKE 'innodb_max_undo_log_size';   -- 单个 undo log 大小

-- ============================================================
-- Demo 4: 验证 undo log 配合回滚
-- ============================================================
SELECT '--- Demo 4: undo log 回滚演示 ---' AS step;

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('Charlie', 35);
UPDATE t_user SET age = 36 WHERE name = 'Charlie';
-- 准备回滚（undo log 已经记录原始值 age=35）
ROLLBACK;

SELECT * FROM t_user WHERE name = 'Charlie';
-- 期望：Charlie 不存在（被回滚）
-- 0 rows（如果上一句是 ROLLBACK）


-- ============================================================
-- Demo 5: 模拟崩溃恢复（用 FLUSH + 杀进程演示不可能，用 REDO log 行为说明）
-- ============================================================
SELECT '--- Demo 5: redo log 持久性演示 ---' AS step;

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('David', 40);
COMMIT;
-- 此时 redo log 已刷盘（即使 MySQL 崩溃，重启后 David 还在）

-- 查看 redo log 是否活跃
SHOW ENGINE INNODB STATUS\G  -- 看 LOG 部分

-- ============================================================
-- Demo 6: 查看 binlog 文件列表
-- ============================================================
SELECT '--- Demo 6: binlog 文件 ---' AS step;

-- MySQL 8.0 默认开启 binlog（log_bin = ON）
SHOW BINARY LOGS;  -- 列出所有 binlog 文件
SHOW MASTER STATUS;  -- 当前 binlog 位置

-- 查看某个 binlog 内容（需要 mysqlbinlog 工具，这里用 SHOW BINLOG EVENTS）
SHOW BINLOG EVENTS IN 'binlog.000001' LIMIT 10;

-- ============================================================
-- Demo 7: 模拟两阶段提交的过程（用显式 prepare）
-- ============================================================
SELECT '--- Demo 7: 两阶段提交原理 ---' AS step;

-- 用户看到的：START TRANSACTION ... COMMIT
-- 实际 InnoDB 做的：
-- 1. 写 undo log
-- 2. 修改 Buffer Pool
-- 3. 写 redo log（标记 PREPARE）
-- 4. 写 binlog
-- 5. 写 redo log（标记 COMMIT）

-- 验证：开启 general_log 看 SQL 执行顺序
SET GLOBAL general_log = 1;
SET GLOBAL log_output = 'TABLE';

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('Eve', 45);
UPDATE t_user SET age = 46 WHERE name = 'Eve';
COMMIT;

-- 查看刚执行的 SQL
SELECT event_time, argument 
FROM mysql.general_log 
WHERE event_time > NOW() - INTERVAL 1 MINUTE
ORDER BY event_time;

-- 关掉 general_log
SET GLOBAL general_log = 0;

-- ============================================================
-- Demo 8: binlog ROW 格式查看（实际行数据）
-- ============================================================
SELECT '--- Demo 8: binlog ROW 格式 ---' AS step;

-- ROW 格式记录的是行数据，不是 SQL
-- 适合用 mysqlbinlog 工具查看
-- 这里用 SHOW BINLOG EVENTS 简要展示

SHOW BINLOG EVENTS LIMIT 5;

-- ============================================================
-- Demo 9: WAL 思想验证
-- ============================================================
SELECT '--- Demo 9: WAL 验证（先写日志再写数据）---' AS step;

-- 设置 InnoDB 刷盘策略
SET GLOBAL innodb_flush_log_at_trx_commit = 1;  -- 最安全

-- 看 Buffer Pool 状态
SHOW ENGINE INNODB STATUS\G  -- 看 BUFFER POOL AND MEMORY 部分

-- ============================================================
-- Demo 10: 慢查询日志
-- ============================================================
SELECT '--- Demo 10: 慢查询日志（额外知识点）---' AS step;

SHOW VARIABLES LIKE 'slow_query_log';
SHOW VARIABLES LIKE 'long_query_time';   -- 阈值（默认 10s）
SHOW VARIABLES LIKE 'log_output';         -- FILE / TABLE

-- 慢查询日志开启
SET GLOBAL slow_query_log = 1;
SET GLOBAL long_query_time = 0;  -- 测试用，所有查询都记

-- 触发慢查询
SELECT SLEEP(1);

-- 查看慢查询日志
SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 3;

-- 关闭
SET GLOBAL slow_query_log = 0;
SET GLOBAL long_query_time = 10;

-- ============================================================
-- 清理
-- ============================================================
SET GLOBAL general_log = 0;
DROP DATABASE IF EXISTS demo_log;
SELECT '--- Demo 结束 ---' AS end;

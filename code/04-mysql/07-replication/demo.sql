-- ============================================================
-- MySQL 主从复制 demo
-- 涵盖：binlog 复制、3 种方式、读写分离、延迟诊断
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- 注意：完整主从复制需要 2 个 MySQL 实例，这里只演示单实例的复制相关配置
-- ============================================================

SELECT '=== MySQL 主从复制 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- ============================================================
-- Demo 1: 查看主从相关配置
-- ============================================================
SELECT '--- Demo 1: binlog 配置 ---' AS step;

SHOW VARIABLES LIKE 'log_bin';                    -- binlog 是否开启
SHOW VARIABLES LIKE 'binlog_format';              -- ROW / STATEMENT / MIXED
SHOW VARIABLES LIKE 'server_id';                 -- 主从唯一标识
SHOW VARIABLES LIKE 'binlog_row_image';           -- 8.0 新增

-- 半同步复制配置
SHOW VARIABLES LIKE 'rpl_semi_sync_master_enabled';
SHOW VARIABLES LIKE 'rpl_semi_sync_slave_enabled';

-- GTID 配置
SHOW VARIABLES LIKE 'gtid_mode';
SHOW VARIABLES LIKE 'enforce_gtid_consistency';

-- ============================================================
-- Demo 2: 查看 Master 状态
-- ============================================================
SELECT '--- Demo 2: 查看 Master 状态 ---' AS step;

-- 准备一个表
DROP DATABASE IF EXISTS demo_repl;
CREATE DATABASE demo_repl DEFAULT CHARACTER SET utf8mb4;
USE demo_repl;
CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL
) ENGINE=InnoDB;

INSERT INTO t_user (name) VALUES ('Alice'), ('Bob');

SHOW MASTER STATUS;
-- 输出: File (binlog 文件名) + Position (位置) + Binlog_Do_DB 等

-- ============================================================
-- Demo 3: 查看 binlog 文件列表
-- ============================================================
SELECT '--- Demo 3: binlog 文件 ---' AS step;

SHOW BINARY LOGS;
-- 列出所有 binlog 文件

-- ============================================================
-- Demo 4: 看 binlog 事件（粗糙版）
-- ============================================================
SELECT '--- Demo 4: binlog 事件 ---' AS step;

INSERT INTO t_user (name) VALUES ('Charlie');
UPDATE t_user SET name = 'Charles' WHERE id = 3;

SHOW BINLOG EVENTS LIMIT 10;
-- 看到 INSERT 和 UPDATE 的事件

-- ============================================================
-- Demo 5: 半同步复制（配置和验证）
-- ============================================================
SELECT '--- Demo 5: 半同步复制配置 ---' AS step;

-- 加载插件（需要动态加载 .so）
-- INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';
-- INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';

-- 查看插件是否已加载
SHOW PLUGINS;

-- 配置
-- SET GLOBAL rpl_semi_sync_master_enabled = 1;
-- SET GLOBAL rpl_semi_sync_slave_enabled = 1;

SELECT '  半同步需要两个 MySQL 实例才能完整验证' AS note;

-- ============================================================
-- Demo 6: GTID 复制（8.0 默认）
-- ============================================================
SELECT '--- Demo 6: GTID 配置 ---' AS step;

SHOW VARIABLES LIKE 'gtid_mode';
-- MySQL 8.0 默认 ON

-- 看自己的 GTID
SHOW MASTER STATUS;
-- 输出会有 Executed_Gtid_Set: uuid:1-100

-- ============================================================
-- Demo 7: 从库配置 SHOW SLAVE STATUS（需要真从库）
-- ============================================================
SELECT '--- Demo 7: 从库配置示例 ---' AS step;

SELECT '  -- Master 配置（mysqld.cnf）' AS config;
-- [mysqld]
-- server-id = 1
-- log_bin = /var/log/mysql/mysql-bin.log
-- binlog_format = ROW
-- gtid_mode = ON
-- enforce_gtid_consistency = ON

SELECT '  -- Slave 配置' AS config;
-- [mysqld]
-- server-id = 2
-- gtid_mode = ON
-- read_only = ON
-- super_read_only = ON  -- 8.0 推荐，防止误写

SELECT '  -- 启动复制（Slave 端）' AS config;
-- CHANGE MASTER TO
--   MASTER_HOST='192.168.1.1',
--   MASTER_USER='repl_user',
--   MASTER_PASSWORD='password',
--   MASTER_AUTO_POSITION=1;  -- GTID 自动找位点
-- START SLAVE;

SELECT '  -- 验证' AS config;
-- SHOW SLAVE STATUS\G
-- 看 Slave_IO_Running: Yes
-- Slave_SQL_Running: Yes
-- Seconds_Behind_Master: 0

-- ============================================================
-- Demo 8: 模拟主从延迟（用 sleep）
-- ============================================================
SELECT '--- Demo 8: 模拟延迟 ---' AS step;

-- 大事务会产生延迟
INSERT INTO t_user (name) 
SELECT CONCAT('user_', seq) FROM 
    (SELECT @row := @row + 1 AS seq FROM 
        (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 
         UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) a,
        (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 
         UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) b,
        (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 
         UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) c,
        (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 
         UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) d,
        (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 
         UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) e
    ) seqs, (SELECT @row := 0) r;

-- 这样就有 10万行，会让从库延迟一会儿
SELECT COUNT(*) AS user_count FROM t_user;

-- ============================================================
-- Demo 9: 多线程复制（8.0 重要改进）
-- ============================================================
SELECT '--- Demo 9: 多线程复制配置 ---' AS step;

SELECT '  -- Slave 端开启' AS config;
-- SET GLOBAL slave_parallel_type = LOGICAL_CLOCK;  -- 8.0 默认值
-- SET GLOBAL slave_parallel_workers = 8;          -- 8 个 worker
-- 前提：ROW 格式 binlog

-- 查看当前配置
SHOW VARIABLES LIKE 'slave_parallel%';

-- ============================================================
-- Demo 10: 延迟从库（防误操作）
-- ============================================================
SELECT '--- Demo 10: 延迟从库（防删库）---' AS step;

SELECT '  延迟从库配置（Slave 端）' AS config;
-- CHANGE MASTER TO MASTER_DELAY = 3600;  -- 延迟 1 小时
-- START SLAVE;

SELECT '  场景：master 执行 DELETE WHERE 1=1 → 1 小时后才传播到从库' AS note;
SELECT '       → 有时间从延迟从库恢复数据' AS note;

-- ============================================================
-- 清理
-- ============================================================
DROP DATABASE IF EXISTS demo_repl;
SELECT '--- Demo 结束 ---' AS end;

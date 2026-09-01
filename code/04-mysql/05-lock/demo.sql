-- ============================================================
-- MySQL 锁机制 demo
-- 涵盖：行锁/表锁/意向锁/Next-Key Lock/死锁
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- 注意：很多锁的演示需要 2 个连接才能完整看到效果
-- ============================================================

SELECT '=== MySQL 锁机制 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- 准备测试库
DROP DATABASE IF EXISTS demo_lock;
CREATE DATABASE demo_lock DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_lock;

CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    version INT NOT NULL DEFAULT 0,  -- 用于乐观锁
    INDEX idx_name (name)
) ENGINE=InnoDB;

INSERT INTO t_user (name, age) VALUES ('Alice', 25), ('Bob', 30), ('Charlie', 35);

-- ============================================================
-- Demo 1: 共享锁 vs 排他锁
-- ============================================================
SELECT '--- Demo 1: 共享锁与排他锁 ---' AS step;

-- SESSION A: 共享锁
START TRANSACTION;
SELECT * FROM t_user WHERE id = 1 LOCK IN SHARE MODE;
-- 加了 id=1 的 S 锁
-- 其他事务可以也加 S 锁，但 X 锁会被阻塞

SELECT * FROM t_user WHERE id = 1;  -- 快照读，不冲突
ROLLBACK;

-- SESSION B: 排他锁（FOR UPDATE）
START TRANSACTION;
SELECT * FROM t_user WHERE id = 1 FOR UPDATE;
-- 加了 id=1 的 X 锁
ROLLBACK;

-- ============================================================
-- Demo 2: 意向锁验证
-- ============================================================
SELECT '--- Demo 2: 意向锁（IX）验证 ---' AS step;

-- 查看意向锁
SELECT * FROM performance_schema.data_locks WHERE OBJECT_NAME = 't_user'\G

-- 实际演示：事务里加 FOR UPDATE（产生行 X 锁 + 表 IX 锁）
-- 这两个 session 需要手动操作

-- ============================================================
-- Demo 3: Next-Key Lock 演示
-- ============================================================
SELECT '--- Demo 3: Next-Key Lock ---' AS step;

TRUNCATE t_user;
INSERT INTO t_user (name, age) VALUES ('A', 10), ('B', 20), ('C', 30);
SELECT * FROM t_user;

-- RR 级别下，SELECT FOR UPDATE 会加 Next-Key Lock
-- Session A 执行：
START TRANSACTION;
SELECT * FROM t_user WHERE id = 2 FOR UPDATE;
-- 加锁范围：(1, 2], (2, 3]
-- Session B 在 (1, 3) 范围内 INSERT 会被阻塞：
-- INSERT INTO t_user (name, age) VALUES ('X', 15);  -- 阻塞
ROLLBACK;

-- ============================================================
-- Demo 4: 行锁基于索引的陷阱
-- ============================================================
SELECT '--- Demo 4: 行锁基于索引 ---' AS step;

-- ✅ 用索引：只锁单行
EXPLAIN SELECT * FROM t_user WHERE id = 1 FOR UPDATE;
-- key = PRIMARY, rows = 1

-- ❌ 全表扫：可能锁全表
-- 假设 name 字段没有索引（实际 demo 里建了，会走 idx_name）
EXPLAIN SELECT * FROM t_user WHERE name = 'Alice' FOR UPDATE;
-- key = idx_name, rows = 1（有索引就好）

-- 加个测试：无索引的列
CREATE TABLE t_no_idx (
    id INT PRIMARY KEY AUTO_INCREMENT,
    no_index_col VARCHAR(50)
) ENGINE=InnoDB;
INSERT INTO t_no_idx (no_index_col) VALUES ('a'), ('b'), ('c');

EXPLAIN SELECT * FROM t_no_idx WHERE no_index_col = 'a' FOR UPDATE;
-- key = NULL, rows = 3 → 全表扫！会锁所有行
-- 这就是大事务的根源

-- ============================================================
-- Demo 5: 乐观锁实现
-- ============================================================
SELECT '--- Demo 5: 乐观锁 ---' AS step;

-- 1. 读取（带 version）
SELECT id, name, age, version FROM t_user WHERE id = 1;
-- 假设读到 version = 0

-- 2. 更新（带 version 条件）
START TRANSACTION;
UPDATE t_user SET age = 26, version = version + 1
WHERE id = 1 AND version = 0;
-- 影响行数 = 1 → 成功
-- 如果影响行数 = 0 → 版本冲突，需要重试
COMMIT;

-- 3. 再次用旧 version 更新（会失败）
UPDATE t_user SET age = 27
WHERE id = 1 AND version = 0;
-- 影响行数 = 0（因为 version 已经是 1）
-- 这就是乐观锁的 CAS 思想

SELECT id, name, age, version FROM t_user WHERE id = 1;

-- ============================================================
-- Demo 6: 死锁演示（需要 2 个 session）
-- ============================================================
SELECT '--- Demo 6: 死锁演示（理论）---' AS step;

-- === Session A ===
-- START TRANSACTION;
-- UPDATE t_user SET name = 'X' WHERE id = 1;  -- 锁 id=1
-- -- 不提交
-- UPDATE t_user SET name = 'Y' WHERE id = 2;  -- 等 id=2 的锁
-- 
-- === Session B（同时）===
-- START TRANSACTION;
-- UPDATE t_user SET name = 'P' WHERE id = 2;  -- 锁 id=2
-- -- 不提交
-- UPDATE t_user SET name = 'Q' WHERE id = 1;  -- 等 id=1 的锁
-- 
-- === 死锁！MySQL 自动检测 ===
-- ERROR 1213 (40001): Deadlock found when trying to get lock

-- ============================================================
-- Demo 7: 查看死锁日志
-- ============================================================
SELECT '--- Demo 7: 查看死锁日志 ---' AS step;

-- 通常用：SHOW ENGINE INNODB STATUS\G
-- 找 LATEST DETECTED DEADLOCK 段
-- 这里用脚本模拟：手动触发锁等待
START TRANSACTION;
SELECT * FROM t_user WHERE id = 1 FOR UPDATE;

-- 锁信息查询（Session B 锁 id=2 后会形成等待）
SELECT 
    engine_transaction_id,
    object_name,
    lock_type,
    lock_mode,
    lock_status,
    lock_data
FROM performance_schema.data_locks 
WHERE object_name = 't_user'\G

ROLLBACK;

-- ============================================================
-- Demo 8: SKIP LOCKED 实战
-- ============================================================
SELECT '--- Demo 8: SKIP LOCKED 抢任务 ---' AS step;

-- 模拟任务表
DROP TABLE IF EXISTS t_task;
CREATE TABLE t_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    status ENUM('pending', 'processing', 'done') NOT NULL DEFAULT 'pending',
    payload VARCHAR(100)
) ENGINE=InnoDB;

INSERT INTO t_task (status, payload) VALUES
('pending', 'task1'), ('pending', 'task2'), ('pending', 'task3'),
('pending', 'task4'), ('pending', 'task5');

-- Worker 抢任务
-- SELECT * FROM t_task WHERE status = 'pending'
--   ORDER BY id LIMIT 2 FOR UPDATE SKIP LOCKED;
-- 跳过被其他 worker 锁的，每 worker 拿不同任务

SELECT '  实际场景：多个 worker 抢任务，SKIP LOCKED 防止争抢' AS note;

-- ============================================================
-- Demo 9: NOWAIT 立即失败
-- ============================================================
SELECT '--- Demo 9: NOWAIT 立即失败 ---' AS step;

-- Session A 锁住 id=1
-- START TRANSACTION;
-- SELECT * FROM t_user WHERE id = 1 FOR UPDATE;

-- Session B 用 NOWAIT（不等）
-- SELECT * FROM t_user WHERE id = 1 FOR UPDATE NOWAIT;
-- ERROR 3572: lock(s) could not be acquired immediately

-- ============================================================
-- Demo 10: MDL 锁（坑）
-- ============================================================
SELECT '--- Demo 10: MDL 锁 ---' AS step;

-- Session A 启动事务
-- START TRANSACTION;
-- SELECT * FROM t_user LIMIT 1;
-- （不加 FOR UPDATE，但 MDL 读锁自动加）

-- Session B 改表结构
-- ALTER TABLE t_user ADD COLUMN dummy INT;
-- 阻塞！因为 Session A 还持有 MDL 读锁

-- 解决：Session A 要么提交/回滚，要么设锁超时
-- SET GLOBAL lock_wait_timeout = 10;  -- 8.0 新增

-- ============================================================
-- Demo 11: 自增锁模式
-- ============================================================
SELECT '--- Demo 11: 自增锁 ---' AS step;

SHOW VARIABLES LIKE 'innodb_autoinc_lock_mode';
-- 0: 传统模式（每次 INSERT 都锁，最安全）
-- 1: 连续模式（默认，简单 INSERT 用轻量级锁，批量 INSERT 才锁）
-- 2: 混合模式（最宽松，性能最好，但简单 INSERT 也有锁）

-- 实际生产推荐 1

-- ============================================================
-- 清理
-- ============================================================
DROP DATABASE IF EXISTS demo_lock;
SELECT '--- Demo 结束 ---' AS end;

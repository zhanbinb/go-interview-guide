-- ============================================================
-- MySQL 事务 demo
-- 涵盖：ACID、4 隔离级别、MVCC、Next-Key Lock
-- ============================================================
-- 用法：mysql -uroot -p < demo.sql
-- 要求：MySQL 8.0+
-- ============================================================

SELECT '=== MySQL 事务 Demo ===' AS title;
SELECT VERSION() AS mysql_version;

-- 准备测试库
DROP DATABASE IF EXISTS demo_tx;
CREATE DATABASE demo_tx DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE demo_tx;

CREATE TABLE t_user (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL,
    INDEX idx_age (age)
) ENGINE=InnoDB;

INSERT INTO t_user (name, age) VALUES
('Alice', 25), ('Bob', 30), ('Charlie', 35),
('David', 28), ('Eve', 32);

SELECT '--- 准备数据 ---' AS step;
SELECT * FROM t_user;

-- ============================================================
-- Demo 1: 查看和设置隔离级别
-- ============================================================
SELECT '--- Demo 1: 隔离级别 ---' AS step;

SELECT @@global.transaction_isolation AS global_level;
SELECT @@session.transaction_isolation AS session_level;

-- 重要：MySQL 默认 RR，SQL Server 默认 RC
-- 面试必问：MySQL 为什么默认 RR？
-- 答：RR 级别通过 Next-Key Lock 解决幻读，平衡并发度和一致性

-- 切换隔离级别
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
SELECT @@session.transaction_isolation AS now_isolation;
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;  -- 恢复默认

-- ============================================================
-- Demo 2: ACID 演示
-- ============================================================
SELECT '--- Demo 2: ACID 之原子性（Atomicity）---' AS step;

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('Frank', 40);
INSERT INTO t_user (name, age) VALUES ('Grace', 45);
-- 模拟错误，回滚
ROLLBACK;

SELECT '  回滚后，Frank 和 Grace 不存在' AS result;
SELECT COUNT(*) AS user_count FROM t_user WHERE name IN ('Frank', 'Grace');

-- ============================================================
SELECT '--- Demo 2.2: ACID 之持久性（Durability）---' AS step;

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('Helen', 50);
COMMIT;
-- 即使 MySQL 重启，Helen 也在（靠 redo log 恢复）

SELECT * FROM t_user WHERE name = 'Helen';

-- ============================================================
-- Demo 3: 脏读（READ UNCOMMITTED 才会有）
-- ============================================================
SELECT '--- Demo 3: 脏读测试 ---' AS step;

-- Session 1: 设置 RU，启动事务，插入但不提交
-- Session 2: 立即查询（应该看到未提交的数据 = 脏读）
-- （这里只是说明，需要 2 个连接才能演示）

-- 实际生产用 RU 是危险的，几乎不用
SET SESSION TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;
START TRANSACTION;
SELECT '  RU 隔离级别下能脏读，但实际几乎不用' AS warning;
ROLLBACK;
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- ============================================================
-- Demo 4: 不可重复读（RC 隔离级别）
-- ============================================================
SELECT '--- Demo 4: 不可重复读（RC 级别）---' AS step;

SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- Session A 模拟：先读一次
START TRANSACTION;
SELECT name, age FROM t_user WHERE id = 1;  -- Alice, 25
-- 这里 Session B 修改并提交
-- （实际演示需要 2 个连接）
SELECT '  RC 级别：同一事务内再次读同一行，结果可能不同' AS explain;

ROLLBACK;
SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;

-- ============================================================
-- Demo 5: RR 级别快照读（核心！MVCC 体现）
-- ============================================================
SELECT '--- Demo 5: RR 级别快照读（不阻塞）---' AS step;

SET SESSION TRANSACTION ISOLATION LEVEL REPEATABLE READ;
START TRANSACTION;

-- 事务开始时 Alice 是 25
SELECT id, name, age FROM t_user WHERE id = 1;  -- Alice, 25

-- 假设另一个连接把 Alice 改成 30 并提交
UPDATE t_user SET age = 30 WHERE id = 1;  -- 这会在当前连接改，但演示用
COMMIT;
-- 注意：MySQL 同一个 session 的 UPDATE 会影响 MVCC 视图

-- 用新的事务读，应该还是 25（快照读）
START TRANSACTION;
SELECT id, name, age FROM t_user WHERE id = 1;  -- 取决于事务开始时间
ROLLBACK;

-- 还原
UPDATE t_user SET age = 25 WHERE id = 1;

-- ============================================================
-- Demo 6: 当前读（FOR UPDATE）能读到最新数据
-- ============================================================
SELECT '--- Demo 6: 当前读能读到最新数据 ---' AS step;

START TRANSACTION;
-- 假设另一个连接把 age 改成 30 并提交
UPDATE t_user SET age = 30 WHERE id = 1;
COMMIT;

-- 当前读：读最新已提交
SELECT id, name, age FROM t_user WHERE id = 1 FOR UPDATE;
-- 看到 age = 30（最新）

ROLLBACK;
-- 还原
UPDATE t_user SET age = 25 WHERE id = 1;

-- ============================================================
-- Demo 7: 幻读与 Next-Key Lock
-- ============================================================
SELECT '--- Demo 7: 幻读与 Next-Key Lock ---' AS step;

-- 准备测试数据
TRUNCATE t_user;
INSERT INTO t_user (name, age) VALUES ('A', 10), ('C', 30), ('E', 50);
SELECT * FROM t_user;

-- 演示间隙锁（需要 2 个 session 才能完整演示）
-- 这里说明原理：
-- 假设数据 id: 10, 30, 50
-- Session A: SELECT * FROM t_user WHERE id BETWEEN 20 AND 40 FOR UPDATE;
-- 加的锁：
--   - 记录锁: id=30
--   - 间隙锁: (10, 30), (30, 50)
-- Session B: INSERT INTO t_user (id) VALUES (25);
--   会被阻塞！因为 25 在间隙 (10, 30) 里
-- Session B: INSERT INTO t_user (id) VALUES (60);
--   不会被阻塞，因为 60 在间隙 (50, +∞) 之外

-- ============================================================
-- Demo 8: 事务传播（嵌套 savepoint）
-- ============================================================
SELECT '--- Demo 8: Savepoint 嵌套事务 ---' AS step;

START TRANSACTION;
INSERT INTO t_user (name, age) VALUES ('Outer1', 20);

SAVEPOINT sp1;
INSERT INTO t_user (name, age) VALUES ('Inner1', 30);
INSERT INTO t_user (name, age) VALUES ('Inner2', 40);

-- 回滚到 savepoint（保留 Outer1，回滚 Inner1/2）
ROLLBACK TO sp1;
INSERT INTO t_user (name, age) VALUES ('Outer2', 50);

COMMIT;

SELECT * FROM t_user ORDER BY id;
-- 期望: Outer1, Outer2（Inner1/2 被回滚了）

-- ============================================================
-- Demo 9: 死锁演示（理论，实际需要 2 个 session）
-- ============================================================
SELECT '--- Demo 9: 死锁（理论）---' AS step;

-- Session A: 锁住 id=1 的行
-- BEGIN; SELECT * FROM t_user WHERE id = 1 FOR UPDATE;

-- Session B: 锁住 id=2 的行
-- BEGIN; SELECT * FROM t_user WHERE id = 2 FOR UPDATE;

-- Session A: 尝试锁 id=2（被 Session B 锁了，等待）
-- SELECT * FROM t_user WHERE id = 2 FOR UPDATE;

-- Session B: 尝试锁 id=1（被 Session A 锁了，等待）
-- SELECT * FROM t_user WHERE id = 1 FOR UPDATE;
-- 死锁！MySQL 自动检测，回滚一个事务

-- 查看死锁日志
SHOW ENGINE INNODB STATUS\G

-- ============================================================
-- 清理
-- ============================================================
DROP DATABASE IF EXISTS demo_tx;
SELECT '--- Demo 结束 ---' AS end;

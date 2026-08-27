# MySQL 面试复习指南

> 配套原文：[mao888/golang-guide - mysql/](https://github.com/mao888/golang-guide/tree/main/mysql)
> 你的背景：熟悉 SQL Server，所以会重点对比两者差异
> 目标：精简主流，应付面试

## 目录

### 🔥 必问核心
- [01-架构：一条 SQL 的执行流程](01-architecture/)
- [02-存储引擎：InnoDB vs MyISAM](02-engine/)
- [03-索引：B+Tree + 联合 + 覆盖 + EXPLAIN](03-index/)
- [04-事务：ACID + 4 隔离级别 + MVCC](04-transaction/)
- [05-锁：行锁 + 意向锁 + Next-Key Lock + 死锁](05-lock/)
- [06-日志：binlog + redo log + undo log + 两阶段提交](06-log/)
- [07-主从复制：binlog 原理 + 3 种复制方式](07-replication/)
- [08-性能优化：EXPLAIN + 慢 SQL + 索引 + 分库分表](08-optimization/)

### ⭐ 重点掌握
- [09-MySQL vs SQL Server 对比](09-vs-sqlserver/) ⭐ 专门为你准备
- [10-SQL 基础：三大范式 + JOIN 原理](11-sql-basics/)
- [11-MySQL 8.0 新特性：窗口函数 + CTE](10-new-features/)

### 💡 选学
- [12-性能调优：Buffer Pool + 慢查询 + 锁等待](12-tuning/)

## 跑法（跟 Go 项目一样）

每个子目录都有：
- `README.md` - 面试速记卡
- `demo.sql` - 可执行 SQL 示例
- `test.sql` - 验证 SQL

本地执行（需要 MySQL）:
    mysql -uroot -p < demo.sql

或者用 Docker 快速起一个:
    docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root mysql:8.0

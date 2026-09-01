# MySQL 主从复制（binlog + 读写分离）

> 面试**必问**章节，跟 06-log 强相关
> 重点：binlog 复制原理、3 种复制方式、读写分离、延迟解决
> 实际生产中所有大厂项目都在用

---

## 🔥 必问核心

### 1. 什么是主从复制？

**简单说**：把 Master 的数据自动同步到 Slave

```
┌─────────┐      binlog      ┌──────────┐
│  Master │  ──────────────→  │  Slave 1 │  ← 读
│  (写)   │                  │  (读)   │  ← 应用读
└─────────┘                  └──────────┘
                              ┌──────────┐
                              │  Slave 2 │  ← 备份
                              │  (读)   │
                              └──────────┘
```

**目的**：
1. **读写分离**（Master 写，Slave 读）
2. **数据备份**（Slave 是热备）
3. **高可用**（Master 挂了 Slave 顶上）
4. **负载均衡**（多个 Slave 分摊读压力）

### 2. 复制原理（3 步，必背）

```
Master 端:                          Slave 端:
                                    
  客户端写 SQL                     
  │                                 
  ↓                                 
  Master 执行事务                   
  │                                 
  ↓                                 
  写 binlog (按事务顺序)            
  │                                 
  ↓                                 
  ┌───── dump thread ───┐            
  │ 把 binlog 推给 Slave │   ──────→  Slave IO Thread
  └────────────────────┘             │  把 binlog 写入 relay log
                                     │
                                     ↓
                                   Slave SQL Thread
                                     │  重放 relay log
                                     ↓
                                   Slave 数据跟 Master 一致
```

**3 步**：
1. **Master 写 binlog**（顺序写，很快）
2. **Slave IO Thread 拉 binlog** → 写入 **relay log**（中继日志）
3. **Slave SQL Thread 重放 relay log** → 同步到 Slave 表

**关键词速记**：**binlog + relay log + 重放**

### 3. binlog 三种格式（06-log 复习）

| 格式 | 内容 | 优点 | 缺点 | 推荐 |
|------|------|------|------|------|
| STATEMENT | 原始 SQL | 简单，文件小 | `now()` 不一致 | ❌ |
| **ROW** | 每行数据变化 | 准确 | 文件大 | ✅ **生产默认** |
| MIXED | 混合模式 | 折中 | 看场景 | ⚠️ |

**ROW 模式必用**：主从数据严格一致，但 binlog 文件会变大。

### 4. 3 种复制方式（核心区分！）

| 方式 | 数据一致性 | 性能 | 场景 |
|------|----------|------|------|
| **异步复制**（默认）| ❌ 可能丢数据 | ✅ 最快 | 性能优先，可丢少量数据 |
| **半同步复制** | ✅ 至少 1 个 slave 收到 | ⚠️ 略慢 | 数据不能丢（推荐）|
| **同步复制** | ✅ 全部 slave 收到 | ❌ 最慢 | 金融级别（极少用）|

**异步复制的问题**：
```
Master 写完 binlog 就返回客户端"成功"
但 Master 此时挂了 → binlog 没推给 Slave → 数据丢失
```

**半同步复制**（MySQL 5.7+ 推荐）：
```
Master 写完 binlog → 等待至少 1 个 Slave 收到 → 才返回"成功"
至少保证 1 个 Slave 有这份数据
```

**配置**：
```sql
-- Master 端安装半同步插件
INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';
SET GLOBAL rpl_semi_sync_master_enabled = 1;

-- Slave 端安装
INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';
SET GLOBAL rpl_semi_sync_slave_enabled = 1;
```

### 5. 主从延迟（实战问题）

**为什么会延迟**：
- Master 写完 binlog 后异步推给 Slave
- Slave 单线程重放（默认 `slave_parallel_workers=0`）
- 大事务 / DDL / 大表 JOIN 都会让延迟变大

**查询延迟**：
```sql
SHOW SLAVE STATUS\G
-- 看 Seconds_Behind_Master
```

**解决延迟的 7 个方法**：
1. **大事务拆小事务**（最重要）
2. **多线程复制**：`SET GLOBAL slave_parallel_workers = 8`（MySQL 5.6+）
3. **半同步复制**（保证数据不丢，但延迟会变大）
4. **强制走主库**（核心业务查询，MySQL 8.0 不需要看延迟）
5. **延迟从库**（专门用于报表等不敏感读）
6. **缓存**（Redis 缓存热点数据）
7. **业务拆分**（不同业务用不同主从）

### 6. 读写分离（实战架构）

```
                  ┌────→ Slave 1 (应用读)
Client ───→ Proxy ┼────→ Slave 2 (报表读)
                  │        ↑
                  └────→ Master (应用写)
```

**实现方式**：
| 方式 | 优缺点 |
|------|--------|
| 应用层判断 | 简单（Service 注解），但耦合 |
| 中间件（MyCat / Sharding-JDBC）| 透明，但多一层 |
| 驱动层（go-mysql-driver）| Go 生态常用 |

**注意点**：
- 主从延迟时可能读到旧数据
- 写后立即读应该走 Master（用 `@Master` 注解）
- 用 `semi-sync` 减少延迟但不能完全消除

---

## ⭐ 重点掌握

### 7. GTID 复制（MySQL 5.6+，推荐）

**GTID = Global Transaction Identifier**
- 每个事务一个全局唯一 ID
- 格式：`server_uuid:transaction_id`

**好处**：
- 主从切换不需要找 binlog 位点
- 主从状态一致性自动检测
- 配置简单（不用 `master_log_file + master_log_pos`）

**配置**：
```ini
# my.cnf
gtid_mode = ON
enforce_gtid_consistency = ON
```

**MySQL 8.0 默认就是 GTID 复制**

### 8. 复制模式

**基于位置（传统）**：
- `CHANGE MASTER TO MASTER_LOG_FILE='binlog.000001', MASTER_LOG_POS=4;`
- 手动指定位点

**基于 GTID（推荐）**：
- `CHANGE MASTER TO MASTER_AUTO_POSITION=1;`
- 自动找同步位点

### 9. 一主多从 vs 多主多从

| 架构 | 优点 | 缺点 | 适用 |
|------|------|------|------|
| **一主多从** | 简单、读扩展 | 单点（master 挂了要切换）| 读多写少 |
| **MHA + 一主多从** | 30 秒自动切换 | 复杂度高 | 生产推荐 |
| **MGR (MySQL Group Replication)** | 多主，高可用 | 性能有损 | MySQL 5.7+，金融 |
| **主主复制** | 双写 | 冲突解决复杂 | ❌ 不推荐 |
| **级联复制** | 减轻 Master 压力 | 延迟更大 | 数据仓库场景 |

### 10. 主从切换（故障转移）

**手动切换**：
```sql
-- Slave 端
STOP SLAVE;
RESET MASTER;
-- 修改应用连接串到新 Master
```

**自动切换**（MHA / Orchestrator / ProxySQL）：
```
Master 挂了
  → MHA 选数据最新的 Slave 提升为新 Master
  → 其他 Slave 重新指向新 Master
  → 应用连接串切换
  → 整个过程 10-30 秒
```

### 11. 复制问题排查

**常见问题**：
| 现象 | 原因 | 解决 |
|------|------|------|
| `Seconds_Behind_Master: NULL` | 复制中断 | `START SLAVE` |
| 主键冲突 | 多 slave 写 | 只让 master 写 |
| 丢数据 | 异步复制 + Master 挂 | 半同步 |
| Slave 卡住 | 大事务 / 锁 | 拆小事务 / `STOP SLAVE; SET GLOBAL slave_parallel_workers=4; START SLAVE` |

**诊断命令**：
```sql
-- 主从状态
SHOW MASTER STATUS;            -- Master
SHOW SLAVE STATUS\G          -- Slave

-- 主从延迟
SHOW SLAVE STATUS\G          -- 看 Seconds_Behind_Master

-- 复制错误
SHOW SLAVE STATUS\G          -- 看 Last_IO_Error / Last_SQL_Error

-- 性能元数据
SELECT * FROM performance_schema.replication_applier_status_by_worker;
```

### 12. SQL Server vs MySQL 复制对比

| 维度 | SQL Server | MySQL |
|------|-----------|-------|
| 机制 | 发布/订阅（pub/sub）| binlog + relay log 重放 |
| 延迟 | 接近实时 | 默认异步（可能延迟）|
| 双向同步 | 原生支持（合并复制）| ❌ 不推荐（GTID + 多主可勉强做）|
| 高可用 | Always On Availability Group | MHA / MGR / Orchestrator |
| 监控 | SSMS 内置 | 需要第三方工具（Orchestrator 等）|

---

## 💡 选学

### 13. 并行复制（MySQL 5.7+ 重要改进）

```sql
-- Slave 端
SET GLOBAL slave_parallel_workers = 8;        -- 8 个 worker 并行
SET GLOBAL slave_parallel_type = LOGICAL_CLOCK; -- 按组并行
```

**前提**：ROW 格式 binlog（STATEMENT 不能并行）
**效果**：延迟从 几小时 降到 几秒

### 14. MGR（MySQL Group Replication，MySQL 5.7+）

**特点**：
- 基于 Paxos 协议
- 多主模式（任何节点都能写）
- 自动故障检测
- 数据强一致性

**配置复杂**，但**多主 + 高可用**场景有用

**适用**：金融、订单等强一致场景

### 15. 延迟从库（delayed replica）

```sql
CHANGE MASTER TO MASTER_DELAY = 3600;  -- 延迟 1 小时
```

**场景**：防止误操作（比如 `DELETE WHERE 1=1`）
- 1 小时后从库才执行这条 SQL
- 有时间从延迟从库恢复

---

## 🎯 面试最常被问的 5 个问题

1. **MySQL 主从复制原理？**
   - 答：Master 写 binlog → Slave IO Thread 拉 → 写 relay log → SQL Thread 重放 → 同步到 Slave

2. **binlog 三种格式？生产用哪个？**
   - 答：STATEMENT / ROW / MIXED。生产用 ROW（准确，8.0 默认）

3. **3 种复制方式？区别？**
   - 答：异步（可能丢数据，最快）、半同步（至少 1 个 slave 收到，推荐）、同步（全部收到，最慢）

4. **主从延迟怎么解决？**
   - 答：大事务拆小、多线程复制、半同步、关键读走主库、用缓存、业务拆分

5. **主从切换怎么做？**
   - 答：手动（停从库→改连接串）或自动（MHA / Orchestrator 检测 Master 挂 → 选最新 Slave 提升）

---

## 📋 速查表

```
| 方式       | 数据一致性 | 性能 | 场景 |
|------------|----------|------|------|
| 异步       | 可能丢   | 最快 | 性能优先 |
| 半同步     | 不丢     | 中等 | 生产推荐 |
| 同步       | 强一致   | 最慢 | 金融 |

3 步复制：binlog → relay log → 重放
ROW 格式：必用
GTID：8.0 默认，推荐
多线程复制：slave_parallel_workers = 8
读写分离：应用层 / ProxySQL / Sharding-JDBC
延迟：SHOW SLAVE STATUS 看 Seconds_Behind_Master
```

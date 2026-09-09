# Redis 数据类型（5 大基础 + 3 大特殊）

> 面试**必问**章节
> 重点：每个类型的应用场景 + 命令时间复杂度
> 实战：高频命令 + 真实场景举例

---

## 🔥 必问核心

### 1. Redis 5 大基础数据类型（必背）

| 类型 | 底层数据结构 | 典型命令 | 时间复杂度 | 核心场景 |
|------|-------------|---------|-----------|----------|
| **String** | SDS（简单动态字符串）| `GET/SET/INCR/MSET/MGET` | O(1) | 缓存、计数器、分布式锁、Session |
| **List** | quicklist（3.2+）/linkedlist | `LPUSH/RPUSH/LPOP/LRANGE` | 头尾 O(1)，中间 O(N) | 消息队列、栈、最新列表 |
| **Hash** | dict（hashtable）| `HSET/HGET/HMSET/HGETALL/HINCRBY` | O(1) | 对象存储、用户资料、配置 |
| **Set** | dict（value 为 NULL 的 hashtable）| `SADD/SMEMBERS/SINTER/SUNION/SDIFF` | O(1) | 去重、共同好友、标签 |
| **ZSet** | skiplist + dict | `ZADD/ZRANGE/ZRANGEBYSCORE/ZINCRBY` | O(log N) | 排行榜、延迟队列、滑动窗口 |

**口诀**：String 字符串、List 列表、Hash 哈希、Set 集合、ZSet 有序集合

### 2. 3 大特殊数据类型（加分）

| 类型 | 底层 | 命令 | 场景 |
|------|------|------|------|
| **Bitmap** | String（位操作）| `SETBIT/GETBIT/BITCOUNT` | 用户签到、日活统计、布隆过滤器 |
| **GEO** | zset（用 GeoHash 编码）| `GEOADD/GEODIST/GEORADIUS` | 附近的人、骑手距离 |
| **HyperLogLog** | 概率数据结构 | `PFADD/PFCOUNT/PFMERGE` | UV 统计（基数估算，误差 < 1%）|

### 3. String 类型详解（最常用）

**底层结构**：SDS（Simple Dynamic String），不是 C 字符串！

**SDS 优势**：
- O(1) 取长度（C 字符串 O(N)）
- 杜绝缓冲区溢出（自动扩容）
- 减少内存重分配（空间预分配 + 惰性释放）

**常用命令**：
```bash
SET key value                  # 设置
GET key                        # 读取
INCR counter                   # +1（原子）
DECR counter                   # -1（原子）
INCRBY counter 5               # +5
MSET k1 v1 k2 v2              # 批量设（不是原子的）
MGET k1 k2                    # 批量读
SETEX key 60 value             # 设置 + 60s 过期
SETNX key value                # 不存在才设（分布式锁用）
APPEND key "world"             # 追加
GETRANGE key 0 -1             # 取子串
```

**3 种特殊编码**（Redis 内部优化）：
- `int`：8 字节长整型
- `embstr`：≤ 44 字节的字符串
- `raw`：> 44 字节

**应用场景**：
| 场景 | 实现 |
|------|------|
| 缓存 | `SET key json` + `EXPIRE key 3600` |
| 计数器 | `INCR article:1:views` |
| 分布式锁 | `SETNX lock + EXPIRE` |
| Session | `SET session:xxx` |
| 全局唯一 ID | `INCR order:id` |
| 限流 | `INCR rate:ip` + 检查 |

### 4. Hash 类型详解

**底层结构**：dict（hashtable + 渐进式 rehash）

**常用命令**：
```bash
HSET user:1 name "Alice" age 25     # 设置字段
HMSET user:1 name "Alice" age 25   # 批量设置（已废弃，用 HSET 多字段）
HGET user:1 name                    # 取一个字段
HMGET user:1 name age              # 批量取
HGETALL user:1                     # 取所有
HINCRBY user:1 age 1               # 字段 +1
HDEL user:1 age                    # 删字段
HEXISTS user:1 name                # 判断字段
HKEYS user:1                       # 所有字段名
HVALS user:1                       # 所有值
```

**应用场景**：
- **用户资料**：`HSET user:1 name "Alice" age 25`（单 key 存整个对象）
- **购物车**：`HSET cart:user:1 item:001 2`（用户购物车里 item 001 的数量）
- **配置**：`HSET config:app theme dark lang zh`

**注意**：
- Hash 适合**字段经常变**的场景（如用户资料）
- 如果字段不会变 → 用 String + JSON 更省内存（省 hash 字段名）
- Hash 不适合字段数特别多的场景（hashtable 冲突）

### 5. List 类型详解

**底层结构**：quicklist（3.2+，listpack + linkedlist 混合）

**常用命令**：
```bash
LPUSH list v1              # 左推入
RPUSH list v2              # 右推入
LPOP list                  # 左弹出
RPOP list                  # 右弹出
LRANGE list 0 -1           # 取所有
LLEN list                  # 长度
LINDEX list 0              # 按索引取
LREM list 2 v              # 删 2 个 v
```

**应用场景**：
| 场景 | 实现 |
|------|------|
| 消息队列 | `LPUSH` + `RPOP`（FIFO）|
| 栈 | `LPUSH` + `LPOP`（LIFO）|
| 最新列表 | `LPUSH` + `LTRIM 0 99`（只保留 100 条）|
| 阻塞队列 | `BRPOP` 代替 `RPOP`（阻塞）|

**性能注意**：
- 头尾操作 O(1)
- 中间操作 O(N)（别做 `LINDEX 1000`）

### 6. Set 类型详解

**底层结构**：dict（value 全为 NULL 的 hashtable）

**常用命令**：
```bash
SADD set m1 m2 m3           # 加
SMEMBERS set                 # 所有成员
SISMEMBER set m1            # 判断成员
SCARD set                    # 元素个数
SREM set m1                  # 删
SPOP set 3                   # 随机抽 3 个（抽奖用）
SRANDMEMBER set 3           # 不删除的随机抽

SINTER set1 set2             # 交集
SUNION set1 set2             # 并集
SDIFF set1 set2              # 差集（set1 - set2）
```

**应用场景**：
| 场景 | 实现 |
|------|------|
| 去重 | `SADD tag:user:1 "go" "java"` |
| 共同好友 | `SINTER friend:1 friend:2` |
| 推荐关注 | `SDIFF other:read:user:1` |
| 抽奖 | `SPOP lottery 3` |
| UV 统计（精确）| `SCARD uv:date:20260801` |

### 7. ZSet（Sorted Set）类型详解

**底层结构**：**skiplist（跳表）+ dict**（双结构，O(log N) 排序和 O(1) 查找）

**常用命令**：
```bash
ZADD zset 100 "Alice"        # 加
ZSCORE zset "Alice"          # 取分数
ZRANGE zset 0 9              # 按分数升序取前 10
ZREVRANGE zset 0 9           # 倒序
ZRANGEBYSCORE zset 80 100    # 按分数范围
ZINCRBY zset 5 "Alice"       # 分数 +5
ZREM zset "Alice"            # 删
ZCARD zset                   # 个数

ZINTERSTORE out 2 z1 z2       # 多 zset 交集（按分数）
```

**应用场景**（超多）：
| 场景 | 实现 |
|------|------|
| 排行榜 | `ZADD rank score member` + `ZREVRANGE 0 9` |
| 延迟队列 | score = 执行时间戳，循环 `ZRANGEBYSCORE` 取到期任务 |
| 滑动窗口限流 | score = 时间戳，`ZREMRANGEBYSCORE` 清旧 + `ZCARD` 计数 |
| 商品价格排序 | `ZADD price:sku price sku_id` |
| 排行榜 + 排名 | `ZREVRANK` |

### 8. 时间复杂度对比表（必背）

| 操作 | String | List | Hash | Set | ZSet |
|------|--------|------|------|-----|------|
| 增 | O(1) | O(1) | O(1) | O(1) | O(log N) |
| 删 | O(1) | O(N) | O(1) | O(1) | O(log N) |
| 改 | O(1) | O(N) | O(1) | O(1) | O(log N) |
| 查 | O(1) | O(N) | O(1) | O(1) | O(log N) |
| 范围 | O(N) | O(N) | O(N) | O(N) | O(log N + M) |

**记住**：ZSet 比其他类型慢（因为要维护有序），但仍是 O(log N)，不会变 O(N)

### 9. 类型选择速查

```
要存字符串/计数器/锁     → String
要存对象（多个字段）      → Hash（字段变化多）或 String + JSON（字段稳定）
要存列表/队列             → List
要去重/共同好友/抽奖       → Set
要排行榜/按分数排序       → ZSet
要布隆过滤器/签到          → Bitmap
要附近的人/距离计算        → GEO
要 UV 估算（误差 1% OK）  → HyperLogLog
```

---

## ⭐ 重点掌握

### 10. 特殊类型应用场景详解

**Bitmap（位图）**：
```bash
SETBIT user:1:login 0 1    # 第 0 天登录
SETBIT user:1:login 1 1    # 第 1 天登录
BITCOUNT user:1:login     # 总共登录几天
GETBIT user:1:login 100   # 第 100 天是否登录
```

**GEO**：
```bash
GEOADD city 116.40 39.90 "Beijing"   # 加位置
GEODIST city "Beijing" "Shanghai" km # 距离
GEORADIUS city 116.40 39.90 1000 km WITHDIST # 附近 1000km
```

**HyperLogLog**：
```bash
PFADD uv:2026-09-01 user1 user2 user3   # 加
PFCOUNT uv:2026-09-01                  # 估算基数（~0.81% 误差）
PFMERGE uv:week37 uv:day1 uv:day2 ...    # 合并
```
**注意**：HyperLogLog 是**估算**，不精确；UV 统计用它省内存（12KB 存几亿基数）

### 11. 底层数据结构速览（02 章节详解）

| 类型 | 底层 | 章节 |
|------|------|------|
| String | SDS（简单动态字符串）| 02 |
| List | quicklist（3.2+）| 02 |
| Hash | dict（hashtable）| 02 |
| Set | dict（value 为 NULL）| 02 |
| ZSet | skiplist + dict | 02 |

---

## 💡 选学（实战细节）

### 12. 大 Key / 热 Key 问题

**大 Key 危害**（面试实战）：
- `KEYS *` 阻塞（生产禁止）
- 单 Key 删除慢（DEL bigkey）
- 集群迁移困难（migrate slot）

**检测**：
```bash
redis-cli --bigkeys    # 扫描大 key
```

**解决**：
- 大 List 拆成多个小 List
- Hash 用分片：`user:1:hash:tag1`, `user:1:hash:tag2`
- 用 `UNLINK` 替代 `DEL`（异步删除）

### 13. 慢查询

```bash
CONFIG SET slowlog-log-slower-than 10000  # 10ms
CONFIG SET slowlog-max-len 128
SLOWLOG GET
```

---

## 🎯 面试最常被问的 5 个问题

1. **Redis 有哪些数据类型？底层分别是什么？**
   - 答：5 大基础（String/SDS、List/quicklist、Hash/dict、Set/dict、ZSet/skiplist+dict）+ 3 特殊（Bitmap/HyperLogLog/GEO）

2. **ZSet 底层是什么？为什么用跳表不用红黑树？**
   - 答：skiplist + dict。跳表实现简单、范围查询 O(log N)、并发友好（局部锁）

3. **Hash 和 String + JSON 怎么选？**
   - 答：Hash 字段会变（部分更新省流量）；String + JSON 整体读（一次拿全）

4. **Bitmap 怎么用？场景？**
   - 答：用户签到、日活统计、节省空间（1 亿用户每天状态只用 12MB）

5. **HyperLogLog 误差多少？**
   - 答：标准误差 0.81%，可用在 UV 统计（精确 vs 估算的权衡）

---

## 📋 速查表

```
| 类型      | 命令前缀 | 应用               | 时间复杂度 |
|----------|---------|-------------------|-----------|
| String   | SET/GET  | 缓存、计数器、锁     | O(1)      |
| List     | LPUSH    | 队列、最新列表       | 头尾 O(1)  |
| Hash     | HSET     | 对象存储             | O(1)      |
| Set      | SADD     | 去重、共同好友       | O(1)      |
| ZSet     | ZADD     | 排行榜、延迟队列     | O(log N)  |
| Bitmap   | SETBIT   | 签到、日活          | O(1)      |
| GEO      | GEOADD   | 附近的人            | O(log N)  |
| HyperLogLog| PFADD  | UV 估算             | O(1)      |
```

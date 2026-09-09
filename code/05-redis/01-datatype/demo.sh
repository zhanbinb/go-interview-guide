#!/bin/bash
# ============================================================
# Redis 数据类型 demo
# ============================================================
# 用法: redis-cli < demo.sh  (或 bash demo.sh)
# 要求: Redis 5.0+, redis-cli 已安装
# 测试: redis-cli ping  # 返回 PONG
# ============================================================

echo "=== Redis 数据类型 Demo ==="
redis-cli ping

# 清空测试数据
redis-cli FLUSHDB > /dev/null

# ============================================================
# Demo 1: String（最常用）
# ============================================================
echo ""
echo "--- Demo 1: String ---"

# 基础 SET/GET
redis-cli SET user:1:name "Alice"
redis-cli GET user:1:name

# 批量
redis-cli MSET user:1:age 25 user:1:city "Beijing"
redis-cli MGET user:1:age user:1:city

# INCR 计数器（原子操作，分布式唯一 ID / 浏览量）
redis-cli SET article:1:views 0
redis-cli INCR article:1:views       # +1 → 1
redis-cli INCRBY article:1:views 10   # +10 → 11
redis-cli GET article:1:views

# SETEX 设置带过期（缓存标准用法）
redis-cli SETEX user:session:abc 60 "session_data_xxx"
redis-cli TTL user:session:abc   # 看剩余秒数

# SETNX 分布式锁
redis-cli DEL lock:order:123
redis-cli SETNX lock:order:123 "owner_1"   # 成功返回 1
redis-cli SETNX lock:order:123 "owner_2"   # 失败返回 0
redis-cli DEL lock:order:123

# ============================================================
# Demo 2: Hash（对象存储）
# ============================================================
echo ""
echo "--- Demo 2: Hash ---"

# 用户资料
redis-cli HSET user:1 name "Alice" age 25 city "Beijing" email "alice@example.com"
redis-cli HGET user:1 name
redis-cli HMGET user:1 name age
redis-cli HGETALL user:1
redis-cli HKEYS user:1
redis-cli HVALS user:1
redis-cli HLEN user:1

# 字段原子加减（HINCRBY）
redis-cli HINCRBY user:1 age 1      # age 25 → 26
redis-cli HGET user:1 age

# HEXISTS / HDEL
redis-cli HEXISTS user:1 email      # 1
redis-cli HDEL user:1 email
redis-cli HEXISTS user:1 email      # 0

# 购物车示例（field = 商品 ID, value = 数量）
redis-cli HSET cart:user:1 sku:001 2 sku:002 1 sku:003 3
redis-cli HGETALL cart:user:1
redis-cli HINCRBY cart:user:1 sku:001 1   # 商品 001 数量 +1
redis-cli HGET cart:user:1 sku:001

# ============================================================
# Demo 3: List（队列/最新列表）
# ============================================================
echo ""
echo "--- Demo 3: List ---"

redis-cli DEL queue:msg
redis-cli LPUSH queue:msg "msg1" "msg2" "msg3"     # 左推入
redis-cli RPUSH queue:msg "msg4" "msg5"            # 右推入
redis-cli LRANGE queue:msg 0 -1                     # 看所有
redis-cli LLEN queue:msg

# FIFO 队列（LPUSH + RPOP）
redis-cli LPUSH order:queue "order_1"
redis-cli LPUSH order:queue "order_2"
redis-cli LPUSH order:queue "order_3"
redis-cli RPOP order:queue                           # 弹出最早入队的

# LIFO 栈（LPUSH + LPOP）
redis-cli LPUSH stack:test "a" "b" "c"
redis-cli LPOP stack:test                            # 弹出最新的

# 最新 100 条（消息流常见用法）
redis-cli DEL news:latest
redis-cli LPUSH news:latest "news_1" "news_2" "news_3" "news_4" "news_5"
redis-cli LTRIM news:latest 0 99                    # 只保留 100 条
redis-cli LRANGE news:latest 0 -1

# ============================================================
# Demo 4: Set（去重/共同好友/抽奖）
# ============================================================
echo ""
echo "--- Demo 4: Set ---"

# 去重
redis-cli SADD tags:article:1 "go" "redis" "mysql" "redis"  # redis 重复会自动去重
redis-cli SMEMBERS tags:article:1
redis-cli SCARD tags:article:1                # 实际只有 3 个

# 共同好友（交集）
redis-cli SADD friend:1 "Alice" "Bob" "Charlie" "David"
redis-cli SADD friend:2 "Bob" "Charlie" "Eve"
redis-cli SINTER friend:1 friend:2             # 共同好友: Bob Charlie
redis-cli SCARD friend:1 friend:2            # 交集元素个数

# 推荐关注（差集）
redis-cli SDIFF friend:1 friend:2             # friend:1 独有的: Alice David

# 抽奖（随机弹出，不重复）
redis-cli SADD lottery:users "u1" "u2" "u3" "u4" "u5" "u6" "u7" "u8" "u9" "u10"
redis-cli SPOP lottery:users 3                # 抽 3 个

# ============================================================
# Demo 5: ZSet（排行榜/延迟队列）
# ============================================================
echo ""
echo "--- Demo 5: ZSet ---"

# 排行榜
redis-cli ZADD game:rank 1500 "player_A"
redis-cli ZADD game:rank 2300 "player_B"
redis-cli ZADD game:rank 1800 "player_C"
redis-cli ZADD game:rank 2100 "player_D"

# 按分数升序（低→高）
redis-cli ZRANGE game:rank 0 -1 WITHSCORES
# 按分数倒序（高→低）— 排行榜前 N
redis-cli ZREVRANGE game:rank 0 2 WITHSCORES

# 修改分数（玩家 B 又得 100 分）
redis-cli ZINCRBY game:rank 100 "player_B"
redis-cli ZREVRANGE game:rank 0 2 WITHSCORES    # B 升到 2400

# 排名
redis-cli ZREVRANK game:rank "player_B"   # B 排名第 0（最高分）
redis-cli ZREVRANK game:rank "player_C"   # C 排名第 2

# 分数范围查询（1800-2200 分的玩家）
redis-cli ZRANGEBYSCORE game:rank 1800 2200 WITHSCORES

# 延迟队列（score = 执行时间戳）
NOW=$(date +%s)
DELAY=$((NOW + 5))    # 5 秒后执行
redis-cli ZADD delay:tasks $DELAY "send_email_to_user_1"
redis-cli ZRANGE delay:tasks -inf +inf WITHSCORES  # 看队列

# ============================================================
# Demo 6: 特殊类型 - Bitmap
# ============================================================
echo ""
echo "--- Demo 6: Bitmap（签到/日活）---"

# 用户 100 一年（365 天）签到记录
redis-cli DEL sign:user:100
redis-cli SETBIT sign:user:100 0 1     # 第 1 天签到
redis-cli SETBIT sign:user:100 5 1     # 第 6 天签到
redis-cli SETBIT sign:user:100 100 1    # 第 101 天签到

# 总共签到几天？
redis-cli BITCOUNT sign:user:100       # 3

# 100 天这天是否签到？
redis-cli GETBIT sign:user:100 100     # 1

# 日活统计（多个用户的签到位做 OR）
redis-cli SETBIT active:20260901 100 1   # 用户 100 当天活跃
redis-cli SETBIT active:20260901 200 1   # 用户 200 当天活跃
redis-cli BITCOUNT active:20260901       # 2 个 DAU

# ============================================================
# Demo 7: 特殊类型 - HyperLogLog（UV 估算）
# ============================================================
echo ""
echo "--- Demo 7: HyperLogLog（UV 估算）---"

# 模拟 100 万用户访问（不真插 100 万，用集合操作加速）
redis-cli DEL uv:20260901
for i in $(seq 1 1000); do
    redis-cli PFADD uv:20260901 "user_$i" > /dev/null
done
redis-cli PFCOUNT uv:20260901             # 接近 1000

# HyperLogLog 内存：12KB 就能存 2^64 个不同值
redis-cli DEBUG OBJECT uv:20260901 | head -2

# ============================================================
# Demo 8: 特殊类型 - GEO（附近的人）
# ============================================================
echo ""
echo "--- Demo 8: GEO（地理位置）---"

redis-cli GEOADD cities 116.40 39.90 "Beijing"     # 经度 纬度 名称
redis-cli GEOADD cities 121.47 31.23 "Shanghai"
redis-cli GEOADD cities 113.27 23.13 "Guangzhou"
redis-cli GEOADD cities 114.06 22.54 "Shenzhen"

# 北京到上海的距离（km）
redis-cli GEODIST cities "Beijing" "Shanghai" km     # ≈ 1067 km

# 北京附近 1100km 内的城市
redis-cli GEOSEARCH cities FROMLONLAT 116.40 39.90 BYRADIUS 1100 km WITHDIST

# ============================================================
# Demo 9: 通用命令
# ============================================================
echo ""
echo "--- Demo 9: 通用命令 ---"

# EXISTS 判断 key 是否存在
redis-cli EXISTS user:1:name        # 1
redis-cli EXISTS no:such:key       # 0

# TYPE 看类型
redis-cli TYPE user:1:name         # string
redis-cli TYPE user:1             # hash
redis-cli TYPE queue:msg          # list
redis-cli TYPE game:rank          # zset

# DBSIZE 看当前 DB key 总数
redis-cli DBSIZE

# INFO 看服务信息（重点看 memory / stats）
redis-cli INFO memory | head -10
redis-cli INFO stats | head -10

# OBJECT 看 key 内部编码（验证 Redis 优化）
redis-cli OBJECT ENCODING user:1:name     # embstr/raw/int
redis-cli OBJECT IDLETIME user:1:name     # 空闲秒数

echo ""
echo "--- Demo 结束 ---"

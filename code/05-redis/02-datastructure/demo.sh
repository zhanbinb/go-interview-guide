#!/bin/bash
# ============================================================
# Redis 底层数据结构 demo
# ============================================================
# 用法: redis-cli < demo.sh  (或 bash demo.sh)
# 要求: Redis 7.0+, redis-cli 已安装
# 重点: 用 OBJECT ENCODING 验证每种类型的底层编码
# ============================================================

echo "=== Redis 底层数据结构 Demo ==="
redis-cli ping

# 清空测试数据
redis-cli FLUSHDB > /dev/null

# ============================================================
# Demo 1: SDS — 三种字符串编码 (int / embstr / raw)
# ============================================================
echo ""
echo "--- Demo 1: SDS 字符串编码 ---"

# 1.1 int 编码：值是整数
redis-cli SET s:int 12345
redis-cli OBJECT ENCODING s:int            # "int"

# 1.2 embstr 编码：≤ 44 字节
redis-cli SET s:embstr "hello"
redis-cli OBJECT ENCODING s:embstr         # "embstr"

# 1.3 raw 编码：> 44 字节（"64 - 16 - 3 - 1 = 44"）
redis-cli SET s:raw "this_is_a_very_long_string_that_exceeds_44_bytes_for_sure"
redis-cli OBJECT ENCODING s:raw            # "raw"

# 1.4 浮点数不算 int（直接是 embstr/raw 形式的字符串）
redis-cli SET s:float "3.14"
redis-cli OBJECT ENCODING s:float          # "embstr"（不是 int）

# 1.5 整数被改写成字符串后，编码会从 int 转成 embstr
redis-cli SET s:mut 100
redis-cli OBJECT ENCODING s:mut            # "int"
redis-cli APPEND s:mut "abc"               # 写字符串后
redis-cli OBJECT ENCODING s:mut            # "raw"

# ============================================================
# Demo 2: Hash — listpack ↔ dict 编码转换
# ============================================================
echo ""
echo "--- Demo 2: Hash 编码 (listpack ↔ dict) ---"

# 2.1 默认 128 字段以内 → listpack（连续内存，省内存）
redis-cli HSET h:small name "Alice" age 25 city "Beijing"
redis-cli OBJECT ENCODING h:small          # "listpack"

# 2.2 加到 128 字段后 → 升级为 dict（hashtable）
redis-cli EVAL "
  for i = 1, 130 do
    redis.call('HSET', 'h:big', 'field_' .. i, 'value_' .. i)
  end
  return 'done'
" 0
redis-cli OBJECT ENCODING h:big            # "listpack"（默认阈值 128 没超）
redis-cli HLEN h:big                       # 130

# 改阈值：把 entries 阈值调成 64，触发升级
redis-cli CONFIG SET hash-max-listpack-entries 64 > /dev/null
redis-cli OBJECT ENCODING h:big            # "listpack"（只在 HSET 时判断）
# 强制重新评估 → 写一次
redis-cli HSET h:big extra "trigger"
redis-cli OBJECT ENCODING h:big            # "dict"
# 还原
redis-cli CONFIG SET hash-max-listpack-entries 128 > /dev/null

# 2.3 value 长度超阈值也会升级（默认 64 字节）
redis-cli CONFIG SET hash-max-listpack-value 32 > /dev/null
redis-cli HSET h:bigvalue field1 "this_is_a_long_value_exceeding_32_bytes_for_sure"
redis-cli OBJECT ENCODING h:bigvalue       # "dict"
redis-cli CONFIG SET hash-max-listpack-value 64 > /dev/null

# ============================================================
# Demo 3: List — quicklist（永远是 listpack 节点）
# ============================================================
echo ""
echo "--- Demo 3: List 编码 (quicklist / listpack) ---"

# 3.1 小 list：单节点 listpack
redis-cli DEL l:small
for i in 1 2 3 4 5; do redis-cli RPUSH l:small "v$i" > /dev/null; done
redis-cli OBJECT ENCODING l:small          # "listpack"（7.0+）

# 3.2 大 list：多个 quicklistNode 串联（每个内部还是 listpack）
redis-cli DEL l:big
for i in $(seq 1 1000); do redis-cli RPUSH l:big "v$i" > /dev/null; done
redis-cli OBJECT ENCODING l:big            # "listpack"（quicklist 节点永远 listpack）
redis-cli LLEN l:big                       # 1000

# 3.3 查看 quicklist 的结构（DEBUG 内部信息）
redis-cli DEBUG OBJECT l:big               # 看 serializedlength / lru 等

# 3.4 修改节点大小参数
redis-cli CONFIG GET list-max-listpack-size     # 默认 -2 (8KB)
redis-cli CONFIG GET list-compress-depth        # 默认 0

# ============================================================
# Demo 4: Set — intset ↔ dict 编码转换
# ============================================================
echo ""
echo "--- Demo 4: Set 编码 (intset ↔ dict) ---"

# 4.1 全是整数 → intset
redis-cli DEL s:intset
redis-cli SADD s:intset 1 2 3 4 5
redis-cli OBJECT ENCODING s:intset         # "intset"

# 4.2 intset 升级：加非整数 → 升级为 hashtable（不会降级）
redis-cli SADD s:intset "abc"
redis-cli OBJECT ENCODING s:intset         # "hashtable"

# 4.3 元素数超 set-max-intset-entries（默认 512）→ 升级为 dict
redis-cli DEL s:big
for i in $(seq 1 600); do redis-cli SADD s:big $i > /dev/null; done
redis-cli OBJECT ENCODING s:big            # "hashtable"
redis-cli SCARD s:big

# 4.4 演示整型范围升级（int16 → int32 → int64）
redis-cli DEL s:upgrade
redis-cli SADD s:upgrade 100               # int16 encoding
redis-cli OBJECT ENCODING s:upgrade        # "intset"
redis-cli SADD s:upgrade 100000            # 触发升级到 int32
redis-cli OBJECT ENCODING s:upgrade        # "intset"
redis-cli SADD s:upgrade 9999999999        # 触发升级到 int64
redis-cli OBJECT ENCODING s:upgrade        # "intset"
redis-cli SADD s:upgrade "string"          # 加字符串 → hashtable
redis-cli OBJECT ENCODING s:upgrade        # "hashtable"

# ============================================================
# Demo 5: ZSet — listpack ↔ skiplist 编码转换
# ============================================================
echo ""
echo "--- Demo 5: ZSet 编码 (listpack ↔ skiplist) ---"

# 5.1 少量元素 → listpack
redis-cli DEL z:small
redis-cli ZADD z:small 100 "Alice" 200 "Bob" 300 "Charlie"
redis-cli OBJECT ENCODING z:small          # "listpack"

# 5.2 大量元素 → skiplist
redis-cli DEL z:big
for i in $(seq 1 200); do redis-cli ZADD z:big $i "member_$i" > /dev/null; done
redis-cli OBJECT ENCODING z:big            # "skiplist"

# 5.3 改阈值触发提前升级
redis-cli CONFIG SET zset-max-listpack-entries 64 > /dev/null
redis-cli OBJECT ENCODING z:big            # "skiplist"（只在 ZADD 时判断）

# 还原
redis-cli CONFIG SET zset-max-listpack-entries 128 > /dev/null

# 5.4 验证 ZSet 同时有 skiplist + dict
redis-cli ZSCORE z:big 100                 # O(1)（dict 查）
redis-cli ZRANGE z:big 0 9 WITHSCORES      # O(log N)（skiplist 遍历）

# ============================================================
# Demo 6: 哈希冲突演示（DEBUG 命令）
# ============================================================
echo ""
echo "--- Demo 6: 哈希冲突与 rehash ---"

# 6.1 插入大量 key 触发 rehash
redis-cli DEL h:rehash
for i in $(seq 1 100); do redis-cli HSET h:rehash "f$i" "v$i" > /dev/null; done
redis-cli OBJECT ENCODING h:rehash         # 看编码
redis-cli HLEN h:rehash

# 6.2 渐进式 rehash：观察 hash 表信息
redis-cli DEBUG OBJECT h:rehash             # 看 internal refcount / encoding

# 6.3 模拟扩缩容：删除到只剩 10 个
for i in $(seq 11 100); do redis-cli HDEL h:rehash "f$i" > /dev/null; done
redis-cli OBJECT ENCODING h:rehash         # 编码（缩容后还可能是 listpack）

# ============================================================
# Demo 7: 内部统计信息
# ============================================================
echo ""
echo "--- Demo 7: 内存与统计 ---"

# 7.1 各种 key 占用的内存字节数
redis-cli SET demo:tiny "x"
redis-cli SET demo:big "$(printf 'a%.0s' {1..100})"
redis-cli MEMORY USAGE demo:tiny
redis-cli MEMORY USAGE demo:big

# 7.2 看每个 key 的空闲时间（OBJECT IDLETIME）
redis-cli OBJECT IDLETIME demo:tiny
redis-cli OBJECT IDLETIME demo:big

# 7.3 命中率与统计
redis-cli INFO stats | grep -E "keyspace_hits|keyspace_misses|used_memory_human"
redis-cli INFO memory | grep -E "used_memory_human|maxmemory_human|mem_fragmentation_ratio"

# ============================================================
# Demo 8: 一次性看所有 demo key 的编码
# ============================================================
echo ""
echo "--- Demo 8: 编码总览 ---"

# 遍历 db 里所有 key，打印它们的 type + encoding
for key in $(redis-cli --scan --pattern '*' | head -50); do
  type=$(redis-cli TYPE "$key")
  enc=$(redis-cli OBJECT ENCODING "$key" 2>/dev/null)
  printf "  %-30s  type=%-8s  encoding=%s\n" "$key" "$type" "$enc"
done

echo ""
echo "--- Demo 结束 ---"

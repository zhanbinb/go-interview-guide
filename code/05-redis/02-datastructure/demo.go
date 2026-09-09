// ============================================================
// Redis 底层数据结构 demo (Go)
// ============================================================
// 用法: go run demo.go
// 要求: 本地 Redis 7.0+ 已起 (redis-server)
// 重点: 用 go-redis 调 OBJECT ENCODING 验证每种类型的底层编码
// ============================================================

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("Redis 连接失败:", err)
		return
	}

	fmt.Println("=== Redis 底层数据结构 Demo (Go) ===")
	rdb.FlushDB(ctx)

	enc := func(key string) string {
		// OBJECT ENCODING key
		s, err := rdb.Do(ctx, "OBJECT", "ENCODING", key).Result()
		if err != nil {
			return "?"
		}
		return fmt.Sprintf("%v", s)
	}

	// ============================================================
	// Demo 1: SDS — int / embstr / raw 三种字符串编码
	// ============================================================
	fmt.Println("\n--- Demo 1: SDS 字符串编码 (int / embstr / raw) ---")

	// 整数 → int 编码（直接用 long 存，省 SDS header）
	rdb.Set(ctx, "s:int", 12345, 0)
	fmt.Printf("  SET s:int 12345       → encoding = %s\n", enc("s:int"))

	// 短字符串（≤ 44 字节）→ embstr 编码（一次 malloc，header+buf 连续）
	rdb.Set(ctx, "s:embstr", "hello", 0)
	fmt.Printf("  SET s:embstr 'hello'  → encoding = %s\n", enc("s:embstr"))

	// 长字符串（> 44 字节）→ raw 编码（两次 malloc）
	longStr := strings.Repeat("a", 100)
	rdb.Set(ctx, "s:raw", longStr, 0)
	fmt.Printf("  SET s:raw 100 字符     → encoding = %s\n", enc("s:raw"))

	// 浮点数字符串 → embstr（不是 int！）
	rdb.Set(ctx, "s:float", "3.14", 0)
	fmt.Printf("  SET s:float '3.14'    → encoding = %s (注意：浮点不算 int)\n", enc("s:float"))

	// int 编码被修改后会"降级"成 raw
	rdb.Set(ctx, "s:mut", 100, 0)
	fmt.Printf("  SET s:mut 100         → encoding = %s\n", enc("s:mut"))
	rdb.Append(ctx, "s:mut", "abc")
	fmt.Printf("  APPEND s:mut 'abc'    → encoding = %s (int → raw)\n", enc("s:mut"))

	// ============================================================
	// Demo 2: Hash — listpack ↔ dict 自动升级
	// ============================================================
	fmt.Println("\n--- Demo 2: Hash 编码 (listpack ↔ dict) ---")

	// 少字段 → listpack
	rdb.HSet(ctx, "h:small", map[string]interface{}{
		"name": "Alice", "age": 25, "city": "Beijing",
	})
	fmt.Printf("  HSET h:small 3 fields → encoding = %s\n", enc("h:small"))

	// 多字段超阈值 → dict（默认阈值 128）
	bigHash := make(map[string]interface{}, 130)
	for i := 0; i < 130; i++ {
		bigHash[fmt.Sprintf("f%d", i)] = fmt.Sprintf("v%d", i)
	}
	rdb.HSet(ctx, "h:big", bigHash)
	fmt.Printf("  HSET h:big 130 fields → encoding = %s\n", enc("h:big"))

	// 调小阈值 + 触发升级
	rdb.ConfigSet(ctx, "hash-max-listpack-entries", "64")
	rdb.HSet(ctx, "h:big", "trigger", "force-reencode")
	fmt.Printf("  阈值=64 后再 HSET    → encoding = %s (升级为 dict)\n", enc("h:big"))
	rdb.ConfigSet(ctx, "hash-max-listpack-entries", "128")

	// value 长度超阈值也升级（默认 64 字节）
	rdb.ConfigSet(ctx, "hash-max-listpack-value", "32")
	rdb.HSet(ctx, "h:bigvalue", "k1", strings.Repeat("v", 100))
	fmt.Printf("  value=100字节         → encoding = %s (value 超阈值 → dict)\n", enc("h:bigvalue"))
	rdb.ConfigSet(ctx, "hash-max-listpack-value", "64")

	// ============================================================
	// Demo 3: List — quicklist（永远是 listpack 节点）
	// ============================================================
	fmt.Println("\n--- Demo 3: List 编码 (quicklist / listpack) ---")

	// 小 list → 单个 listpack 节点的 quicklist
	rdb.Del(ctx, "l:small")
	for i := 1; i <= 5; i++ {
		rdb.RPush(ctx, "l:small", fmt.Sprintf("v%d", i))
	}
	fmt.Printf("  RPUSH 5 个元素        → encoding = %s\n", enc("l:small"))

	// 大 list → 多个 quicklistNode 串联（每个内部还是 listpack）
	rdb.Del(ctx, "l:big")
	for i := 1; i <= 1000; i++ {
		rdb.RPush(ctx, "l:big", fmt.Sprintf("v%d", i))
	}
	fmt.Printf("  RPUSH 1000 个元素     → encoding = %s (永远 listpack)\n", enc("l:big"))

	// 看 quicklist 节点大小参数
	listSizeCfg, _ := rdb.ConfigGet(ctx, "list-max-listpack-size").Result()
	compressCfg, _ := rdb.ConfigGet(ctx, "list-compress-depth").Result()
	fmt.Printf("  list-max-listpack-size = %v (每节点字节, -2=8KB)\n", listSizeCfg["list-max-listpack-size"])
	fmt.Printf("  list-compress-depth    = %v (0=全不压缩)\n", compressCfg["list-compress-depth"])

	// ============================================================
	// Demo 4: Set — intset ↔ dict 自动升级
	// ============================================================
	fmt.Println("\n--- Demo 4: Set 编码 (intset ↔ dict) ---")

	// 全整数 → intset（连续内存，超省空间）
	rdb.Del(ctx, "s:intset")
	rdb.SAdd(ctx, "s:intset", 1, 2, 3, 4, 5)
	fmt.Printf("  SADD 全整数 5 个       → encoding = %s\n", enc("s:intset"))

	// 加非整数 → 升级为 hashtable（不可逆）
	rdb.SAdd(ctx, "s:intset", "abc")
	fmt.Printf("  SADD 字符串 'abc'      → encoding = %s (升级)\n", enc("s:intset"))

	// 元素数超阈值（默认 512）→ 升级为 dict
	rdb.Del(ctx, "s:big")
	bigSet := make([]interface{}, 600)
	for i := 0; i < 600; i++ {
		bigSet[i] = i + 1
	}
	rdb.SAdd(ctx, "s:big", bigSet...)
	fmt.Printf("  SADD 600 个整数        → encoding = %s (超 512 → dict)\n", enc("s:big"))

	// 整数范围升级：int16 → int32 → int64（升级不可逆）
	rdb.Del(ctx, "s:upgrade")
	rdb.SAdd(ctx, "s:upgrade", 100) // int16
	fmt.Printf("  SADD 100 (int16范围)   → encoding = %s\n", enc("s:upgrade"))
	rdb.SAdd(ctx, "s:upgrade", 100000) // 升级到 int32
	fmt.Printf("  SADD 100000 (int32)    → encoding = %s\n", enc("s:upgrade"))
	rdb.SAdd(ctx, "s:upgrade", 9999999999) // 升级到 int64
	fmt.Printf("  SADD 9999999999 (int64)→ encoding = %s\n", enc("s:upgrade"))

	// ============================================================
	// Demo 5: ZSet — listpack ↔ skiplist 编码转换
	// ============================================================
	fmt.Println("\n--- Demo 5: ZSet 编码 (listpack ↔ skiplist) ---")

	// 少元素 → listpack
	rdb.Del(ctx, "z:small")
	rdb.ZAdd(ctx, "z:small",
		redis.Z{Score: 100, Member: "Alice"},
		redis.Z{Score: 200, Member: "Bob"},
		redis.Z{Score: 300, Member: "Charlie"},
	)
	fmt.Printf("  ZADD 3 个元素          → encoding = %s\n", enc("z:small"))

	// 多元素 → skiplist
	rdb.Del(ctx, "z:big")
	zMembers := make([]redis.Z, 200)
	for i := 0; i < 200; i++ {
		zMembers[i] = redis.Z{Score: float64(i + 1), Member: fmt.Sprintf("m%d", i)}
	}
	rdb.ZAdd(ctx, "z:big", zMembers...)
	fmt.Printf("  ZADD 200 个元素        → encoding = %s\n", enc("z:big"))

	// 改阈值
	rdb.ConfigSet(ctx, "zset-max-listpack-entries", "64")
	fmt.Printf("  阈值改 64 后 encoding  = %s (不会自动降级)\n", enc("z:big"))
	rdb.ConfigSet(ctx, "zset-max-listpack-entries", "128")

	// 验证 ZSet 的双结构：ZSCORE O(1) + ZRANGE O(log N)
	score, _ := rdb.ZScore(ctx, "z:big", "m100").Result()
	rng, _ := rdb.ZRangeWithScores(ctx, "z:big", 0, 4).Result()
	fmt.Printf("  ZSCORE m100  = %.0f  (dict 查, O(1))\n", score)
	fmt.Printf("  ZRANGE 0..4  = %v  (skiplist 遍历, O(log N + M))\n", len(rng))

	// ============================================================
	// Demo 6: 哈希冲突 & rehash 演示
	// ============================================================
	fmt.Println("\n--- Demo 6: 哈希冲突与 rehash ---")

	// 6.1 DEBUG OBJECT 看内部信息（refcount / encoding / lru 等）
	rdb.Del(ctx, "h:rehash")
	hashFields := make(map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		hashFields[fmt.Sprintf("f%d", i)] = fmt.Sprintf("v%d", i)
	}
	rdb.HSet(ctx, "h:rehash", hashFields)
	fmt.Printf("  HSET 100 fields        → encoding = %s\n", enc("h:rehash"))

	// 6.2 DEBUG OBJECT 看内部
	debugObj, _ := rdb.Do(ctx, "DEBUG", "OBJECT", "h:rehash").Result()
	fmt.Printf("  DEBUG OBJECT h:rehash  → %v\n", debugObj)

	// 6.3 删到 10 个字段触发缩容
	for i := 11; i <= 100; i++ {
		rdb.HDel(ctx, "h:rehash", fmt.Sprintf("f%d", i))
	}
	fmt.Printf("  HDEL 剩 10 字段        → encoding = %s (缩容后)\n", enc("h:rehash"))

	// ============================================================
	// Demo 7: 内存使用 & 内部统计
	// ============================================================
	fmt.Println("\n--- Demo 7: 内存与统计 ---")

	rdb.Set(ctx, "demo:tiny", "x", 0)
	rdb.Set(ctx, "demo:big", strings.Repeat("a", 100), 0)
	tinyMem, _ := rdb.MemoryUsage(ctx, "demo:tiny").Result()
	bigMem, _ := rdb.MemoryUsage(ctx, "demo:big").Result()
	fmt.Printf("  demo:tiny (1 字节值)  → MEMORY USAGE = %d 字节\n", tinyMem)
	fmt.Printf("  demo:big  (100 字节值) → MEMORY USAGE = %d 字节\n", bigMem)

	// INFO 看内存 & 命中率
	memInfo, _ := rdb.Info(ctx, "memory").Result()
	statsInfo, _ := rdb.Info(ctx, "stats").Result()
	fmt.Printf("  used_memory_human      = %s\n", extractField(memInfo, "used_memory_human"))
	fmt.Printf("  mem_fragmentation_ratio= %s\n", extractField(memInfo, "mem_fragmentation_ratio"))
	fmt.Printf("  keyspace_hits          = %s\n", extractField(statsInfo, "keyspace_hits"))
	fmt.Printf("  keyspace_misses        = %s\n", extractField(statsInfo, "keyspace_misses"))

	// ============================================================
	// Demo 8: 一次性扫所有 key 打印 type + encoding
	// ============================================================
	fmt.Println("\n--- Demo 8: 编码总览 (扫库) ---")

	keys, _ := rdb.Keys(ctx, "*").Result()
	for _, k := range keys {
		t, _ := rdb.Type(ctx, k).Result()
		e, _ := rdb.Do(ctx, "OBJECT", "ENCODING", k).Result()
		fmt.Printf("  %-25s  type=%-8s  encoding=%v\n", k, t, e)
	}

	// 清理
	rdb.FlushDB(ctx)
	fmt.Println("\n=== Demo 结束 ===")
}

// extractField 从 INFO 输出里提取一行（形如 "key:value"）
func extractField(info, key string) string {
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, key+":") {
			return strings.TrimPrefix(line, key+":")
		}
	}
	return "(not found)"
}

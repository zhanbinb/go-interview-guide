// ============================================================
// Redis 底层数据结构 - 实战练习
// ============================================================
// 任务清单（每个 run 一个练习，注释掉其他）：
//   1) 猜编码：给一个 key，先猜 OBJECT ENCODING 是什么，再验证
//   2) 触发升级：把 Hash/Set/ZSet 喂大，看 listpack/intset 何时变 dict/skiplist
//   3) 缩容观察：删字段看 dict 是否变 listpack
//   4) SDS 边界：测试 int 编码被 Append 字符串后变 raw
//   5) intset 整数范围升级：int16 → int32 → int64
// 跑法: go run main.go
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
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("Redis 连接失败:", err)
		return
	}

	rdb.FlushDB(ctx)
	fmt.Println("=== Redis 底层数据结构 - 练习 ===")

	// 工具函数：取一个 key 的编码
	enc := func(key string) string {
		s, _ := rdb.Do(ctx, "OBJECT", "ENCODING", key).Result()
		return fmt.Sprintf("%v", s)
	}

	// ============================================================
	// 练习 1: 猜编码（先在脑子里想，再看实际输出）
	// ============================================================
	fmt.Println("\n[练习 1] 猜编码：每个 SET/HSET/SADD/ZADD 之后编码是什么？")

	rdb.Set(ctx, "ex:int", 100, 0)
	fmt.Printf("  整数 100              → %s  (猜：int)\n", enc("ex:int"))

	rdb.Set(ctx, "ex:short", "hi", 0)
	fmt.Printf("  字符串 'hi'            → %s  (猜：embstr)\n", enc("ex:short"))

	rdb.Set(ctx, "ex:long", strings.Repeat("a", 200), 0)
	fmt.Printf("  200 字节字符串         → %s  (猜：raw)\n", enc("ex:long"))

	rdb.HSet(ctx, "ex:hash_small", "f1", "v1", "f2", "v2")
	fmt.Printf("  2 字段 hash            → %s  (猜：listpack)\n", enc("ex:hash_small"))

	rdb.SAdd(ctx, "ex:set_int", 1, 2, 3)
	fmt.Printf("  全整数 set 3 个        → %s  (猜：intset)\n", enc("ex:set_int"))

	rdb.ZAdd(ctx, "ex:zset_small", redis.Z{Score: 1, Member: "a"})
	fmt.Printf("  1 元素 zset            → %s  (猜：listpack)\n", enc("ex:zset_small"))

	// ============================================================
	// 练习 2: 触发升级 — Hash 128 字段阈值
	// ============================================================
	fmt.Println("\n[练习 2] Hash 升级：listpack → dict")
	fmt.Println("  观察: 加到 128 字段后，下一次 HSET 触发升级")

	rdb.Del(ctx, "ex:hash_big")
	batch := 10
	for i := 1; i <= 200; i++ {
		rdb.HSet(ctx, "ex:hash_big", fmt.Sprintf("f%d", i), "v")
		// 每加 10 个打一次当前编码
		if i%batch == 0 {
			fmt.Printf("  字段数=%3d  encoding=%s  (HLEN=%d)\n", i, enc("ex:hash_big"), rdb.HLen(ctx, "ex:hash_big").Val())
		}
	}

	// ============================================================
	// 练习 3: SDS int → raw 边界
	// ============================================================
	fmt.Println("\n[练习 3] SDS int 编码被覆盖后")
	rdb.Set(ctx, "ex:sds", 999, 0)
	fmt.Printf("  初始 int              → %s\n", enc("ex:sds"))
	rdb.Append(ctx, "ex:sds", "x")
	fmt.Printf("  APPEND 'x' 后         → %s (int → raw)\n", enc("ex:sds"))
	rdb.Set(ctx, "ex:sds", 999, 0)
	fmt.Printf("  重新 SET 999           → %s (再变回 int)\n", enc("ex:sds"))

	// ============================================================
	// 练习 4: intset 整数范围升级
	// ============================================================
	fmt.Println("\n[练习 4] intset 整数范围升级（不可逆）")

	rdb.Del(ctx, "ex:intset")
	rdb.SAdd(ctx, "ex:intset", 100) // int16
	fmt.Printf("  100        (int16)   → %s\n", enc("ex:intset"))
	rdb.SAdd(ctx, "ex:intset", 70000) // 升级到 int32
	fmt.Printf("  +70000     (int32)   → %s (升级)\n", enc("ex:intset"))
	rdb.SAdd(ctx, "ex:intset", 9999999999) // 升级到 int64
	fmt.Printf("  +9999999999(int64)   → %s (升级)\n", enc("ex:intset"))
	rdb.SAdd(ctx, "ex:intset", 100) // 不会降级
	fmt.Printf("  +100 (回 int16 范围)  → %s (不降级！)\n", enc("ex:intset"))
	rdb.SAdd(ctx, "ex:intset", "string")
	fmt.Printf("  +'string'  (非整数)  → %s (升级 hashtable)\n", enc("ex:intset"))

	// ============================================================
	// 练习 5: ZSet 升级
	// ============================================================
	fmt.Println("\n[练习 5] ZSet 升级：listpack → skiplist")

	rdb.Del(ctx, "ex:zset_big")
	for i := 1; i <= 200; i++ {
		rdb.ZAdd(ctx, "ex:zset_big", redis.Z{Score: float64(i), Member: fmt.Sprintf("m%d", i)})
		if i%20 == 0 {
			fmt.Printf("  元素数=%3d  encoding=%s\n", i, enc("ex:zset_big"))
		}
	}

	// ============================================================
	// 练习 6: 缩容观察
	// ============================================================
	fmt.Println("\n[练习 6] 缩容: dict 删除大量字段后会变回 listpack 吗？")
	rdb.Del(ctx, "ex:hash_shrink")
	for i := 1; i <= 200; i++ {
		rdb.HSet(ctx, "ex:hash_shrink", fmt.Sprintf("f%d", i), "v")
	}
	fmt.Printf("  200 字段               → %s (dict)\n", enc("ex:hash_shrink"))
	for i := 11; i <= 200; i++ {
		rdb.HDel(ctx, "ex:hash_shrink", fmt.Sprintf("f%d", i))
	}
	fmt.Printf("  删到剩 10 字段         → %s (注意：dict 缩容后一般不降级回 listpack)\n", enc("ex:hash_shrink"))

	// ============================================================
	// 练习 7: 内存对比（不同编码占多少字节）
	// ============================================================
	fmt.Println("\n[练习 7] 内存对比")
	rdb.Del(ctx, "ex:mem1", "ex:mem2", "ex:mem3", "ex:mem4", "ex:mem5")
	rdb.Set(ctx, "ex:mem1", 100, 0)                                                      // int
	rdb.Set(ctx, "ex:mem2", "hi", 0)                                                     // embstr
	rdb.Set(ctx, "ex:mem3", strings.Repeat("a", 200), 0)                                 // raw
	rdb.HSet(ctx, "ex:mem4", "f1", "v1")                                                 // listpack
	rdb.HSet(ctx, "ex:mem5", map[string]interface{}{"f1": "v1", "f2": "v2", "f3": "v3"}) // listpack
	for _, k := range []string{"ex:mem1", "ex:mem2", "ex:mem3", "ex:mem4", "ex:mem5"} {
		mu, _ := rdb.MemoryUsage(ctx, k).Result()
		fmt.Printf("  %-10s encoding=%-10s MEMORY USAGE = %d 字节\n", k, enc(k), mu)
	}

	// 清理
	time.Sleep(100 * time.Millisecond)
	rdb.FlushDB(ctx)
	fmt.Println("\n=== 练习结束 ===")
}

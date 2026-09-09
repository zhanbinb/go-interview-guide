// ============================================================
// Redis 数据类型 demo (Go 代码示例)
// ============================================================
// 用法: 启动一个 Redis 服务，然后 go run demo.go
// 要求: Redis 5.0+, go-redis/v9
// 安装: go mod init demo && go get github.com/redis/go-redis/v9
// ============================================================

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	ctx := context.Background()

	// 测试连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("Redis 连接失败:", err)
		return
	}
	fmt.Println("=== Redis 数据类型 Demo (Go) ===")
	rdb.FlushDB(ctx)

	// ----- String -----
	fmt.Println("\n--- String ---")

	// 简单 SET/GET
	rdb.Set(ctx, "user:1:name", "Alice", 0)
	val, _ := rdb.Get(ctx, "user:1:name").Result()
	fmt.Println("  Get user:1:name =", val)

	// 计数器
	rdb.Set(ctx, "article:1:views", 0, 0)
	rdb.Incr(ctx, "article:1:views")       // +1
	rdb.IncrBy(ctx, "article:1:views", 10) // +10
	views, _ := rdb.Get(ctx, "article:1:views").Int()
	fmt.Println("  Views =", views) // 11

	// 分布式锁（SETNX + EXPIRE）
	lockKey := "lock:order:123"
	ok, _ := rdb.SetNX(ctx, lockKey, "owner_1", 10*time.Second).Result()
	fmt.Println("  SetNX success:", ok)
	rdb.Del(ctx, lockKey)

	// ----- Hash -----
	fmt.Println("\n--- Hash ---")

	// 用户资料
	rdb.HSet(ctx, "user:1", map[string]interface{}{
		"name":  "Alice",
		"age":   25,
		"city":  "Beijing",
		"email": "alice@example.com",
	})
	user, _ := rdb.HGetAll(ctx, "user:1").Result()
	fmt.Printf("  user:1 = %v\n", user)

	// 字段原子加减
	rdb.HIncrBy(ctx, "user:1", "age", 1)
	age, _ := rdb.HGet(ctx, "user:1", "age").Int()
	fmt.Println("  age +1 =", age)

	// ----- List（消息队列）-----
	fmt.Println("\n--- List (消息队列) ---")

	rdb.Del(ctx, "queue:msg")
	// 生产端：LPUSH 从左侧推入
	rdb.LPush(ctx, "queue:msg", "msg1", "msg2", "msg3")
	length, _ := rdb.LLen(ctx, "queue:msg").Result()
	fmt.Println("  [生产者] LPUSH 后 queue 长度:", length)

	messages, _ := rdb.LRange(ctx, "queue:msg", 0, -1).Result()
	fmt.Println("  queue 内容:", messages) // [msg3 msg2 msg1] 左侧插入顺序

	// 消费端1：RPOP 非阻塞，从右侧弹出（FIFO 队列）
	popMsg, _ := rdb.RPop(ctx, "queue:msg").Result()
	fmt.Println("  [消费者] RPOP 弹出:", popMsg) // msg1
	popMsg, _ = rdb.RPop(ctx, "queue:msg").Result()
	fmt.Println("  [消费者] RPOP 弹出:", popMsg) // msg2
	popMsg, _ = rdb.RPop(ctx, "queue:msg").Result()
	fmt.Println("  [消费者] RPOP 弹出:", popMsg) // msg3
	popMsg, err := rdb.RPop(ctx, "queue:msg").Result()
	fmt.Printf("  [消费者] 队空 RPOP: msg=%q, err=%v\n", popMsg, err)

	// 生产端：再次推入
	rdb.LPush(ctx, "queue:msg", "msgA", "msgB")

	// 消费端2：BRPOP 阻塞等待（生产环境推荐）
	// 超时 1 秒，有消息立即返回，没消息超时退出
	fmt.Println("  [消费者] BRPOP 阻塞 1s...")
	result, _ := rdb.BRPop(ctx, 1*time.Second, "queue:msg").Result()
	// result = [queue:msg, msgA] —— 第一个元素是 key，第二个是值
	fmt.Printf("  BRPOP 收到: key=%s, value=%s\n", result[0], result[1])

	result, _ = rdb.BRPop(ctx, 1*time.Second, "queue:msg").Result()
	fmt.Printf("  BRPOP 收到: key=%s, value=%s\n", result[0], result[1])

	// 队空时 BRPOP 超时返回空切片 + redis.Nil
	result, err = rdb.BRPop(ctx, 500*time.Millisecond, "queue:msg").Result()
	fmt.Printf("  BRPOP 超时: result=%v, err=%v\n", result, err)

	// ----- Set -----
	fmt.Println("\n--- Set ---")

	// 去重（自动）+ 共同好友（交集）
	rdb.SAdd(ctx, "friend:1", "Alice", "Bob", "Charlie", "David")
	rdb.SAdd(ctx, "friend:2", "Bob", "Charlie", "Eve")
	common, _ := rdb.SInter(ctx, "friend:1", "friend:2").Result()
	fmt.Println("  共同好友:", common) // [Bob Charlie]

	// 推荐关注（差集）
	recommend, _ := rdb.SDiff(ctx, "friend:1", "friend:2").Result()
	fmt.Println("  推荐关注（friend:1 独有的）:", recommend) // [Alice David]

	// ----- ZSet（排行榜）-----
	fmt.Println("\n--- ZSet (排行榜) ---")

	rdb.ZAdd(ctx, "game:rank",
		redis.Z{Score: 1500, Member: "player_A"},
		redis.Z{Score: 2300, Member: "player_B"},
		redis.Z{Score: 1800, Member: "player_C"},
	)
	// 玩家 B 加 100 分
	rdb.ZIncrBy(ctx, "game:rank", 100, "player_B")

	// Top 3
	top3, _ := rdb.ZRevRangeWithScores(ctx, "game:rank", 0, 2).Result()
	fmt.Printf("  Top 3: %v\n", top3)

	// 玩家 B 的排名（0-based）
	rank, _ := rdb.ZRevRank(ctx, "game:rank", "player_B").Result()
	fmt.Println("  player_B 排名:", rank) // 0（最高分）

	// ----- Bitmap（日活）-----
	fmt.Println("\n--- Bitmap (日活) ---")

	rdb.SetBit(ctx, "active:20260901", 100, 1)
	rdb.SetBit(ctx, "active:20260901", 200, 1)
	dau, _ := rdb.BitCount(ctx, "active:20260901", nil).Result()
	fmt.Println("  2026-09-01 DAU:", dau) // 2

	// ----- HyperLogLog（UV 估算）-----
	fmt.Println("\n--- HyperLogLog (UV 估算) ---")

	rdb.Del(ctx, "uv:20260901")
	for i := 0; i < 1000; i++ {
		rdb.PFAdd(ctx, "uv:20260901", fmt.Sprintf("user_%d", i))
	}
	uv, _ := rdb.PFCount(ctx, "uv:20260901").Result()
	fmt.Println("  UV 估算:", uv) // 接近 1000

	// ----- GEO（附近）-----
	fmt.Println("\n--- GEO (附近) ---")

	rdb.GeoAdd(ctx, "cities", &redis.GeoLocation{
		Name:      "Beijing",
		Longitude: 116.40,
		Latitude:  39.90,
	})
	rdb.GeoAdd(ctx, "cities", &redis.GeoLocation{
		Name:      "Shanghai",
		Longitude: 121.47,
		Latitude:  31.23,
	})
	dist, _ := rdb.GeoDist(ctx, "cities", "Beijing", "Shanghai", "km").Result()
	fmt.Println("  北京→上海 距离:", dist, "km")

	// ----- 类型对比 -----
	fmt.Println("\n--- 类型速查 ---")
	rdb.Set(ctx, "demo:string", "x", 0)
	rdb.HSet(ctx, "demo:hash", "f", "v")
	rdb.LPush(ctx, "demo:list", "v")
	rdb.SAdd(ctx, "demo:set", "v")
	rdb.ZAdd(ctx, "demo:zset", redis.Z{Score: 1, Member: "v"})

	fmt.Println("  string type:", rdb.Type(ctx, "demo:string").Val())
	fmt.Println("  hash   type:", rdb.Type(ctx, "demo:hash").Val())
	fmt.Println("  list   type:", rdb.Type(ctx, "demo:list").Val())
	fmt.Println("  set    type:", rdb.Type(ctx, "demo:set").Val())
	fmt.Println("  zset   type:", rdb.Type(ctx, "demo:zset").Val())

	// 清理
	rdb.FlushDB(ctx)
	fmt.Println("\n=== Demo 结束 ===")
}

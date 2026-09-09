package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	var ctx = context.Background()

	var rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

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

	//计数器
	rdb.Set(ctx, "article:1:views", 0, 0)
	rdb.Incr(ctx, "article:1:views")       //+1
	rdb.IncrBy(ctx, "article:1:views", 10) //+10
	views, _ := rdb.Get(ctx, "article:1:views").Int()
	fmt.Println("  Views =", views) // 11

	//分布式锁 (SETNX + EXPIRE)
	lockKey := "lock:order:123"
	ok, _ := rdb.SetNX(ctx, lockKey, "owner_1", 10*time.Second).Result()
	fmt.Println("  SetNX success:", ok)
	rdb.Del(ctx, lockKey)

	// ----- Hash -----
	fmt.Println("\n--- Hash ---")
	//用户资料
	rdb.HSet(ctx, "user:1", map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"city":  "New York",
		"email": "alice@example.com",
	})

	user, _ := rdb.HGetAll(ctx, "user:1").Result()
	fmt.Println("  Get user:1 =", user)

	rdb.HIncrBy(ctx, "user:1", "age", 10)
	age, _ := rdb.HGet(ctx, "user:1", "age").Int()
	fmt.Println("  Age =", age)

	// ----- List(消息队列) -----
	fmt.Println("\n--- List ---")

	// LPUSH
	rdb.Del(ctx, "queue:msg")
	rdb.LPush(ctx, "queue:msg", "msg1", "msg2", "msg3")
	fmt.Println("  LPUSH msg1,msg2,msg3")

	len, _ := rdb.LLen(ctx, "queue:msg").Result()
	fmt.Println("  queue 长度 =", len)

	msgs, _ := rdb.LRange(ctx, "queue:msg", 0, -1).Result()
	fmt.Println("  queue 消息 =", msgs)

}

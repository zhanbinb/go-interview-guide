package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// request 模拟一个 HTTP 请求
type request struct {
	id int
}

// process 模拟处理请求（耗时 100ms）
func process(ctx context.Context, r request) {
	select {
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("    处理请求 %d 完成\n", r.id)
	case <-ctx.Done():
		fmt.Printf("    请求 %d 取消: %v\n", r.id, ctx.Err())
	}
}

// DemoFanout 演示 Worker Pool + 限流的高级模式
//
// 实战场景：限流的并发处理
//   - 10 个任务，但只有 2 个 worker
//   - worker 池复用 sync.Pool 的"工作上下文"
//   - context 取消 → 所有 worker 退出
func DemoFanout() {
	fmt.Println("=== 带限流的 Pool + Pipeline ===")
	fmt.Println()

	// 实验 1：限流 (2 个 worker 处理 10 个任务)
	fmt.Println("【实验 1】限流: 10 任务 / 2 worker")
	jobs := make(chan request, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// 只开 2 个 worker (限流!)
	for w := 1; w <= 2; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for r := range jobs {
				fmt.Printf("    worker %d 拿到请求 %d\n", workerID, r.id)
				process(ctx, r)
			}
		}(w)
	}

	start := time.Now()
	for i := 1; i <= 10; i++ {
		jobs <- request{id: i}
	}
	close(jobs)
	wg.Wait()
	fmt.Printf("  10 个任务完成用时: %v (2 worker 并发, ~500ms)\n\n", time.Since(start))

	// 实验 2：sync.Pool 复用 buffer (HTTP body 场景)
	fmt.Println("【实验 2】Pool 复用 buffer")
	bufPool := &sync.Pool{
		New: func() any {
			return make([]byte, 0, 1024)
		},
	}

	getBuf := func() []byte {
		b := bufPool.Get().([]byte)
		return b[:0] // reset 但保留 capacity
	}
	putBuf := func(b []byte) {
		bufPool.Put(b)
	}

	// 模拟 10000 次请求复用 buffer
	start = time.Now()
	for i := 0; i < 10000; i++ {
		buf := getBuf()
		buf = append(buf, []byte("hello world")...)
		_ = buf
		putBuf(buf)
	}
	fmt.Printf("  10000 次 buffer 复用: %v\n", time.Since(start))
	fmt.Println()

	fmt.Println("实战组合:")
	fmt.Println("  - Worker Pool 限制并发数")
	fmt.Println("  - sync.Pool 复用 buffer/request 对象")
	fmt.Println("  - context 控制超时/取消")
	fmt.Println("  - WaitGroup 等所有 worker")
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"time"

	_ "net/http/pprof" // 注册 /debug/pprof/ 端点
)

// DemoLeak 演示 pprof 排查 goroutine 泄漏
//
// 步骤：
//  1. 运行 go run . leak（这个 demo 制造泄漏）
//  2. 另开一个终端：curl http://localhost:6060/debug/pprof/goroutine?debug=2
//  3. 看输出：会显示所有 goroutine 的栈，泄漏的 goroutine 一目了然
//  4. 或者 go tool pprof http://localhost:6060/debug/pprof/goroutine
func DemoLeak() {
	fmt.Println("=== Goroutine 泄漏演示 ===")
	fmt.Println("这个程序会持续创建泄漏的 goroutine")
	fmt.Println("另开终端跑: curl http://localhost:6060/debug/pprof/goroutine?debug=2")
	fmt.Println()

	// 模拟业务：不断接收任务
	tasks := make(chan int, 100)
	var wg sync.WaitGroup

	// 启动 3 个 worker
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for task := range tasks {
				_ = process(task, id)
			}
		}(i)
	}

	// 持续投任务
	go func() {
		for i := 0; ; i++ {
			tasks <- i
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// 持续打印 goroutine 数量
	go func() {
		for {
			fmt.Printf("  [监控] goroutines = %d\\n", runtime.NumGoroutine())
			time.Sleep(2 * time.Second)
		}
	}()

	// 主线程：启动 pprof
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		http.ListenAndServe(":6060", mux)
	}()

	// 关键 BUG：process 函数忘记 return，context 取消也不退出
	// 实际场景：处理 HTTP 请求时 ctx 取消，goroutine 还在跑
	select {}
}

// process 模拟"忘记退出"的 goroutine
// 实际代码场景：HTTP handler 启动 goroutine 处理，但请求结束后没通知退出
func process(task, workerID int) error {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel // BUG: 忘记 defer cancel()
	_ = ctx

	// 永远在等某事件（永远不会发生）
	ch := make(chan struct{})
	<-ch
	_ = ch
	return nil
}

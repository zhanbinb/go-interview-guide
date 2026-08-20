package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	// var wg sync.WaitGroup
	// for i := 0; i < 3; i++ {
	// 	wg.Add(1)
	// 	go func(i int) { // ← i 作为函数参数
	// 		defer wg.Done()
	// 		fmt.Printf("  goroutine 看到 i = %d\n", i)
	// 	}(i) // ← 把外层 i 的值 复制 进去
	// }
	// wg.Wait()
	// for i := 0; i < 3; i++ {
	// 	go func() { fmt.Println(i) }() // 3 个 goroutine 都捕获"同一个 i"
	// }
	// time.Sleep(100 * time.Millisecond)

	// var wg sync.WaitGroup
	// fmt.Println("场景 2：显式传参 go func(i int) {...}(i) （仍然正确，向后兼容）")
	// for i := 0; i < 3; i++ {
	// 	wg.Add(1)
	// 	go func(x int) {
	// 		defer wg.Done()
	// 		fmt.Printf("  goroutine 看到 i = %d\n", x)
	// 	}(i)
	// }
	// wg.Wait()
	// fmt.Println()

	// items := []string{"apple", "banana", "orange"}
	// for _, item := range items {
	// 	wg.Add(1)
	// 	go func(x string) {
	// 		defer wg.Done()
	// 		fmt.Printf("  goroutine 看到 item = %s\n", x)
	// 	}(item)
	// }
	// wg.Wait()
	// fmt.Println()

	//demoChannelBlocking()
	//demoMutexLock()
	demoLeak()
}

// demoChannelBlocking 演示阻塞 goroutine
func demoChannelBlocking() {
	ch := make(chan int)
	start := time.Now()
	go func() {
		time.Sleep(300 * time.Millisecond)
		ch <- 42
	}()
	v := <-ch
	fmt.Printf(">> 从 goroutine 收到 %d\n", v)
	fmt.Printf(">> 耗时: %v\n", time.Since(start))
}
func demoMutexLock() {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		time.Sleep(300 * time.Millisecond)
		defer mu.Unlock()
	}()

	start := time.Now()
	mu.Lock()
	mu.Unlock()
	fmt.Printf("  Mutex 已释放 (等待 %v)\n\n", time.Since(start))
}

func demoLeak() {
	fmt.Println("=== Goroutine 泄漏演示 ===")
	fmt.Println()

	report := func(label string) {
		fmt.Printf("  %s: NumGoroutine = %d\n", label, runtime.NumGoroutine())
	}
	report("初始")
	// ---------- 泄漏 1：channel 没人接收 ----------
	fmt.Println("\n【泄漏 1】向无缓冲 channel 发送数据，但没人接收")
	leakCh := make(chan int)
	for i := 0; i < 3; i++ {
		go func(id int) {
			fmt.Printf("  goroutine %d 准备发送\n", id)
			leakCh <- id // 永久阻塞
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 1 后")
	fmt.Println("\n【泄漏 2】context 没 cancel，子 goroutine 等 ctx.Done()")
	for i := 0; i < 3; i++ {
		go func(id int) {
			ctx, _ := context.WithCancel(context.Background())
			_ = ctx
			<-make(chan struct{}) // 永久阻塞
			_ = id
		}(i)
	}
	time.Sleep(100 * time.Millisecond)
	report("泄漏 2 后")

	// ---------- 泄漏 3：WaitGroup Add 没 Done ----------
	fmt.Println("\n【泄漏 3】WaitGroup.Add(1) 但 goroutine 永远不 Done")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			// 永远不会调用 wg.Done()
			select {} // 死循环
			_ = id
		}(i)
	}
	// 注意：这里故意不调用 wg.Wait()
	_ = wg
	time.Sleep(100 * time.Millisecond)
	report("泄漏 3 后")
}

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DemoPatterns 演示 5 种 channel 经典模式
//
// ============================================================================
// 1. Producer-Consumer: 一生产者一消费者
// 2. Fan-out:           一生产者多消费者（负载分摊）
// 3. Fan-in:            多生产者一消费者（结果汇总）
// 4. Pipeline:          阶段式数据流
// 5. Quit signal:       done channel 协调退出
// ============================================================================
func DemoPatterns() {
	fmt.Println("=== Channel 5 种经典模式 ===")
	fmt.Println()

	// ---------- 1. Producer-Consumer ----------
	fmt.Println("【模式 1】Producer-Consumer")
	jobs := make(chan int, 5)
	done := make(chan struct{})

	go producer("P1", jobs, 5)
	consumer("C1", jobs, done)
	<-done
	fmt.Println()

	// ---------- 2. Fan-out ----------
	fmt.Println("【模式 2】Fan-out (1 producer → 3 consumers)")
	jobs = make(chan int, 10)
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("    [worker-%d] 处理 job=%d\n", id, job)
				time.Sleep(20 * time.Millisecond)
			}
		}(w)
	}
	go func() {
		for i := 1; i <= 6; i++ {
			jobs <- i
		}
		close(jobs)
	}()
	wg.Wait()
	fmt.Println()

	// ---------- 3. Fan-in ----------
	fmt.Println("【模式 3】Fan-in (3 producers → 1 consumer)")
	in := make(chan int)
	var pwg sync.WaitGroup
	for p := 1; p <= 3; p++ {
		pwg.Add(1)
		go func(id int) {
			defer pwg.Done()
			for i := 0; i < 3; i++ {
				in <- id*10 + i
			}
		}(p)
	}
	go func() {
		pwg.Wait()
		close(in)
	}()
	var results []int
	for v := range in {
		results = append(results, v)
	}
	fmt.Printf("    汇总结果: %v\n\n", results)

	// ---------- 4. Pipeline ----------
	fmt.Println("【模式 4】Pipeline (3 阶段)")
	stage1 := make(chan int)
	stage2 := make(chan int)
	stage3 := make(chan int)

	// stage 1: 生成 1..5
	go func() {
		defer close(stage1)
		for i := 1; i <= 5; i++ {
			stage1 <- i
		}
	}()

	// stage 2: ×10
	go func() {
		defer close(stage2)
		for v := range stage1 {
			stage2 <- v * 10
		}
	}()

	// stage 3: +1
	go func() {
		defer close(stage3)
		for v := range stage2 {
			stage3 <- v + 1
		}
	}()

	// consumer
	for v := range stage3 {
		fmt.Printf("    pipeline 输出: %d\n", v)
	}
	fmt.Println()

	// ---------- 5. Quit signal ----------
	fmt.Println("【模式 5】Quit signal (done channel 协调退出)")
	quit := make(chan struct{})
	work := make(chan int)
	go func() {
		for {
			select {
			case <-quit:
				fmt.Println("    [worker] 收到 quit，退出")
				return
			case j := <-work:
				fmt.Printf("    [worker] 处理 job=%d\n", j)
			}
		}
	}()
	work <- 1
	work <- 2
	close(quit) // 通知退出
	time.Sleep(50 * time.Millisecond)
	fmt.Println()

	// ---------- 6. context 超时取消 ----------
	fmt.Println("【模式 6】context 超时取消（更优雅的 quit）")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done2 := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("    [worker] ctx 取消: %v\n", ctx.Err())
				close(done2)
				return
			case <-time.After(50 * time.Millisecond):
				fmt.Println("    [worker] 工作...")
			}
		}
	}()
	<-done2
	fmt.Println()

	fmt.Println("📌 模式选择速查:")
	fmt.Println("   - 简单一对多:    Producer-Consumer")
	fmt.Println("   - CPU 密集分摊: Fan-out")
	fmt.Println("   - 结果汇总:     Fan-in")
	fmt.Println("   - 阶段式数据流: Pipeline")
	fmt.Println("   - 退出协调:     done channel / context")
}

// producer 生成 jobs 个任务
func producer(name string, jobs chan<- int, n int) {
	for i := 1; i <= n; i++ {
		fmt.Printf("    [%s] 生成 job=%d\n", name, i)
		jobs <- i
	}
	close(jobs) // 关键：关闭让 consumer range 退出
}

// consumer 消费 jobs 直到 channel 关闭
func consumer(name string, jobs <-chan int, done chan<- struct{}) {
	for job := range jobs {
		fmt.Printf("    [%s] 消费 job=%d\n", name, job)
	}
	close(done)
}

package main

import (
	"fmt"
	"sync"
	"time"
)

// Job 任务定义
type Job struct {
	ID   int
	Data string
}

// DemoWorkerPool 演示 Worker Pool 模式
//
// 模式：固定数量 worker 处理任务
//   - 任务通过 channel 传递
//   - worker 数量通常 = CPU 核数
//   - WaitGroup 等所有 worker 完成
//   - 生产推荐: github.com/panjf2000/ants
func DemoWorkerPool() {
	fmt.Println("=== Worker Pool 演示 ===")
	fmt.Println()

	const (
		workerCount = 3
		jobCount    = 10
	)

	// 实验 1：基本实现
	fmt.Println("【实验 1】基本 Worker Pool")
	jobs := make(chan Job, jobCount)
	var wg sync.WaitGroup

	// 启动 worker
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("    worker %d 处理 job %d (%s)\n",
					workerID, job.ID, job.Data)
				time.Sleep(50 * time.Millisecond) // 模拟工作
			}
		}(w)
	}

	// 派发任务
	for i := 1; i <= jobCount; i++ {
		jobs <- Job{ID: i, Data: fmt.Sprintf("task-%d", i)}
	}
	close(jobs) // 关键：关闭让 worker range 退出

	wg.Wait()
	fmt.Printf("  完成 %d 个任务，%d 个 worker\n\n", jobCount, workerCount)

	// 实验 2：Worker Pool + 超时控制
	fmt.Println("【实验 2】Worker Pool + Context 超时")
	jobs2 := make(chan Job, jobCount)
	done := make(chan struct{})

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return // 收到全局取消
				case job, ok := <-jobs2:
					if !ok {
						return // channel 关闭
					}
					fmt.Printf("    worker %d 处理 job %d\n", workerID, job.ID)
				}
			}
		}(w)
	}

	for i := 1; i <= 3; i++ {
		jobs2 <- Job{ID: i}
	}
	time.Sleep(100 * time.Millisecond)
	close(done) // 通知所有 worker 退出

	wg.Wait()
	fmt.Println("  通过 close(done) 通知 worker 退出")
	fmt.Println()

	fmt.Println("池中的对象随时可能被 GC 回收，Worker Pool 不存在此问题")
	fmt.Println("Worker Pool 适用: 任何需要\"并发处理任务\"的场景")
}

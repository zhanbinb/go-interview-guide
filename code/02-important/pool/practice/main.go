package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	DemoWorerPool()
}

func DemoWorerPool() {
	fmt.Println("=== Worker Pool ===")
	fmt.Println()

	const (
		workerCount = 3
		jobCount    = 10
	)

	type Job struct {
		ID   int
		Data string
	}

	// 实验 1：基本实现
	fmt.Println("【实验 1】基本 Worker Pool")
	jobs := make(chan Job, jobCount)

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("Worker %d processed job %d\n", workerId, job.ID)
				time.Sleep(time.Millisecond * 100)
				fmt.Printf("Worker %d processed job %d done\n", workerId, job.ID)
			}
		}(w)
	}
	for j := 0; j < jobCount; j++ {
		jobs <- Job{ID: j, Data: fmt.Sprintf("Job %d", j)}
	}
	close(jobs)
	wg.Wait()

	fmt.Println("All jobs processed")
	fmt.Println()

}

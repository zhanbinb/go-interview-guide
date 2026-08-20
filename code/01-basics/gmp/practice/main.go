package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	//go hello()
	goMaxProcs()
}

func goMaxProcs() {
	fmt.Printf("默认 P 数量 (GOMAXPROCS):    %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("当前 goroutine 数:            %d\n", runtime.NumGoroutine())
	fmt.Println()

	// 修改 GOMAXPROCS
	prev := runtime.GOMAXPROCS(2)
	fmt.Printf(">> runtime.GOMAXPROCS(2) 返回旧值: %d\n", prev)
	fmt.Printf(">> 当前 GOMAXPROCS: %d（注意 P 数量变少了）\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	for i := range 8 {
		go func(id int) {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("goroutine %d\n", id)
		}(i)
	}
	time.Sleep(50 * time.Millisecond) // 让 goroutine 进入 sleep
	fmt.Printf(">> 此时 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println(">> 用 GODEBUG=schedtrace=100 跑本 demo，可以观察到 threads > gomaxprocs")
	fmt.Println()
}

func hello() {
	fmt.Println("hello world")
}

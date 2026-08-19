package main

import (
	"fmt"
	"runtime"
	"time"
)

// DemoGOMAXPROCS 演示 GOMAXPROCS 与 NumCPU 的关系
//
// 关键结论：
//   - NumCPU    = OS 报告的逻辑 CPU 数（硬件决定，不可改）
//   - GOMAXPROCS = P 的数量（默认 = NumCPU，可调用 runtime.GOMAXPROCS(n) 修改）
//   - M 的数量  ≥ GOMAXPROCS（syscall 时会临时创建更多 M）
//
// 面试问法："G、P、M 的个数问题？"
func DemoGOMAXPROCS() {
	fmt.Printf("硬件 CPU 核心数 (NumCPU):    %d\n", runtime.NumCPU())
	fmt.Printf("默认 P 数量 (GOMAXPROCS):    %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("当前 goroutine 数:            %d\n", runtime.NumGoroutine())
	fmt.Println()

	// 修改 GOMAXPROCS
	prev := runtime.GOMAXPROCS(2)
	fmt.Printf(">> runtime.GOMAXPROCS(2) 返回旧值: %d\n", prev)
	fmt.Printf(">> 当前 GOMAXPROCS: %d（注意 P 数量变少了）\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 验证：M 的数量可以 > GOMAXPROCS（通过让 goroutine 进入 syscall）
	fmt.Println(">> 创建 5 个 syscall 中的 goroutine（time.Sleep）...")
	for i := 0; i < 5; i++ {
		go func(id int) {
			time.Sleep(500 * time.Millisecond) // 进入 syscall
			_ = id
		}(i)
	}

	time.Sleep(50 * time.Millisecond) // 让 goroutine 进入 sleep
	fmt.Printf(">> 此时 goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Println(">> 用 GODEBUG=schedtrace=100 跑本 demo，可以观察到 threads > gomaxprocs")
	fmt.Println()

	time.Sleep(600 * time.Millisecond) // 等待 sleep 完成
	fmt.Println("✅ 完成。提示：再跑一遍时把 GOMAXPROCS 改大/改小，观察 P 数变化。")
}

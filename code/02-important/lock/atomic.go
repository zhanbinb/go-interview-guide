package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// DemoAtomic 演示 atomic 包
//
// ============================================================================
// atomic 包提供的原子操作：
//
//   Add / Sub:        原子加减（int32/64, uint32/64）
//   Load / Store:     原子读/写
//   CAS (CompAndSwap):乐观锁核心
//   Swap:             原子交换
//
// 底层：CPU 原子指令（如 x86 的 LOCK CMPXCHG）
// 用途：替代简单类型的 Mutex，性能更好
// ============================================================================
func DemoAtomic() {
	fmt.Println("=== atomic 包 ===")
	fmt.Println()

	// 实验 1：原子加法
	fmt.Println("【实验 1】atomic.AddInt64 - 无锁计数器")
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()
	fmt.Printf("  1000 次 AddInt64 后: %d\n\n", counter)

	// 实验 2：CAS (CompareAndSwap)
	fmt.Println("【实验 2】CAS - 乐观锁")
	var value int32 = 100
	// 原子地将 value 从 100 改成 200
	if atomic.CompareAndSwapInt32(&value, 100, 200) {
		fmt.Printf("  CAS 成功: 100 → %d\n", value)
	} else {
		fmt.Printf("  CAS 失败，当前值: %d\n", value)
	}
	// 再次尝试（期望失败）
	if !atomic.CompareAndSwapInt32(&value, 100, 300) {
		fmt.Printf("  第二次 CAS 失败（值已不是 100），当前: %d\n", value)
	}
	fmt.Println()

	// 实验 3：CAS 自旋锁模式
	fmt.Println("【实验 3】CAS 实现自旋锁（适合短临界区）")
	var lock int32 // 0 = unlocked, 1 = locked
	criticalSection := func(id int) {
		// 自旋等待
		for !atomic.CompareAndSwapInt32(&lock, 0, 1) {
			// 忙等
		}
		// 临界区
		fmt.Printf("  goroutine %d 进入临界区\n", id)
		atomic.StoreInt32(&lock, 0)
	}
	for i := 0; i < 3; i++ {
		go criticalSection(i)
	}
	fmt.Println()

	// 实验 4：atomic vs mutex
	fmt.Println("【实验 4】atomic vs mutex 性能对比")
	var atomicCnt int64
	var mu sync.Mutex
	mxCnt := 0

	// 这里只是占位，性能测试看 Benchmark
	_ = atomicCnt
	_ = mu
	_ = mxCnt
	fmt.Println("  性能对比看 lock_test.go 的 BenchmarkAtomic vs BenchmarkMutex")
	fmt.Println()

	fmt.Println("📌 atomic 要点:")
	fmt.Println("   - 只支持简单类型 (int32/64, uint32/64, pointer)")
	fmt.Println("   - 底层是 CPU 原子指令，比 Mutex 快一个数量级")
	fmt.Println("   - 适合简单计数器、标志位、状态切换")
	fmt.Println("   - 复杂结构还是用 Mutex")
}

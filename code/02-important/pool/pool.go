package main

import (
	"fmt"
	"sync"
	"time"
)

// buf 模拟一个需要复用的大对象
type buf struct {
	data []byte
}

// DemoPool 演示 sync.Pool 用法
//
// sync.Pool 核心特性：
//   - Get/Put 复用对象
//   - 池中的对象随时可能被 GC 回收（不能持久化）
//   - 每个 P 独立本地池，无锁
//   - 用 New 字段提供"池空"时的初始化函数
//   - 经典场景：fmt.Sprintf 内部用 Pool 缓存 pp
func DemoPool() {
	fmt.Println("=== sync.Pool 演示 ===")
	fmt.Println()

	// 实验 1：基本用法
	fmt.Println("【实验 1】基本用法 Get/Put")
	pool := &sync.Pool{
		New: func() any {
			fmt.Println("    (New: 创建新对象)")
			return &buf{data: make([]byte, 1024)}
		},
	}

	// 第一次 Get：池空，会调用 New
	b1 := pool.Get().(*buf)
	fmt.Println("    第 1 次 Get: &buf{} (从 New 创建)")

	// Put 归还
	pool.Put(b1)
	fmt.Println("    Put 归还")

	// 第二次 Get：从池里拿
	b2 := pool.Get().(*buf)
	fmt.Printf("    第 2 次 Get: b2 == b1 ? %v (从池里拿)\n\n", b2 == b1)

	// 实验 2：性能对比（Pool vs 直接 new）
	fmt.Println("【实验 2】性能对比: Pool vs 直接 new")
	const N = 100000

	// 直接分配
	start := time.Now()
	for i := 0; i < N; i++ {
		_ = &buf{data: make([]byte, 1024)}
	}
	tDirect := time.Since(start)

	// 用 Pool
	pool2 := &sync.Pool{New: func() any { return &buf{data: make([]byte, 1024)} }}
	start = time.Now()
	for i := 0; i < N; i++ {
		b := pool2.Get().(*buf)
		pool2.Put(b)
	}
	tPool := time.Since(start)

	fmt.Printf("  直接 new %d 次: %v\n", N, tDirect)
	fmt.Printf("  Pool 复用 %d 次:  %v\n", N, tPool)
	fmt.Printf("  Pool 加速比: %.2fx\n\n", float64(tDirect)/float64(tPool))

	// 实验 3：使用规范
	fmt.Println("【实验 3】使用规范")
	fmt.Println("  1. Put 前可选 Reset（防止旧数据泄露）")
	fmt.Println("  2. Get 后类型断言: x := pool.Get().(*MyType)")
	fmt.Println("  3. New 字段必须设置（Get 时池空才用）")
	fmt.Println("  4. 不要在 Pool 里放带状态的连接/资源（会被 GC 回收）")
	fmt.Println()

	fmt.Println("池中的对象随时可能被 GC 回收，不能依赖！")
	fmt.Println("sync.Pool 适合: 临时对象（bytes.Buffer、protobuf message）")
}

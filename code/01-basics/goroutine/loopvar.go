package main

import (
	"fmt"
	"sync"
	"time"
)

// DemoLoopVar 演示 Go 1.22+ 循环变量语义变更
//
// 面试高频问题："for range 的时候它的地址会发生变化么？" / "go 关键字传参要注意什么？"
//
// ============================================================================
// 历史背景：
//   Go 1.22 之前：
//     for i := 0; i < 3; i++ {
//         go func() { fmt.Println(i) }()  // ❌ 几乎都打印 3
//     }
//   原因是：循环里只有一个 i 变量（地址不变），所有 goroutine 闭包共享。
//   解决：必须显式传参 go func(i int) { ... }(i)
//
// Go 1.22 起（你正在用的 Go 1.25）：
//   每次迭代 i 都是新变量（地址也变），闭包捕获自动正确。
//   上面的写法不再有陷阱。
//
// 面试要点：能讲清"为什么之前会出问题"+"Go 1.22 怎么改的"+"怎么验证"
// ============================================================================
func DemoLoopVar() {
	fmt.Println("=== Go 1.22+ 循环变量语义演示 ===")
	fmt.Printf("Go version: see runtime.Version()\n")
	fmt.Println()

	// ---------- 场景 1：经典 for 循环 + 闭包 ----------
	fmt.Println("场景 1：for i := 0; i < 3; i++ + 闭包（Go 1.22+ 自动正确）")
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  goroutine 看到 i = %d\n", i)
		}()
	}
	wg.Wait()
	fmt.Println()

	// ---------- 场景 2：显式传参（兼容写法） ----------
	fmt.Println("场景 2：显式传参 go func(i int) {...}(i) （仍然正确，向后兼容）")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("  goroutine 看到 i = %d\n", i)
		}(i)
	}
	wg.Wait()
	fmt.Println()

	// ---------- 场景 3：range over slice ----------
	fmt.Println("场景 3：for _, v := range slice（Go 1.22+ 自动正确）")
	items := []string{"apple", "banana", "cherry"}
	for _, v := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("  goroutine 看到 v = %s\n", v)
		}()
	}
	wg.Wait()
	fmt.Println()

	// ---------- 场景 4：对比演示 - 用 time.Sleep 模拟老版本的"竞争窗口" ----------
	fmt.Println("场景 4：竞争窗口演示（Go 1.22+ 即使 sleep 后读也是正确的）")
	fmt.Println("    即使闭包内部 time.Sleep(1ms) 让 goroutine 排队执行，")
	fmt.Println("    Go 1.22+ 也会读到自己那一轮的 i，而不是最后一轮的。")
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			fmt.Printf("  goroutine 醒来后看到 i = %d\n", i)
		}()
	}
	wg.Wait()
	fmt.Println()

	fmt.Println("✅ 全部输出 0/1/2（而不是 3/3/3）说明 Go 1.22+ 循环变量语义生效。")
	fmt.Println()
	fmt.Println("📌 面试时强调：")
	fmt.Println("   - Go 1.22 之前要传参（坑爹但安全）")
	fmt.Println("   - Go 1.22 之后会自动修复（更直观，但要注意老教程/老代码可能误导）")
	fmt.Println("   - 即使在新版本，显式传参依然是推荐写法（更清晰）")
}

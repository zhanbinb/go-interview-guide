package main

import "fmt"

// DemoTriColor 演示三色标记示意（概念层面）
//
// ============================================================================
// 三色标记是 Go GC 的核心算法：
//   白色：未扫描过（垃圾的候选）
//   灰色：已发现，但子引用未扫描
//   黑色：已扫描完，子引用都处理过
//
// 流程：
//   1. 初始：所有对象白色
//   2. 从根对象（栈、全局变量）开始，标记为灰色
//   3. 反复：取一个灰色对象 → 标黑 → 它的子引用标灰
//   4. 结束：剩下的白色对象 = 垃圾，回收
//
// 为什么需要三色（而不是两色或四色）：
//   - 两色（黑/白）无法解决并发标记问题（GC 和用户程序同时跑）
//   - 三色"灰色"提供了"已发现但未处理完"的中间状态，是并发标记的关键
// ============================================================================
func DemoTriColor() {
	fmt.Println("=== 三色标记示意 ===")
	fmt.Println()
	fmt.Println("三色含义:")
	fmt.Println("  ⚪ 白 (white): 未扫描（垃圾候选）")
	fmt.Println("  🔘 灰 (gray):  已发现，子引用未扫描")
	fmt.Println("  ⚫ 黑 (black): 已扫描，子引用处理完")
	fmt.Println()

	fmt.Println("标记流程（简化）:")
	fmt.Println()
	fmt.Println("  初始状态：所有对象都是白的")
	fmt.Println("    ⚪ A    ⚪ B    ⚪ C    ⚪ D    ⚪ E")
	fmt.Println()
	fmt.Println("  第 1 步：从根出发，标灰")
	fmt.Println("    🔘 A    ⚪ B    ⚪ C    ⚪ D    ⚪ E   (根 → A)")
	fmt.Println()
	fmt.Println("  第 2 步：扫描 A，标黑；A 的子引用标灰")
	fmt.Println("    ⚫ A    🔘 B    🔘 C    ⚪ D    ⚪ E   (A→B, A→C)")
	fmt.Println()
	fmt.Println("  第 3 步：扫描 B，标黑")
	fmt.Println("    ⚫ A    ⚫ B    🔘 C    ⚪ D    ⚪ E")
	fmt.Println()
	fmt.Println("  第 4 步：扫描 C，标黑")
	fmt.Println("    ⚫ A    ⚫ B    ⚫ C    ⚪ D    ⚪ E")
	fmt.Println()
	fmt.Println("  结束：剩下的 ⚪ D、⚪ E 是垃圾，回收")
	fmt.Println()
	fmt.Println("📌 面试要点:")
	fmt.Println("   - 三色标记是并发标记算法（GC 和用户程序一起跑）")
	fmt.Println("   - 灰色是中间状态：已发现但还没扫完")
	fmt.Println("   - 不变式：黑色不能指向白色（写屏障就是为了维持这个）")
}

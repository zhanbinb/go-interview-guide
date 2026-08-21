package main

import "fmt"

// DemoBarrier 演示为什么需要写屏障
//
// ============================================================================
// 问题：GC 和用户程序并发执行，会出问题
//
// 场景（假设 GC 标记 A，B 从 A 引用 D）：
//   - GC 已标记 A 为黑（已扫描）
//   - GC 已扫描 B 标黑
//   - 用户程序：D = nil  （把 D 从 B 的引用断开）
//   - 然后：A.D = newD  （让 A 指向一个新对象）
//   - 结果：newD 没被任何黑色对象引用，GC 看不到，会被错误回收！
//
// 这就是"漏标"问题，写屏障就是解决它的。
//
// 写屏障 = 在用户程序读写对象时插入一段代码，通知 GC
// ============================================================================
func DemoBarrier() {
	fmt.Println("=== 写屏障为什么需要 ===")
	fmt.Println()

	fmt.Println("【场景】漏标问题（GC 和用户程序并发）:")
	fmt.Println()
	fmt.Println("  初始状态:")
	fmt.Println("    ⚫ A (已扫)")
	fmt.Println("      ├── D (⚪ 未扫)")
	fmt.Println("    ⚫ B (已扫)")
	fmt.Println("      └── D (⚪ 未扫)")
	fmt.Println()
	fmt.Println("  用户程序:")
	fmt.Println("    B.D = nil       // D 不再被 B 引用")
	fmt.Println("    A.D = newD      // A 现在指向 newD")
	fmt.Println()
	fmt.Println("  GC 视角:")
	fmt.Println("    A 是黑的，不再扫描")
	fmt.Println("    B 是黑的，不再扫描")
	fmt.Println("    newD 没被任何黑色对象引用 → 被当作垃圾回收 ❌")
	fmt.Println("    （但 A.D 还在引用 newD → 用户程序 panic）")
	fmt.Println()

	fmt.Println("【解决】写屏障（Write Barrier）:")
	fmt.Println()
	fmt.Println("  插入屏障（Dijkstra）: 被引用时立即标灰")
	fmt.Println("    A.D = newD 时，触发写屏障 → newD 立即标灰")
	fmt.Println("    GC 后续会扫描灰色 newD，不会漏 ✅")
	fmt.Println()
	fmt.Println("  删除屏障（Yuasa）: 引用断开时标灰")
	fmt.Println("    B.D = nil 时，触发写屏障 → 旧 D 标灰")
	fmt.Println("    GC 后续会扫描灰色旧 D，不会漏 ✅")
	fmt.Println()
	fmt.Println("  Go 1.8+ 混合写屏障：两种都用，最大化保险")
	fmt.Println()

	fmt.Println("📌 面试要点:")
	fmt.Println("   - 写屏障是 GC 和用户程序并发执行的桥梁")
	fmt.Println("   - 每次写操作都额外付出一点代价（Go 1.8+ 做了大量优化）")
	fmt.Println("   - Go 1.8 用混合屏障后，STW 几乎可以忽略")
}

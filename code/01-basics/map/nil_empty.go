package main

import "fmt"

// DemoNilEmpty 演示 nil map 和空 map 的区别
//
// 关键差异：
//   - nil map (var m map[K]V): 不能写（panic），可以读（返回零值）
//   - 空 map (make/m:= map{}): 可读可写
//
// 共同点：都可以 len() 和 delete()
func DemoNilEmpty() {
	fmt.Println("=== nil map vs 空 map ===")
	fmt.Println()

	// 场景 1：nil map 的行为
	fmt.Println("【场景 1】nil map: var m map[string]int")
	var nilMap map[string]int
	fmt.Printf("  nilMap == nil: %v, len=%d\n", nilMap == nil, len(nilMap))

	// 读 nil map：OK，返回零值
	v, ok := nilMap["key"]
	fmt.Printf("  v, ok := nilMap[\"key\"]: v=%d, ok=%v (OK, 不 panic)\n", v, ok)

	// len(nilMap): OK
	fmt.Printf("  len(nilMap): %d (OK)\n", len(nilMap))

	// delete(nilMap, "key"): OK（no-op）
	delete(nilMap, "key")
	fmt.Printf("  delete(nilMap, \"key\"): OK (no-op)\n\n")

	// 场景 2：写 nil map 会 panic
	fmt.Println("【场景 2】写 nil map 会 panic")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  nilMap[\"k\"] = 1: 💥 PANIC: %v\n", r)
			}
		}()
		nilMap["k"] = 1
	}()
	fmt.Println()

	// 场景 3：空 map (make 后)
	fmt.Println("【场景 3】空 map: make(map[string]int)")
	emptyMap := make(map[string]int)
	fmt.Printf("  emptyMap == nil: %v, len=%d\n", emptyMap == nil, len(emptyMap))
	emptyMap["k"] = 100
	fmt.Printf("  emptyMap[\"k\"] = 100 后: %v, len=%d\n", emptyMap, len(emptyMap))
	fmt.Println()

	// 场景 4：空字面量 map
	fmt.Println("【场景 4】空字面量 map: m := map[string]int{}")
	literalMap := map[string]int{}
	fmt.Printf("  literalMap == nil: %v\n", literalMap == nil)
	literalMap["k"] = 200
	fmt.Printf("  literalMap[\"k\"] = 200 后: %v\n", literalMap)
	fmt.Println()

	// 场景 5：JSON 反序列化常用 nil map
	fmt.Println("【场景 5】JSON 反序列化时，nil field vs 空 map")
	type Config struct {
		Tags map[string]string `json:"tags"` // nil → JSON null
	}
	c1 := Config{}
	fmt.Printf("  c1.Tags == nil: %v (JSON 输出 null)\n", c1.Tags == nil)

	// safeWrite: 防止 nil map panic
	fmt.Println("\n【场景 6】safe write 模式")
	safeWrite := func(m map[string]int, k string, v int) {
		if m == nil {
			m = make(map[string]int) // 但这改了局部副本，外面还是 nil
		}
		m[k] = v
	}
	var maybeNil map[string]int
	safeWrite(maybeNil, "x", 1)
	fmt.Printf("  maybeNil 仍是 nil: %v (因为改了局部)\n", maybeNil == nil)
	fmt.Println("  → 正确做法: 在外面初始化或返回新 map")
	fmt.Println()

	fmt.Println("📌 关键差异表:")
	fmt.Println("  ┌─────────────┬──────┬──────┐")
	fmt.Println("  │   操作      │ nil  │ empty │")
	fmt.Println("  ├─────────────┼──────┼──────┤")
	fmt.Println("  │ 读          │  ✅  │  ✅  │")
	fmt.Println("  │ 写          │  ❌  │  ✅  │")
	fmt.Println("  │ delete      │  ✅  │  ✅  │")
	fmt.Println("  │ len()       │  ✅  │  ✅  │")
	fmt.Println("  │ range       │  ✅  │  ✅  │")
	fmt.Println("  └─────────────┴──────┴──────┘")
}

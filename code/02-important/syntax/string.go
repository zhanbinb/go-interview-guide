package main

import "fmt"

// DemoString 演示三种引号的区别
func DemoString() {
	fmt.Println("=== 单/双/反引号区别 ===")
	fmt.Println()

	// 单引号：rune 字面量（int32）
	fmt.Println("【单引号】rune 字面量（int32）")
	var c rune = 'A'
	fmt.Printf("  c = 'A'，类型: %T, 值: %d (65)\\n", c, c)
	// 中文字符是 rune 类型（4 字节）
	var cn rune = '中'
	fmt.Printf("  cn = '中'，类型: %T, 值: %d (Unicode 码点)\\n\\n", cn, cn)

	// 双引号：普通 string（处理转义）
	fmt.Println("【双引号】普通 string（处理转义）")
	s1 := "hello\\nworld" // \n 会被解析为换行符
	fmt.Printf("  \"hello\\\\nworld\" = %q (含换行符)\\n", s1)
	s2 := "中文测试"
	fmt.Printf("  \"中文测试\" = %q (UTF-8 编码，每个汉字 3 字节)\\n\\n", s2)

	// 反引号：raw string（不处理转义）
	fmt.Println("【反引号】raw string（不处理转义）")
	s3 := `hello\nworld` // \n 是字面两个字符
	fmt.Printf("  `hello\\\\nworld` = %q (\\\\n 是两个字符)\\n", s3)
	s4 := `
多行
字符串`
	fmt.Printf("  反引号支持直接换行:\\n%s\\n", s4)

	fmt.Println("📌 选择:")
	fmt.Println("   - 'a'   : 单字符 rune (Unicode 码点)")
	fmt.Println("   - \"abc\": 字符串（解析转义）")
	fmt.Println("   - `abc` : 多行/不转义字符串")
	fmt.Println("   - 字符串底层是 UTF-8 编码的 []byte")
}

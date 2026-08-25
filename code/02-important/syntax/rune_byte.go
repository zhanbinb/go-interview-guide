package main

import (
	"fmt"
	"unicode/utf8"
)

// DemoRuneByte 演示 rune vs byte + uint 溢出
func DemoRuneByte() {
	fmt.Println("=== rune/byte + uint 溢出 ===")
	fmt.Println()

	// byte vs rune
	fmt.Println("【实验 1】byte vs rune")
	var b byte = 'A'        // byte = uint8
	var r rune = '中'        // rune = int32
	fmt.Printf("  byte = 'A':   类型=%T, 大小=%d 字节, 值=%d\\n", b, 1, b)
	fmt.Printf("  rune = '中':  类型=%T, 大小=%d 字节, 值=%d (Unicode 码点)\\n", r, 4, r)
	fmt.Println()

	// rune 处理字符串
	fmt.Println("【实验 2】rune 处理字符串（中英文混合）")
	s := "Hello, 世界!"
	fmt.Printf("  s = %q, len=%d (字节数)\\n", s, len(s))
	runes := []rune(s)
	fmt.Printf("  []rune(s) 长度=%d (字符数)\\n", len(runes))
	fmt.Printf("  utf8.RuneCountInString = %d\\n", utf8.RuneCountInString(s))
	fmt.Println()

	// uint 溢出
	fmt.Println("【实验 3】uint 溢出（回卷）")
	var u8 uint8 = 255
	fmt.Printf("  uint8 = 255, 加 1 后: %d (回卷到 0!)\\n", u8+1)

	// 安全的判断
	fmt.Println("\\n  边界判断的正确写法:")
	var counter int = 200
	if counter >= 200 {
		fmt.Println("    counter 已到上限")
	}
	fmt.Println()

	fmt.Println("📌 关键:")
	fmt.Println("   - byte 处理 ASCII / UTF-8 字节流")
	fmt.Println("   - rune 处理 Unicode 码点（遍历中文字符串用 []rune）")
	fmt.Println("   - uint 溢出是 wrap，不是 panic（无错误处理）")
	fmt.Println("   - 边界判断显式写，不要依赖 uint wrap")
}

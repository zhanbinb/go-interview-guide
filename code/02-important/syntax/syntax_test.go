package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestRuneVsByte 验证 rune 和 byte 大小
func TestRuneVsByte(t *testing.T) {
	var b byte = 'A'
	var r rune = 'A'
	if b != byte(r) {
		t.Errorf("byte and rune should be equal for ASCII: b=%d r=%d", b, r)
	}
	// rune 实际是 int32
	var rn rune = '中'
	if rn <= 0 {
		t.Error("rune for 中 should be positive")
	}
}

// TestStringLen 验证字符串字节数 vs 字符数
func TestStringLen(t *testing.T) {
	s := "Hello"
	if len(s) != 5 {
		t.Errorf("expected len=5, got %d", len(s))
	}
	// 中文字符在 UTF-8 是 3 字节
	cn := "中"
	if len(cn) != 3 {
		t.Errorf("expected len=3 for 中, got %d", len(cn))
	}
	if utf8.RuneCountInString(cn) != 1 {
		t.Errorf("expected 1 rune, got %d", utf8.RuneCountInString(cn))
	}
}

// TestRawString 验证反引号不处理转义
func TestRawString(t *testing.T) {
	raw := `hello\nworld`
	if !strings.Contains(raw, `\n`) {
		t.Errorf("raw string should contain literal \\n, got %s", raw)
	}
	if strings.Contains(raw, "\n") {
		t.Error("raw string should not contain actual newline")
	}
}

// TestUintOverflow 验证 uint 溢出是 wrap
func TestUintOverflow(t *testing.T) {
	var u uint8 = 255
	// wrap 到 0（不是 panic）
	if u+1 != 0 {
		t.Errorf("expected uint8(255)+1 == 0, got %d", u+1)
	}
}

// TestPassByValue 验证传值不修改外部
func TestPassByValue(t *testing.T) {
	type S struct{ N int }
	s := S{N: 10}
	modify := func(s S) {
		s.N = 20
	}
	modify(s)
	if s.N != 10 {
		t.Errorf("pass by value should not modify external: %d", s.N)
	}
}

// TestPassByPointer 验证传指针修改外部
func TestPassByPointer(t *testing.T) {
	type S struct{ N int }
	s := S{N: 10}
	modify := func(s *S) {
		s.N = 20
	}
	modify(&s)
	if s.N != 20 {
		t.Errorf("pass by pointer should modify external: %d", s.N)
	}
}

// TestSliceSharedReference 验证 slice 传引用语义
func TestSliceSharedReference(t *testing.T) {
	s := []int{1, 2, 3}
	modify := func(s []int) {
		s[0] = 999
	}
	modify(s)
	if s[0] != 999 {
		t.Errorf("slice should share underlying: %d", s[0])
	}
}

// TestStructVisibility 验证大小写可见性
func TestStructVisibility(t *testing.T) {
	// LowercaseName 不可导出
	type lower struct {
		name string // lowercase 字段在包外不可见
	}
	l := lower{name: "test"}
	_ = l
	// 编译期检查：跨包访问 l.name 会失败
}

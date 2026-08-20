package main

import (
	"reflect"
	"testing"
	"unsafe"
)

// TestSliceStruct 验证 slice 三元组
func TestSliceStruct(t *testing.T) {
	s := make([]int, 3, 5)
	if len(s) != 3 {
		t.Errorf("expected len=3, got %d", len(s))
	}
	if cap(s) != 5 {
		t.Errorf("expected cap=5, got %d", cap(s))
	}
	// unsafe.Sizeof 应该是 24（ptr + len + cap = 8+8+8）
	if got := unsafe.Sizeof(s); got != 24 {
		t.Errorf("expected sizeof=24, got %d", got)
	}
}

// TestAppendNoExpand 验证 cap 够时不扩容（ptr 不变）
func TestAppendNoExpand(t *testing.T) {
	s := make([]int, 0, 3)
	s = append(s, 1) // 先 append 一个，&s[0] 才合法
	ptrBefore := &s[0]
	s = append(s, 2)
	if &s[0] != ptrBefore {
		t.Error("append should not allocate new array when cap is enough")
	}
}

// TestAppendExpand 验证 cap 不够时扩容（ptr 变）
func TestAppendExpand(t *testing.T) {
	s := make([]int, 0, 2)
	s = append(s, 1)
	ptrBefore := &s[0]
	s = append(s, 2, 3, 4) // 触发扩容
	if &s[0] == ptrBefore {
		t.Error("append should allocate new array when cap is insufficient")
	}
	if cap(s) < 4 {
		t.Errorf("new cap should be >= 4, got %d", cap(s))
	}
}

// TestShareUnderlying 验证截取 slice 共享底层数组
func TestShareUnderlying(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	s2 := s1[1:3]
	s2[0] = 999
	if s1[1] != 999 {
		t.Errorf("expected s1[1]=999, got %d (slice should share underlying array)", s1[1])
	}
}

// TestCopyIsolated 验证 copy 后两个 slice 独立
func TestCopyIsolated(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	s2 := make([]int, len(s1))
	copy(s2, s1)
	s2[0] = 999
	if s1[0] != 1 {
		t.Errorf("expected s1[0]=1, got %d (copy should isolate slices)", s1[0])
	}
}

// TestParamShare 验证函数内修改影响外部
func TestParamShare(t *testing.T) {
	s := []int{1, 2, 3}
	modifyFirst(s)
	if s[0] != 999 {
		t.Errorf("expected s[0]=999 (shared underlying), got %d", s[0])
	}
}

// TestParamNoAffect 验证函数内 append 扩容后不影响外部
func TestParamNoAffect(t *testing.T) {
	s := []int{1, 2, 3} // cap=3
	appendBeyond(s)      // 触发扩容
	if len(s) != 3 {
		t.Errorf("expected len=3 (external unchanged), got %d", len(s))
	}
}

// TestExpandRules 验证扩容规则（append 足够多元素强制扩容）
func TestExpandRules(t *testing.T) {
	tests := []struct {
		oldCap  int
		minNew  int // 期望至少 >= 这个值
	}{
		{1, 2},   // 翻倍
		{2, 4},   // 翻倍
		{4, 8},   // 翻倍
		{100, 200}, // 翻倍
		{256, 320}, // 进入阶梯式
		{512, 512}, // 阶梯式：约 1.25x
	}
	for _, tt := range tests {
		// 创建 len=cap 的 slice（满的），append 一个元素必扩容
		s := make([]int, tt.oldCap, tt.oldCap)
		s = append(s, 0)
		if cap(s) < tt.minNew {
			t.Errorf("oldCap=%d: expected newCap >= %d, got %d", tt.oldCap, tt.minNew, cap(s))
		}
	}
}

// TestNilAndEmpty 验证 nil slice 和空 slice 区别
func TestNilAndEmpty(t *testing.T) {
	var nilSlice []int
	emptySlice := []int{}
	emptySliceMade := make([]int, 0)

	if nilSlice != nil {
		t.Error("nil slice should be nil")
	}
	if emptySlice == nil {
		t.Error("empty []int{} should not be nil")
	}
	if emptySliceMade == nil {
		t.Error("make([]int, 0) should not be nil")
	}

	// nil slice 可以 append
	tmpSlice := append(nilSlice, 1)
	if len(tmpSlice) != 1 {
		t.Error("nil slice should be appendable")
	}

	// nil slice 不能直接索引（用一个新的 nil slice 测试）
	func() {
		defer func() {
			if recover() == nil {
				t.Error("indexing nil slice should panic")
			}
		}()
		var nilForIndex []int
		_ = nilForIndex[0] // 期望 panic
	}()
}

// TestSliceEqual 验证 slice 相等比较
func TestSliceEqual(t *testing.T) {
	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	// slice 不能直接 == 比较
	if reflect.DeepEqual(a, b) != true {
		t.Error("DeepEqual should be true for identical slices")
	}
	// 编译就会报错：if a == b { ... } // invalid
}

// BenchmarkAppendGrow benchmark append 触发扩容
func BenchmarkAppendGrow(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := make([]int, 0)
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
	}
}

// BenchmarkAppendPrealloc benchmark 预分配 cap 的 append
func BenchmarkAppendPrealloc(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 1000)
		for j := 0; j < 1000; j++ {
			s = append(s, j)
		}
	}
}

// helpers used by tests
func modifyFirst(s []int) {
	s[0] = 999
}

func appendBeyond(s []int) {
	s = append(s, 4, 5, 6) // cap=3，必扩容
}

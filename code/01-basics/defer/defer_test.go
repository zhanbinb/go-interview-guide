package main

import (
	"testing"
)

// TestOrder 验证 LIFO 顺序
// defer 按 LIFO 执行，所以验证 defer 必须最先注册（最后执行）
func TestOrder(t *testing.T) {
	result := make([]int, 0, 3)
	// 先注册验证 defer（最后才执行，append 已经做完）
	defer func() {
		want := []int{3, 2, 1}
		if len(result) != len(want) {
			t.Errorf("expected len=%d, got %d (result=%v)", len(want), len(result), result)
			return
		}
		for i := range want {
			if result[i] != want[i] {
				t.Errorf("at index %d: expected %d, got %d", i, want[i], result[i])
			}
		}
	}()
	// 再注册 append defers（先注册的后执行 = LIFO）
	for i := 1; i <= 3; i++ {
		defer func(n int) { result = append(result, n) }(i)
	}
}

// TestAnonymousReturn 验证匿名返回值 defer 改不了
func TestAnonymousReturn(t *testing.T) {
	f := func() int {
		ret := 1
		defer func() { ret = 999 }()
		return ret
	}
	if got := f(); got != 1 {
		t.Errorf("expected 1 (anonymous return, defer can't change), got %d", got)
	}
}

// TestNamedReturn 验证命名返回值 defer 能改
func TestNamedReturn(t *testing.T) {
	f := func() (ret int) {
		ret = 1
		defer func() { ret = 999 }()
		return
	}
	if got := f(); got != 999 {
		t.Errorf("expected 999 (named return, defer can change), got %d", got)
	}
}

// TestClosureImmediate 验证 defer 参数立即求值
func TestClosureImmediate(t *testing.T) {
	x := 1
	defer func(n int) {
		if n != 1 {
			t.Errorf("expected 1 (immediate evaluation), got %d", n)
		}
	}(x)
	x = 999
}

// TestClosureDelayed 验证闭包延迟求值
func TestClosureDelayed(t *testing.T) {
	y := 1
	defer func() {
		if y != 999 {
			t.Errorf("expected 999 (closure reads latest), got %d", y)
		}
	}()
	y = 999
}

// TestRecover 验证 defer + recover 捕获 panic
func TestRecover(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to be recovered")
		}
	}()
	panic("test panic")
}

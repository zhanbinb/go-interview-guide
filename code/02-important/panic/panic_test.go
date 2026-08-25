package main

import (
	"errors"
	"testing"
)

// TestArrayIndexOutOfRange 验证数组越界触发 panic
func TestArrayIndexOutOfRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic from array index out of range")
		}
	}()
	s := []int{1, 2, 3}
	_ = s[10]
}

// TestNilPointer 验证 nil 指针触发 panic
func TestNilPointer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic from nil pointer")
		}
	}()
	var p *int
	_ = *p
}

// TestTypeAssertionFail 验证类型断言失败触发 panic
func TestTypeAssertionFail(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic from type assertion failure")
		}
	}()
	var i interface{} = "hello"
	_ = i.(int)
}

// TestCloseClosedChannel 验证关闭已关闭 channel 触发 panic
func TestCloseClosedChannel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic from closing closed channel")
		}
	}()
	ch := make(chan int)
	close(ch)
	close(ch)
}

// TestRecoverInDefer 验证 defer 中 recover 生效
func TestRecoverInDefer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected recover to capture panic")
		}
	}()
	panic("test")
}

// TestRecoverOutsideDefer 验证不在 defer 中 recover 无效
func TestRecoverOutsideDefer(t *testing.T) {
	// 这种情况 recover 返回 nil（无效）
	r := recover()
	if r != nil {
		t.Error("recover outside defer should return nil")
	}
}

// TestConvertPanicToError 验证把 panic 转成 error
func TestConvertPanicToError(t *testing.T) {
	fn := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New("from panic")
			}
		}()
		panic("boom")
	}
	err := fn()
	if err == nil || err.Error() != "from panic" {
		t.Errorf("expected error from panic, got %v", err)
	}
}

// TestRePanic 验证 re-panic 模式
func TestRePanic(t *testing.T) {
	caught := false
	inner := func() {
		defer func() {
			r := recover()
			if r != nil {
				caught = true
				panic(r) // re-panic
			}
		}()
		panic("inner")
	}
	outer := func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Error("outer should catch re-panic")
			}
		}()
		inner()
	}
	outer()
	if !caught {
		t.Error("inner should have caught first panic")
	}
}

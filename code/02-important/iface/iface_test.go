package main

import (
	"testing"
)

// TestNilInterface 验证完全 nil 的接口
func TestNilInterface(t *testing.T) {
	var i interface{}
	if i != nil {
		t.Error("uninitialized interface should be nil")
	}
}

// TestTypedNilInterface 验证有类型但值是 nil 的接口
func TestTypedNilInterface(t *testing.T) {
	var p *int
	var i interface{} = p
	if i == nil {
		t.Error("interface with typed nil should NOT equal nil")
	}
	// 类型断言回来后是 nil
	if v, ok := i.(*int); !ok || v != nil {
		t.Errorf("typed-nil: ok=%v, v=%v", ok, v)
	}
}

// TestTypeAssertionSuccess 验证类型断言成功
func TestTypeAssertionSuccess(t *testing.T) {
	var i interface{} = "hello"
	v, ok := i.(string)
	if !ok || v != "hello" {
		t.Errorf("expected hello, got %v ok=%v", v, ok)
	}
}

// TestTypeAssertionFail 验证类型断言失败（带 ok 不 panic）
func TestTypeAssertionFail(t *testing.T) {
	var i interface{} = "hello"
	v, ok := i.(int)
	if ok {
		t.Errorf("expected ok=false, got ok=true, v=%v", v)
	}
}

// TestTypeAssertionPanic 验证不带 ok 的断言失败会 panic
func TestTypeAssertionPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic from type assertion failure")
		}
	}()
	var i interface{} = "hello"
	_ = i.(int)
}

// TestTypeSwitch 验证类型 switch
func TestTypeSwitch(t *testing.T) {
	check := func(i interface{}) string {
		switch v := i.(type) {
		case int:
			return "int"
		case string:
			_ = v
			return "string"
		default:
			return "other"
		}
	}
	if got := check(42); got != "int" {
		t.Errorf("42: expected int, got %s", got)
	}
	if got := check("hi"); got != "string" {
		t.Errorf("hi: expected string, got %s", got)
	}
	if got := check(3.14); got != "other" {
		t.Errorf("3.14: expected other, got %s", got)
	}
}


// TestPolymorphism 验证多态（复用 polymorphism.go 里定义的类型）
func TestPolymorphism(t *testing.T) {
	var s Speaker = Dog{Name: "test"}
	got := s.Speak()
	if got != "test: 汪！" {
		t.Errorf("expected test: 汪！, got %s", got)
	}
}

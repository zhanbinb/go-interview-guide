package main

import (
	"context"
	"testing"
	"time"
)

// TestBackground 验证 Background 不取消
func TestBackground(t *testing.T) {
	ctx := context.Background()
	if ctx.Err() != nil {
		t.Errorf("Background.Err() should be nil, got %v", ctx.Err())
	}
	if _, ok := ctx.Deadline(); ok {
		t.Error("Background.Deadline() should have ok=false")
	}
}

// TestWithCancel 验证 WithCancel 能取消
func TestWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ctx.Err() == nil {
		t.Error("expected ctx.Err() to be non-nil after cancel")
	}
	// Done() 应该已经关闭
	select {
	case <-ctx.Done():
		// 期望
	default:
		t.Error("expected ctx.Done() to be closed")
	}
}

// TestWithTimeout 验证超时自动取消
func TestWithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("expected timeout to fire within 50ms")
	}
}

// TestWithValue 验证 WithValue 存取
func TestWithValue(t *testing.T) {
	type myKey struct{}
	ctx := context.WithValue(context.Background(), myKey{}, "hello")
	if got := ctx.Value(myKey{}); got != "hello" {
		t.Errorf("expected hello, got %v", got)
	}
}

// TestCascadingCancel 验证级联取消
func TestCascadingCancel(t *testing.T) {
	parent, pcancel := context.WithCancel(context.Background())
	child, ccancel := context.WithCancel(parent)
	defer ccancel()

	pcancel() // 取消父 ctx

	// 子 ctx 应该在合理时间内也被取消
	select {
	case <-child.Done():
		// 期望
	case <-time.After(100 * time.Millisecond):
		t.Error("child ctx should be canceled after parent cancel")
	}
}

// TestDeadlines 验证 Deadline 可读
func TestDeadlines(t *testing.T) {
	d := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("expected Deadline to return ok=true for WithTimeout")
	}
	// deadline 应该在 100ms 之后
	if time.Until(deadline) > d {
		t.Errorf("deadline %v is more than %v from now", deadline, d)
	}
}

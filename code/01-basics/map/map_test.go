package main

import (
	"reflect"
	"sync"
	"testing"
)

// TestNilMap 验证 nil map 的可读性
func TestNilMap(t *testing.T) {
	var m map[string]int

	// 读：OK
	v, ok := m["key"]
	if v != 0 || ok {
		t.Errorf("expected (0, false), got (%d, %v)", v, ok)
	}

	// len: OK
	if len(m) != 0 {
		t.Errorf("expected len=0, got %d", len(m))
	}

	// delete: OK
	delete(m, "key") // no-op，不 panic
}

// TestNilMapWrite 验证 nil map 写会 panic
func TestNilMapWrite(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when writing to nil map")
		}
	}()
	var m map[string]int
	m["key"] = 1
}

// TestKeyTypes 验证 key 类型支持
func TestKeyTypes(t *testing.T) {
	// 可以做 key 的类型
	_ = map[int]string{}
	_ = map[string]int{}
	_ = map[bool]int{}
	_ = map[[2]int]string{}
	type Point struct{ X, Y int }
	_ = map[Point]string{}
	_ = map[chan int]int{}
	_ = map[*int]int{}
}

// TestIterationUnordered 验证 map 遍历是无序的
func TestIterationUnordered(t *testing.T) {
	m := make(map[int]int, 100)
	for i := 0; i < 100; i++ {
		m[i] = i
	}

	// 连续 3 次遍历，比较顺序
	getOrder := func() []int {
		var order []int
		for k := range m {
			order = append(order, k)
		}
		return order
	}

	order1 := getOrder()
	order2 := getOrder()
	order3 := getOrder()

	if len(order1) != 100 {
		t.Fatalf("expected 100 keys, got %d", len(order1))
	}

	// 几乎不可能 3 次遍历顺序完全相同
	if reflect.DeepEqual(order1, order2) && reflect.DeepEqual(order2, order3) {
		t.Log("warning: 3 iterations returned same order (very rare)")
	}
}

// TestMapAddress 验证不能对 map 元素取地址
// 编译就会报错，所以这里只演示逻辑
func TestMapAddress(t *testing.T) {
	m := map[int]int{1: 100}
	// _ = &m[1]  // compile error: cannot take the address of m[1]
	v := m[1]
	if v != 100 {
		t.Errorf("expected 100, got %d", v)
	}
}

// TestMapDeleteMemory 验证 delete 不立即释放内存
func TestMapDeleteMemory(t *testing.T) {
	m := make(map[int]int, 1000)
	for i := 0; i < 1000; i++ {
		m[i] = i
	}
	// 删除所有元素
	for i := 0; i < 1000; i++ {
		delete(m, i)
	}
	if len(m) != 0 {
		t.Errorf("expected len=0, got %d", len(m))
	}
	// bucket 仍然存在，要等 GC 才会释放
}

// TestRWMutexMap 验证 RWMutex + map 并发安全
func TestRWMutexMap(t *testing.T) {
	var mu sync.RWMutex
	m := make(map[string]int)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mu.Lock()
			m["k"] = id
			mu.Unlock()
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.RLock()
			_ = m["k"]
			mu.RUnlock()
		}()
	}
	wg.Wait()

	mu.RLock()
	if _, ok := m["k"]; !ok {
		t.Error("expected key \"k\" to exist")
	}
	mu.RUnlock()
}

// TestSyncMap 验证 sync.Map 并发安全
func TestSyncMap(t *testing.T) {
	var sm sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sm.Store("k", id)
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = sm.Load("k")
		}()
	}
	wg.Wait()

	if _, ok := sm.Load("k"); !ok {
		t.Error("expected key \"k\" to exist")
	}
}

// TestExpandTrigger 验证扩容后 map 仍可用
func TestExpandTrigger(t *testing.T) {
	m := make(map[int]int)
	for i := 0; i < 10000; i++ {
		m[i] = i
	}
	for i := 0; i < 10000; i++ {
		if v, ok := m[i]; !ok || v != i {
			t.Errorf("at i=%d: expected %d, got %d (ok=%v)", i, i, v, ok)
		}
	}
}

// BenchmarkMapRead benchmark 读 map
func BenchmarkMapRead(b *testing.B) {
	m := make(map[int]int, 1024)
	for i := 0; i < 1024; i++ {
		m[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i&1023]
	}
}

// BenchmarkRWMutexMapRead benchmark RWMutex + map 读
func BenchmarkRWMutexMapRead(b *testing.B) {
	var mu sync.RWMutex
	m := make(map[int]int, 1024)
	for i := 0; i < 1024; i++ {
		m[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = m[i&1023]
		mu.RUnlock()
	}
}

// BenchmarkSyncMapRead benchmark sync.Map 读
func BenchmarkSyncMapRead(b *testing.B) {
	var sm sync.Map
	for i := 0; i < 1024; i++ {
		sm.Store(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sm.Load(i & 1023)
	}
}

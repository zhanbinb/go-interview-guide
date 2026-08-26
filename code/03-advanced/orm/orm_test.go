package main

import (
	"strings"
	"testing"
)

// TestCreate 验证创建
func TestCreate(t *testing.T) {
	db := newMockDB()
	u := User{Name: "Alice", Email: "a@x.com", Age: 25}
	db.Create(u)
	if len(db.tables["user"]) != 1 {
		t.Errorf("expected 1 user, got %d", len(db.tables["user"]))
	}
}

// TestFirstFound 验证 First 找到
func TestFirstFound(t *testing.T) {
	db := newMockDB()
	db.Create(User{Name: "Alice", Email: "a@x.com", Age: 25})
	if !db.First(&User{}, 1) {
		t.Error("First should find user with id=1")
	}
}

// TestFirstNotFound 验证 First 找不到
func TestFirstNotFound(t *testing.T) {
	db := newMockDB()
	if db.First(&User{}, 999) {
		t.Error("First should return false for non-existent id")
	}
}

// TestSave 验证 Save（更新）
func TestSave(t *testing.T) {
	db := newMockDB()
	u := User{Model: Model{ID: 1}, Name: "Alice", Email: "a@x.com", Age: 25}
	db.Create(u)
	u.Age = 26
	if !db.Save(u) {
		t.Error("Save should return true for existing id")
	}
}

// TestDeleteSoft 验证软删除
func TestDeleteSoft(t *testing.T) {
	db := newMockDB()
	u := User{Model: Model{ID: 1}}
	db.Create(u)
	db.Delete(u)
	// 软删除后行还在
	if len(db.tables["user"]) != 1 {
		t.Errorf("soft delete should keep row, got %d", len(db.tables["user"]))
	}
	// 但 deleted_at 不为空
	row := db.tables["user"][0]
	if row["deleted_at"] == "" || row["deleted_at"] == nil {
		t.Error("deleted_at should be set after soft delete")
	}
}

// TestBeforeCreateHook 验证 BeforeCreate 钩子
func TestBeforeCreateHook(t *testing.T) {
	a := &Article{Title: "Hello World 测试"}
	if err := a.BeforeCreate(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(a.Slug, "hello-world") {
		t.Errorf("expected slug to contain 'hello-world', got %s", a.Slug)
	}
}

// TestTransactionCommit 验证事务提交
func TestTransactionCommit(t *testing.T) {
	db := newMockDB()
	tx := db.Begin()
	tx.Create(User{Name: "Alice", Email: "a@x.com", Age: 25})
	tx.Create(User{Name: "Bob", Email: "b@x.com", Age: 30})
	tx.Commit()
	if len(db.tables["user"]) != 2 {
		t.Errorf("expected 2 users after commit, got %d", len(db.tables["user"]))
	}
}

// TestTransactionRollback 验证事务回滚
func TestTransactionRollback(t *testing.T) {
	db := newMockDB()
	tx := db.Begin()
	tx.Create(User{Name: "Alice", Email: "a@x.com", Age: 25})
	tx.Rollback()
	if len(db.tables["user"]) != 0 {
		t.Errorf("expected 0 users after rollback, got %d", len(db.tables["user"]))
	}
}

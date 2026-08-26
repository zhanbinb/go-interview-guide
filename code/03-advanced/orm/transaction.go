package main

import "fmt"

// mockTx 模拟事务（带快照回滚）
type mockTx struct {
	db       *mockDB
	snapshot map[string][]map[string]any
	commits  int
	rollbacks int
}

// Begin 开启事务（保存快照）
func (db *mockDB) Begin() *mockTx {
	fmt.Println("  BEGIN TRANSACTION")
	snapshot := make(map[string][]map[string]any)
	for k, v := range db.tables {
		snap := make([]map[string]any, len(v))
		copy(snap, v)
		snapshot[k] = snap
	}
	return &mockTx{db: db, snapshot: snapshot}
}

// Commit 提交
func (tx *mockTx) Commit() {
	tx.commits++
	fmt.Println("  COMMIT")
}

// Rollback 回滚（恢复快照）
func (tx *mockTx) Rollback() {
	tx.rollbacks++
	fmt.Println("  ROLLBACK")
	// 恢复快照
	tx.db.tables = tx.snapshot
}

// Create 在事务中执行（用 tx.db 还是原 db？简化处理都用 db）
func (tx *mockTx) Create(model any) {
	tx.db.Create(model)
}

// DemoTx 演示事务
func DemoTx() {
	fmt.Println("=== 事务 ===")
	fmt.Println()

	// 1. 基本事务
	fmt.Println("【1】基本事务 - 全部成功")
	db := newMockDB()
	tx := db.Begin()
	tx.Create(User{Name: "Alice", Email: "a@x.com", Age: 25})
	tx.Create(User{Name: "Bob", Email: "b@x.com", Age: 30})
	tx.Commit()
	fmt.Printf("  ✓ 提交: %d 个用户\\n\\n", len(db.tables["user"]))

	// 2. 失败回滚
	fmt.Println("【2】失败回滚")
	db2 := newMockDB()
	tx2 := db2.Begin()
	tx2.Create(User{Name: "Alice", Email: "a@x.com", Age: 25})
	// 模拟失败
	fmt.Println("  [业务报错: 比如余额不足]")
	tx2.Rollback()
	fmt.Printf("  ✗ 回滚: users 表有 %d 行（应该是 0）\\n\\n", len(db2.tables["user"]))

	// 3. GORM 实际事务写法
	fmt.Println("【3】GORM 实际事务（错误自动回滚）")
	fmt.Println("  err := db.Transaction(func(tx *gorm.DB) error {")
	fmt.Println("      if err := tx.Create(&user1).Error; err != nil {")
	fmt.Println("          return err  // 自动 rollback")
	fmt.Println("      }")
	fmt.Println("      if err := tx.Create(&user2).Error; err != nil {")
	fmt.Println("          return err  // 自动 rollback")
	fmt.Println("      }")
	fmt.Println("      return nil  // 自动 commit")
	fmt.Println("  })")
	fmt.Println()

	fmt.Println("📌 事务要点:")
	fmt.Println("   - ACID: Atomic / Consistent / Isolated / Durable")
	fmt.Println("   - 多个写操作必须用事务（避免部分失败）")
	fmt.Println("   - GORM Transaction 函数：返回 error 自动 rollback")
	fmt.Println("   - 事务粒度不要太大（影响并发）")
}

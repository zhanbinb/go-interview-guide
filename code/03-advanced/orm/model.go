package main

import "fmt"

// Model 模拟 gorm.Model（基础字段）
// GORM 的 gorm.Model 是这样的：
//   type Model struct {
//       ID        uint      `gorm:"primarykey"`
//       CreatedAt time.Time `gorm:"autoCreateTime"`
//       UpdatedAt time.Time `gorm:"autoUpdateTime"`
//       DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
//   }
type Model struct {
	ID        int
	CreatedAt string
	UpdatedAt string
	DeletedAt *string // 软删除：null = 未删除
}

// User 用户模型
type User struct {
	Model
	Name  string `db:"user_name"`  // 列名映射
	Email string `db:"email"`
	Age   int    `db:"age"`
}

// Post 文章模型（关联关系演示）
type Post struct {
	Model
	Title  string `db:"title"`
	Body   string `db:"body"`
	UserID int    `db:"user_id"` // 外键
}

// DemoModel 演示模型定义
func DemoModel() {
	fmt.Println("=== 模型定义 ===")
	fmt.Println()

	// 1. 基础 Model
	fmt.Println("【1】基础 Model (gorm.Model 内置字段)")
	fmt.Println("  ID / CreatedAt / UpdatedAt / DeletedAt")
	fmt.Println()

	// 2. Struct Tag 映射
	fmt.Println("【2】Struct Tag 映射数据库列名")
	u := User{
		Model: Model{ID: 1, CreatedAt: "2026-01-01"},
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	fmt.Printf("  User{ID: %d, Name: %q, Email: %q, Age: %d}\n",
		u.ID, u.Name, u.Email, u.Age)
	fmt.Println("  → Name 字段用 db:\"user_name\" 映射到 user_name 列")
	fmt.Println()

	// 3. 软删除字段
	fmt.Println("【3】软删除 (DeletedAt)")
	now := "2026-01-15"
	u2 := User{Model: Model{ID: 2, DeletedAt: &now}}
	fmt.Printf("  DeletedAt = %v 表示已删除\\n", u2.DeletedAt)
	fmt.Printf("  DeletedAt = nil 表示未删除\\n\\n", )

	// 4. 关联关系
	fmt.Println("【4】关联关系（外键）")
	p := Post{
		Model:  Model{ID: 100},
		Title:  "Hello",
		UserID: u.ID,
	}
	fmt.Printf("  Post{ID: %d, Title: %q, UserID: %d} → 关联到 User.ID=%d\\n",
		p.ID, p.Title, p.UserID, u.ID)
	fmt.Println()

	// 5. GORM 实际用法
	fmt.Println("【5】GORM 实际用法（对比）")
	fmt.Println("  type User struct {")
	fmt.Println("      gorm.Model")
	fmt.Println("      Name string `gorm:\"column:user_name;size:100\"`")
	fmt.Println("  }")
	fmt.Println("  db.Create(&user)  // 自动生成 INSERT SQL")
	fmt.Println("  db.First(&user, 1) // 自动生成 SELECT WHERE id=1 SQL")
	fmt.Println()

	fmt.Println("📌 ORM 让 SQL 自动生成，避免手写 SQL 字符串")
	fmt.Println("   代价: 学习成本 + 性能略低于手写 SQL")
}

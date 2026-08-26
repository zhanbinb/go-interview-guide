package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// mockDB 模拟数据库（用内存 map 代替真 DB）
type mockDB struct {
	tables map[string][]map[string]any
	idGen  int
}

func newMockDB() *mockDB {
	return &mockDB{
		tables: make(map[string][]map[string]any),
		idGen:  0,
	}
}

// Create 模拟 db.Create
func (db *mockDB) Create(model any) {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name())
	now := time.Now().Format("2006-01-02")

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	row := make(map[string]any)
	db.idGen++
	row["id"] = db.idGen
	row["created_at"] = now
	row["updated_at"] = now

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("db")
		if tag == "" {
			tag = strings.ToLower(field.Name)
		}
		row[tag] = v.Field(i).Interface()
	}

	db.tables[tableName] = append(db.tables[tableName], row)
	fmt.Printf("  INSERT INTO %s -> row id=%d\n", tableName, row["id"])
}

// First 模拟 db.First
func (db *mockDB) First(model any, id int) bool {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name())
	for _, row := range db.tables[tableName] {
		if row["id"] == id {
			fmt.Printf("  SELECT * FROM %s WHERE id=%d -> 找到\n", tableName, id)
			return true
		}
	}
	fmt.Printf("  SELECT * FROM %s WHERE id=%d -> 未找到\n", tableName, id)
	return false
}

// Find 模拟 db.Find
func (db *mockDB) Find(model any, where string, args ...any) int {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name())
	fmt.Printf("  SELECT * FROM %s WHERE %s %v -> 找到 %d 行\n",
		tableName, where, args, len(db.tables[tableName]))
	return len(db.tables[tableName])
}

// Save 模拟 db.Save
func (db *mockDB) Save(model any) bool {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name())
	now := time.Now().Format("2006-01-02")
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	id := int(v.FieldByName("ID").Int())
	for _, row := range db.tables[tableName] {
		if row["id"] == id {
			row["updated_at"] = now
			fmt.Printf("  UPDATE %s SET ... WHERE id=%d\n", tableName, id)
			return true
		}
	}
	fmt.Printf("  UPDATE %s WHERE id=%d -> 未找到\n", tableName, id)
	return false
}

// Delete 模拟 db.Delete（软删除）
func (db *mockDB) Delete(model any) bool {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableName := strings.ToLower(t.Name())
	now := time.Now().Format("2006-01-02")
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	id := int(v.FieldByName("ID").Int())
	for _, row := range db.tables[tableName] {
		if row["id"] == id {
			row["deleted_at"] = now
			fmt.Printf("  UPDATE %s SET deleted_at=%q WHERE id=%d (软删除)\n",
				tableName, now, id)
			return true
		}
	}
	return false
}

// DemoORM 演示 CRUD 操作
func DemoORM() {
	fmt.Println("=== ORM CRUD 操作 ===")
	fmt.Println()

	db := newMockDB()

	// Create
	fmt.Println("【1】Create 插入")
	db.Create(User{Name: "Alice", Email: "alice@example.com", Age: 25})
	db.Create(User{Name: "Bob", Email: "bob@example.com", Age: 30})
	db.Create(User{Name: "Charlie", Email: "charlie@example.com", Age: 35})
	fmt.Println()

	// First
	fmt.Println("【2】First 按 ID 查一条")
	db.First(&User{}, 1)
	db.First(&User{}, 999)
	fmt.Println()

	// Find
	fmt.Println("【3】Find 按条件查多条")
	db.Find(&[]User{}, "age > ?", 28)
	fmt.Println()

	// Save
	fmt.Println("【4】Save 更新")
	u := User{Model: Model{ID: 1}, Name: "Alice", Email: "alice@example.com", Age: 26}
	db.Save(u)
	fmt.Println()

	// Delete
	fmt.Println("【5】Delete 软删除")
	u2 := User{Model: Model{ID: 2}}
	db.Delete(u2)
	fmt.Println("  -> u2 还在表里，但 deleted_at 不为空")
	fmt.Println()

	fmt.Println("GORM 完整 CRUD:")
	fmt.Println("  db.Create(&user)")
	fmt.Println("  db.First(&user, 1)")
	fmt.Println("  db.Find(&users, \"age > ?\", 20)")
	fmt.Println("  db.Model(&user).Update(\"age\", 30)")
	fmt.Println("  db.Delete(&user)")
}

package main

import "fmt"

// Hook 钩子接口（类似 GORM 的 hooks）
type Hook interface {
	BeforeCreate() error
	AfterCreate() error
	BeforeUpdate() error
	AfterUpdate() error
}

// Article 带钩子的文章
type Article struct {
	ID      int
	Title   string
	Content string
	Slug    string // 自动生成的 URL slug
}

// BeforeCreate 创建前自动生成 slug
func (a *Article) BeforeCreate() error {
	if a.Slug == "" {
		a.Slug = slugify(a.Title)
		fmt.Printf("  [hook BeforeCreate] 自动生成 slug: %q\\n", a.Slug)
	}
	return nil
}

// AfterCreate 创建后打日志（实际场景：发通知/发事件）
func (a *Article) AfterCreate() error {
	fmt.Printf("  [hook AfterCreate] 文章已创建: id=%d\\n", a.ID)
	return nil
}

// slugify 简化的 slug 生成（实际用 github.com/gosimple/slug）
func slugify(s string) string {
	result := ""
	for _, c := range s {
		switch {
		case c == ' ':
			result += "-"
		case c >= 'A' && c <= 'Z':
			result += string(c + 32) // 转小写
		case c >= 'a' && c <= 'z':
			result += string(c)
		case c >= '0' && c <= '9':
			result += string(c)
		}
	}
	if result == "" {
		result = "untitled"
	}
	return result
}

// DemoHooks 演示钩子机制
func DemoHooks() {
	fmt.Println("=== 钩子机制 ===")
	fmt.Println()

	// 1. 创建时自动调用钩子
	fmt.Println("【1】BeforeCreate 自动生成 slug")
	a := &Article{
		Title:   "Hello World 教程",
		Content: "...",
	}
	fmt.Println("  调用 a.BeforeCreate() 之前:")
	fmt.Printf("    Slug = %q\\n", a.Slug)

	// 手动调用钩子（演示用，ORM 会自动调用）
	a.BeforeCreate()
	fmt.Printf("  调用 a.BeforeCreate() 之后:\\n    Slug = %q\\n", a.Slug)
	a.ID = 1
	a.AfterCreate()
	fmt.Println()

	// 2. Update 钩子
	fmt.Println("【2】Update 钩子（实际场景：同步 ES、更新缓存）")
	fmt.Println("  BeforeUpdate: 检查权限")
	fmt.Println("  AfterUpdate:  失效缓存")
	fmt.Println()

	// 3. GORM 实际钩子列表
	fmt.Println("【3】GORM 完整钩子")
	fmt.Println("  BeforeCreate / AfterCreate")
	fmt.Println("  BeforeUpdate / AfterUpdate")
	fmt.Println("  BeforeSave  / AfterSave")
	fmt.Println("  BeforeDelete / AfterDelete")
	fmt.Println("  AfterFind")
	fmt.Println()

	fmt.Println("📌 钩子实战用途:")
	fmt.Println("   - 自动填充字段（slug, uuid, created_at）")
	fmt.Println("   - 数据校验（BeforeCreate 检查必填）")
	fmt.Println("   - 副作用（AfterUpdate 失效缓存）")
	fmt.Println("   - 审计日志（谁改了什么）")
	fmt.Println("   - 软删除关联清理")
}

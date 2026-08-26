# ORM 演示（精简版）

> 对应复习指南：§19 ORM

## 说明

sandbox 无网络下载不了 GORM，所以本 demo 用标准库 + mock 演示 ORM 核心模式。

## 面试只需了解 5 个点

1. ORM = Object-Relational Mapping（对象-关系映射）
2. Struct tag 映射数据库字段（如 gorm:"column:user_name"）
3. CRUD 四种操作（Create / Read / Update / Delete）
4. 钩子（BeforeCreate / AfterUpdate 等）用于自动填充字段
5. 软删除（用 deleted_at 字段实现）

## 文件清单

- main.go: 菜单
- model.go: 模型定义（struct + tag）
- orm.go: 简化版 ORM（CRUD + 自动 SQL 生成）
- hooks.go: 钩子机制
- transaction.go: 事务
- orm_test.go: 测试

## 面试速记

GORM 核心：
  - gorm.Model: 内置 ID/CreatedAt/UpdatedAt/DeletedAt
  - db.Create(&user): 插入
  - db.First(&user, id): 查一条
  - db.Find(&users): 查多条
  - db.Where("name = ?", "x"): 条件
  - db.Save(&user): 更新
  - db.Delete(&user): 删除（软删除）
  - db.Transaction(func(tx) { ... }): 事务
  - Hooks: BeforeCreate/AfterCreate 等

GORM Gen:
  - 代码生成 ORM（类型安全）
  - 不需要字符串 SQL

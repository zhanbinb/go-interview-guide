package main

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Name  string  `db:"name"`
	Price float64 `db:"price"`
}

type User struct {
	gorm.Model
	Name     string    `db:"name"`
	Age      int       `db:"age"`
	Birthday time.Time `db:"birthday"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}

	//ctx := context.Background()

	db.AutoMigrate(&Product{}, &User{})

	//创建
	// db.Create(&Product{
	// 	Name:  "test",
	// 	Price: 100,
	// })

	// db.Create(&Product{
	// 	Name:  "test2",
	// 	Price: 200,
	// })

	// var product Product
	// db.First(&product, 1)

	// fmt.Println(product)

	// db.Model(product).Update("price", 500)

	// product.ID = 3
	// db.Where("id =?", 3).Delete(&product)
	// // 创建用户

	user := User{Name: "Kevin", Age: 30, Birthday: time.Now()}

	result := db.Debug().Create(&user)

	fmt.Println(user.ID)
	// //返回插入记录数
	fmt.Println("插入记录数:", result.RowsAffected)

	var users []User
	db.Debug().Unscoped().Find(&users)
	for _, u := range users {
		fmt.Println(u.Name)
	}

	var users2 []User
	db.Raw("SELECT * FROM users").Scan(&users2)
	for _, u := range users2 {
		fmt.Println(u.Age)
	}

	//db.Exec("Drop table users")

	fmt.Println("创建成功")
}

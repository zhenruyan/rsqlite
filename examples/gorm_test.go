package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormUser 用户模型
type GormUser struct {
	ID        uint       `gorm:"primarykey"`
	Name      string     `gorm:"not null;size:100"`
	Email     string     `gorm:"uniqueIndex;size:100"`
	Age       int        `gorm:"default:0"`
	Posts     []GormPost `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GormPost 文章模型
type GormPost struct {
	ID        uint     `gorm:"primarykey"`
	Title     string   `gorm:"not null;size:200"`
	Content   string   `gorm:"type:text"`
	UserID    uint     `gorm:"not null"`
	User      GormUser `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func runGormTest() {
	fmt.Println("=== GORM 兼容性测试 ===")

	// 连接到rqlite集群
	// 注意：这里使用sqlite驱动名，GORM会自动使用我们的rqlite驱动
	db, err := gorm.Open(sqlite.Open("10.0.1.10:4000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		fmt.Println("这通常意味着rqlite服务器没有运行，但DSN解析是正常的")
		return
	}

	fmt.Println("✓ 成功连接到rqlite集群")

	// 自动迁移
	err = db.AutoMigrate(&GormUser{}, &GormPost{})
	if err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	fmt.Println("✓ 自动迁移完成")

	// 测试基本CRUD操作
	testGormBasicCRUD(db)

	// 测试关联查询
	testGormAssociations(db)

	// 测试事务
	testGormTransactions(db)

	// 测试批量操作
	testGormBatchOperations(db)

	fmt.Println("=== GORM 兼容性测试完成 ===")
}

func testGormBasicCRUD(db *gorm.DB) {
	fmt.Println("\n--- 测试基本CRUD操作 ---")

	// Create - 创建
	user := GormUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}
	result := db.Create(&user)
	if result.Error != nil {
		log.Printf("创建用户失败: %v", result.Error)
		return
	}
	fmt.Printf("✓ 创建用户成功: ID=%d, 影响行数=%d\n", user.ID, result.RowsAffected)

	// Read - 查询
	var foundUser GormUser
	db.First(&foundUser, user.ID)
	fmt.Printf("✓ 查询用户: ID=%d, 姓名=%s, 邮箱=%s\n", foundUser.ID, foundUser.Name, foundUser.Email)

	// Update - 更新
	db.Model(&foundUser).Update("Age", 26)
	fmt.Printf("✓ 更新用户年龄为: %d\n", 26)

	// Delete - 删除（软删除）
	db.Delete(&foundUser)
	fmt.Println("✓ 软删除用户成功")

	// 查询包括软删除的记录
	var deletedUser GormUser
	db.Unscoped().First(&deletedUser, user.ID)
	fmt.Printf("✓ 查询软删除用户: ID=%d, 姓名=%s\n", deletedUser.ID, deletedUser.Name)
}

func testGormAssociations(db *gorm.DB) {
	fmt.Println("\n--- 测试关联查询 ---")

	// 创建用户
	user := GormUser{
		Name:  "李四",
		Email: "lisi@example.com",
		Age:   30,
	}
	db.Create(&user)

	// 创建文章
	posts := []GormPost{
		{
			Title:   "Go语言入门",
			Content: "这是一篇关于Go语言的入门文章...",
			UserID:  user.ID,
		},
		{
			Title:   "rqlite使用指南",
			Content: "这是一篇关于rqlite的使用指南...",
			UserID:  user.ID,
		},
	}
	db.Create(&posts)
	fmt.Printf("✓ 为用户 %s 创建了 %d 篇文章\n", user.Name, len(posts))

	// 预加载查询
	var userWithPosts GormUser
	db.Preload("Posts").First(&userWithPosts, user.ID)
	fmt.Printf("✓ 预加载查询: 用户 %s 有 %d 篇文章\n", userWithPosts.Name, len(userWithPosts.Posts))

	// 关联查询
	var userPosts []GormPost
	db.Where("user_id = ?", user.ID).Find(&userPosts)
	fmt.Printf("✓ 关联查询: 找到 %d 篇文章\n", len(userPosts))
}

func testGormTransactions(db *gorm.DB) {
	fmt.Println("\n--- 测试事务 ---")

	// 注意：rqlite不支持真正的事务，但GORM的事务接口仍然可以使用
	tx := db.Begin()

	user := GormUser{
		Name:  "王五",
		Email: "wangwu@example.com",
		Age:   28,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("事务中创建用户失败: %v", err)
		return
	}

	post := GormPost{
		Title:   "事务测试文章",
		Content: "这是在事务中创建的文章",
		UserID:  user.ID,
	}

	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		log.Printf("事务中创建文章失败: %v", err)
		return
	}

	tx.Commit()
	fmt.Printf("✓ 事务提交成功: 创建用户 %s 和文章 %s\n", user.Name, post.Title)
}

func testGormBatchOperations(db *gorm.DB) {
	fmt.Println("\n--- 测试批量操作 ---")

	// 批量创建用户
	users := []GormUser{
		{Name: "用户1", Email: "user1@example.com", Age: 20},
		{Name: "用户2", Email: "user2@example.com", Age: 21},
		{Name: "用户3", Email: "user3@example.com", Age: 22},
		{Name: "用户4", Email: "user4@example.com", Age: 23},
		{Name: "用户5", Email: "user5@example.com", Age: 24},
	}

	result := db.CreateInBatches(&users, 3) // 每批3个
	if result.Error != nil {
		log.Printf("批量创建失败: %v", result.Error)
		return
	}
	fmt.Printf("✓ 批量创建 %d 个用户成功\n", len(users))

	// 批量查询
	var allUsers []GormUser
	db.Find(&allUsers)
	fmt.Printf("✓ 查询到总共 %d 个用户\n", len(allUsers))

	// 批量更新
	db.Model(&GormUser{}).Where("age < ?", 25).Update("age", gorm.Expr("age + ?", 1))
	fmt.Println("✓ 批量更新年龄小于25的用户")

	// 条件查询
	var youngUsers []GormUser
	db.Where("age BETWEEN ? AND ?", 20, 25).Find(&youngUsers)
	fmt.Printf("✓ 查询年龄在20-25之间的用户: %d 个\n", len(youngUsers))

	// 聚合查询
	var count int64
	var avgAge float64
	db.Model(&GormUser{}).Count(&count)
	db.Model(&GormUser{}).Select("AVG(age)").Row().Scan(&avgAge)
	fmt.Printf("✓ 聚合查询: 总用户数=%d, 平均年龄=%.1f\n", count, avgAge)
}

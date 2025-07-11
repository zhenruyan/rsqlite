package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
	"xorm.io/xorm"
)

// XormUser XORM用户模型
type XormUser struct {
	Id        int64     `xorm:"pk autoincr 'id'"`
	Name      string    `xorm:"varchar(100) notnull 'name'"`
	Email     string    `xorm:"varchar(100) unique notnull 'email'"`
	Age       int       `xorm:"default 0 'age'"`
	CreatedAt time.Time `xorm:"created 'created_at'"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'"`
}

// XormPost XORM文章模型
type XormPost struct {
	Id        int64     `xorm:"pk autoincr 'id'"`
	Title     string    `xorm:"varchar(200) notnull 'title'"`
	Content   string    `xorm:"text 'content'"`
	UserId    int64     `xorm:"notnull 'user_id'"`
	CreatedAt time.Time `xorm:"created 'created_at'"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'"`
}

// TableName 设置表名
func (XormUser) TableName() string {
	return "xorm_users"
}

func (XormPost) TableName() string {
	return "xorm_posts"
}

func runXormTest() {
	fmt.Println("=== XORM 兼容性测试 ===")

	// 连接到rqlite集群
	engine, err := xorm.NewEngine("sqlite", "10.0.1.10:4000")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		fmt.Println("这通常意味着rqlite服务器没有运行，但DSN解析是正常的")
		return
	}

	// 设置日志级别
	engine.ShowSQL(true)

	fmt.Println("✓ 成功连接到rqlite集群")

	// 同步表结构
	err = engine.Sync2(new(XormUser), new(XormPost))
	if err != nil {
		log.Fatalf("同步表结构失败: %v", err)
	}
	fmt.Println("✓ 同步表结构完成")

	// 测试基本CRUD操作
	testXormBasicCRUD(engine)

	// 测试关联查询
	testXormAssociations(engine)

	// 测试事务
	testXormTransactions(engine)

	// 测试批量操作
	testXormBatchOperations(engine)

	fmt.Println("=== XORM 兼容性测试完成 ===")
}

func testXormBasicCRUD(engine *xorm.Engine) {
	fmt.Println("\n--- 测试基本CRUD操作 ---")

	// Create - 创建
	user := &XormUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	affected, err := engine.Insert(user)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 创建用户成功: ID=%d, 影响行数=%d\n", user.Id, affected)

	// Read - 查询
	foundUser := &XormUser{}
	has, err := engine.ID(user.Id).Get(foundUser)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		return
	}
	if !has {
		log.Printf("用户不存在: ID=%d", user.Id)
		return
	}
	fmt.Printf("✓ 查询用户: ID=%d, 姓名=%s, 邮箱=%s\n", foundUser.Id, foundUser.Name, foundUser.Email)

	// Update - 更新
	foundUser.Age = 26
	affected, err = engine.ID(foundUser.Id).Update(foundUser)
	if err != nil {
		log.Printf("更新用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 更新用户年龄为: %d, 影响行数=%d\n", foundUser.Age, affected)

	// Delete - 删除
	affected, err = engine.ID(foundUser.Id).Delete(&XormUser{})
	if err != nil {
		log.Printf("删除用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 删除用户成功, 影响行数=%d\n", affected)
}

func testXormAssociations(engine *xorm.Engine) {
	fmt.Println("\n--- 测试关联查询 ---")

	// 创建用户
	user := &XormUser{
		Name:  "李四",
		Email: "lisi@example.com",
		Age:   30,
	}
	_, err := engine.Insert(user)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return
	}

	// 创建文章
	posts := []XormPost{
		{
			Title:   "Go语言入门",
			Content: "这是一篇关于Go语言的入门文章...",
			UserId:  user.Id,
		},
		{
			Title:   "rqlite使用指南",
			Content: "这是一篇关于rqlite的使用指南...",
			UserId:  user.Id,
		},
	}

	affected, err := engine.Insert(&posts)
	if err != nil {
		log.Printf("创建文章失败: %v", err)
		return
	}
	fmt.Printf("✓ 为用户 %s 创建了 %d 篇文章, 影响行数=%d\n", user.Name, len(posts), affected)

	// 关联查询 - 通过JOIN查询
	type UserWithPostCount struct {
		XormUser  `xorm:"extends"`
		PostCount int `xorm:"post_count"`
	}

	var userWithCount []UserWithPostCount
	err = engine.Table("xorm_users").
		Select("xorm_users.*, COUNT(xorm_posts.id) as post_count").
		Join("LEFT", "xorm_posts", "xorm_users.id = xorm_posts.user_id").
		Where("xorm_users.id = ?", user.Id).
		GroupBy("xorm_users.id").
		Find(&userWithCount)

	if err != nil {
		log.Printf("关联查询失败: %v", err)
		return
	}

	if len(userWithCount) > 0 {
		fmt.Printf("✓ 关联查询: 用户 %s 有 %d 篇文章\n", userWithCount[0].Name, userWithCount[0].PostCount)
	}

	// 查询用户的所有文章
	var userPosts []XormPost
	err = engine.Where("user_id = ?", user.Id).Find(&userPosts)
	if err != nil {
		log.Printf("查询文章失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询到 %d 篇文章\n", len(userPosts))
}

func testXormTransactions(engine *xorm.Engine) {
	fmt.Println("\n--- 测试事务 ---")

	// 注意：rqlite不支持真正的事务，但XORM的事务接口仍然可以使用
	session := engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		log.Printf("开始事务失败: %v", err)
		return
	}

	user := &XormUser{
		Name:  "王五",
		Email: "wangwu@example.com",
		Age:   28,
	}

	_, err = session.Insert(user)
	if err != nil {
		session.Rollback()
		log.Printf("事务中创建用户失败: %v", err)
		return
	}

	post := &XormPost{
		Title:   "事务测试文章",
		Content: "这是在事务中创建的文章",
		UserId:  user.Id,
	}

	_, err = session.Insert(post)
	if err != nil {
		session.Rollback()
		log.Printf("事务中创建文章失败: %v", err)
		return
	}

	err = session.Commit()
	if err != nil {
		log.Printf("事务提交失败: %v", err)
		return
	}

	fmt.Printf("✓ 事务提交成功: 创建用户 %s 和文章 %s\n", user.Name, post.Title)
}

func testXormBatchOperations(engine *xorm.Engine) {
	fmt.Println("\n--- 测试批量操作 ---")

	// 批量创建用户
	users := []XormUser{
		{Name: "用户1", Email: "user1@example.com", Age: 20},
		{Name: "用户2", Email: "user2@example.com", Age: 21},
		{Name: "用户3", Email: "user3@example.com", Age: 22},
		{Name: "用户4", Email: "user4@example.com", Age: 23},
		{Name: "用户5", Email: "user5@example.com", Age: 24},
	}

	affected, err := engine.Insert(&users)
	if err != nil {
		log.Printf("批量创建失败: %v", err)
		return
	}
	fmt.Printf("✓ 批量创建 %d 个用户成功, 影响行数=%d\n", len(users), affected)

	// 批量查询
	var allUsers []XormUser
	err = engine.Find(&allUsers)
	if err != nil {
		log.Printf("批量查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询到总共 %d 个用户\n", len(allUsers))

	// 批量更新
	affected, err = engine.Where("age < ?", 25).Incr("age", 1).Update(&XormUser{})
	if err != nil {
		log.Printf("批量更新失败: %v", err)
		return
	}
	fmt.Printf("✓ 批量更新年龄小于25的用户, 影响行数=%d\n", affected)

	// 条件查询
	var youngUsers []XormUser
	err = engine.Where("age BETWEEN ? AND ?", 20, 25).Find(&youngUsers)
	if err != nil {
		log.Printf("条件查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询年龄在20-25之间的用户: %d 个\n", len(youngUsers))

	// 聚合查询
	count, err := engine.Count(&XormUser{})
	if err != nil {
		log.Printf("聚合查询失败: %v", err)
		return
	}

	type AvgResult struct {
		AvgAge float64 `xorm:"avg_age"`
	}
	var avgResult AvgResult
	_, err = engine.Table("xorm_users").Select("AVG(age) as avg_age").Get(&avgResult)
	if err != nil {
		log.Printf("平均年龄查询失败: %v", err)
		return
	}

	fmt.Printf("✓ 聚合查询: 总用户数=%d, 平均年龄=%.1f\n", count, avgResult.AvgAge)
}

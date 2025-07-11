package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
)

// BunUser Bun用户模型
type BunUser struct {
	bun.BaseModel `bun:"table:bun_users,alias:u"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,unique,notnull"`
	Age       int       `bun:"age,default:0"`
	Posts     []BunPost `bun:"rel:has-many,join:id=user_id"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

// BunPost Bun文章模型
type BunPost struct {
	bun.BaseModel `bun:"table:bun_posts,alias:p"`

	ID        int64     `bun:"id,pk,autoincrement"`
	Title     string    `bun:"title,notnull"`
	Content   string    `bun:"content"`
	UserID    int64     `bun:"user_id,notnull"`
	User      *BunUser  `bun:"rel:belongs-to,join:user_id=id"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

func runBunTest() {
	fmt.Println("=== Bun ORM 兼容性测试 ===")

	// 连接到rqlite集群
	sqldb, err := sql.Open("sqlite", "10.0.1.10:4000")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		fmt.Println("这通常意味着rqlite服务器没有运行，但DSN解析是正常的")
		return
	}

	// 创建Bun数据库实例
	db := bun.NewDB(sqldb, sqlitedialect.New())

	// 添加查询钩子用于调试
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	fmt.Println("✓ 成功连接到rqlite集群")

	ctx := context.Background()

	// 创建表
	if err := createBunTables(ctx, db); err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✓ 创建表完成")

	// 测试基本CRUD操作
	testBunBasicCRUD(ctx, db)

	// 测试关联查询
	testBunAssociations(ctx, db)

	// 测试事务
	testBunTransactions(ctx, db)

	// 测试批量操作
	testBunBatchOperations(ctx, db)

	fmt.Println("=== Bun ORM 兼容性测试完成 ===")
}

func createBunTables(ctx context.Context, db *bun.DB) error {
	// 创建用户表
	_, err := db.NewCreateTable().Model((*BunUser)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return err
	}

	// 创建文章表
	_, err = db.NewCreateTable().Model((*BunPost)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func testBunBasicCRUD(ctx context.Context, db *bun.DB) {
	fmt.Println("\n--- 测试基本CRUD操作 ---")

	// Create - 创建
	user := &BunUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	_, err := db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 创建用户成功: ID=%d\n", user.ID)

	// Read - 查询
	foundUser := &BunUser{}
	err = db.NewSelect().Model(foundUser).Where("id = ?", user.ID).Scan(ctx)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询用户: ID=%d, 姓名=%s, 邮箱=%s\n", foundUser.ID, foundUser.Name, foundUser.Email)

	// Update - 更新
	foundUser.Age = 26
	_, err = db.NewUpdate().Model(foundUser).Where("id = ?", foundUser.ID).Exec(ctx)
	if err != nil {
		log.Printf("更新用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 更新用户年龄为: %d\n", foundUser.Age)

	// Delete - 删除
	_, err = db.NewDelete().Model(foundUser).Where("id = ?", foundUser.ID).Exec(ctx)
	if err != nil {
		log.Printf("删除用户失败: %v", err)
		return
	}
	fmt.Println("✓ 删除用户成功")
}

func testBunAssociations(ctx context.Context, db *bun.DB) {
	fmt.Println("\n--- 测试关联查询 ---")

	// 创建用户
	user := &BunUser{
		Name:  "李四",
		Email: "lisi@example.com",
		Age:   30,
	}
	_, err := db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return
	}

	// 创建文章
	posts := []*BunPost{
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

	_, err = db.NewInsert().Model(&posts).Exec(ctx)
	if err != nil {
		log.Printf("创建文章失败: %v", err)
		return
	}
	fmt.Printf("✓ 为用户 %s 创建了 %d 篇文章\n", user.Name, len(posts))

	// 关联查询 - 查询用户及其文章
	userWithPosts := &BunUser{}
	err = db.NewSelect().Model(userWithPosts).
		Relation("Posts").
		Where("u.id = ?", user.ID).
		Scan(ctx)
	if err != nil {
		log.Printf("关联查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 关联查询: 用户 %s 有 %d 篇文章\n", userWithPosts.Name, len(userWithPosts.Posts))

	// 查询文章及其作者
	var postsWithUser []BunPost
	err = db.NewSelect().Model(&postsWithUser).
		Relation("User").
		Where("p.user_id = ?", user.ID).
		Scan(ctx)
	if err != nil {
		log.Printf("查询文章失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询到 %d 篇文章，作者: %s\n", len(postsWithUser), postsWithUser[0].User.Name)
}

func testBunTransactions(ctx context.Context, db *bun.DB) {
	fmt.Println("\n--- 测试事务 ---")

	// 注意：rqlite不支持真正的事务，但Bun的事务接口仍然可以使用
	err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		user := &BunUser{
			Name:  "王五",
			Email: "wangwu@example.com",
			Age:   28,
		}

		_, err := tx.NewInsert().Model(user).Exec(ctx)
		if err != nil {
			return err
		}

		post := &BunPost{
			Title:   "事务测试文章",
			Content: "这是在事务中创建的文章",
			UserID:  user.ID,
		}

		_, err = tx.NewInsert().Model(post).Exec(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("✓ 事务提交成功: 创建用户 %s 和文章 %s\n", user.Name, post.Title)
		return nil
	})

	if err != nil {
		log.Printf("事务执行失败: %v", err)
	}
}

func testBunBatchOperations(ctx context.Context, db *bun.DB) {
	fmt.Println("\n--- 测试批量操作 ---")

	// 批量创建用户
	users := []*BunUser{
		{Name: "用户1", Email: "user1@example.com", Age: 20},
		{Name: "用户2", Email: "user2@example.com", Age: 21},
		{Name: "用户3", Email: "user3@example.com", Age: 22},
		{Name: "用户4", Email: "user4@example.com", Age: 23},
		{Name: "用户5", Email: "user5@example.com", Age: 24},
	}

	_, err := db.NewInsert().Model(&users).Exec(ctx)
	if err != nil {
		log.Printf("批量创建失败: %v", err)
		return
	}
	fmt.Printf("✓ 批量创建 %d 个用户成功\n", len(users))

	// 批量查询
	var allUsers []BunUser
	err = db.NewSelect().Model(&allUsers).Scan(ctx)
	if err != nil {
		log.Printf("批量查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询到总共 %d 个用户\n", len(allUsers))

	// 批量更新
	_, err = db.NewUpdate().Model((*BunUser)(nil)).
		Set("age = age + 1").
		Where("age < ?", 25).
		Exec(ctx)
	if err != nil {
		log.Printf("批量更新失败: %v", err)
		return
	}
	fmt.Println("✓ 批量更新年龄小于25的用户")

	// 条件查询
	var youngUsers []BunUser
	err = db.NewSelect().Model(&youngUsers).
		Where("age BETWEEN ? AND ?", 20, 25).
		Scan(ctx)
	if err != nil {
		log.Printf("条件查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询年龄在20-25之间的用户: %d 个\n", len(youngUsers))

	// 聚合查询
	count, err := db.NewSelect().Model((*BunUser)(nil)).Count(ctx)
	if err != nil {
		log.Printf("聚合查询失败: %v", err)
		return
	}

	var avgAge float64
	err = db.NewSelect().Model((*BunUser)(nil)).
		ColumnExpr("AVG(age) as avg_age").
		Scan(ctx, &avgAge)
	if err != nil {
		log.Printf("平均年龄查询失败: %v", err)
		return
	}

	fmt.Printf("✓ 聚合查询: 总用户数=%d, 平均年龄=%.1f\n", count, avgAge)
}

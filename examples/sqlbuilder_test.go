package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
)

// SimpleUser 简单用户模型
type SimpleUser struct {
	ID        int64
	Name      string
	Email     string
	Age       int
	CreatedAt time.Time
}

// SimplePost 简单文章模型
type SimplePost struct {
	ID        int64
	Title     string
	Content   string
	UserID    int64
	CreatedAt time.Time
}

// SimpleQueryBuilder 简单的查询构建器
type SimpleQueryBuilder struct {
	table      string
	selectCols []string
	whereCond  []string
	whereArgs  []interface{}
	orderBy    string
	limit      int
}

func NewQueryBuilder(table string) *SimpleQueryBuilder {
	return &SimpleQueryBuilder{
		table:      table,
		selectCols: []string{"*"},
		whereCond:  []string{},
		whereArgs:  []interface{}{},
	}
}

func (qb *SimpleQueryBuilder) Select(cols ...string) *SimpleQueryBuilder {
	qb.selectCols = cols
	return qb
}

func (qb *SimpleQueryBuilder) Where(condition string, args ...interface{}) *SimpleQueryBuilder {
	qb.whereCond = append(qb.whereCond, condition)
	qb.whereArgs = append(qb.whereArgs, args...)
	return qb
}

func (qb *SimpleQueryBuilder) OrderBy(order string) *SimpleQueryBuilder {
	qb.orderBy = order
	return qb
}

func (qb *SimpleQueryBuilder) Limit(limit int) *SimpleQueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *SimpleQueryBuilder) Build() (string, []interface{}) {
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(qb.selectCols, ", "), qb.table)

	if len(qb.whereCond) > 0 {
		query += " WHERE " + strings.Join(qb.whereCond, " AND ")
	}

	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	return query, qb.whereArgs
}

func runSQLBuilderTest() {
	fmt.Println("=== SQL Builder 兼容性测试 ===")

	// 连接到rqlite集群
	db, err := sql.Open("sqlite", "localhost:4001,localhost:4002,localhost:4003")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		fmt.Println("这通常意味着rqlite服务器没有运行，但DSN解析是正常的")
		return
	}
	defer db.Close()

	fmt.Println("✓ 成功连接到rqlite集群")

	// 创建表
	if err := createTables(db); err != nil {
		log.Fatalf("创建表失败: %v", err)
	}
	fmt.Println("✓ 创建表完成")

	// 测试基本CRUD操作
	testSQLBuilderCRUD(db)

	// 测试复杂查询
	testSQLBuilderComplexQueries(db)

	// 测试批量操作
	testSQLBuilderBatchOperations(db)

	fmt.Println("=== SQL Builder 兼容性测试完成 ===")
}

func createTables(db *sql.DB) error {
	// 创建用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS simple_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		age INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := db.Exec(userTable)
	if err != nil {
		return err
	}

	// 创建文章表
	postTable := `
	CREATE TABLE IF NOT EXISTS simple_posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES simple_users(id)
	)`

	_, err = db.Exec(postTable)
	return err
}

func testSQLBuilderCRUD(db *sql.DB) {
	fmt.Println("\n--- 测试基本CRUD操作 ---")

	// Create - 创建用户
	insertSQL := "INSERT INTO simple_users (name, email, age) VALUES (?, ?, ?)"
	result, err := db.Exec(insertSQL, "张三", "zhangsan@example.com", 25)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
		return
	}

	userID, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()
	fmt.Printf("✓ 创建用户成功: ID=%d, 影响行数=%d\n", userID, affected)

	// Read - 使用查询构建器查询
	qb := NewQueryBuilder("simple_users").
		Select("id", "name", "email", "age").
		Where("id = ?", userID)

	query, args := qb.Build()
	fmt.Printf("✓ 构建查询: %s, 参数: %v\n", query, args)

	var user SimpleUser
	err = db.QueryRow(query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.Age)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
		return
	}
	fmt.Printf("✓ 查询用户: ID=%d, 姓名=%s, 邮箱=%s, 年龄=%d\n", user.ID, user.Name, user.Email, user.Age)

	// Update - 更新用户
	updateSQL := "UPDATE simple_users SET age = ? WHERE id = ?"
	result, err = db.Exec(updateSQL, 26, userID)
	if err != nil {
		log.Printf("更新用户失败: %v", err)
		return
	}
	affected, _ = result.RowsAffected()
	fmt.Printf("✓ 更新用户年龄, 影响行数=%d\n", affected)

	// Delete - 删除用户
	deleteSQL := "DELETE FROM simple_users WHERE id = ?"
	result, err = db.Exec(deleteSQL, userID)
	if err != nil {
		log.Printf("删除用户失败: %v", err)
		return
	}
	affected, _ = result.RowsAffected()
	fmt.Printf("✓ 删除用户, 影响行数=%d\n", affected)
}

func testSQLBuilderComplexQueries(db *sql.DB) {
	fmt.Println("\n--- 测试复杂查询 ---")

	// 先创建一些测试数据
	users := []SimpleUser{
		{Name: "李四", Email: "lisi@example.com", Age: 30},
		{Name: "王五", Email: "wangwu@example.com", Age: 28},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 32},
	}

	var userIDs []int64
	for _, user := range users {
		result, err := db.Exec("INSERT INTO simple_users (name, email, age) VALUES (?, ?, ?)",
			user.Name, user.Email, user.Age)
		if err != nil {
			log.Printf("创建用户失败: %v", err)
			continue
		}
		id, _ := result.LastInsertId()
		userIDs = append(userIDs, id)
	}
	fmt.Printf("✓ 创建了 %d 个测试用户\n", len(userIDs))

	// 为每个用户创建文章
	for i, userID := range userIDs {
		_, err := db.Exec("INSERT INTO simple_posts (title, content, user_id) VALUES (?, ?, ?)",
			fmt.Sprintf("文章标题%d", i+1), fmt.Sprintf("这是第%d篇文章的内容", i+1), userID)
		if err != nil {
			log.Printf("创建文章失败: %v", err)
		}
	}
	fmt.Println("✓ 为每个用户创建了文章")

	// 复杂查询1: 年龄范围查询
	qb1 := NewQueryBuilder("simple_users").
		Select("name", "age").
		Where("age BETWEEN ? AND ?", 25, 35).
		OrderBy("age DESC").
		Limit(10)

	query1, args1 := qb1.Build()
	fmt.Printf("✓ 年龄范围查询: %s\n", query1)

	rows, err := db.Query(query1, args1...)
	if err != nil {
		log.Printf("执行查询失败: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var name string
		var age int
		rows.Scan(&name, &age)
		count++
	}
	fmt.Printf("✓ 查询到 %d 个用户在年龄范围内\n", count)

	// 复杂查询2: JOIN查询（使用原生SQL）
	joinQuery := `
	SELECT u.name, u.email, p.title, p.created_at
	FROM simple_users u
	JOIN simple_posts p ON u.id = p.user_id
	WHERE u.age > ?
	ORDER BY p.created_at DESC
	`

	rows2, err := db.Query(joinQuery, 25)
	if err != nil {
		log.Printf("JOIN查询失败: %v", err)
		return
	}
	defer rows2.Close()

	joinCount := 0
	for rows2.Next() {
		var userName, userEmail, postTitle string
		var createdAt time.Time
		rows2.Scan(&userName, &userEmail, &postTitle, &createdAt)
		joinCount++
	}
	fmt.Printf("✓ JOIN查询到 %d 条用户-文章记录\n", joinCount)
}

func testSQLBuilderBatchOperations(db *sql.DB) {
	fmt.Println("\n--- 测试批量操作 ---")

	// 批量插入用户
	batchUsers := []SimpleUser{
		{Name: "批量用户1", Email: "batch1@example.com", Age: 20},
		{Name: "批量用户2", Email: "batch2@example.com", Age: 21},
		{Name: "批量用户3", Email: "batch3@example.com", Age: 22},
		{Name: "批量用户4", Email: "batch4@example.com", Age: 23},
		{Name: "批量用户5", Email: "batch5@example.com", Age: 24},
	}

	// 使用事务进行批量插入
	tx, err := db.Begin()
	if err != nil {
		log.Printf("开始事务失败: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO simple_users (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		log.Printf("准备语句失败: %v", err)
		return
	}
	defer stmt.Close()

	for _, user := range batchUsers {
		_, err = stmt.Exec(user.Name, user.Email, user.Age)
		if err != nil {
			tx.Rollback()
			log.Printf("批量插入失败: %v", err)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("提交事务失败: %v", err)
		return
	}
	fmt.Printf("✓ 批量插入 %d 个用户成功\n", len(batchUsers))

	// 聚合查询
	var totalUsers int
	var avgAge float64

	err = db.QueryRow("SELECT COUNT(*), AVG(age) FROM simple_users").Scan(&totalUsers, &avgAge)
	if err != nil {
		log.Printf("聚合查询失败: %v", err)
		return
	}
	fmt.Printf("✓ 聚合查询: 总用户数=%d, 平均年龄=%.1f\n", totalUsers, avgAge)

	// 批量更新
	result, err := db.Exec("UPDATE simple_users SET age = age + 1 WHERE age < ?", 25)
	if err != nil {
		log.Printf("批量更新失败: %v", err)
		return
	}
	affected, _ := result.RowsAffected()
	fmt.Printf("✓ 批量更新年龄小于25的用户, 影响行数=%d\n", affected)
}

// 运行SQL Builder测试
// 使用方法: go run sqlbuilder_test.go
func init() {
	// 这个文件可以作为库使用，或者单独运行
}

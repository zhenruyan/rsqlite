package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zhenruyan/rsqlite"
	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
)

func main() {
	// 演示如何使用rqlite驱动
	fmt.Println("=== rqlite Go驱动使用示例 ===")

	// 1. 演示DSN解析
	fmt.Println("\n1. DSN解析示例:")
	demonstrateDSNParsing()

	// 2. 演示驱动注册
	fmt.Println("\n2. 驱动注册检查:")
	checkDriverRegistration()

	// 3. 演示连接配置
	fmt.Println("\n3. 连接配置示例:")
	demonstrateConnectionConfig()

	// 4. 演示实际使用（需要rqlite服务器运行）
	fmt.Println("\n4. 实际使用示例:")
	fmt.Println("   注意：以下示例需要rqlite服务器运行在localhost:4001")
	fmt.Println("   如果没有运行rqlite服务器，会显示连接错误，这是正常的")
	demonstrateActualUsage()

	fmt.Println("\n=== 示例完成 ===")
}

func demonstrateDSNParsing() {
	// 各种DSN格式的解析示例
	dsnExamples := []string{
		"localhost:4001",
		"localhost:4001,localhost:4002,localhost:4003",
		"user:pass@localhost:4001",
		"localhost:4001?consistency=strong&timeout=60s",
		"user:pass@host1:4001,host2:4002?consistency=strong&timeout=30s",
	}

	for i, dsn := range dsnExamples {
		fmt.Printf("  示例 %d: %s\n", i+1, dsn)
		cfg, err := rsqlite.ParseDSN(dsn)
		if err != nil {
			fmt.Printf("    错误: %v\n", err)
			continue
		}

		fmt.Printf("    节点数量: %d\n", len(cfg.Nodes))
		fmt.Printf("    节点列表: %v\n", cfg.Nodes)
		if cfg.Username != "" {
			fmt.Printf("    用户名: %s\n", cfg.Username)
		}
		if cfg.Password != "" {
			fmt.Printf("    密码: %s\n", cfg.Password)
		}
		fmt.Printf("    一致性级别: %s\n", cfg.ConsistencyLevel)
		fmt.Printf("    超时时间: %s\n", cfg.Timeout)
		fmt.Println()
	}
}

func checkDriverRegistration() {
	drivers := sql.Drivers()
	fmt.Printf("  已注册的驱动: %v\n", drivers)

	// 检查我们的驱动是否已注册
	sqliteFound := false
	rqliteFound := false

	for _, driver := range drivers {
		if driver == "sqlite" {
			sqliteFound = true
		}
		if driver == "rqlite" {
			rqliteFound = true
		}
	}

	if sqliteFound {
		fmt.Println("  ✓ sqlite驱动已注册（用于ORM兼容性）")
	}
	if rqliteFound {
		fmt.Println("  ✓ rqlite驱动已注册（用于明确使用）")
	}
}

func demonstrateConnectionConfig() {
	// 演示不同的连接配置
	configs := []struct {
		name string
		dsn  string
		desc string
	}{
		{
			name: "单节点",
			dsn:  "localhost:4001",
			desc: "连接到单个rqlite节点",
		},
		{
			name: "多节点集群",
			dsn:  "node1:4001,node2:4001,node3:4001",
			desc: "连接到rqlite集群，自动选主",
		},
		{
			name: "强一致性",
			dsn:  "localhost:4001?consistency=strong",
			desc: "使用强一致性级别",
		},
		{
			name: "带认证",
			dsn:  "admin:secret@localhost:4001",
			desc: "使用用户名和密码认证",
		},
		{
			name: "完整配置",
			dsn:  "user:pass@node1:4001,node2:4001?consistency=strong&timeout=30s",
			desc: "包含所有配置选项",
		},
	}

	for _, config := range configs {
		fmt.Printf("  %s:\n", config.name)
		fmt.Printf("    DSN: %s\n", config.dsn)
		fmt.Printf("    描述: %s\n", config.desc)

		// 解析配置
		cfg, err := rsqlite.ParseDSN(config.dsn)
		if err != nil {
			fmt.Printf("    错误: %v\n", err)
		} else {
			fmt.Printf("    解析成功: %d个节点, %s一致性, %s超时\n",
				len(cfg.Nodes), cfg.ConsistencyLevel, cfg.Timeout)
		}
		fmt.Println()
	}
}

func demonstrateActualUsage() {
	// 尝试连接到rqlite（如果服务器运行的话）
	db, err := sql.Open("sqlite", "localhost:4001")
	if err != nil {
		fmt.Printf("  连接失败: %v\n", err)
		return
	}
	defer db.Close()

	// 设置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		fmt.Printf("  连接测试失败: %v\n", err)
		fmt.Println("  这通常意味着rqlite服务器没有运行")
		return
	}

	fmt.Println("  ✓ 成功连接到rqlite服务器")

	// 创建示例表
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS demo_users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		fmt.Printf("  创建表失败: %v\n", err)
		return
	}

	fmt.Println("  ✓ 成功创建示例表")

	// 插入数据
	result, err := db.Exec("INSERT INTO demo_users (name, email) VALUES (?, ?)",
		"测试用户", "test@example.com")
	if err != nil {
		fmt.Printf("  插入数据失败: %v\n", err)
		return
	}

	id, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()
	fmt.Printf("  ✓ 成功插入数据: ID=%d, 影响行数=%d\n", id, affected)

	// 查询数据
	rows, err := db.Query("SELECT id, name, email FROM demo_users WHERE id = ?", id)
	if err != nil {
		fmt.Printf("  查询数据失败: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			fmt.Printf("  扫描数据失败: %v\n", err)
			return
		}
		fmt.Printf("  ✓ 查询到数据: ID=%d, 姓名=%s, 邮箱=%s\n", id, name, email)
	}

	// 演示事务（注意：rqlite中事务是无操作的）
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("  开始事务失败: %v\n", err)
		return
	}

	_, err = tx.Exec("INSERT INTO demo_users (name, email) VALUES (?, ?)",
		"事务用户", "tx@example.com")
	if err != nil {
		tx.Rollback()
		fmt.Printf("  事务执行失败: %v\n", err)
		return
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("  事务提交失败: %v\n", err)
		return
	}

	fmt.Println("  ✓ 事务执行成功")
}

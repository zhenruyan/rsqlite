package main

import (
	"fmt"

	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
	"xorm.io/xorm"
	xormlog "xorm.io/xorm/log"
)

func main() {
	fmt.Println("=== XORM 兼容性测试 ===")

	// 连接到rqlite集群
	engine, err := xorm.NewEngine("sqlite", "10.0.1.10:4000")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}

	// 设置日志级别
	engine.SetLogLevel(xormlog.LOG_INFO)
	engine.ShowSQL(true)

	fmt.Println("✓ 成功连接到rqlite集群")

	// 测试基本的数据库操作
	fmt.Println("\n--- 测试基本数据库操作 ---")

	// 1. 测试查询sqlite_master表
	fmt.Println("1. 测试查询sqlite_master表:")
	tables := make([]map[string][]byte, 0)
	err = engine.SQL("SELECT name FROM sqlite_master WHERE type='table'").Find(&tables)
	if err != nil {
		fmt.Printf("   查询sqlite_master失败: %v\n", err)
	} else {
		fmt.Printf("   查询sqlite_master成功，找到 %d 个表\n", len(tables))
		for _, table := range tables {
			fmt.Printf("   表名: %s\n", string(table["name"]))
		}
	}

	// 2. 测试创建简单表
	fmt.Println("\n2. 测试创建简单表:")
	_, err = engine.Exec("CREATE TABLE IF NOT EXISTS test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		fmt.Printf("   创建表失败: %v\n", err)
	} else {
		fmt.Println("   创建表成功")
	}

	// 3. 测试插入数据
	fmt.Println("\n3. 测试插入数据:")
	_, err = engine.Exec("INSERT INTO test_table (name) VALUES (?)", "测试数据")
	if err != nil {
		fmt.Printf("   插入数据失败: %v\n", err)
	} else {
		fmt.Println("   插入数据成功")
	}

	// 4. 测试查询数据
	fmt.Println("\n4. 测试查询数据:")
	results := make([]map[string][]byte, 0)
	err = engine.SQL("SELECT * FROM test_table").Find(&results)
	if err != nil {
		fmt.Printf("   查询数据失败: %v\n", err)
	} else {
		fmt.Printf("   查询数据成功，找到 %d 条记录\n", len(results))
		for _, result := range results {
			fmt.Printf("   ID: %s, Name: %s\n", string(result["id"]), string(result["name"]))
		}
	}

	// 5. 现在测试XORM的Sync2功能
	fmt.Println("\n5. 测试XORM Sync2功能:")

	type SimpleModel struct {
		Id   int64  `xorm:"pk autoincr 'id'"`
		Name string `xorm:"varchar(100) 'name'"`
	}

	err = engine.Sync2(new(SimpleModel))
	if err != nil {
		fmt.Printf("   Sync2失败: %v\n", err)

		// 详细分析错误
		fmt.Println("\n--- 详细错误分析 ---")

		// 测试XORM内部使用的查询
		fmt.Println("测试XORM内部查询:")

		// 尝试直接查询
		rows, err := engine.DB().DB.Query("SELECT name FROM sqlite_master WHERE type='table'")
		if err != nil {
			fmt.Printf("直接查询失败: %v\n", err)
		} else {
			fmt.Println("直接查询成功")
			defer rows.Close()
			for rows.Next() {
				var name string
				err = rows.Scan(&name)
				if err != nil {
					fmt.Printf("扫描失败: %v\n", err)
				} else {
					fmt.Printf("表名: %s\n", name)
				}
			}
		}
	} else {
		fmt.Println("   Sync2成功")
	}

	fmt.Println("\n=== 测试完成 ===")
}

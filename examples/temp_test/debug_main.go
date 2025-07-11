package main

import (
	"database/sql"
	"fmt"

	_ "github.com/zhenruyan/rsqlite" // 导入rqlite驱动
)

func main() {
	fmt.Println("=== 调试 nil 值问题 ===")

	// 连接到rqlite集群
	db, err := sql.Open("sqlite", "10.0.1.10:4000")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("✓ 成功连接到rqlite集群")

	// 测试连接
	err = db.Ping()
	if err != nil {
		fmt.Printf("连接测试失败: %v\n", err)
		fmt.Println("rqlite服务器可能没有运行，但我们可以继续测试驱动的其他功能")
	} else {
		fmt.Println("✓ 连接测试成功")
	}

	// 创建一个包含NULL值的测试表
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS null_test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 插入一些包含NULL值的数据
	_, err = db.Exec("INSERT INTO null_test (name, age) VALUES (?, ?)", "Alice", 25)
	if err != nil {
		fmt.Printf("插入数据1失败: %v\n", err)
		return
	}

	_, err = db.Exec("INSERT INTO null_test (name, age) VALUES (?, ?)", nil, 30)
	if err != nil {
		fmt.Printf("插入数据2失败: %v\n", err)
		return
	}

	_, err = db.Exec("INSERT INTO null_test (name, age) VALUES (?, ?)", "Charlie", nil)
	if err != nil {
		fmt.Printf("插入数据3失败: %v\n", err)
		return
	}

	fmt.Println("✓ 插入测试数据成功")

	// 现在查询数据，看看NULL值如何处理
	fmt.Println("\n--- 查询包含NULL值的数据 ---")
	rows, err := db.Query("SELECT id, name, age FROM null_test")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("获取列名失败: %v\n", err)
		return
	}
	fmt.Printf("列名: %v\n", columns)

	for rows.Next() {
		// 使用sql.NullString和sql.NullInt64来处理NULL值
		var id int64
		var name sql.NullString
		var age sql.NullInt64

		err = rows.Scan(&id, &name, &age)
		if err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}

		fmt.Printf("ID: %d, Name: %v (valid: %t), Age: %v (valid: %t)\n",
			id, name.String, name.Valid, age.Int64, age.Valid)
	}

	// 测试使用interface{}来扫描
	fmt.Println("\n--- 使用interface{}扫描 ---")
	rows2, err := db.Query("SELECT id, name, age FROM null_test")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var id, name, age interface{}
		err = rows2.Scan(&id, &name, &age)
		if err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}

		fmt.Printf("ID: %v (%T), Name: %v (%T), Age: %v (%T)\n",
			id, id, name, name, age, age)
	}

	// 测试查询sqlite_master
	fmt.Println("\n--- 测试查询sqlite_master ---")
	rows3, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		fmt.Printf("查询sqlite_master失败: %v\n", err)
		return
	}
	defer rows3.Close()

	for rows3.Next() {
		var name interface{}
		err = rows3.Scan(&name)
		if err != nil {
			fmt.Printf("扫描失败: %v\n", err)
			continue
		}

		fmt.Printf("表名: %v (%T)\n", name, name)
	}

	fmt.Println("\n=== 调试完成 ===")
}

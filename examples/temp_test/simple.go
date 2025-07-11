package main

import (
	"database/sql"
	"fmt"

	_ "github.com/zhenruyan/rsqlite"
)

func main() {
	fmt.Println("=== 简单查询测试 ===")

	db, err := sql.Open("sqlite", "10.0.1.10:4000")
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer db.Close()

	// 创建表
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS simple_test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}
	fmt.Println("✓ 创建表成功")

	// 插入数据
	result, err := db.Exec("INSERT INTO simple_test (name) VALUES (?)", "test1")
	if err != nil {
		fmt.Printf("插入数据失败: %v\n", err)
		return
	}

	id, _ := result.LastInsertId()
	affected, _ := result.RowsAffected()
	fmt.Printf("✓ 插入数据成功: ID=%d, 影响行数=%d\n", id, affected)

	// 查询数据
	fmt.Println("\n--- 查询数据 ---")
	rows, err := db.Query("SELECT id, name FROM simple_test")
	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("查询执行成功，开始遍历结果...")

	count := 0
	for rows.Next() {
		count++
		var id int64
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			fmt.Printf("扫描第%d行失败: %v\n", count, err)
			continue
		}

		fmt.Printf("第%d行: ID=%d, Name=%s\n", count, id, name)
	}

	if count == 0 {
		fmt.Println("没有查询到任何数据")
	} else {
		fmt.Printf("总共查询到 %d 行数据\n", count)
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		fmt.Printf("遍历过程中发生错误: %v\n", err)
	}

	fmt.Println("\n=== 测试完成 ===")
}

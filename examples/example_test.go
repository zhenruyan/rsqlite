package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/zhenruyan/rsqlite"
)

func ExampleBasicUsage() {
	// 演示DSN解析和驱动注册
	// 注意：这个示例不会实际连接到rqlite服务器

	// 测试DSN解析
	dsn := "10.0.1.10:4000"
	cfg, err := rsqlite.ParseDSN(dsn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("解析的节点数量: %d\n", len(cfg.Nodes))
	fmt.Printf("一致性级别: %s\n", cfg.ConsistencyLevel)
	fmt.Printf("超时时间: %s\n", cfg.Timeout)

	// 演示驱动注册
	drivers := sql.Drivers()
	found := false
	for _, driver := range drivers {
		if driver == "sqlite" {
			found = true
			break
		}
	}

	if found {
		fmt.Println("rqlite驱动已成功注册为sqlite")
	}

	// Output:
	// 解析的节点数量: 3
	// 一致性级别: weak
	// 超时时间: 30s
	// rqlite驱动已成功注册为sqlite
}

func TestDriverRegistration(t *testing.T) {
	// 测试驱动是否正确注册
	drivers := sql.Drivers()
	found := false
	for _, driver := range drivers {
		if driver == "sqlite" {
			found = true
			break
		}
	}
	if !found {
		t.Error("sqlite driver not registered")
	}
}
func BenchmarkQuery(b *testing.B) {
	// 性能测试
	db, err := sql.Open("sqlite", "10.0.1.10:4000")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// 准备测试数据
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS bench_test (id INTEGER PRIMARY KEY, value TEXT)")
	if err != nil {
		b.Fatal(err)
	}

	_, err = db.Exec("INSERT INTO bench_test (value) VALUES (?)", "test_value")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rows, err := db.Query("SELECT id, value FROM bench_test LIMIT 1")
			if err != nil {
				b.Fatal(err)
			}
			for rows.Next() {
				var id int
				var value string
				rows.Scan(&id, &value)
			}
			rows.Close()
		}
	})
}

func TestClusterFailover(t *testing.T) {
	// 测试集群故障转移
	// 这个测试需要实际的rqlite集群才能运行
	db, err := sql.Open("sqlite", "10.0.1.10:4000")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 测试在节点故障时的自动故障转移
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行一些操作来测试故障转移
	for i := 0; i < 10; i++ {
		if err := db.PingContext(ctx); err != nil {
			t.Errorf("Ping failed on iteration %d: %v", i, err)
		}
		time.Sleep(time.Second)
	}
}

func TestXorm(t *testing.T) {
	runXormTest()
}

func TestBun(t *testing.T) {
	runBunTest()
}

func TestGorm(t *testing.T) {
	runGormTest()
}

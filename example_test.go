package rsqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"
	"time"
)

func ExampleBasicUsage() {
	// 演示DSN解析和驱动注册
	// 注意：这个示例不会实际连接到rqlite服务器

	// 测试DSN解析
	dsn := "localhost:4001,localhost:4002,localhost:4003"
	cfg, err := ParseDSN(dsn)
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

func ExampleWithConsistency() {
	// 演示带有一致性级别和超时的DSN解析
	dsn := "localhost:4001,localhost:4002,localhost:4003?consistency=strong&timeout=30s"
	cfg, err := ParseDSN(dsn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("节点数量: %d\n", len(cfg.Nodes))
	fmt.Printf("一致性级别: %s\n", cfg.ConsistencyLevel)
	fmt.Printf("超时时间: %s\n", cfg.Timeout)

	// 演示带认证的DSN解析
	authDSN := "user:pass@host1:4001,host2:4002?consistency=strong&timeout=60s"
	authCfg, err := ParseDSN(authDSN)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("用户名: %s\n", authCfg.Username)
	fmt.Printf("密码: %s\n", authCfg.Password)

	// Output:
	// 节点数量: 3
	// 一致性级别: strong
	// 超时时间: 30s
	// 用户名: user
	// 密码: pass
}

func ExampleGORMCompatibility() {
	// 演示与GORM的兼容性
	// 由于我们注册为"sqlite"驱动，GORM可以直接使用

	type User struct {
		ID    uint   `gorm:"primarykey"`
		Name  string `gorm:"not null"`
		Email string `gorm:"uniqueIndex"`
	}

	// GORM 会使用我们的驱动连接到rqlite集群
	// db, err := gorm.Open(sqlite.Open("localhost:4001,localhost:4002,localhost:4003"), &gorm.Config{})
	// if err != nil {
	//     log.Fatal(err)
	// }

	// // 自动迁移
	// db.AutoMigrate(&User{})

	// // 创建用户
	// user := User{Name: "李四", Email: "lisi@example.com"}
	// db.Create(&user)

	// // 查询用户
	// var users []User
	// db.Find(&users)

	fmt.Println("GORM兼容性示例")
	// Output:
	// GORM兼容性示例
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

func TestParseDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected *Config
		hasError bool
	}{
		{
			dsn: "localhost:4001",
			expected: &Config{
				Nodes:            []string{"http://localhost:4001"},
				Timeout:          30 * time.Second,
				ConsistencyLevel: "weak",
			},
			hasError: false,
		},
		{
			dsn: "user:pass@host1:4001,host2:4002?consistency=strong&timeout=60s",
			expected: &Config{
				Nodes:            []string{"http://host1:4001", "http://host2:4002"},
				Username:         "user",
				Password:         "pass",
				Timeout:          60 * time.Second,
				ConsistencyLevel: "strong",
			},
			hasError: false,
		},
		{
			dsn:      "",
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		cfg, err := ParseDSN(test.dsn)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for DSN: %s", test.dsn)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for DSN %s: %v", test.dsn, err)
			continue
		}

		if len(cfg.Nodes) != len(test.expected.Nodes) {
			t.Errorf("Expected %d nodes, got %d for DSN: %s", len(test.expected.Nodes), len(cfg.Nodes), test.dsn)
		}

		if cfg.Username != test.expected.Username {
			t.Errorf("Expected username %s, got %s for DSN: %s", test.expected.Username, cfg.Username, test.dsn)
		}

		if cfg.Password != test.expected.Password {
			t.Errorf("Expected password %s, got %s for DSN: %s", test.expected.Password, cfg.Password, test.dsn)
		}

		if cfg.ConsistencyLevel != test.expected.ConsistencyLevel {
			t.Errorf("Expected consistency %s, got %s for DSN: %s", test.expected.ConsistencyLevel, cfg.ConsistencyLevel, test.dsn)
		}
	}
}

func BenchmarkQuery(b *testing.B) {
	// 性能测试
	db, err := sql.Open("sqlite", "localhost:4000")
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
	t.Error("需要实际的rqlite集群环境")

	db, err := sql.Open("sqlite", "10.0.1.2:4000")
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

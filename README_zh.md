# rqlite Go 驱动

这是一个用Go语言编写的rqlite数据库驱动，完全兼容Go的`database/sql`接口，支持多节点配置、自动选主，并且完全兼容SQLite语法。驱动注册名为`sqlite`，因此各种ORM框架可以无缝使用。

## 特性

- ✅ **完全兼容 database/sql 接口** - 可以直接替换SQLite使用
- ✅ **多节点支持** - 支持配置多个rqlite节点
- ✅ **自动选主** - 自动发现和连接到集群的leader节点
- ✅ **故障转移** - 当leader节点不可用时自动切换到其他节点
- ✅ **ORM兼容** - 与GORM、XORM等ORM框架无缝集成
- ✅ **连接池支持** - 支持Go标准的数据库连接池
- ✅ **事务支持** - 支持数据库事务（注意：rqlite的事务是无操作的）
- ✅ **参数化查询** - 支持预编译语句和参数绑定
- ✅ **多种一致性级别** - 支持strong、weak、none一致性级别

## 安装

```bash
go get github.com/zhenruyan/rsqlite
```

## 快速开始

### 基本使用

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/zhenruyan/rsqlite" // 导入驱动
)

func main() {
    // 连接到rqlite集群
    db, err := sql.Open("sqlite", "localhost:4001,localhost:4002,localhost:4003")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 测试连接
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    // 创建表
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE
    )`)
    if err != nil {
        log.Fatal(err)
    }

    // 插入数据
    result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        "张三", "zhangsan@example.com")
    if err != nil {
        log.Fatal(err)
    }

    id, _ := result.LastInsertId()
    fmt.Printf("插入用户ID: %d\n", id)

    // 查询数据
    rows, err := db.Query("SELECT id, name, email FROM users WHERE name = ?", "张三")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    for rows.Next() {
        var id int
        var name, email string
        if err := rows.Scan(&id, &name, &email); err != nil {
            log.Fatal(err)
        }
        fmt.Printf("用户: ID=%d, 姓名=%s, 邮箱=%s\n", id, name, email)
    }
}
```

### 使用认证和配置

```go
// 使用用户名密码认证，设置一致性级别和超时
db, err := sql.Open("sqlite", "user:password@host1:4001,host2:4002?consistency=strong&timeout=30s")
```

### 与GORM集成

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    _ "github.com/zhenruyan/rsqlite"
)

type User struct {
    ID    uint   `gorm:"primarykey"`
    Name  string `gorm:"not null"`
    Email string `gorm:"uniqueIndex"`
}

func main() {
    // GORM会自动使用我们的rqlite驱动
    db, err := gorm.Open(sqlite.Open("localhost:4001,localhost:4002,localhost:4003"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // 自动迁移
    db.AutoMigrate(&User{})

    // 创建
    db.Create(&User{Name: "李四", Email: "lisi@example.com"})

    // 查询
    var user User
    db.First(&user, "name = ?", "李四")
}
```

## DSN (数据源名称) 格式

```
[username:password@]host1:port1[,host2:port2,...][?param1=value1&param2=value2]
```

### 参数说明

- `username:password` - 可选的认证信息
- `host:port` - rqlite节点地址，支持多个节点用逗号分隔
- `consistency` - 一致性级别：`strong`、`weak`（默认）、`none`
- `timeout` - 连接超时时间，如：`30s`、`1m`

### DSN 示例

```go
// 单节点，默认配置
"localhost:4001"

// 多节点集群
"node1:4001,node2:4001,node3:4001"

// 使用认证
"admin:password@localhost:4001"

// 强一致性，60秒超时
"localhost:4001?consistency=strong&timeout=60s"

// 完整配置
"user:pass@node1:4001,node2:4001,node3:4001?consistency=strong&timeout=30s"
```

## 一致性级别

rqlite支持三种一致性级别：

- **strong**: 强一致性，读写都通过leader节点，保证线性一致性
- **weak**: 弱一致性（默认），写通过leader，读可能有延迟
- **none**: 无一致性保证，读写都可能不是最新数据，但性能最好

## 故障处理

驱动会自动处理以下故障情况：

1. **Leader选举** - 自动发现新的leader节点
2. **节点故障** - 自动重试其他可用节点
3. **网络分区** - 在网络恢复后自动重连
4. **连接超时** - 支持配置连接和查询超时

## 限制和注意事项

1. **事务支持**: rqlite不支持传统的ACID事务，`Begin()`、`Commit()`、`Rollback()`是无操作的
2. **并发写入**: 只有leader节点可以处理写入操作
3. **SQL兼容性**: 支持SQLite的SQL语法，但某些高级特性可能不可用
4. **连接管理**: 建议使用连接池来管理数据库连接

## 性能建议

1. **使用连接池**: 通过`SetMaxOpenConns()`和`SetMaxIdleConns()`配置连接池
2. **批量操作**: 使用事务将多个操作组合在一起
3. **合适的一致性级别**: 根据业务需求选择合适的一致性级别
4. **预编译语句**: 对于重复执行的查询使用`Prepare()`

```go
// 配置连接池
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// 批量插入示例
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

for i := 0; i < 1000; i++ {
    _, err = tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        fmt.Sprintf("用户%d", i), fmt.Sprintf("user%d@example.com", i))
    if err != nil {
        tx.Rollback()
        log.Fatal(err)
    }
}

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

## 示例和测试

项目包含完整的示例代码和测试用例：

```bash
# 运行测试
go test -v .

# 运行示例
cd examples && go run basic_usage.go
```

## 贡献

欢迎贡献代码！请确保：

1. 代码通过所有测试
2. 遵循Go代码规范
3. 添加适当的测试用例
4. 更新相关文档

## 许可证

本项目采用MIT许可证。详情请参阅LICENSE文件。 
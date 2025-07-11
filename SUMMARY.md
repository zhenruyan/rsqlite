# rqlite Go驱动 - 项目总结

## 项目概述

这是一个用Go语言编写的rqlite数据库驱动程序，完全兼容Go的`database/sql`接口。该驱动支持多节点配置、自动选主、故障转移，并且注册为"sqlite"驱动名称，因此可以与各种ORM框架（如GORM、XORM等）无缝集成。

## 主要特性

✅ **完全兼容 database/sql 接口** - 可以直接替换SQLite使用  
✅ **多节点支持** - 支持配置多个rqlite节点  
✅ **自动选主** - 自动发现和连接到集群的leader节点  
✅ **故障转移** - 当leader节点不可用时自动切换到其他节点  
✅ **ORM兼容** - 与GORM、XORM等ORM框架无缝集成  
✅ **连接池支持** - 支持Go标准的数据库连接池  
✅ **事务支持** - 支持数据库事务（注意：rqlite的事务是无操作的）  
✅ **参数化查询** - 支持预编译语句和参数绑定  
✅ **多种一致性级别** - 支持strong、weak、none一致性级别  

## 项目结构

```
rsqlite/
├── go.mod              # 项目模块定义
├── go.sum              # 依赖版本锁定
├── README.md           # 项目文档
├── SUMMARY.md          # 项目总结
├── driver.go           # 主驱动实现
├── connection.go       # 连接管理
├── leader.go           # 选主算法
├── result.go           # 结果处理
├── statement.go        # 语句处理
├── example_test.go     # 测试用例
└── examples/           # 使用示例
    ├── go.mod          # 示例模块定义
    └── basic_usage.go  # 基本使用示例
```

## 核心组件

### 1. 驱动注册 (driver.go)
- 实现`database/sql/driver.Driver`接口
- 注册为"sqlite"和"rqlite"两个驱动名称
- DSN解析支持多种格式

### 2. 连接管理 (connection.go)
- 实现`database/sql/driver.Conn`接口
- 支持连接池
- 自动重连和故障转移

### 3. 选主算法 (leader.go)
- 基于健康检查的节点选择
- 响应时间加权算法
- 自动故障检测和切换

### 4. 结果处理 (result.go)
- 实现`database/sql/driver.Rows`和`database/sql/driver.Result`接口
- 类型转换和数据映射
- 兼容gorqlite API

### 5. 语句处理 (statement.go)
- 实现`database/sql/driver.Stmt`接口
- 参数绑定支持
- 预编译语句（无操作实现）

## DSN格式支持

```
[username:password@]host1:port1[,host2:port2,...][?param1=value1&param2=value2]
```

### 支持的参数
- `consistency`: 一致性级别（strong、weak、none）
- `timeout`: 连接超时时间

### 示例
```go
// 单节点
"localhost:4001"

// 多节点集群
"node1:4001,node2:4001,node3:4001"

// 带认证
"user:pass@localhost:4001"

// 完整配置
"user:pass@node1:4001,node2:4001?consistency=strong&timeout=30s"
```

## 使用方法

### 基本使用
```go
import (
    "database/sql"
    _ "github.com/free/rsqlite"
)

db, err := sql.Open("sqlite", "localhost:4001,localhost:4002,localhost:4003")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 正常的数据库操作
rows, err := db.Query("SELECT * FROM users")
```

### 与GORM集成
```go
import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    _ "github.com/free/rsqlite"
)

db, err := gorm.Open(sqlite.Open("localhost:4001,localhost:4002"), &gorm.Config{})
```

## 测试覆盖

- ✅ 驱动注册测试
- ✅ DSN解析测试
- ✅ 连接配置测试
- ✅ 示例代码测试
- ⚠️ 集群故障转移测试（需要实际rqlite集群）

## 性能考虑

1. **连接池配置**: 推荐设置适当的连接池大小
2. **一致性级别**: 根据业务需求选择合适的一致性级别
3. **批量操作**: 使用事务进行批量操作
4. **预编译语句**: 对重复查询使用预编译语句

## 限制和注意事项

1. **事务支持**: rqlite不支持传统的ACID事务
2. **并发写入**: 只有leader节点可以处理写入操作
3. **SQL兼容性**: 支持SQLite语法，但某些高级特性可能不可用
4. **网络依赖**: 需要稳定的网络连接到rqlite集群

## 依赖项

- `github.com/rqlite/gorqlite`: rqlite官方Go客户端库
- Go标准库: `database/sql`, `context`, `net/http`等

## 编译和运行

```bash
# 编译项目
go build -v .

# 运行测试
go test -v .

# 运行示例
cd examples && go run basic_usage.go
```

## 贡献指南

1. 确保代码通过所有测试
2. 添加适当的注释和文档
3. 遵循Go代码规范
4. 提交前运行`go fmt`和`go vet`

## 许可证

本项目采用开源许可证，具体请参考LICENSE文件。

## 联系信息

如有问题或建议，请通过GitHub Issues联系。 
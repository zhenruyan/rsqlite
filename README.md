# rqlite Go Driver

A Go database driver for rqlite, fully compatible with Go's `database/sql` interface. It supports multi-node configuration, automatic leader discovery, and is fully compatible with SQLite syntax. The driver is registered as `sqlite`, making it seamlessly compatible with various ORM frameworks.

[中文文档](README_zh.md) | English

## Features

- ✅ **Full database/sql compatibility** - Drop-in replacement for SQLite
- ✅ **Multi-node support** - Configure multiple rqlite nodes
- ✅ **Automatic leader discovery** - Automatically discover and connect to cluster leader
- ✅ **Fault tolerance** - Automatic failover when leader becomes unavailable
- ✅ **ORM compatibility** - Seamless integration with GORM, XORM, and other ORM frameworks
- ✅ **Connection pooling** - Support for Go's standard database connection pool
- ✅ **Transaction support** - Database transaction support (note: rqlite transactions are no-ops)
- ✅ **Parameterized queries** - Support for prepared statements and parameter binding
- ✅ **Multiple consistency levels** - Support for strong, weak, and none consistency levels

## Installation

```bash
go get github.com/zhenruyan/rsqlite
```

## Quick Start

### Basic Usage

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/zhenruyan/rsqlite" // Import driver
)

func main() {
    // Connect to rqlite cluster
    db, err := sql.Open("sqlite", "localhost:4001,localhost:4002,localhost:4003")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Test connection
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    // Create table
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE
    )`)
    if err != nil {
        log.Fatal(err)
    }

    // Insert data
    result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        "John Doe", "john@example.com")
    if err != nil {
        log.Fatal(err)
    }

    id, _ := result.LastInsertId()
    fmt.Printf("Inserted user ID: %d\n", id)

    // Query data
    rows, err := db.Query("SELECT id, name, email FROM users WHERE name = ?", "John Doe")
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
        fmt.Printf("User: ID=%d, Name=%s, Email=%s\n", id, name, email)
    }
}
```

### Using Authentication and Configuration

```go
// Use username/password authentication, set consistency level and timeout
db, err := sql.Open("sqlite", "user:password@host1:4001,host2:4002?consistency=strong&timeout=30s")
```

### GORM Integration

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
    // GORM will automatically use our rqlite driver
    db, err := gorm.Open(sqlite.Open("localhost:4001,localhost:4002,localhost:4003"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // Auto migrate
    db.AutoMigrate(&User{})

    // Create
    db.Create(&User{Name: "Jane Smith", Email: "jane@example.com"})

    // Query
    var user User
    db.First(&user, "name = ?", "Jane Smith")
}
```

## DSN (Data Source Name) Format

```
[username:password@]host1:port1[,host2:port2,...][?param1=value1&param2=value2]
```

### Parameters

- `username:password` - Optional authentication credentials
- `host:port` - rqlite node addresses, multiple nodes separated by commas
- `consistency` - Consistency level: `strong`, `weak` (default), `none`
- `timeout` - Connection timeout, e.g., `30s`, `1m`

### DSN Examples

```go
// Single node, default configuration
"localhost:4001"

// Multi-node cluster
"node1:4001,node2:4001,node3:4001"

// With authentication
"admin:password@localhost:4001"

// Strong consistency, 60 second timeout
"localhost:4001?consistency=strong&timeout=60s"

// Full configuration
"user:pass@node1:4001,node2:4001,node3:4001?consistency=strong&timeout=30s"
```

## Consistency Levels

rqlite supports three consistency levels:

- **strong**: Strong consistency, all reads and writes go through leader node, guarantees linearizability
- **weak**: Weak consistency (default), writes go through leader, reads may have slight delay
- **none**: No consistency guarantee, reads and writes may not reflect latest data, but best performance

## Fault Handling

The driver automatically handles the following fault scenarios:

1. **Leader election** - Automatically discover new leader nodes
2. **Node failures** - Automatically retry with other available nodes
3. **Network partitions** - Automatically reconnect after network recovery
4. **Connection timeouts** - Support for configurable connection and query timeouts

## Limitations and Notes

1. **Transaction support**: rqlite doesn't support traditional ACID transactions, `Begin()`, `Commit()`, `Rollback()` are no-ops
2. **Concurrent writes**: Only the leader node can handle write operations
3. **SQL compatibility**: Supports SQLite SQL syntax, but some advanced features may not be available
4. **Connection management**: Recommended to use connection pooling for database connections

## Performance Recommendations

1. **Use connection pooling**: Configure connection pool via `SetMaxOpenConns()` and `SetMaxIdleConns()`
2. **Batch operations**: Use transactions to group multiple operations together
3. **Appropriate consistency level**: Choose the right consistency level based on business requirements
4. **Prepared statements**: Use `Prepare()` for repeatedly executed queries

```go
// Configure connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Batch insert example
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

for i := 0; i < 1000; i++ {
    _, err = tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 
        fmt.Sprintf("User%d", i), fmt.Sprintf("user%d@example.com", i))
    if err != nil {
        tx.Rollback()
        log.Fatal(err)
    }
}

if err := tx.Commit(); err != nil {
    log.Fatal(err)
}
```

## Examples and Testing

The project includes complete example code and test cases:

```bash
# Run tests
go test -v .

# Run examples
cd examples && go run basic_usage.go
```

## Contributing

Contributions are welcome! Please ensure:

1. Code passes all tests
2. Follow Go coding conventions
3. Add appropriate test cases
4. Update relevant documentation

## License

This project is licensed under the MIT License. See the LICENSE file for details. 
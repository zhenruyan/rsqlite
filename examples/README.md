# rqlite Go驱动 ORM兼容性测试

这个目录包含了各种主流Go ORM框架与rqlite驱动的兼容性测试示例。

## 测试文件说明

### 1. 基本测试
- **basic_usage.go** - 基本的database/sql接口测试，演示DSN解析、驱动注册等功能

### 2. GORM测试
- **gorm_test.go** - GORM ORM框架兼容性测试

### 3. Bun测试  
- **bun_test.go** - Bun ORM框架兼容性测试

### 4. XORM测试
- **xorm_test.go** - XORM框架兼容性测试

### 5. SQL Builder测试
- **sqlbuilder_test.go** - 自定义SQL查询构建器测试，无需外部依赖

## 运行测试

### 基本测试（无需额外依赖）
```bash
cd examples
go run basic_usage.go
```

### GORM测试
首先安装GORM依赖：
```bash
cd examples
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

然后运行测试：
```bash
go run gorm_test.go
```

### Bun测试
首先安装Bun依赖：
```bash
cd examples
go get github.com/uptrace/bun
go get github.com/uptrace/bun/dialect/sqlitedialect
go get github.com/uptrace/bun/extra/bundebug
```

然后运行测试：
```bash
go run bun_test.go
```

### XORM测试
首先安装XORM依赖：
```bash
cd examples
go get xorm.io/xorm
```

然后运行测试：
```bash
go run xorm_test.go
```

### SQL Builder测试（无需额外依赖）
```bash
cd examples
# 作为库使用，查看代码示例
cat sqlbuilder_test.go

# 或者在你的代码中调用
# runSQLBuilderTest()
```

## 测试功能

每个ORM测试都包含以下功能验证：

### 1. 基本CRUD操作
- Create - 创建记录
- Read - 查询记录
- Update - 更新记录  
- Delete - 删除记录

### 2. 关联查询
- 一对多关系
- 预加载查询
- JOIN查询

### 3. 事务支持
- 事务开始、提交、回滚
- 注意：rqlite不支持真正的ACID事务

### 4. 批量操作
- 批量插入
- 批量查询
- 批量更新
- 聚合查询

## 注意事项

1. **rqlite服务器**: 测试需要rqlite服务器运行在localhost:4001，如果没有运行服务器，测试会显示连接错误但DSN解析仍然正常。

2. **事务限制**: rqlite不支持传统的ACID事务，所有事务操作都是无操作的(no-op)，但ORM的事务接口仍然可以使用。

3. **一致性级别**: 可以通过DSN参数配置一致性级别：
   - `strong` - 强一致性
   - `weak` - 弱一致性（默认）
   - `none` - 无一致性保证

4. **多节点支持**: 所有测试都支持多节点配置，例如：
   ```
   localhost:4001,localhost:4002,localhost:4003
   ```

## DSN格式

```
[username:password@]host1:port1[,host2:port2,...][?param1=value1&param2=value2]
```

### 示例
```bash
# 单节点
localhost:4001

# 多节点集群
localhost:4001,localhost:4002,localhost:4003

# 带认证
user:pass@localhost:4001

# 强一致性
localhost:4001?consistency=strong&timeout=30s
```

## 故障排除

### 依赖问题
如果遇到导入错误，请确保已安装相应的ORM依赖包。

### 连接问题
如果看到连接错误，这是正常的，因为测试假设rqlite服务器运行在localhost:4001。你可以：

1. 启动rqlite服务器：
   ```bash
   rqlited ~/node.1
   ```

2. 或者修改测试文件中的DSN连接字符串指向你的rqlite实例。

### 编译问题
确保你在examples目录中运行测试，并且已经正确设置了Go模块路径。

## 扩展测试

你可以基于这些示例创建自己的测试：

1. 复制现有的测试文件
2. 修改模型结构和测试逻辑
3. 添加你需要的特定测试场景

## 贡献

欢迎提交新的ORM框架兼容性测试！请确保：

1. 遵循现有的代码结构
2. 包含完整的CRUD测试
3. 添加相应的文档说明
4. 测试通过编译（在安装依赖的情况下） 
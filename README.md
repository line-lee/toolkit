# 分表工具包 (Sharding Toolkit)

一个基于时间间隔的 MySQL 分表管理工具包，支持按小时、天、月、年自动创建和管理分表。

## 功能特性

- 🕐 支持多种时间粒度分表（小时、天、月、年）
- 🔒 基于 Redis 的分布式锁，避免并发建表冲突
- 💾 内存缓存机制，减少数据库元数据查询
- 🏗️ 构建器模式，提供流畅的 API 接口
- 📊 查询参数自动生成，支持时间范围查询
- 🔧 自动表结构复制，确保分表结构一致

## 注意事项

1. **MySQL 连接必需** - 用于分表存在性检查和自动创建分表结构
2. **基础表必须存在** - 工具会根据基础表结构创建分表，请确保基础表已提前创建
3. **Redis 连接必需** - 用于分布式锁，避免并发建表冲突
4. **时间精度** - 确保传入的时间参数与时区设置一致

## 安装使用

### 安装依赖
```bash
go get github.com/line-lee/toolkit
```

### 快速开始

#### 1. 初始化分表配置

```go
package main

import (
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/line-lee/toolkit/sharding"
	"log"
	"time"
)

func main() {
	// 初始化数据库连接
	mysqlClient, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/my_database?parseTime=true")
	if err != nil {
		log.Fatal("MySQL连接失败:", err)
	}
	defer mysqlClient.Close()

	// 初始化Redis连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // 如果没有密码
		DB:       0,  // 默认数据库
	})
	defer redisClient.Close()

	// 创建分表配置
	builder := sharding.TableBuilder().
		MysqlClient(mysqlClient).
		RedisClient(redisClient).
		DBName("my_database").
		Primary("user_logs").
		ThisTime(time.Now()).
		Type(sharding.Day)
	// 获取分表名（自动创建不存在的表）
	tableName, err := sharding.New(builder).GetTableName()
	if err != nil {
		panic(err)
	}
	fmt.Printf("当前分表: %s\n", tableName)
}
```

#### 2. 生成查询参数
```go
// 生成时间范围查询的分表参数
paramsBuilder := sharding.ParamsBuilder().
    Primary("user_logs").
    Start(time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)).
    End(time.Date(2025, 8, 3, 0, 0, 0, 0, time.UTC)).
    IsEndClose(false).
    Type(sharding.Day)

results, err := sharding.Params(paramsBuilder)
if err != nil {
    panic(err)
}

for _, result := range results {
    fmt.Printf("表名: %s, 时间范围: %s - %s\n", 
        result.TableName, 
        result.Start.Format("2006-01-02"),
        result.End.Format("2006-01-02"))
}
```

## API 参考

### TableBuilder 方法
- `MysqlClient(*sql.DB)` - 设置 MySQL 客户端
- `RedisClient(*redis.Client)` - 设置 Redis 客户端
- `DBName(string)` - 设置数据库名
- `Primary(string)` - 设置基础表名
- `ThisTime(time.Time)` - 设置当前时间
- `Type(Type)` - 设置分表类型

### ParamsBuilder 方法
- `Primary(string)` - 设置基础表名
- `Start(time.Time)` - 设置查询开始时间
- `End(time.Time)` - 设置查询结束时间
- `IsEndClose(bool)` - 设置是否包含结束时间
- `Type(Type)` - 设置分表类型

## 示例输出

### 分表命名示例
- **Hour**: `user_logs_2025082115` (2025年8月21日15时)
- **Day**: `user_logs_20250821` (2025年8月21日)
- **Month**: `user_logs_202508` (2025年8月)
- **Year**: `user_logs_2025` (2025年)

### 查询参数示例
输入时间范围: `2025-08-01` 到 `2025-08-03`
输出分表查询参数:
```
表名: user_logs_20250801, 时间范围: 2025-08-01 - 2025-08-02
表名: user_logs_20250802, 时间范围: 2025-08-02 - 2025-08-03
```



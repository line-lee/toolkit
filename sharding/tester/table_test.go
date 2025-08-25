package tester

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/line-lee/toolkit/sharding"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"testing"
	"time"
)

// 使用 TestContainers 启动 MySQL 容器
func setupMysql(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	ctr, err := mysql.Run(ctx, "mysql:8.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	connectionString, err := ctr.ConnectionString(ctx, "tls=skip-verify")
	require.NoError(t, err)
	db, err := sql.Open("mysql", connectionString)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		err = db.Ping()
		return err == nil
	}, 30*time.Second, 1*time.Second)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	t.Logf("Mysql container started\n")
	return db
}

// 使用 TestContainers 启动 Redis 容器
func setupRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	// 创建 Redis 容器请求
	redisContainer, err := tcredis.Run(ctx, "redis:7")
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)
	// 获取 Redis 连接地址
	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	// 创建 Redis 客户端
	options, err := redis.ParseURL(uri)
	require.NoError(t, err)
	client := redis.NewClient(options)
	// 等待 Redis 真正可用
	require.Eventually(t, func() bool {
		_, err := client.Ping(ctx).Result()
		return err == nil
	}, 30*time.Second, 1*time.Second)
	t.Logf("Redis container started\n")
	return client
}

func TestGetTableName(t *testing.T) {
	mysqlClient, redisClient := setupMysql(t), setupRedis(t)
	var dbName = "test"
	var primary = "user_day"
	t.Run("完整功能测试", func(t *testing.T) {
		// 初始化基础表信息
		creatSql := fmt.Sprintf("CREATE TABLE `test`.`%s` ("+
			"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',"+
			"  `user_id` bigint(20) unsigned NOT NULL COMMENT '用户ID',"+
			"  `day_date` date NOT NULL COMMENT '统计日期',"+
			"  `login_count` int(11) NOT NULL DEFAULT '0' COMMENT '当日登录次数',"+
			"  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',"+
			"  PRIMARY KEY (`id`),"+
			"  UNIQUE KEY `uk_user_date` (`user_id`,`day_date`),"+
			"  KEY `idx_date` (`day_date`)"+
			") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户每日统计表';", primary)

		_, err := mysqlClient.Exec(creatSql)
		require.NoError(t, err, creatSql)

		testCases := []struct {
			name      string
			time      time.Time
			shardType sharding.Type
			expected  string
		}{
			{
				name:      "按小时分表",
				time:      time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC),
				shardType: sharding.Hour,
				expected:  "user_day_2024081514",
			},
			{
				name:      "按天分表",
				time:      time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC),
				shardType: sharding.Day,
				expected:  "user_day_20240815",
			},
			{
				name:      "按月分表",
				time:      time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC),
				shardType: sharding.Month,
				expected:  "user_day_202408",
			},
			{
				name:      "按年分表",
				time:      time.Date(2024, 8, 15, 14, 30, 0, 0, time.UTC),
				shardType: sharding.Year,
				expected:  "user_day_2024",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := sharding.TableBuilder().
					MysqlClient(mysqlClient).
					RedisClient(redisClient).
					DBName(dbName).
					Primary(primary).
					ThisTime(tc.time).
					Type(tc.shardType)

				tableOption := sharding.New(builder)
				tableName, err := tableOption.GetTableName()
				require.NoError(t, err)

				// 功能实现
				require.Equal(t, tableName, tc.expected)
				// 分表数据增加
				insertSql := fmt.Sprintf("INSERT INTO `test`.`%s` (`user_id`, `day_date`, `login_count`) VALUES (12345, '2024-01-15', 3)", tableName)
				_, err = mysqlClient.Exec(insertSql)
				require.NoError(t, err, insertSql)
				// 分表数据查询
				var id, userId int64
				querySql := fmt.Sprintf("SELECT id,user_id FROM `test`.`%s` WHERE user_id = 12345", tableName)
				err = mysqlClient.QueryRow(querySql).Scan(&id, &userId)
				require.NoError(t, err, querySql)
				require.Equal(t, int64(12345), userId)
				require.NotZero(t, id)
			})
		}
	})

	t.Run("TableBuilder参数测试", func(t *testing.T) {
		t.Run("缺少必填参数-MysqlClient", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				RedisClient(redisClient).
				DBName(dbName).
				Primary(primary).
				ThisTime(thisTime).
				Type(sharding.Day)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithMysqlClient 必填")
		})

		t.Run("缺少必填参数-RedisClient", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				MysqlClient(new(sql.DB)).
				DBName(dbName).
				Primary(primary).
				ThisTime(thisTime).
				Type(sharding.Day)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithRedisClient 必填")
		})

		t.Run("缺少必填参数-DBName", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				MysqlClient(mysqlClient).
				RedisClient(redisClient).
				Primary(primary).
				ThisTime(thisTime).
				Type(sharding.Day)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithDBName 必填")
		})

		t.Run("缺少必填参数-Primary", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				MysqlClient(mysqlClient).
				RedisClient(redisClient).
				DBName(dbName).
				ThisTime(thisTime).
				Type(sharding.Day)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithPrimary 必填")
		})

		t.Run("缺少必填参数-ThisTime", func(t *testing.T) {
			builder := sharding.TableBuilder().
				MysqlClient(mysqlClient).
				RedisClient(redisClient).
				DBName(dbName).
				Primary(primary).
				Type(sharding.Day)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithThisTime 必填")
		})

		t.Run("缺少必填参数-Type", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				MysqlClient(mysqlClient).
				RedisClient(redisClient).
				DBName(dbName).
				Primary(primary).
				ThisTime(thisTime)
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "option WithThisTime 必填")
		})

		t.Run("未知分表类型", func(t *testing.T) {
			thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
			builder := sharding.TableBuilder().
				MysqlClient(mysqlClient).
				RedisClient(redisClient).
				DBName(dbName).
				Primary(primary).
				ThisTime(thisTime).
				Type(999) // 未知类型
			tableOption := sharding.New(builder)
			_, err := tableOption.GetTableName()
			require.Error(t, err)
			require.Contains(t, err.Error(), "分表类型不识别")
		})
	})
}

// TestGetTableNameEdgeCases 测试GetTableName方法的边界情况
func TestGetTableNameEdgeCases(t *testing.T) {
	mysqlClient, redisClient := setupMysql(t), setupRedis(t)

	// 创建数据库
	_, err := mysqlClient.Exec("CREATE DATABASE IF NOT EXISTS test")
	require.NoError(t, err)

	t.Run("基础表不存在", func(t *testing.T) {
		thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
		builder := sharding.TableBuilder().
			MysqlClient(mysqlClient).
			RedisClient(redisClient).
			DBName("test").
			Primary("nonexistent_table").
			ThisTime(thisTime).
			Type(sharding.Day)

		tableOption := sharding.New(builder)
		_, err := tableOption.GetTableName()
		// 基础表不存在应该报错
		require.Error(t, err)
	})

	t.Run("数据库不存在", func(t *testing.T) {
		// 先创建基础表
		createTableSQL := "CREATE TABLE `test`.`test_table` (`id` INT PRIMARY KEY, `name` VARCHAR(50))"
		_, err := mysqlClient.Exec(createTableSQL)
		require.NoError(t, err)

		thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
		builder := sharding.TableBuilder().
			MysqlClient(mysqlClient).
			RedisClient(redisClient).
			DBName("nonexistent_db").
			Primary("test_table").
			ThisTime(thisTime).
			Type(sharding.Day)

		tableOption := sharding.New(builder)
		_, err = tableOption.GetTableName()
		require.Error(t, err)
		// 数据库不存在应该报错
	})

	t.Run("表名包含特殊字符", func(t *testing.T) {
		// 创建包含特殊字符的表名
		specialTableName := "test-table_with.special$chars"
		createTableSQL := fmt.Sprintf("CREATE TABLE `test`.`%s` (`id` INT PRIMARY KEY, `name` VARCHAR(50))", specialTableName)
		_, err := mysqlClient.Exec(createTableSQL)
		require.NoError(t, err)

		thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
		builder := sharding.TableBuilder().
			MysqlClient(mysqlClient).
			RedisClient(redisClient).
			DBName("test").
			Primary(specialTableName).
			ThisTime(thisTime).
			Type(sharding.Day)

		tableOption := sharding.New(builder)
		tableName, err := tableOption.GetTableName()
		require.NoError(t, err)
		expectedName := fmt.Sprintf("%s_%s", specialTableName, thisTime.Format("20060102"))
		require.Equal(t, expectedName, tableName)
	})

	t.Run("极端时间值", func(t *testing.T) {
		// 创建基础表
		createTableSQL := "CREATE TABLE `test`.`extreme_time_table` (`id` INT PRIMARY KEY, `name` VARCHAR(50))"
		_, err := mysqlClient.Exec(createTableSQL)
		require.NoError(t, err)

		testCases := []struct {
			name           string
			time           time.Time
			shardType      sharding.Type
			expectedSuffix string
		}{
			{
				name:           "最小时间值",
				time:           time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
				shardType:      sharding.Year,
				expectedSuffix: "1970",
			},
			{
				name:           "未来时间",
				time:           time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC),
				shardType:      sharding.Month,
				expectedSuffix: "209912",
			},
			{
				name:           "闰年2月29日",
				time:           time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC),
				shardType:      sharding.Day,
				expectedSuffix: "20240229",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				builder := sharding.TableBuilder().
					MysqlClient(mysqlClient).
					RedisClient(redisClient).
					DBName("test").
					Primary("extreme_time_table").
					ThisTime(tc.time).
					Type(tc.shardType)

				tableOption := sharding.New(builder)
				tableName, err := tableOption.GetTableName()
				require.NoError(t, err)
				expectedName := fmt.Sprintf("extreme_time_table_%s", tc.expectedSuffix)
				require.Equal(t, expectedName, tableName)
			})
		}
	})

	t.Run("并发创建同一分表", func(t *testing.T) {
		// 创建基础表
		createTableSQL := "CREATE TABLE `test`.`concurrent_table` (`id` INT PRIMARY KEY, `name` VARCHAR(50))"
		_, err := mysqlClient.Exec(createTableSQL)
		require.NoError(t, err)

		thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)

		// 启动多个goroutine并发创建同一分表
		resultChan := make(chan struct {
			tableName string
			err       error
		}, 3)

		for i := 0; i < 3; i++ {
			go func() {
				builder := sharding.TableBuilder().
					MysqlClient(mysqlClient).
					RedisClient(redisClient).
					DBName("test").
					Primary("concurrent_table").
					ThisTime(thisTime).
					Type(sharding.Day)

				tableOption := sharding.New(builder)
				tableName, err := tableOption.GetTableName()
				resultChan <- struct {
					tableName string
					err       error
				}{tableName, err}
			}()
		}

		// 收集所有结果
		var results []struct {
			tableName string
			err       error
		}
		for i := 0; i < 3; i++ {
			result := <-resultChan
			results = append(results, result)
		}

		// 验证结果
		expectedTableName := "concurrent_table_20240815"
		successCount := 0
		for _, result := range results {
			if result.err == nil {
				require.Equal(t, expectedTableName, result.tableName)
				successCount++
			}
		}

		// 至少应该有一个成功创建
		require.Greater(t, successCount, 0)
	})
}

// TestTableCreationWithComplexStructure 测试复杂表结构的分表创建
func TestTableCreationWithComplexStructure(t *testing.T) {
	mysqlClient, redisClient := setupMysql(t), setupRedis(t)

	// 创建数据库
	_, err := mysqlClient.Exec("CREATE DATABASE IF NOT EXISTS test")
	require.NoError(t, err)

	t.Run("复杂表结构复制", func(t *testing.T) {
		// 创建包含多种数据类型和约束的复杂表
		complexTableSQL := `CREATE TABLE test.complex_table (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
			user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
			username VARCHAR(50) NOT NULL COMMENT '用户名',
			email VARCHAR(100) UNIQUE COMMENT '邮箱',
			balance DECIMAL(10,2) DEFAULT 0.00 COMMENT '余额',
			status TINYINT DEFAULT 1 COMMENT '状态',
			metadata JSON COMMENT 'JSON数据',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			PRIMARY KEY (id),
			UNIQUE KEY uk_username (username),
			KEY idx_user_id (user_id),
			KEY idx_status_created (status, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='复杂用户表'`

		_, err := mysqlClient.Exec(complexTableSQL)
		require.NoError(t, err)

		thisTime := time.Date(2024, 8, 15, 10, 30, 0, 0, time.UTC)
		builder := sharding.TableBuilder().
			MysqlClient(mysqlClient).
			RedisClient(redisClient).
			DBName("test").
			Primary("complex_table").
			ThisTime(thisTime).
			Type(sharding.Month)

		tableOption := sharding.New(builder)
		tableName, err := tableOption.GetTableName()
		require.NoError(t, err)
		require.Equal(t, "complex_table_202408", tableName)

		// 验证分表是否成功创建且结构正确
		var count int
		err = mysqlClient.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'test' AND table_name = ?", tableName).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, "分表应该被成功创建")

		// 验证分表结构
		rows, err := mysqlClient.Query("DESCRIBE test." + tableName)
		require.NoError(t, err)
		defer rows.Close()

		columnCount := 0
		for rows.Next() {
			columnCount++
		}
		require.Greater(t, columnCount, 8, "分表应该包含所有原表的列")
	})
}

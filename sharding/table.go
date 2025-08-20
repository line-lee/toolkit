package sharding

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/line-lee/toolkit/stringkit"
	"log"
	"strings"
	"time"
)

// New 分表初始化对象
// ops 建表参数
func New(builder *TableOptionsBuilder) *TableOption {
	option := new(TableOption)
	for _, op := range builder.funcs {
		op(option)
	}
	if option.mysqlClient == nil {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithMysqlClient 必填")}
	}
	if option.redisClient == nil {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithRedisClient 必填")}
	}
	if stringkit.IsBlank(option.db) {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithDBName 必填")}
	}
	if stringkit.IsBlank(option.primary) {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithPrimary 必填")}
	}
	if option.thisTime.IsZero() {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithThisTime 必填")}
	}
	if option.t == 0 {
		return &TableOption{err: errors.New("分表初始化对象,New()参数中， option WithThisTime 必填")}
	}
	var suffix string
	switch option.t {
	// 2006-01-02 15:04:05
	case Hour:
		suffix = option.thisTime.Format("2006010215")
	case Day:
		suffix = option.thisTime.Format("20060102")
	case Month:
		suffix = option.thisTime.Format("200601")
	case Year:
		suffix = option.thisTime.Format("2006")
	default:
		return &TableOption{err: fmt.Errorf("mysql分表，分表类型不识别，shard type %d", option.t)}
	}
	option.expect = fmt.Sprintf("%s_%s", option.primary, suffix)
	return option
}

type TableOption struct {
	// 数据库连接，用于分表检查和创建分表
	mysqlClient *sql.DB
	// redis连接，用于建表分布式锁，避免并发异常
	redisClient *redis.Client
	// db 库名
	db string
	// primary 初始表名
	primary string
	// 当前时间
	thisTime time.Time
	// 分表类型
	t Type

	// expect 分表名
	expect string
	// 错误传递
	err error
}

type TableOptionsBuilder struct {
	funcs []TableOptionFunc
}

func TableBuilder() *TableOptionsBuilder {
	return &TableOptionsBuilder{}
}

type TableOptionFunc func(*TableOption)

func (tb *TableOptionsBuilder) MysqlClient(mysqlClient *sql.DB) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.mysqlClient = mysqlClient
	})
	return tb
}

func (tb *TableOptionsBuilder) RedisClient(client *redis.Client) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.redisClient = client
	})
	return tb
}

func (tb *TableOptionsBuilder) DBName(dbName string) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.db = dbName
	})
	return tb
}

func (tb *TableOptionsBuilder) Primary(primary string) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.primary = primary
	})
	return tb
}

func (tb *TableOptionsBuilder) ThisTime(thisTime time.Time) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.thisTime = thisTime
	})
	return tb
}

func (tb *TableOptionsBuilder) Type(t Type) *TableOptionsBuilder {
	tb.funcs = append(tb.funcs, func(opt *TableOption) {
		opt.t = t
	})
	return tb
}

// 内存中查询存在的数据库表
var tm = make(map[string]bool)

func (to *TableOption) GetTableName() (string, error) {
	if to.err != nil {
		return "", to.err
	}
	if !to.lock() {
		log.Printf("mysql分表，表存在检查，分布式锁获取失败，able:%s,db:%s\n", to.expect, to.db)
		return "", errors.New("mysql分表，表存在检查，分布式锁获取失败")
	}
	defer to.unlock()
	if isExist := tm[to.expect]; !isExist {
		// 内存不存在，继续查库
		tableCheckSql := fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = ?", to.db)
		var tableName string
		err := to.mysqlClient.QueryRow(tableCheckSql, to.expect).Scan(&tableName)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Printf("mysql分表，表存在检查，information_schema.TABLES 查询错误,table:%s,db:%s, err:%v\n", to.expect, to.db, err)
			return "", err
		}
		if errors.Is(err, sql.ErrNoRows) {
			// 表不存在，初始建表结构，新建表
			showCreateSql := fmt.Sprintf("SHOW CREATE TABLE %s.%s", to.db, to.primary)
			var showTableName, createSql string
			err = to.mysqlClient.QueryRow(showCreateSql).Scan(&showTableName, &createSql)
			if err != nil {
				log.Printf("mysql分表，表存在检查，information_schema.TABLES 查询错误,table:%s,db:%s, err:%v\n", to.expect, to.db, err)
				return "", err
			}
			createSql = strings.ReplaceAll(createSql, fmt.Sprintf("`%s`", to.primary), fmt.Sprintf("`%s`.`%s`", to.db, to.expect))
			_, err = to.mysqlClient.Exec(createSql)
			if err != nil {
				log.Printf("mysql分表，创建新表报错,table:%s,db:%s, err:%v\n", to.expect, to.db, err)
				return "", err
			}
		}
		tm[to.expect] = true
	}
	return to.expect, nil
}

func (to *TableOption) lock() bool {
	// count：重试计数器；retry：重试次数
	var count, retry = 1, 50
	for !to.redisClient.SetNX(fmt.Sprintf("SHARDING_TABLE_LOCK_%s_%s", to.db, to.primary), 1234, 5*time.Second).Val() {
		if count > retry {
			return false
		}
		count++
		time.Sleep(100 * time.Millisecond)
	}
	return true
}

func (to *TableOption) unlock() {
	to.redisClient.Del(fmt.Sprintf("SHARDING_TABLE_LOCK_%s_%s", to.db, to.primary))
}

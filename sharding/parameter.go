package sharding

import (
	"errors"
	"fmt"
	"github.com/line-lee/toolkit/stringkit"
	"time"
)

type ParamsResult struct {
	TableName  string
	Start      time.Time
	End        time.Time
	IsEndClose bool // 是否闭合，false就是<end，true就使用<=end
}

func Params(builder ParamsOptionsBuilder) ([]*ParamsResult, error) {
	option := new(ParamsOption)
	for _, opf := range builder.funcs {
		opf(option)
	}
	if stringkit.IsBlank(option.primary) {
		return nil, errors.New("primary option is required，使用 WithParamsPrimary 传入option参数")
	}
	if option.start.IsZero() {
		return nil, errors.New("start option is required，使用 WithParamsStart 传入option参数")
	}
	if option.end.IsZero() {
		return nil, errors.New("end option is required，使用 WithParamsEnd 传入option参数")
	}
	if option.end.Before(option.start) {
		return nil, errors.New("WARNING:star > end")
	}
	if option.t == 0 {
		return nil, errors.New("t option is required，使用 WithParamsType 传入option参数")
	}
	switch option.t {
	case Hour:
		return option.hour()
	case Day:
		return option.day()
	case Month:
		return option.month()
	case Year:
		return option.year()
	default:
		return nil, errors.New("WARNING：type unknown")
	}

}

// ParamsOption 所有参数，由option方法传入，比如primary，由 WithParamsPrimary() 写入参数
type ParamsOption struct {
	// 原始数据库名，或者叫做分表前缀，例如分表有：driver_hour_202508,driver_hour_202509.....，取driver_hour传入
	primary string
	// 查询开始时间
	start time.Time
	// 查询结束时间
	end time.Time
	// 是否包含结束时间
	isEndClose bool
	// 分表类型，传入定义枚举，
	t Type
}

type ParamsOptionsBuilder struct {
	funcs []ParamsOptionFunc
}

func ParamsBuilder() *ParamsOptionsBuilder {
	return &ParamsOptionsBuilder{}
}

type ParamsOptionFunc func(opt *ParamsOption)

func (pb *ParamsOptionsBuilder) Primary(primary string) *ParamsOptionsBuilder {
	pb.funcs = append(pb.funcs, func(option *ParamsOption) {
		option.primary = primary
	})
	return pb
}

func (pb *ParamsOptionsBuilder) Start(start time.Time) *ParamsOptionsBuilder {
	pb.funcs = append(pb.funcs, func(option *ParamsOption) {
		option.start = start
	})
	return pb
}

func (pb *ParamsOptionsBuilder) End(end time.Time) *ParamsOptionsBuilder {
	pb.funcs = append(pb.funcs, func(option *ParamsOption) {
		option.end = end
	})
	return pb
}

func (pb *ParamsOptionsBuilder) IsEndClose(isClose bool) *ParamsOptionsBuilder {
	pb.funcs = append(pb.funcs, func(option *ParamsOption) {
		option.isEndClose = isClose
	})
	return pb
}

func (pb *ParamsOptionsBuilder) Type(t Type) *ParamsOptionsBuilder {
	pb.funcs = append(pb.funcs, func(option *ParamsOption) {
		option.t = t
	})
	return pb
}

func (po *ParamsOption) hour() ([]*ParamsResult, error) {
	const timeFormat = "2006010215"
	var result = make([]*ParamsResult, 0)
	if po.start.Year() == po.end.Year() &&
		po.start.Month() == po.end.Month() &&
		po.start.Day() == po.end.Day() &&
		po.start.Hour() == po.end.Hour() {
		// 同年同月同日同时
		var tableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
		return []*ParamsResult{{TableName: tableName, Start: po.start, End: po.end, IsEndClose: po.isEndClose}}, nil
	}
	// 举个栗子：按月分表，按小时统计，查询时间是 2025-08-19 17:45:00到2025-10-01 10:20:00
	// 开始需要增加参数 2025-08-19 17:00到2025-08-19 18:00，左闭右开
	var startTableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
	var next = po.start.Add(time.Hour)
	var nextStart = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())
	result = append(result, &ParamsResult{TableName: startTableName, Start: po.start, End: nextStart, IsEndClose: false})
	// 循环增加参数 2025-08-19 18:00:00到2025-10-01 10:00:00，左闭右开
	for nextStart.Year() != po.end.Year() ||
		nextStart.Month() != po.end.Month() ||
		nextStart.Day() != po.end.Day() ||
		nextStart.Hour() != po.end.Hour() {
		var tableName = fmt.Sprintf("%s_%s", po.primary, nextStart.Format(timeFormat))
		var nextEnd = nextStart.Add(time.Hour)
		result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: nextEnd, IsEndClose: false})
		nextStart = nextEnd
	}
	// 结尾
	// ********************特殊情况是在边界上********************
	// 当请求的isEndClose=false，且 end 刚好取值在新表开始时间，例如：end=2025-10-01 10:00
	// 这个时候不再对 driver_hour_2025100110这张分表做查询
	if !po.isEndClose &&
		po.end.Minute() == 0 &&
		po.end.Second() == 0 {
		return result, nil
	}
	// 结尾增加参数 2025-10-01 10:00:00到2025-10-01 10:20:00，左闭右闭
	var tableName = fmt.Sprintf("%s_%s", po.primary, po.end.Format(timeFormat))
	result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: po.end, IsEndClose: po.isEndClose})
	return result, nil
}

func (po *ParamsOption) day() ([]*ParamsResult, error) {
	const timeFormat = "20060102"
	var result = make([]*ParamsResult, 0)
	if po.start.Year() == po.end.Year() &&
		po.start.Month() == po.end.Month() &&
		po.start.Day() == po.end.Day() {
		// 同年同月同日
		var tableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
		return []*ParamsResult{{TableName: tableName, Start: po.start, End: po.end, IsEndClose: po.isEndClose}}, nil
	}
	// 举个栗子：按月分表，按小时统计，查询时间是 2025-08-19 17:45:00到2025-10-01 10:20:00
	// 开始需要增加参数 2025-08-19 17:00到2025-08-19 18:00，左闭右开
	var startTableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
	var next = po.start.AddDate(0, 0, 1)
	var nextStart = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	result = append(result, &ParamsResult{TableName: startTableName, Start: po.start, End: nextStart, IsEndClose: false})
	// 循环增加参数 2025-08-19 18:00:00到2025-10-01 10:00:00，左闭右开
	for nextStart.Year() != po.end.Year() ||
		nextStart.Month() != po.end.Month() ||
		nextStart.Day() != po.end.Day() {
		var tableName = fmt.Sprintf("%s_%s", po.primary, nextStart.Format(timeFormat))
		var nextEnd = nextStart.AddDate(0, 0, 1)
		result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: nextEnd, IsEndClose: false})
		nextStart = nextEnd
	}
	// 结尾
	// ********************特殊情况是在边界上********************
	// 当请求的isEndClose=false，且 end 刚好取值在新表开始时间，例如：end=2025-10-01 00:00
	// 这个时候不再对 driver_hour_20251001这张分表做查询
	if !po.isEndClose &&
		po.end.Hour() == 0 &&
		po.end.Minute() == 0 &&
		po.end.Second() == 0 {
		return result, nil
	}
	// 结尾增加参数 2025-10-01 00:00:00到2025-10-01 10:20:00，左闭右闭
	var tableName = fmt.Sprintf("%s_%s", po.primary, po.end.Format("20060102"))
	result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: po.end, IsEndClose: po.isEndClose})
	return result, nil
}

func (po *ParamsOption) month() ([]*ParamsResult, error) {
	const timeFormat = "200601"
	var result = make([]*ParamsResult, 0)
	if po.start.Year() == po.end.Year() &&
		po.start.Month() == po.end.Month() {
		// 同年同月
		var tableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
		return []*ParamsResult{{TableName: tableName, Start: po.start, End: po.end, IsEndClose: po.isEndClose}}, nil
	}
	// 举个栗子：按月分表，按小时统计，查询时间是 2025-08-19 17:00到2025-10-01 10:00
	// 开始需要增加参数 2025-08-19 17:00到2025-09-01 00:00，左闭右开
	var startTableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
	var next = po.start.AddDate(0, 1, 0)
	var nextStart = time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, next.Location())
	result = append(result, &ParamsResult{TableName: startTableName, Start: po.start, End: nextStart, IsEndClose: false})
	// 循环增加参数 2025-09-01 00:00到2025-10-01 00:00，左闭右开
	for nextStart.Year() != po.end.Year() || nextStart.Month() != po.end.Month() {
		var tableName = fmt.Sprintf("%s_%s", po.primary, nextStart.Format(timeFormat))
		var nextEnd = nextStart.AddDate(0, 1, 0)
		result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: nextEnd, IsEndClose: false})
		nextStart = nextEnd
	}
	// 结尾
	// ********************特殊情况是在边界上********************
	// 当请求的isEndClose=false，且 end 刚好取值在新表开始时间，例如：end=2025-10-01 00:00:00
	// 这个时候不再对 driver_hour_202510 这张分表做查询
	if !po.isEndClose &&
		po.end.Day() == 1 &&
		po.end.Minute() == 0 &&
		po.end.Hour() == 0 &&
		po.end.Minute() == 0 &&
		po.end.Second() == 0 {
		return result, nil
	}
	// 结尾增加参数 2025-10-01 00:00:00到2025-10-01 10:00:00，左闭右闭
	var tableName = fmt.Sprintf("%s_%s", po.primary, po.end.Format(timeFormat))
	result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: po.end, IsEndClose: po.isEndClose})
	return result, nil
}

func (po *ParamsOption) year() ([]*ParamsResult, error) {
	const timeFormat = "2006"
	var result = make([]*ParamsResult, 0)
	if po.start.Year() == po.end.Year() {
		// 同年
		var tableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
		return []*ParamsResult{{TableName: tableName, Start: po.start, End: po.end, IsEndClose: po.isEndClose}}, nil
	}
	// 举个栗子：按月分表，按小时统计，查询时间是 2025-08-19 17:45:00到2027-10-01 10:20:00
	// 开始需要增加参数 2025-08-19 17:45:00到2026-01-01 00:00:00，左闭右开
	var startTableName = fmt.Sprintf("%s_%s", po.primary, po.start.Format(timeFormat))
	var next = po.start.AddDate(1, 0, 0)
	var nextStart = time.Date(next.Year(), time.January, 1, 0, 0, 0, 0, next.Location())
	result = append(result, &ParamsResult{TableName: startTableName, Start: po.start, End: nextStart, IsEndClose: false})
	// 循环增加参数 2026-01-01 00:00:00到2027-01-01 00:00:00，左闭右开
	for nextStart.Year() != po.end.Year() {
		var tableName = fmt.Sprintf("%s_%s", po.primary, nextStart.Format(timeFormat))
		var nextEnd = nextStart.AddDate(1, 0, 0)
		result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: nextEnd, IsEndClose: false})
		nextStart = nextEnd
	}
	// 结尾
	// ********************特殊情况是在边界上********************
	// 当请求的isEndClose=false，且 end 刚好取值在新表开始时间，例如：end=2027-01-01 00:00:00
	// 这个时候不再对 driver_hour_2027这张分表做查询
	if !po.isEndClose &&
		po.end.Month() == time.January &&
		po.end.Day() == 1 &&
		po.end.Hour() == 0 &&
		po.end.Minute() == 0 &&
		po.end.Second() == 0 {
		return result, nil
	}
	// 结尾增加参数 2027-01-01 00:00:00到2027-10-01 10:20:00，左闭右闭
	var tableName = fmt.Sprintf("%s_%s", po.primary, po.end.Format(timeFormat))
	result = append(result, &ParamsResult{TableName: tableName, Start: nextStart, End: po.end, IsEndClose: po.isEndClose})
	return result, nil
}

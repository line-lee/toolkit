package sharding

// Type  分表粒度，按年，月，日....分表
type Type int

const (
	Hour  Type = 10 // 按小时分表
	Day   Type = 20 // 按天分表
	Month Type = 30 // 按月分表
	Year  Type = 40 // 按年分表
)

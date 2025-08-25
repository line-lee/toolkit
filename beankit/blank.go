package beankit

// IsStringBlank 判断字符串是否为空，true：空；false不为空
func IsStringBlank(str string) bool {
	return str == "" || len(str) == 0
}

// IsSliceEmpty 判断切片是否为空，true：空；false不为空
func IsSliceEmpty[T any](s []T) bool {
	return s == nil || len(s) == 0
}

package stringkit

func IsBlank(str string) bool {
	if str == "" || len(str) == 0 {
		return true
	}
	return false
}

func IsNotBlank(str string) bool {
	if str != "" && len(str) != 0 {
		return true
	}
	return false
}

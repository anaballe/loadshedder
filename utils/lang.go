package utils

func Contains[T comparable](arr []T, val T) bool {
	for _, item := range arr {
		if item == val {
			return true
		}
	}
	return false
}

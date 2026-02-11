package lodash

func ValueOr[T comparable](val, defaultValue T) T {
	var zero T
	if val == zero {
		return defaultValue
	}
	return val
}

package utils

func OR[T comparable](value, fallback T) T {
	var zero T
	if value == zero {
		return fallback
	}
	return value
}

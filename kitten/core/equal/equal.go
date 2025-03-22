// Package equal 判断值是否相同
package equal

import (
	"maps"
	"slices"
	"time"
)

// IsSame 判断值是否相同
func IsSame[T comparable](v ...T) bool {
	return IsSameByFunc(func(a, b T) bool { return a == b }, v...)
}

// IsSameByFunc 用函数判断值是否相同
func IsSameByFunc[T any](f func(T, T) bool, v ...T) bool {
	for i := range len(v) - 1 {
		if !f(v[0], v[i+1]) {
			return false
		}
	}
	return true
}

// IsSameSlice 判断切片是否相同
func IsSameSlice[T comparable](s ...[]T) bool {
	return IsSameByFunc(slices.Equal, s...)
}

// IsSameMap 判断映射是否相同
func IsSameMap[K comparable, V comparable](m ...map[K]V) bool {
	return IsSameByFunc(maps.Equal, m...)
}

// 判断两个时间是否在同一天
func equalDate(t1, t2 time.Time) bool {
	return t1.YearDay() == t2.YearDay() && t1.Year() == t2.Year()
}

// IsSameDate 判断时间是否在同一天
func IsSameDate(t ...time.Time) bool {
	return IsSameByFunc(equalDate, t...)
}

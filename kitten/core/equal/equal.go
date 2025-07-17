// Package equal 判断值是否相同
package equal

import (
	"maps"
	"slices"
	"time"
)

// IsSame 判断值是否相同
func IsSame[T comparable](v ...T) bool {
	return IsSameFunc(func(a, b T) bool { return a == b }, v...)
}

// IsSameSlice 判断切片是否相同
func IsSameSlice[T comparable](s ...[]T) bool {
	return IsSameFunc(slices.Equal, s...)
}

// IsSameMap 判断映射是否相同
func IsSameMap[K comparable, V comparable](m ...map[K]V) bool {
	return IsSameFunc(maps.Equal, m...)
}

// IsSameDate 判断时间是否在同一天
func IsSameDate(t ...time.Time) bool {
	return IsSameFunc(equalDate, t...)
}

// IsSameDate4AM 判断时间是否在同一天，界限为 4:00
func IsSameDate4AM(t ...time.Time) bool {
	return IsSameFunc(equalDate4AM, t...)
}

// 判断两个时间是否在同一天，界限为 4:00
func equalDate4AM(t1, t2 time.Time) bool {
	t1 = t1.Add(-4 * time.Hour)
	t2 = t2.Add(-4 * time.Hour)
	return equalDate(t1, t2)
}

// 判断两个时间是否在同一天
func equalDate(t1, t2 time.Time) bool {
	t1 = t1.Local()
	t2 = t2.Local()
	return t1.YearDay() == t2.YearDay() && t1.Year() == t2.Year()
}

// IsSameFunc 用函数判断值是否相同
func IsSameFunc[T any](f func(T, T) bool, v ...T) bool {
	for i := range len(v) - 1 {
		if !f(v[0], v[i+1]) {
			return false
		}
	}
	return true
}

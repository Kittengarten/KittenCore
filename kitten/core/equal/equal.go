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
	return equalDate(t1.Add(-4*time.Hour), t2.Add(-4*time.Hour))
}

// 判断两个时间是否在同一天
func equalDate(t1, t2 time.Time) bool {
	return CmpDay(t1, t2) == 0
}

// CmpDay4AM 比较天数差值 t2 - t1
func CmpDay4AM(t1, t2 time.Time) int {
	return CmpDay(t1.Add(-4*time.Hour), t2.Add(-4*time.Hour))
}

// CmpDay 比较天数差值 t2 - t1
func CmpDay(t1, t2 time.Time) int {
	var (
		y1, m1, d1 = t1.Local().Date()
		y2, m2, d2 = t2.Local().Date()
	)
	return int(
		time.Date(y2, m2, d2,
			0, 0, 0, 0, time.Local).
			Sub(time.Date(y1, m1, d1,
				0, 0, 0, 0, time.Local)))
}

// CmpWeek 比较 ISO 周数差值 t2 - t1
func CmpWeek(t1, t2 time.Time) int {
	t1 = t1.Local()
	t2 = t2.Local()
	return int(t2.AddDate(
		0,
		0,
		-int((int(t2.Weekday())+6)%7),
	).Sub(t1.AddDate(
		0,
		0,
		-int((int(t1.Weekday())+6)%7),
	)).Hours()/24) / 7
}

// CmpMonth 比较月数差值 t2 - t1
func CmpMonth(t1, t2 time.Time) int {
	var (
		y1, m1, _ = t1.Local().Date()
		y2, m2, _ = t2.Local().Date()
	)
	return (y2-y1)*12 + int(m2) - int(m1)
}

// FullYears 获取周年数 t2 - t1
func FullYears(t1, t2 time.Time) int {
	y := CmpYear(t1, t2)
	if t1.AddDate(y, 0, 0).After(t2) {
		y--
	}
	return y
}

// CmpYear 比较年数差值 t2 - t1
func CmpYear(t1, t2 time.Time) int {
	return t2.Local().Year() - t1.Local().Year()
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

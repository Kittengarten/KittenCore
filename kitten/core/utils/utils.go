// Package utils 工具
package utils

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
)

var (
	// ErrInvalidData 无效的数据喵！
	ErrInvalidData = errors.New(`无效的数据喵！`)
	// ErrInvalidArgument 无效的参数喵！
	ErrInvalidArgument = errors.New(`无效的参数喵！`)
	// ErrNoMatch 正则表达式没有匹配到喵！
	ErrNoMatch = errors.New(`正则表达式没有匹配到喵！`)
)

// GenerateRandomNumber 生成 n 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start, end, n int) ([]int, error) {
	// 范围检查
	if start >= end {
		return nil, fmt.Errorf(`上限 %d 必须大于下限 %d：%w`,
			end, start, ErrInvalidArgument)
	}
	if (end - start) < n {
		return nil, fmt.Errorf(`下限 %d 和上限 %d 之间的数字只有 %d 个，`+
			`不满足 %d 个的要求：%w`,
			start, end, end-start, n, ErrInvalidArgument)
	}
	if n <= 0 {
		return nil, fmt.Errorf(`个数 %d 不是正整数：%w`,
			n, ErrInvalidArgument)
	}
	// 存放不重复结果的集合
	set := make(map[int]struct{}, n)
	for len(set) < n {
		// 生成随机数
		//nolint:gosec
		set[rand.N(end-start)+start] = struct{}{}
	}
	// 集合转换为切片
	return slices.Collect(maps.Keys(set)), nil
}

// ConvertSlice 将 src 中的每个元素由 T 类型转换为 U 类型
func ConvertSlice[T any, U any](src []T, f func(T) U) []U {
	dst := make([]U, len(src))
	for i, v := range src {
		dst[i] = f(v)
	}
	return dst
}

// RemoveDuplicates 去除切片中的重复元素
func RemoveDuplicates[T comparable](slice []T) (result []T) {
	seen := make(map[T]struct{})
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// RemoveDuplicatesFunc 去除切片中的重复元素
func RemoveDuplicatesFunc[T any](slice []T, f func(T, T) bool) (result []T) {
	for _, v := range slice {
		if !slices.ContainsFunc(result, func(e T) bool {
			return f(e, v)
		}) {
			result = append(result, v)
		}
	}
	return result
}

// Round 保留小数点后 n 位
func Round(f float64, n int) float64 {
	pow10N := math.Pow10(n)
	return math.RoundToEven(f*pow10N) / pow10N
}

// BoolToString 将布尔值转换为字符串
func BoolToString(b bool) string {
	if b {
		return `true`
	}
	return `false`
}

// BoolToInt 将布尔值转换为数字
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// GetTypeName 获取任意类型变量的类型名
func GetTypeName(value any) string {
	// 获取 reflect.Type
	t := reflect.TypeOf(value)
	// 如果是指针类型，获取其指向的元素类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// 返回类型名
	return t.Name()
}

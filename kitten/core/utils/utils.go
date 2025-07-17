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

	"golang.org/x/exp/constraints"
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
		return nil, fmt.Errorf(`下限 %d 必须小于上限 %d：%w`,
			start, end, ErrInvalidArgument)
	}
	if n <= 0 {
		return nil, fmt.Errorf(`个数 %d 不是正整数：%w`,
			n, ErrInvalidArgument)
	}
	if end-start < n {
		return nil, fmt.Errorf(`下限 %d 和上限 %d 之间的数字只有 %d 个，`+
			`不满足 %d 个的要求：%w`,
			start, end, end-start, n, ErrInvalidArgument)
	}
	// 抽取个数占范围的比例很低
	// 或者范围极大
	// 使用集合法降低开销
	if r := float64(end - start); r/float64(n) > 1+16/math.Log10(r) {
		return grnSet(start, end, n), nil
	}
	// 抽取个数占范围的比例较低且范围较大
	// 或者范围过大
	// 使用蓄水池抽样法降低内存需求
	if end-start > 2*n && end-start > 1<<2 || end-start > 1<<4 {
		return gnrReservoir(start, end, n), nil
	}
	// 使用洗牌法
	return gnrShuffle(start, end, n), nil
}

// 集合法生成 n 个 [start, end) 范围的不重复的随机数
func grnSet(start, end, n int) []int {
	// 存放不重复结果的集合
	set := make(map[int]struct{}, n)
	for len(set) < n {
		// 生成随机数
		//nolint:gosec
		set[rand.N(end-start)+start] = struct{}{}
	}
	// 集合转换为切片
	return slices.Collect(maps.Keys(set))
}

// 蓄水池抽样法生成 n 个 [start, end) 范围的不重复的随机数
func gnrReservoir(start, end, n int) []int {
	reservoir := make([]int, n)
	// 初始化蓄水池为前k个元素
	for i := range n {
		reservoir[i] = start + i
	}
	// 遍历范围外的元素
	for i := n; i < end-start; i++ {
		//nolint:gosec
		if j := rand.N(i + 1); j < n {
			reservoir[j] = start + i
		}
	}
	return reservoir
}

// 洗牌法生成 n 个 [start, end) 范围的不重复的随机数
func gnrShuffle(start, end, n int) []int {
	// 构造 [start, end) 范围内的切片
	nums := make([]int, end-start)
	for i := range nums {
		nums[i] = start + i
	}
	// 打乱顺序
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})
	// 返回前n个元素
	return nums[:n]
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
func RemoveDuplicates[T comparable](slice []T) []T {
	var (
		seen   = make(map[T]struct{})
		result = make([]T, 0, len(slice))
	)
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

// BoolToInt 将布尔值转换为数字
func BoolToInt[T constraints.Integer](b bool) T {
	if b {
		return 1
	}
	return 0
}

// BoolToIntStr 将布尔值转换为数字字符串
func BoolToIntStr(b bool) string {
	if b {
		return `1`
	}
	return `0`
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

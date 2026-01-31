// Package utils 工具
package utils

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"math"
	"math/rand/v2"
	"os"
	"reflect"
	"runtime/debug"
	"slices"
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

type (
	// Object 空对象
	Object = struct{}
	// Set 集合
	Set[T comparable] = map[T]Object
)

// PlatformBits 平台位数
const PlatformBits = 32 << (^uint(0) >> 63)

var (
	// ErrInvalidData 无效的数据喵！
	ErrInvalidData = errors.New(`无效的数据喵！`)
	// ErrInvalidArgument 无效的参数喵！
	ErrInvalidArgument = errors.New(`无效的参数喵！`)
	// ErrNoMatch 正则表达式没有匹配到喵！
	ErrNoMatch = errors.New(`正则表达式没有匹配到喵！`)
)

// GenerateRandomNumber 生成 n 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start, end, n int) (Set[int], error) {
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
func grnSet(start, end, n int) Set[int] {
	// 存放不重复结果的集合
	set := make(Set[int], n)
	for len(set) < n {
		// 生成随机数
		//nolint:gosec
		set[rand.N(end-start)+start] = Object{}
	}
	return set
}

// 蓄水池抽样法生成 n 个 [start, end) 范围的不重复的随机数
func gnrReservoir(start, end, n int) Set[int] {
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
	return maps.Collect(keys(reservoir))
}

// 洗牌法生成 n 个 [start, end) 范围的不重复的随机数
func gnrShuffle(start, end, n int) Set[int] {
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
	return maps.Collect(keys(nums[:n]))
}
func keys[Slice ~[]E, E any](s Slice) iter.Seq2[E, struct{}] {
	return func(yield func(E, struct{}) bool) {
		for _, e := range s {
			if !yield(e, struct{}{}) {
				return
			}
		}
	}
}

// ConvertSlice 将 src 中的每个元素由 E1 类型转换为 E2 类型
func ConvertSlice[E1 any, E2 any](src []E1, f func(E1) E2) []E2 {
	dst := make([]E2, len(src))
	for i, e := range src {
		dst[i] = f(e)
	}
	return dst
}

// RemoveDuplicates 去除切片中的重复元素
func RemoveDuplicates[E comparable](slice []E) []E {
	var (
		seen   = make(Set[E])
		result = make([]E, 0, len(slice))
	)
	for _, e := range slice {
		if _, ok := seen[e]; !ok {
			seen[e] = Object{}
			result = append(result, e)
		}
	}
	return result
}

// RemoveDuplicatesFunc 去除切片中的重复元素
func RemoveDuplicatesFunc[E any](slice []E, f func(E, E) bool) (result []E) {
	for _, e := range slice {
		if !slices.ContainsFunc(result, func(er E) bool {
			return f(er, e)
		}) {
			result = append(result, e)
		}
	}
	return result
}

// RemoveDuplicateIter 去除迭代器中的重复元素
func RemoveDuplicateIter[E comparable](iter iter.Seq[E]) (result iter.Seq[E]) {
	return func(yield func(E) bool) {
		seen := make(Set[E])
		for e := range iter {
			if _, ok := seen[e]; !ok {
				seen[e] = Object{}
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
}

// RemoveDuplicateIterFunc 去除迭代器中的重复元素
func RemoveDuplicateIterFunc[E any](iter iter.Seq[E], f func(E, E) bool) (result iter.Seq[E]) {
	return func(yield func(E) bool) {
		s := make([]E, 0)
		for e := range iter {
			if !slices.ContainsFunc(s, func(es E) bool {
				return f(e, es)
			}) {
				s = append(s, e)
				continue
			}
			if !yield(e) {
				return
			}
		}
	}
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
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// 返回类型名
	return t.Name()
}

// GoroutineSeq 协程序列号
var GoroutineSeq atomic.Uint64

// Go 运行一个 goroutine
func Go(name string, f func()) {
	go func() {
		var (
			n = slog.String(`名称`, name)
			s = slog.Uint64(`序列号`, GoroutineSeq.Add(1))
		)
		// 处理 panic，防止程序崩溃
		defer HandlePanic(n, s)
		slog.Info(`正在启动协程……`, n, s)
		defer slog.Info(`协程已结束！`, n, s)
		f()
	}()
}

// 崩溃信息路径
var Crash string

// HandlePanic 处理 panic
func HandlePanic(name, seq slog.Attr) {
	if err := recover(); err != nil {
		slog.Info(`协程从 panic 恢复……`,
			name,
			seq,
			slog.Any(`错误`, err),
			slog.String(`堆栈`, string(debug.Stack())),
		)
		file, err := os.Create(Crash)
		if err != nil {
			slog.Error(`创建`, slog.String(`路径`, Crash),
				slog.Any(`错误`, err))
			return
		}
		defer file.Close()
		if _, err := fmt.Fprintf(file, "panic: %v\n%s\n", err, string(debug.Stack())); err != nil {
			slog.Error(`写入`, slog.String(`路径`, Crash),
				slog.Any(`错误`, err))
		}
	}
}

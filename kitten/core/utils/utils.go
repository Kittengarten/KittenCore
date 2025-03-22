// Package utils 工具
package utils

import (
	"fmt"
	"maps"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/Kittengarten/KittenCore/internal/wr"
)

type (
	// Choicer 随机项目的抽象接口
	Choicer interface {
		GetID() int             // 该项目的 ID
		GetInformation() string // 该项目的信息
	}

	// ChoicerW 带权重的随机项目的抽象接口
	ChoicerW interface {
		Choicer
		GetWeight() int // 该项目的权重
	}

	// Choicers 由随机项目的抽象接口组成的切片
	Choicers []Choicer

	// ChoicersW 由带权重的随机项目的抽象接口组成的切片
	ChoicersW []ChoicerW
)

// Choose 按权重抽取一个项目的序号
func (c ChoicersW) Choose() (int, error) {
	chooser, err := wr.NewChooser(
		ConvertSlice(
			c,
			func(ch ChoicerW) wr.Choice[int, int] {
				return wr.Choice[int, int]{Item: ch.GetID(), Weight: ch.GetWeight()}
			},
		)...,
	)
	if err != nil {
		return -1, err
	}
	return chooser.Pick(), nil
}

// GenerateRandomNumber 生成 count 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start, end, count int) ([]int, error) {
	// 范围检查
	if start >= end {
		return nil, fmt.Errorf(`上限 %d 必须大于下限 %d 喵！`, end, start)
	}
	if (end - start) < count {
		return nil, fmt.Errorf(`下限 %d 和上限 %d 之间的数字只有 %d 个，不满足 %d 个的要求喵！`, start, end, end-start, count)
	}
	if count <= 0 {
		return nil, fmt.Errorf(`个数 %d 不是正整数喵！`, count)
	}
	// 存放不重复结果的集合
	set := make(map[int]struct{}, count)
	for len(set) < count {
		// 生成随机数
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

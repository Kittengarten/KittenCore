package choice

import (
	"errors"
	"math/rand/v2"

	"github.com/Kittengarten/KittenCore/internal/wr"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"golang.org/x/exp/constraints"
)

type (
	// Choicer 随机项目的抽象接口
	Choicer[T any] interface {
		ID() T        // ID 该项目的标识符
		Info() string // Info 该项目的信息
	}

	// WeightedChoicer 带权重的随机项目的抽象接口
	WeightedChoicer[T any, W constraints.Integer] interface {
		Choicer[T]
		Weight() W // 该项目的权重
	}

	// Choicers 由随机项目的抽象接口组成的切片
	//nolint:misspell
	Choicers[T any] []Choicer[T]

	// WeightedChoicers 由带权重的随机项目的抽象接口组成的切片
	WeightedChoicers[T any, W constraints.Integer] []WeightedChoicer[T, W]

	// WeightedChooser 由带权重的随机项目的抽象接口组成的切片
	WeightedChooser[T any, W constraints.Integer] = *wr.Chooser[T, W]
)

// NewChoicers 创建项目切片
func NewChoicers[T any, C Choicer[T]](items ...C) Choicers[T] {
	return Choicers[T](utils.ConvertSlice(
		items,
		func(i C) Choicer[T] {
			return i
		},
	))
}

// NewWeightedChooser 创建带权重项目切片
func NewWeightedChooser[T any, W constraints.Integer, WC WeightedChoicer[T, W]](wc ...WC) (WeightedChooser[T, W], error) {
	return wr.NewChooser(utils.ConvertSlice(
		wc,
		func(wc WC) wr.Choice[T, W] {
			return wr.NewChoice(wc.ID(), wc.Weight())
		},
	)...)
}

// MaxWeightProportion 获取最高权重占全部权重的比例
func MaxWeightProportion[T any, W constraints.Integer, WC WeightedChoicer[T, W]](wc ...WC) float64 {
	var maxWeight, sumWeight W
	for _, ch := range wc {
		w := ch.Weight()
		sumWeight += max(0, w)
		maxWeight = max(maxWeight, 0, w)
	}
	if sumWeight == 0 {
		return 0
	}
	return float64(maxWeight) / float64(sumWeight)
}

// ErrNoChoice 没有可用选项喵！
var ErrNoChoice = errors.New(`没有可用选项喵！`)

// Choose 抽取一个项目的标识符
func (c Choicers[T]) Choose() (item T, err error) {
	if len(c) == 0 {
		return item, ErrNoChoice
	}
	return c[rand.N(len(c))].ID(), nil
}

// ChooseInfo 抽取一个项目的信息
func (c Choicers[T]) ChooseInfo() (string, error) {
	if len(c) == 0 {
		return ``, ErrNoChoice
	}
	return c[rand.N(len(c))].Info(), nil
}

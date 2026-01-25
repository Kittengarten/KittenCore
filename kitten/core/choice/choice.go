package choice

import (
	"errors"
	"math/rand/v2"

	"github.com/Kittengarten/KittenCore/internal/wr"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
)

type (
	// Choicer 随机项目的抽象接口
	Choicer interface {
		ID() any      // ID 该项目的标识符
		Info() string // Info 该项目的信息
	}

	// WeightedChoicer 带权重的随机项目的抽象接口
	WeightedChoicer interface {
		Choicer
		Weight() int // 该项目的权重
	}

	// Choicers 由随机项目的抽象接口组成的切片
	//nolint:misspell
	Choicers []Choicer

	// WeightedChooser 由带权重的随机项目的抽象接口组成的切片
	WeightedChooser = *wr.Chooser[any, int]
)

// NewChoicers 创建项目切片
func NewChoicers[T Choicer](items ...T) Choicers {
	return Choicers(utils.ConvertSlice(
		items,
		func(i T) Choicer {
			return i
		},
	))
}

// NewWeightedChooser 创建带权重项目切片
func NewWeightedChooser[T WeightedChoicer](items ...T) (WeightedChooser, error) {
	return wr.NewChooser(utils.ConvertSlice(
		items,
		func(ch T) wr.Choice[any, int] {
			return wr.NewChoice(ch.ID(), ch.Weight())
		},
	)...)
}

// MaxWeightProportion 获取最高权重占全部权重的比例
func MaxWeightProportion[T WeightedChoicer](items ...T) float64 {
	var maxWeight, sumWeight int
	for _, ch := range items {
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
func (c Choicers) Choose() (any, error) {
	if len(c) == 0 {
		return nil, ErrNoChoice
	}
	return c[rand.N(len(c))].ID(), nil
}

// ChooseInfo 抽取一个项目的信息
func (c Choicers) ChooseInfo() (string, error) {
	if len(c) == 0 {
		return ``, ErrNoChoice
	}
	return c[rand.N(len(c))].Info(), nil
}

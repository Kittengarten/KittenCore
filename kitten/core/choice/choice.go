package choice

import (
	"errors"

	"github.com/Kittengarten/KittenCore/internal/wr"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
)

type (
	// Choicer 随机项目的抽象接口
	Choicer interface {
		ID() any      // ID 该项目的标识符
		Info() string // Info 该项目的信息
	}

	// ChoicerW 带权重的随机项目的抽象接口
	ChoicerW interface {
		Choicer
		Weight() int // 该项目的权重
	}

	// Choicers 由随机项目的抽象接口组成的切片
	//nolint:misspell
	Choicers []Choicer

	// ChoicersW 由带权重的随机项目的抽象接口组成的切片
	ChoicersW []ChoicerW
)

// ErrNoChoice 没有可用选项喵！
var ErrNoChoice = errors.New(`没有可用选项喵！`)

// Choose 按权重抽取一个项目的标识符
func (c ChoicersW) Choose() (any, error) {
	if len(c) == 0 {
		return nil, ErrNoChoice
	}
	chooser, err := wr.NewChooser(
		utils.ConvertSlice(
			c,
			func(ch ChoicerW) wr.Choice[any, int] {
				return wr.Choice[any, int]{Item: ch.ID(), Weight: ch.Weight()}
			},
		)...,
	)
	if err != nil {
		return nil, err
	}
	return chooser.Pick(), nil
}

// MaxWeightProportion 获取最高权重占全部权重的比例
func (c ChoicersW) MaxWeightProportion() float64 {
	var maxWeight, sumWeight int
	for _, ch := range c {
		w := ch.Weight()
		sumWeight += w
		maxWeight = max(maxWeight, w)
	}
	if sumWeight == 0 {
		return 0
	}
	return float64(maxWeight) / float64(sumWeight)
}

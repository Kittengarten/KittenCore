package choice

import (
	"github.com/Kittengarten/KittenCore/internal/wr"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
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
		utils.ConvertSlice(
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

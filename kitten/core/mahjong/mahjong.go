// Package mahjong 麻将
package mahjong

import (
	"maps"
	"math/rand/v2"
	"slices"
)

var (
	// 配牌数量
	deal = map[bool]int{
		true:  14,
		false: 13,
	}
	// 麻将牌
	mahjong = map[string]rune{
		`东风`: 0x1F000,
		`南风`: 0x1F001,
		`西风`: 0x1F002,
		`北风`: 0x1F003,
		`中`:  0x1F004,
		`发`:  0x1F005,
		`白`:  0x1F006,
		`一万`: 0x1F007,
		`二万`: 0x1F008,
		`三万`: 0x1F009,
		`四万`: 0x1F00A,
		`五万`: 0x1F00B,
		`六万`: 0x1F00C,
		`七万`: 0x1F00D,
		`八万`: 0x1F00E,
		`九万`: 0x1F00F,
		`一索`: 0x1F010,
		`二索`: 0x1F011,
		`三索`: 0x1F012,
		`四索`: 0x1F013,
		`五索`: 0x1F014,
		`六索`: 0x1F015,
		`七索`: 0x1F016,
		`八索`: 0x1F017,
		`九索`: 0x1F018,
		`一筒`: 0x1F019,
		`二筒`: 0x1F01A,
		`三筒`: 0x1F01B,
		`四筒`: 0x1F01C,
		`五筒`: 0x1F01D,
		`六筒`: 0x1F01E,
		`七筒`: 0x1F01F,
		`八筒`: 0x1F020,
		`九筒`: 0x1F021,
	}
	// 基础牌山（每种牌一张）
	baseTiles = slices.Collect(maps.Values(mahjong))
)

// New 配牌
func New(dealer bool) []rune {
	m := NewWall()[:deal[dealer]]
	slices.Sort(m)
	return m
}

// NewWall 新的牌山
func NewWall() []rune {
	m := slices.Repeat(baseTiles, 4)
	rand.Shuffle(len(m), func(i, j int) {
		m[i], m[j] = m[j], m[i]
	})
	return m
}

// Package qqemoji QQ 表情
package qqemoji

import (
	_ "embed"
	"iter"
	"maps"
	"math/rand/v2"
	"slices"

	"github.com/goccy/go-yaml"
	"go.uber.org/zap"
)

var (
	//go:embed qq_emoji.yaml
	qqEmojiYAML []byte
	// QQ 表情
	qqEmoji map[string]rune
)

func init() {
	if err := yaml.Unmarshal(qqEmojiYAML, &qqEmoji); err != nil {
		zap.S().Error(err)
	}
}

// All 获取 QQ 表情键值的迭代器
func All() iter.Seq2[string, rune] {
	return maps.All(qqEmoji)
}

// Get 获取 QQ 表情
func Get(emoji string) rune {
	return qqEmoji[emoji]
}

// Random 随机获取 QQ 表情
func Random() rune {
	//nolint:gosec
	return slices.Collect(maps.Values(qqEmoji))[rand.N(len(qqEmoji))]
}

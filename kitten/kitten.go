// Package kitten 包含了 KittenCore 以及各插件的核心依赖结构体、方法和函数
package kitten

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

const path Path = `config.yaml` // 配置文件名

var (
	// Configs 来自 Bot 的配置文件
	Configs = LoadMainConfig()
	// Bot 实例
	Bot *zero.Ctx
)

// Package core KittenCore 基础依赖
package core

import (
	zero "github.com/wdvxdr1123/ZeroBot"
)

// PlatformBits 平台位数
const PlatformBits = 32 << (^uint(0) >> 63)

// NotOnlyToMe 不是（@ 自己 | 以自己的名字之一开头 | 私聊）任何之一
func NotOnlyToMe(ctx *zero.Ctx) bool {
	return !zero.OnlyToMe(ctx)
}

// NotCommand 不是命令
func NotCommand(ctx *zero.Ctx) bool {
	return !zero.CommandRule(``)(ctx)
}

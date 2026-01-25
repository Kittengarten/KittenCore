// Package core KittenCore 基础依赖
package core

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// Timeout 超时时间 5 分钟
const Timeout = 5 * time.Minute

// NotOnlyToMe 不是（@ 自己 | 以自己的名字之一开头 | 私聊）任何之一
func NotOnlyToMe(ctx *zero.Ctx) bool {
	return !zero.OnlyToMe(ctx)
}

// NotCommand 不是命令
func NotCommand(ctx *zero.Ctx) bool {
	return !zero.CommandRule(``)(ctx)
}

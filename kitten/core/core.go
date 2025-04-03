// Package core KittenCore 基础依赖
package core

import (
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"

	trshttp "github.com/fumiama/terasu/http"
	trshttp2 "github.com/fumiama/terasu/http2"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// PlatformBits 平台位数
const PlatformBits = 32 << (^uint(0) >> 63)

func init() {
	shttp.TLSClient = trshttp.DefaultClient
	shttp.TLSHTTP2Client = trshttp2.DefaultClient
}

// NotOnlyToMe 不是（@ 自己 | 以自己的名字之一开头 | 私聊）任何之一
func NotOnlyToMe(ctx *zero.Ctx) bool {
	return !zero.OnlyToMe(ctx)
}

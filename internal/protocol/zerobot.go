package protocol

import (
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

// 代理类型
type proxy bool

const (
	Forward proxy = true  // 正向代理
	Reverse proxy = false // 反向代理
)

// Runbot 启动机器人
func RunBot(p proxy) {
	config := kitten.MainConfig()
	zero.RunAndBlock(&zero.Config{
		NickName:      config.NickName,
		CommandPrefix: config.CommandPrefix,
		SuperUsers: utils.ConvertSlice(
			config.SuperUsers,
			func(v kitten.QQ) int64 { return v.Int() },
		),
		AddSpaceAfterAt: true,
		Driver: []zero.Driver{
			wsDriver(p, config.WebSocket),
		},
	}, process.GlobalInitMutex.Unlock)
}

// 获取 WebSocket 驱动
func wsDriver(p proxy, ws kitten.WebSocket) zero.Driver {
	if p {
		// OneBot 正向 WS 默认使用 6700 端口
		return driver.NewWebSocketClient(ws.URL, ws.AccessToken)
	}
	// OneBot 反向 WS 默认使用 5140 端口
	return driver.NewWebSocketServer(16, ws.URL, ws.AccessToken)
}

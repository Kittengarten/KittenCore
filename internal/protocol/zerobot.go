package protocol

import (
	"strings"

	"github.com/Kittengarten/KittenCore/internal/config"

	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

const (
	Forward = `forward` // Forward 正向代理
	Reverse = `reverse` // Reverse 反向代理
	HTTP    = `http`    // HTTP ...
)

// Runbot 启动机器人
func RunBot() {
	config := config.Get()
	zero.RunAndBlock(&zero.Config{
		NickName:        config.NickName,
		CommandPrefix:   config.CommandPrefix,
		SuperUsers:      config.SuperUsers,
		AddSpaceAfterAt: config.AddSpaceAfterAt,
		Driver: []zero.Driver{
			newDriver(config.Protocol),
		},
	}, process.GlobalInitMutex.Unlock)
}

// 获取驱动
func newDriver(p config.Protocol) zero.Driver {
	switch strings.ToLower(p.Type) {
	case Forward:
		// OneBot 正向 WS
		return driver.NewWebSocketClient(p.URL, p.AccessToken)
	case Reverse:
		// OneBot 反向 WS
		return driver.NewWebSocketServer(16, p.URL, p.AccessToken)
	case HTTP:
		return driver.NewHTTPClient(p.URL, p.AccessToken,
			p.CallerURL, p.CallerToken)
	default:
		return nil
	}
}

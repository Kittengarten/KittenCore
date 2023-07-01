// Package kitten 包含了 KittenCore 以及各插件的核心依赖结构体、方法和函数，以及固有的响应（如戳一戳）
package kitten

import (
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	poke     = rate.NewManager[int64](5*time.Minute, 9) // 戳一戳
	nickname = LoadConfig().NickName[0]                 // 昵称
	// Bot 实例
	Bot *zero.Ctx
	// BotSFACGchan 用于传送 Bot 实例的通道
	BotSFACGchan = make(chan *zero.Ctx)
	// Configs 来自 Bot 的配置文件
	Configs = LoadConfig()
)

const (
	randMax      = 100           // 随机数上限（不包含）
	path    Path = `config.yaml` // 配置文件名
)

func init() {
	// 向 SFACG 插件传入 Bot 实例
	go func() {
		for Bot == nil {
			Bot = zero.GetBot(Configs.SelfID)
		}
		BotSFACGchan <- Bot
	}()

	// 戳一戳
	zero.On(`notice/notify/poke`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			g int64              // 本群的群号
			u = ctx.Event.UserID // 发出 poke 的 QQ 号
		)
		if ctx.Event.DetailType == "private" {
			g = -ctx.Event.UserID
		} else {
			g = ctx.Event.GroupID
		}
		switch {
		case poke.Load(g).AcquireN(5):
			// 5 分钟共 8 块命令牌 一次消耗 5 块命令牌
			ctx.SendChain(message.Poke(u))
		case poke.Load(g).AcquireN(3):
			// 5 分钟共 8 块命令牌 一次消耗 3 块命令牌
			ctx.SendChain(message.At(u), TextOf(`请不要拍%s >_<`, nickname))
		case poke.Load(g).Acquire():
			// 5 分钟共 8 块命令牌 一次消耗 1 块命令牌
			ctx.SendChain(message.At(u), TextOf("喂(#`O′) 拍%s干嘛！（好感 - %d）", nickname, rand.Intn(randMax)+1))
		default:
			// 频繁触发，不回复
		}
	})

	// 图片，用于让 Bot 发送图片，可通过 CQ 码、链接等，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.Send(message.Image(ctx.State[`args`].(string)))
	})
}

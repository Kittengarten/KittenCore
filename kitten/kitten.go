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
	config   = LoadConfig()                             // 全局配置文件
	poke     = rate.NewManager[int64](time.Minute*5, 9) // 戳一戳
	nickname = LoadConfig().NickName[0]                 // 昵称
	// Bot 用于传送 Bot 实例的通道
	Bot = make(chan *zero.Ctx)
)

const randMax = 100 // 随机数上限（不包含）

func init() {
	go func() {
		var bot *zero.Ctx
		for bot == nil {
			bot = zero.GetBot(config.SelfID)
		}
		Bot <- bot
	}()

	// 戳一戳
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var (
				gID = ctx.Event.GroupID // 本群的群号
				uID = ctx.Event.UserID  // 发出 poke 的 QQ 号
			)
			switch {
			case poke.Load(gID).AcquireN(5):
				// 5分钟共8块命令牌 一次消耗5块命令牌
				ctx.SendChain(message.Poke(uID))
			case poke.Load(gID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				ctx.SendChain(message.At(uID), TextOf("请不要拍%s >_<", nickname))
			case poke.Load(gID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				ctx.SendChain(message.At(uID), TextOf("喂(#`O′) 拍%s干嘛！（好感 - %d）", nickname, rand.Intn(randMax)+1))
			default:
				// 频繁触发，不回复
			}
		})
}

package kitten

import (
	"fmt"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var poke = rate.NewManager[int64](time.Minute*5, 9) // 戳一戳

func init() {
	zero.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			switch {
			case poke.Load(ctx.Event.GroupID).AcquireN(5):
				// 5分钟共8块命令牌 一次消耗5块命令牌
				ctx.SendChain(message.Poke(ctx.Event.UserID))
			case poke.Load(ctx.Event.GroupID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("请不要拍%s >_<", nickname)))
			case poke.Load(ctx.Event.GroupID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("喂(#`O′) 拍%s干嘛！（好感-1）", nickname)))
			default:
				// 频繁触发，不回复
			}
		})
}

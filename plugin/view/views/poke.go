package views

import (
	"math/rand/v2"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// Poke 戳一戳
func Poke(msgr *kitten.Messager) message.ID {
	o, err := msgr.Object()
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	n, err := o.Name()
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	switch limiter := pokeLimiter(msgr.Ctx); {
	case limiter.AcquireN(5):
		// 5 分钟共 9 块命令牌 一次消耗 5 块命令牌
		times.RandomDelayRange(time.Second, 2*time.Second)
		msgr.Poke()
		msgr.CallAction(`send_like`, map[string]any{
			`user_id`: msgr.Event.UserID,
			`times`:   rand.N(20) + 1, //nolint:gosec
		})
	case limiter.AcquireN(3):
		// 5 分钟共 9 块命令牌 一次消耗 3 块命令牌
		return msgr.SendWithImageFail(`请不要拍`, n, ` >_<`)
	case limiter.Acquire():
		// 5 分钟共 9 块命令牌 一次消耗 1 块命令牌
		//nolint:gosec
		return msgr.SendWithImageFailf("喂(#`O′) 拍%s干嘛！\n（好感 - %d）", n, rand.N(100)+1)
		// 频繁触发，不回复
	}
	return message.ID{}
}

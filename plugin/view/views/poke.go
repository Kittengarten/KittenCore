package views

import (
	"math/rand/v2"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Poke 戳一戳
func Poke(handler *msg.Handler) message.ID {
	o, err := handler.Object()
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	n, err := o.Name(handler)
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	switch limiter := pokeLimiter(handler.Ctx); {
	case limiter.AcquireN(5):
		// 5 分钟共 9 块命令牌 一次消耗 5 块命令牌
		select {
		case <-times.RandDelayRange(time.Second, 2*time.Second):
			handler.Poke()
			handler.CallActionWithContext(
				`send_like`,
				zero.H{
					`user_id`: handler.Event().UserID,
					`times`:   rand.N(20) + 1, //nolint:gosec
				},
			)
		case <-handler.Done():
			handler.SendWithImageFail(handler.Err())
		}
	case limiter.AcquireN(3):
		// 5 分钟共 9 块命令牌 一次消耗 3 块命令牌
		return handler.SendWithImageFail(`请不要拍`, n, ` >_<`)
	case limiter.Acquire():
		// 5 分钟共 9 块命令牌 一次消耗 1 块命令牌
		//nolint:gosec
		return handler.SendWithImageFailf("喂(#`O′) 拍%s干嘛！（好感 - %d）", n, rand.N(100)+1)
	}
	// 频繁触发，不回复
	return message.ID{}
}

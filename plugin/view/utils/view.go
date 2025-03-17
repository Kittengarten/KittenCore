package utils

import (
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/rate"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var pokeLimiter = rate.New(rate.ByGroup, 5*time.Minute, 9) // 戳一戳限速

// View 查看
func View(msgr *kitten.Messager, service string, logFilePath core.Path) message.ID {
	switch name, who := func() (name, who string) {
		name = core.CleanAll(msgr.Args(), false)
		who = name
		for _, n := range kitten.MainConfig().NickName {
			if name == n {
				who = kitten.MainConfig().NickName[0]
			}
		}
		return
	}(); who {
	case kitten.MainConfig().NickName[0]:
		t := cpuTemperature(logFilePath)
		return msgr.
			Reply().
			AtLf().
			Image(core.FilePath(service,
				strconv.Itoa(
					getPerf(cpuPercent(), percent(getMem()), t),
				)+`.png`),
			).
			Text(viewString(msgr, name, t)).
			Send()
	case `鸡汤`:
		return send(msgr, jiTang, false)
	case `情话`:
		return send(msgr, qingHua, false)
	case `疯狂星期四`, `疯四`:
		if time.Now().Weekday() != time.Thursday {
			// 如果不是星期四，则不发送
			return msgr.SendWithImageFail(`今天不是星期四喵！`)
		}
		return send(msgr, kfc, false)
	case `一言`:
		return sendYiYan(msgr)
	case `waifu`, `老婆`, `随机老婆`:
		return sendWaifu(msgr)
	case `麻将`, `庄家`:
		return sendMahjong(msgr, true)
	case `闲家`:
		return sendMahjong(msgr, false)
	default:
		// 花语
		return SendFlower(msgr)
	}
}

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
		core.RandomDelayRange(time.Second, 2*time.Second)
		msgr.Poke()
		msgr.CallAction(`send_like`, map[string]any{
			`user_id`: msgr.Event.UserID,
			`times`:   rand.N(20) + 1,
		})
	case limiter.AcquireN(3):
		// 5 分钟共 9 块命令牌 一次消耗 3 块命令牌
		return msgr.SendWithImageFail(`请不要拍`, n, ` >_<`)
	case limiter.Acquire():
		// 5 分钟共 9 块命令牌 一次消耗 1 块命令牌
		return msgr.SendWithImageFailOf("喂(#`O′) 拍%s干嘛！\n（好感 - %d）", n, rand.N(100)+1)
		// 频繁触发，不回复
	}
	return message.ID{}
}

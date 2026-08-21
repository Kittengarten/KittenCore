// Package views 查看信息
package views

import (
	"strconv"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/plugin/view/img"
	"github.com/Kittengarten/KittenCore/plugin/view/perf"
	"github.com/Kittengarten/KittenCore/plugin/view/text"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 戳一戳限速
var pokeLimiter = rate.New(rate.ByGroup, 5*time.Minute, 9)

// View 查看
func View(handler *msg.Handler, service string, logFilePath fio.Path) message.ID {
	switch name, who := func() (name, who string) {
		o, err := handler.Object()
		if err != nil {
			handler.Quote().AtLf().Text(err).Send()
			return
		}
		name = str.Clean(handler.Args(), false)
		who = name
		for _, n := range zero.BotConfig.NickName {
			if name == n {
				who = zero.BotConfig.NickName[0]
				if err = o.SetName(handler, name); err != nil {
					handler.Quote().AtLf().Text(err).Send()
					return
				}
			}
		}
		return
	}(); who {
	case zero.BotConfig.NickName[0]:
		// TODO: 使用一整张图片返回
		ct := perf.CPUTemperature(logFilePath)
		return handler.Quote().AtLf().
			Image(fio.NewPath(service, strconv.Itoa(perf.Level(handler, ct))+`.png`)).
			Text(perf.ViewString(handler, name, ct, text.Weight())).Send()
	case `鸡汤`:
		return text.SendJiTang(handler)
	case `情话`:
		return text.SendQingHua(handler)
	case `疯狂星期四`, `疯四`:
		return text.SendKFC(handler)
	case `一言`:
		return text.SendYiYan(handler)
	case `waifu`, `老婆`, `随机老婆`:
		return img.SendWaifu(handler)
	case `麻将`, `庄家`:
		return text.SendMahjong(handler, true)
	case `闲家`:
		return text.SendMahjong(handler, false)
	case ``:
		return message.ID{}
	default:
		return handler.DoNotKnow()
	}
}

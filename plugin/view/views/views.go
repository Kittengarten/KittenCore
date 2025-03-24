// Package views 查看信息
package views

import (
	"strconv"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/plugin/view/img"
	"github.com/Kittengarten/KittenCore/plugin/view/perf"
	"github.com/Kittengarten/KittenCore/plugin/view/text"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var pokeLimiter = rate.New(rate.ByGroup, 5*time.Minute, 9) // 戳一戳限速

// View 查看
func View(msgr *kitten.Messager, service string, logFilePath fio.Path) message.ID {
	switch name, who := func() (name, who string) {
		name = str.CleanAll(msgr.Args(), false)
		who = name
		for _, n := range kitten.MainConfig().NickName {
			if name == n {
				who = kitten.MainConfig().NickName[0]
			}
		}
		return
	}(); who {
	case kitten.MainConfig().NickName[0]:
		return msgr.
			Reply().
			AtLf().
			Image(fio.NewPath(service, strconv.Itoa(perf.Level(logFilePath))+`.png`)).
			Text(perf.ViewString(msgr, name, logFilePath)).
			Send()
	case `鸡汤`:
		return text.SendJiTang(msgr)
	case `情话`:
		return text.SendQingHua(msgr)
	case `疯狂星期四`, `疯四`:
		if time.Now().Weekday() != time.Thursday {
			// 如果不是星期四，则不发送
			return msgr.SendWithImageFail(`今天不是星期四喵！`)
		}
		return text.SendKFC(msgr)
	case `一言`:
		return text.SendYiYan(msgr)
	case `waifu`, `老婆`, `随机老婆`:
		return img.SendWaifu(msgr)
	case `麻将`, `庄家`:
		return text.SendMahjong(msgr, true)
	case `闲家`:
		return text.SendMahjong(msgr, false)
	default:
		// 花语
		return text.SendFlower(msgr)
	}
}

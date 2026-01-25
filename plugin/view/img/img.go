// Package img 查看图片
package img

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/msg/seg"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const waifu = `https://www.thiswaifudoesnotexist.net/example-%d.jpg` // AI 随机老婆

// SendWaifu 发送 AI 随机老婆
func SendWaifu(handler *msg.Handler) message.ID {
	//nolint:gosec
	return handler.Quote().AtLf().Image(fio.NewPath(fmt.Sprintf(waifu, rand.N(100001)))).Send()
}

// SendImage 从 ctx 参数中的 URL 发送图片
func SendImage(handler *msg.Handler, su bool) message.ID {
	img := handler.Args()
	if !su && strings.HasPrefix(img, `file://`) {
		return handler.SendWithImageFail(`权限不足喵！`)
	}
	return handler.Quote().AtLf().Image(fio.NewPath(img)).Send()
}

// Scan 扫码
func Scan(handler *msg.Handler, su, mpp func(ctx *zero.Ctx) bool) message.ID {
	// 如果没有提供图片，从链接解析
	img := handler.Args()
	if img == `` || strings.HasPrefix(img, `[`) ||
		!su(handler.Ctx) && strings.HasPrefix(img, `file://`) {
		// 需要一张图片
		if mpp(handler.Ctx) {
			return scanQRCode(handler)
		}
		return handler.SendWithImageFail(`没有收到图片，命令已过期喵！`)
	}
	s, err := msg.ScanQRCode(handler, img)
	if err != nil {
		return handler.SendWithImageFail(`扫描失败喵！`, err)
	}
	return handler.Quote().AtLf().Text(s).Send()
}

// 从上下文的消息所附带的图片中扫描二维码并发送结果（支持多张图片）
func scanQRCode(handler *msg.Handler) message.ID {
	r := make([]string, 0, len(handler.Event().Message))
	for _, e := range handler.Event().Message {
		if e.Type != seg.Image || mio.GetImagePath(e) == `` {
			continue
		}
		s, err := handler.ScanQRCodeInQQ(mio.GetImagePath(e).String())
		if err != nil {
			r = append(r, err.Error())
			continue
		}
		r = append(r, s.String())
	}
	return handler.Quote().AtLf().Text(strings.Join(r, "\n")).Send()
}

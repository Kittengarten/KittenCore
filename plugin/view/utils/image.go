package utils

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/io"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const waifu = `https://www.thiswaifudoesnotexist.net/example-%d.jpg` // AI 随机老婆

// 发送 AI 随机老婆
func sendWaifu(msgr *kitten.Messager) message.ID {
	return msgr.Reply().AtLf().Image(io.Path(fmt.Sprintf(waifu, rand.N(100001)))).Send()
}

// SendImage 从 ctx 参数中的 URL 发送图片
func SendImage(msgr *kitten.Messager, su bool) message.ID {
	img := msgr.Args()
	if !su && strings.HasPrefix(img, `file://`) {
		return msgr.SendWithImageFail(`权限不足喵！`)
	}
	return msgr.Reply().AtLf().Image(io.Path(img)).Send()
}

// Scan 扫码
func Scan(msgr *kitten.Messager, su, hp bool, mpp func() bool) message.ID {
	if hp {
		// 如果提供了图片，直接使用
		return scanQRCode(msgr)
	}
	// 如果没有提供图片，从链接解析
	img := msgr.Args()
	if img == `` || !su && strings.HasPrefix(img, `file://`) {
		// 需要一张图片
		if mpp() {
			return scanQRCode(msgr)
		}
		return msgr.SendWithImageFail(`没有收到图片，命令已过期喵！`)
	}
	s, err := kitten.ScanQRCode(img)
	if err != nil {
		return msgr.SendWithImageFail(`扫描失败喵！`, err)
	}
	return msgr.Reply().AtLf().Text(s).Send()
}

// 从上下文的消息所附带的图片中扫描二维码并发送结果（支持多张图片）
func scanQRCode(msgr *kitten.Messager) message.ID {
	r := make([]string, 0, len(msgr.Event.Message))
	for _, seg := range msgr.Event.Message {
		if seg.Type != `image` || seg.Data[`file`] == `` {
			continue
		}
		s, err := msgr.ScanQRCodeInQQ(seg.Data[`file`])
		if err != nil {
			kitten.Error(err)
			r = append(r, err.Error())
			continue
		}
		r = append(r, s.String())
	}
	return msgr.Reply().AtLf().Text(strings.Join(r, "\n")).Send()
}

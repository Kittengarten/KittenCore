// Package voice 语音
package voice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	alipayvoiceURL = "https://mm.cqu.cc/share/zhifubaodaozhang/mp3/%v.mp3" // 支付宝到账语音
	maxMoney       = 1e8                                                   // 最大金额
	minMoney       = 1e-2                                                  // 最小金额
)

// SendAlipayVoice 发送支付宝到账语音
func SendAlipayVoice(handler *msg.Handler) message.ID {
	var (
		s      = strings.TrimSpace(handler.Args())
		i, err = strconv.ParseFloat(s, 64)
	)
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	if i >= maxMoney {
		return handler.SendWithImageFail(`金额太大，禁止获取喵！`)
	}
	if i < minMoney {
		return handler.SendWithImageFail(`金额至少为`, minMoney, `喵！`)
	}
	return handler.Record(fmt.Sprintf(alipayvoiceURL, s)).Send()
}

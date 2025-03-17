package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	alipayvoiceURL = "https://mm.cqu.cc/share/zhifubaodaozhang/mp3/%v.mp3" // 支付宝到账语音
	maxMoney       = 100_000_000                                           // 最大金额
	minMoney       = 0.01                                                  // 最小金额
)

// SendAlipayVoice 发送支付宝到账语音
func SendAlipayVoice(msgr *kitten.Messager) message.ID {
	var (
		s      = strings.TrimSpace(msgr.Args())
		i, err = strconv.ParseFloat(s, 64)
	)
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	if i >= maxMoney {
		return msgr.SendWithImageFail(`金额太大，禁止获取喵！`)
	}
	if i < minMoney {
		return msgr.SendWithImageFail(`金额至少为`, minMoney, `喵！`)
	}
	return msgr.Record(fmt.Sprintf(alipayvoiceURL, s)).Send()
}

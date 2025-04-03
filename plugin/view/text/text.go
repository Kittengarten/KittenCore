// Package text 查看文本
package text

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/mahjong"

	"github.com/Kittengarten/KittenAnno/wta"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	jiTang  = `https://api.btstu.cn/yan/api.php?charset=utf-8&encode=text` // 鸡汤
	qingHua = `https://xiaobai.klizi.cn/API/other/wtqh.php`                // 情话
	kfc     = `http://api.jixs.cc/api/wenan-fkxqs/index.php`               // 疯狂星期四
	yiYan   = `https://v1.hitokoto.cn/?c=a&c=b&c=c&c=d&c=h&c=i`            // 动漫 漫画 游戏 文学 影视 诗词（一言）
)

// SendHTML 发送网页 HTML 内容，lf 控制内容是否换行
func SendHTML(msgr *kitten.Messager, url string, lf bool) message.ID {
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GET(url)
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	var s strings.Builder
	if _, err := io.Copy(&s, b); err != nil {
		return msgr.SendWithImageFail(err)
	}
	return msgr.Reply().AtLf().Text(str.CleanAll(s.String(), lf)).Send()
}

// SendJiTang 发送鸡汤
func SendJiTang(msgr *kitten.Messager) message.ID {
	return SendHTML(msgr, jiTang, false)
}

// SendQingHua 发送情话
func SendQingHua(msgr *kitten.Messager) message.ID {
	return SendHTML(msgr, qingHua, false)
}

// SendKFC 发送疯狂星期四
func SendKFC(msgr *kitten.Messager) message.ID {
	if time.Now().Weekday() != time.Thursday {
		// 如果不是星期四，则不发送
		return msgr.SendWithImageFail(`今天不是星期四喵！`)
	}
	return SendHTML(msgr, kfc, false)
}

// SendYiYan 发送一言
func SendYiYan(msgr *kitten.Messager) message.ID {
	var (
		// 获取 HTTP 响应体，失败则返回
		b, err = shttp.GET(yiYan)
		rsp    struct {
			Hitokoto string `json:"hitokoto"`
			From     string `json:"from"`
			FromWho  string `json:"from_who"`
		}
	)
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	if err := json.NewDecoder(b).Decode(&rsp); err != nil {
		return msgr.SendWithImageFail(err)
	}
	return msgr.Reply().AtLf().Text(rsp.Hitokoto, `
	出自：`, rsp.From, func() string {
		if rsp.FromWho == `` {
			return ``
		}
		return `
	作者：` + rsp.FromWho
	}()).Send()
}

// SendMahjong 发送麻将配牌
func SendMahjong(msgr *kitten.Messager, dealer bool) message.ID {
	return msgr.Reply().AtLf().Text(string(mahjong.New(dealer))).Send()
}

// GetWTA 返回世界树纪元
func GetWTA(msgr *kitten.Messager) string {
	o, err := msgr.Object()
	if err != nil {
		return err.Error()
	}
	n := str.CleanAll(msgr.Args(), false)
	if err = o.SetName(n); err != nil {
		return err.Error()
	}
	a, _ := wta.GetAnno()
	return n + `报时：
日期：	` + a.DateStr() + `
时间：	` + a.String() + `
琴弦：	` + a.Chord() + `
花卉：	` + a.Flower() + `
` + a.ElementalAndImageryStr()
}

// Weight 返回叠猫猫体重字符串
func Weight() string {
	if kitten.Weight == 0 {
		return ``
	}
	return fmt.Sprintf(`	❤	叠猫猫体重：	%.1f kg`, float64(kitten.Weight)/10)
}

// SendFlower 返回花语
var SendFlower = func(_ *kitten.Messager) message.ID {
	return message.ID{}
}

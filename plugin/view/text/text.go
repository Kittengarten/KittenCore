// Package text 查看文本
package text

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/mahjong"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/Kittengarten/KittenAnno/wta"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	// Checker 检查者
	Checker interface {
		Check(ctx context.Context, name, info string) string
	}
	// Sender 发送者
	Sender interface {
		// // 发送花语
		// SendFlower(handler *msg.Handler) message.ID
	}
)

// Export 导出接口
var Export struct {
	Checker // Checker 检查者
	Sender  // Sender 发送者
}

const (
	jiTang  = `https://api.btstu.cn/yan/api.php?charset=utf-8&encode=text` // 鸡汤
	qingHua = `https://xiaobai.klizi.cn/API/other/wtqh.php`                // 情话
	kfc     = `https://api.pearktrue.cn/api/kfc/`                          // 疯狂星期四
	yiYan   = `https://v1.hitokoto.cn/?c=a&c=b&c=c&c=d&c=h&c=i`            // 动漫 漫画 游戏 文学 影视 诗词（一言）
)

// SendJiTang 发送鸡汤
func SendJiTang(handler *msg.Handler) message.ID {
	return SendHTML(handler, jiTang, false)
}

// SendQingHua 发送情话
func SendQingHua(handler *msg.Handler) message.ID {
	return SendHTML(handler, qingHua, false)
}

// SendKFC 发送疯狂星期四
func SendKFC(handler *msg.Handler) message.ID {
	if time.Now().Weekday() != time.Thursday {
		// 如果不是星期四，则不发送
		return handler.SendWithImageFail(`今天不是星期四喵！`)
	}
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GETWithContext(handler, kfc)
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	defer shttp.Clear(b)
	rsp := new(struct {
		Code int
		Msg  string
		Text string
	})
	if err := json.NewDecoder(b).Decode(rsp); err != nil {
		return handler.SendWithImageFail(err)
	}
	if rsp.Code != 200 || rsp.Msg != `获取成功` {
		return handler.SendWithImageFail(rsp.Code, `：`, rsp.Msg)
	}
	return handler.Quote().AtLf().Text(rsp.Text).Send()
}

// SendYiYan 发送一言
func SendYiYan(handler *msg.Handler) message.ID {
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GETWithContext(handler, yiYan)
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	defer shttp.Clear(b)
	rsp := new(struct {
		Hitokoto string `json:"hitokoto"`
		From     string `json:"from"`
		FromWho  string `json:"from_who"`
	})
	if err := json.NewDecoder(b).Decode(rsp); err != nil {
		return handler.SendWithImageFail(err)
	}
	return handler.Quote().AtLf().Text(rsp.Hitokoto, `
	出自：`, rsp.From, func() string {
		if rsp.FromWho == `` {
			return ``
		}
		return `
	作者：` + rsp.FromWho
	}()).Send()
}

// SendHTML 发送网页 HTML 内容，lf 控制内容是否换行
func SendHTML(handler *msg.Handler, url string, lf bool) message.ID {
	// 获取 HTTP 响应体，失败则返回
	b, err := shttp.GETWithContext(handler, url)
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	defer shttp.Clear(b)
	s := new(strings.Builder)
	s.Grow(2048)
	if _, err := io.Copy(s, b); err != nil {
		return handler.SendWithImageFail(err)
	}
	return handler.Quote().AtLf().Text(str.Clean(s.String(), lf)).Send()
}

// SendMahjong 发送麻将配牌
func SendMahjong(handler *msg.Handler, dealer bool) message.ID {
	return handler.Quote().AtLf().Text(string(mahjong.New(dealer))).Send()
}

// GetWTA 返回世界树纪元
func GetWTA(name string) string {
	a, _ := wta.GetAnno()
	return name + `报时：
日期：	` + a.DateStr() + `
时间：	` + a.String() + `
琴弦：	` + a.Chord() + `
花卉：	` + a.Flower() + `
` + a.ElementalAndImageryStr()
}

// Weight 返回叠猫猫体重字符串
func Weight() string {
	if stat.Weight == 0 {
		return ``
	}
	return fmt.Sprintf(`	❤	叠猫猫体重：	%.1f kg`, float64(stat.Weight)/10)
}

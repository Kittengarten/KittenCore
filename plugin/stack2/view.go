package stack2

import (
	"fmt"
	"strings"
	"time"

	hookcharts "github.com/Kittengarten/KittenCore/internal/hook/charts"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/vicanso/go-charts/v2"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查看叠猫猫，不修改原数据
func (d *data) view(handler *msg.Handler, all bool, r ...replacer) {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	var (
		s = d.getStack() // 获取叠猫猫队列
		l = len(s)       // 叠猫猫队列长度（避免闭包捕获被修改后的值）
	)
	utils.Go(`查看叠猫猫设置群昵称`, func() { setCard(handler, l) })
	_ = sendTextf(handler, r[0], true, `【叠猫猫队列】
现在有 %d 只猫猫
总重量为 %.1f kg
————%s`,
		l,
		i2f(s.totalWeight()),
		func() any {
			// TODO: 私聊无论如何均可以查看全部
			if !all {
				// 查看省略版
				return s.Str()
			}
			// 查看全部（先发送前 50 条）
			sr := s[max(0, l-50):]
			// 剩余部分
			s = s[:l-len(sr)]
			return &sr
		}())
	if !all {
		// 查看省略版，无需发送后续文字
		return
	}
	for len(s) > 0 {
		select {
		case <-times.RandDelayRange(time.Second, 2*time.Second):
			// 发送剩余部分的前 50 条
			sr := s[max(0, len(s)-50):]
			sendText(handler, r[0], false, &sr)
			// 剩余部分的剩余部分
			s = s[:len(s)-len(sr)]
		case <-handler.Done():
			handler.SendWithImageFail(handler.Err())
		}
	}
}

// 查看叠猫猫图片
func (d *data) viewImage(handler *msg.Handler, r ...replacer) message.ID {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	var (
		s = d.getStack() // 获取叠猫猫队列
		l = len(s)       // 叠猫猫队列长度
	)
	if l < 2 {
		return message.ID{}
	}
	var (
		values = [][]float64{make([]float64, l)} // 叠猫猫图示数据
		str    = make([]string, l)               // 叠猫猫图示文字
	)
	for h, m := range s {
		values[0][h] = i2f(m.Weight)
		str[h] = strings.ReplaceAll(func() string {
			h := msg.NewWithContext(handler, globalCtx)
			if r[0].location == cockroach {
				return fmt.Sprintf(`【%s】翼展 %.1f cm`,
					r[0].Replace(m.getType(h).String()),
					i2f(m.Weight),
				)
			}
			return fmt.Sprintf(`%s（%d）%.1f %s %s`,
				m.TitleCardOrNickName(h),
				m.Int(),
				i2f(m.Weight),
				r[0].Replace(`kg`),
				r[0].Replace(m.getType(handler).String()),
			)
		}(), `	`, ``)
	}
	p, err := setViewChart(values, str, l, r...)
	if err != nil {
		return sendWithImageFail(handler, r[0], err)
	}
	return sendImage(handler, p, r...)
}

// 设置查看图表
func setViewChart(v [][]float64, s []string, l int, r ...replacer) (*charts.Painter, error) {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	var (
		width   = max(min(3840, 160+320*l), 960)
		height  = max(min(2160, 45+90*l), 270)
		padding = width >> 5
	)
	return charts.HorizontalBarRender(
		v,
		charts.FontFamilyOptionFunc(hookcharts.FontName),
		charts.WidthOptionFunc(width),
		charts.HeightOptionFunc(height),
		charts.PaddingOptionFunc(charts.Box{
			Top:    padding,
			Right:  padding,
			Bottom: padding,
			Left:   padding,
		}),
		charts.PNGTypeOption(),
		charts.TitleTextOptionFunc(r[0].Replace(`叠猫猫队列`)),
		charts.LegendLabelsOptionFunc([]string{r[0].Replace(`体重（kg）`)}),
		charts.YAxisDataOptionFunc(s),
	)
}

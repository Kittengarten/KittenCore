package stack2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/vicanso/go-charts/v2"

	"github.com/FloatTech/zbputils/img/text"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查看叠猫猫，不修改原数据
func (d *data) view(msgr *kitten.Messager, all bool) {
	s := d.getStack() // 获取叠猫猫队列
	go setCard(msgr, len(s))
	_ = sendTextOf(msgr, `【叠猫猫队列】
现在有 %d 只猫猫
总重量为 %.1f kg
————%s`,
		len(s),
		itof(s.totalWeight()),
		func() any {
			if !all {
				// 查看省略版
				return s.Str()
			}
			// 查看全部（先发送前 50 条）
			sr := s[max(0, len(s)-50):]
			// 剩余部分
			s = s[:len(s)-len(sr)]
			return &sr
		}())
	if !all {
		// 查看省略版，无需发送后续文字
		return
	}
	for len(s) > 0 {
		core.RandomDelayRange(time.Second, 2*time.Second)
		// 发送剩余部分的前 50 条
		sr := s[max(0, len(s)-50):]
		sendText(msgr, &sr)
		// 剩余部分的剩余部分
		s = s[:len(s)-len(sr)]
	}
}

// 初始化字体
func initFont() error {
	// 获取字体数据
	buf, err := core.FilePath(text.GlowSansFontFile).ReadBytes()
	if err != nil {
		return err
	}
	// 安装字体
	const fontName = `glow`
	if err := charts.InstallFont(fontName, buf.Bytes()); err != nil {
		return err
	}
	// 加载字体
	if font, err := charts.GetFont(fontName); err == nil {
		// 设置默认字体
		charts.SetDefaultFont(font)
	}
	return err
}

// 查看叠猫猫图片
func (d *data) viewImage(msgr *kitten.Messager) message.ID {
	var (
		s = d.getStack() // 获取叠猫猫队列
		l = len(s)       // 叠猫猫队列长度
	)
	if l < 2 {
		return message.ID{}
	}
	var (
		values = make([][]float64, 1) // 叠猫猫图示数据
		str    = make([]string, l)    // 叠猫猫图示文字
	)
	values[0] = make([]float64, l) // 初始化二维切片
	for h, m := range s {
		values[0][h] = itof(m.Weight)
		str[h] = strings.ReplaceAll(func() string {
			if globalLocation == cockroach {
				return fmt.Sprintf(`【%s】翼展 %.1f cm`,
					m.getType(GlobalMessager).String(),
					itof(m.Weight),
				)
			}
			return fmt.Sprintf(`%s（%d）%.1f %s %s`,
				m.TitleCardOrNickName(GlobalMessager),
				m.Int(),
				itof(m.Weight),
				l10nReplacer().Replace(`kg`),
				l10nReplacer().Replace(m.getType(msgr).String()),
			)
		}(), `	`, ``)
	}
	p, err := setViewChart(values, str, l)
	if err != nil {
		return sendWithImageFail(msgr, err)
	}
	return sendImage(msgr, p)
}

// 设置查看图表
func setViewChart(v [][]float64, s []string, l int) (*charts.Painter, error) {
	charts.SetDefaultWidth(max(min(3840, 160+320*l), 960))
	charts.SetDefaultHeight(max(min(2160, 90*l), 270))
	return charts.HorizontalBarRender(
		v,
		charts.TitleTextOptionFunc(l10nReplacer().Replace(`叠猫猫队列`)),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  40,
			Bottom: 20,
			Left:   40,
		}),
		charts.LegendLabelsOptionFunc([]string{l10nReplacer().Replace(`体重（kg）`)}),
		charts.YAxisDataOptionFunc(s),
	)
}

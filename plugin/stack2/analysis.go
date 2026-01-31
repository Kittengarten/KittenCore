package stack2

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	hookcharts "github.com/Kittengarten/KittenCore/internal/hook/charts"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/vicanso/go-charts/v2"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	// 概率组
	chance struct {
		p float64 // 压坏概率
		f float64 // 摔下概率
		s float64 // 成功概率
	}

	// 分析文本
	tips struct {
		Flat         []string `yaml:"平地摔"`
		Tiger        []string `yaml:"老虎"`
		RBQ          []string `yaml:"绒布球"`
		Mew          []string `yaml:"奶猫"`
		Press        []string `yaml:"压"`
		Fall         []string `yaml:"摔"`
		Fail         []string `yaml:"危"`
		Normal       []string `yaml:"普通"`
		Rare         []string `yaml:"稀有"`
		Kittengarten []string `yaml:"幼喵园"`
		Custom       []string `yaml:"自定义"` // 任何情况下都可能触发
	}
)

var (
	//go:embed tip.yaml
	tipYAML  []byte
	tipSlice tips // 叠猫猫小贴士
)

// 分析叠猫猫，不修改原数据
func (d *data) analysis(handler *msg.Handler) {
	if globalLocation == cockroach {
		handler.SendWithImageFail(cockroachDoNotAnalysis)
		return
	}
	c, f, img := d.generateAnalysis(handler)
	if !img {
		// 如果不需要图片，什么也不做
		return
	}
	select {
	case <-times.RandDelayRange(time.Second, 2*time.Second):
		d.analysisImage(handler, c, f)
	case <-handler.Done():
		handler.SendWithImageFail(handler.Err())
	}
}

// 生成分析并发送文本
func (d *data) generateAnalysis(handler *msg.Handler) (c chance, flat, img bool) {
	var (
		dr = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		s  = dr.getStack()    // 获取叠猫猫队列
	)
	m, err := dr.pre(handler) // 初始化自身
	if err != nil {
		// 如果不能加入，什么也不做
		return c, flat, img
	}
	// 如果能加入
	// 猫娘少女和成年猫娘以上享有分析图片特权
	img = 猫娘少女 <= m.getTypeID(handler)
	l := len(s) // 叠猫猫队列长度
	utils.Go(`叠猫猫分析设置群昵称`, func() { setCard(handler, l) })
	if l == 0 {
		// 如果是空队列
		flat = true
		c.f = chanceFlat(m) // 平地摔概率
		c.s = 1 - c.f       // 成功概率
		_ = sendTextf(handler, true, `【叠猫猫分析】
当前体重：　	%.1f kg
平地摔概率：	%.2f%%
成功概率：　	%.2f%%
%s`,
			i2f(m.Weight),
			100*c.f,
			100*c.s,
			tipFlat(handler),
		)
		return c, flat, img
	}
	// 如果是非空队列
	sn := slices.Clone(s)
	sn = append(sn, m)              // 用于压坏判定的队列
	c.p = sn.chancePressed(handler) // 压坏概率
	gp := func() float64 {
		if m.getTypeID(handler) <= 抱枕 ||
			s[l-1].getTypeID(handler) >= 幼年猫娘 ||
			s[0].getTypeID(handler) >= 猫车 {
			// 抱枕及以下的猫猫不会导致猫猫摔下去
			// 直接在猫娘以上级别的身上叠猫猫不会摔下去
			// 底座为猫车以上时，不会摔下去
			return 0
		}
		return m.chanceFall(s[l-1])
	}() // 不压坏的情况下，摔下去的概率
	c.f = (1 - c.p) * gp // 摔下概率
	c.s = 1 - c.p - c.f  // 成功概率
	_ = sendTextf(handler, true, `【叠猫猫分析】
猫堆高度：	%d
当前体重：	%.1f kg
%s%s%s%s%s`,
		l,
		i2f(m.Weight),
		chanceOutput(`压坏概率`, c.p),
		chanceOutput(`摔下概率`, c.f),
		chanceOutput(`成功概率`, c.s),
		chanceOutput(`清空概率`, chanceClear(handler, s, m)),
		tip(handler, m.Weight, c),
	)
	return c, flat, img
}

// 输出概率文本
func chanceOutput(s string, c float64) string {
	switch {
	case c == 0:
		return ``
	case c >= 1e-4:
		return fmt.Sprintf("%s：	%.2f%%\n", s, 100*c)
	default:
		return fmt.Sprintf("%s：	%.2E\n", s, c)
	}
}

// 计算清空猫堆的概率
func chanceClear(handler *msg.Handler, s data, m meow) float64 {
	var (
		sn   = append(s, m)              // 用于压坏判定的队列
		p, f = 1.0, 1.0                  // 每次的压坏、摔下概率
		p1   = sn.chancePressed(handler) // 压坏概率
		l    = len(s)                    // 猫堆高度
	)
	if l == 0 {
		// 如果猫堆本来就是空的，清空概率等于平地摔概率
		return chanceFlat(m)
	}
	var (
		cf = func() float64 {
			if m.getTypeID(handler) <= 抱枕 ||
				s[l-1].getTypeID(handler) >= 幼年猫娘 ||
				s[0].getTypeID(handler) >= 猫车 {
				// 抱枕及以下的猫猫不会导致猫猫摔下去
				// 直接在猫娘以上级别的身上叠猫猫不会摔下去
				// 底座为猫车以上时，不会摔下去
				return 0
			}
			return m.chanceFall(s[l-1])
		}() // 在不压坏的情况下，摔下去的概率
		ff = true // 计算摔下概率是否判定叠入的猫猫
	)
	for range sn {
		if p == 0 {
			break
		}
		if len(sn) == 1 {
			// 最后一只猫猫（叠入的猫猫本身），则已经清空
			sn = nil
			break
		}
		p *= sn.chancePressed(handler) // 每次的压坏概率
		if sn[0].getTypeID(handler) >= 猫娘萝莉 && len(sn) > 2 {
			// 底座是猫娘萝莉以上，则不会继续压坏
			// 此时剩余的猫堆高度大于 1，则无法清空
			p = 0
			break
			// 此时剩余的猫堆高度等于 1，则刚好清空
		}
		sn = sn[1:] // 去除压坏的猫猫
	}
	for range s {
		if f == 0 {
			break
		}
		if ff {
			// 如果是判定叠入的猫猫
			f *= cf
		}
		l = len(s)
		if l == 0 {
			break
		}
		if !ff {
			// 如果不是判定叠入的猫猫
			f *= m.chanceFall(s[l-1]) // 在不压坏的情况下，每次摔下去的概率
		}
		ff = false // 从此必然不是判定叠入的猫猫
		m = s[l-1]
		if m.getTypeID(handler) >= 猫娘少女 && l >= 2 {
			// 如果摔下去的是猫娘少女以上级别，则下方的猫猫不会继续摔下去
			f = 0
		}
		s = s[:l-1] // 去除摔下去的猫猫
	}
	return p + (1-p1)*f
}

// 平地摔小贴士
func tipFlat(ctx context.Context) string {
	if tipSlice, err = fio.LoadWithContext[tips](ctx, tipsPath, string(tipYAML)); err != nil {
		return err.Error()
	}
	if l := len(tipSlice.Flat); l != 0 {
		//nolint:gosec
		return strings.TrimSpace(tipSlice.Flat[rand.N(l)])
	}
	return ``
}

// 叠猫猫小贴士，除平地摔以外
func tip(handler *msg.Handler, w int, c chance) string {
	if tipSlice, err = fio.LoadWithContext[tips](handler, tipsPath, string(tipYAML)); err != nil {
		return err.Error()
	}
	t := make([]string, 0, 128)
	if w >= mapMeow[猫娘少女].weight {
		t = append(t, tipSlice.Tiger...)
	}
	if w <= 1 {
		t = append(t, tipSlice.RBQ...)
	}
	if w < 10 {
		t = append(t, tipSlice.Mew...)
	}
	if c.p >= 0.5 {
		t = append(t, tipSlice.Press...)
	}
	if c.f >= 0.5 {
		t = append(t, tipSlice.Fall...)
	}
	if c.s < 0.2 {
		t = append(t, tipSlice.Fail...)
	}
	if 0.5 <= c.s && c.s < 0.8 {
		t = append(t, tipSlice.Normal...)
		//nolint:gosec
		if math.Pow(math.E, math.E)*rand.Float64() < 1 {
			t = append(t, tipSlice.Rare...)
		}
	}
	if len(t) == 0 {
		t = tipSlice.Kittengarten
	}
	h := msg.NewWithContext(handler, globalCtx)
	//nolint:gosec
	return strings.NewReplacer(
		`{player}`,
		usr.NewQQ(h.Event().UserID).CallName(h),
	).Replace(strings.TrimSpace(t[rand.N(len(t))]))
}

// 叠猫猫分析图片
func (d *data) analysisImage(handler *msg.Handler, c chance, flat bool) message.ID {
	p, err := setAnalysisChart(
		func() []float64 {
			if flat {
				return []float64{c.f, c.s}
			}
			return []float64{c.p, c.f, c.s}
		}(),
		flat,
	)
	if err != nil {
		return sendWithImageFail(handler, err)
	}
	return sendImage(handler, p)
}

// 设置分析图表
func setAnalysisChart(v []float64, flat bool) (*charts.Painter, error) {
	return charts.PieRender(
		v,
		charts.FontFamilyOptionFunc(hookcharts.FontName),
		charts.WidthOptionFunc(640),
		charts.HeightOptionFunc(360),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 20,
			Left:   20,
		}),
		charts.PNGTypeOption(),
		charts.TitleOptionFunc(charts.TitleOption{
			Text:    l10n.Replace(`叠猫猫分析`),
			Subtext: `概率`,
			Left:    charts.PositionCenter,
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Orient: charts.OrientVertical,
			Data: func() []string {
				if flat {
					return []string{
						l10n.Replace(`平地摔`),
						`成功`,
					}
				}
				return []string{
					l10n.Replace(`压坏`),
					l10n.Replace(`摔下`),
					`成功`,
				}
			}(),
			Left: charts.PositionLeft,
		}),
		charts.PieSeriesShowLabel(),
	)
}

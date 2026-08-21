package stack2

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"

	hookcharts "github.com/Kittengarten/KittenCore/internal/hook/charts"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/vicanso/go-charts/v2"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 统计组
type quantity struct {
	绒布球, 奶猫, 抱枕, 小可爱, 大可爱, 猫娘, 老虎 int
}

// 叠猫猫排行榜，不修改原数据
func (d *data) rank(handler *msg.Handler, r ...replacer) {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	// 不活跃的猫猫不参与排行，但数据保留
	d.clear(handler, true)
	// 将猫猫按体重顺序排序
	slices.SortFunc(*d, func(m, n meow) int {
		if i := cmp.Compare(m.Weight, n.Weight); i != 0 {
			return i
		}
		return cmp.Compare(n.Int(), m.Int())
	})
	var (
		c  = len(*d) // 猫猫总数
		ra = slices.IndexFunc(*d, func(m meow) bool {
			return handler.Event().UserID == m.Int()
		}) // 发起查询的猫猫排名
	)
	if c == 0 {
		sendWithImageFail(handler, r[0], `还没有猫猫喵！`)
		return
	}
	if ra == -1 {
		sendWithImageFail(handler, r[0], `你太久或者从来没有加入过喵！`)
		return
	}
	var (
		w = make([]int, c) // 保存累计重量的切片
		q = quantity{}
	)
	for i, m := range *d {
		switch m.getTypeID(handler) {
		case 绒布球:
			q.绒布球++
		case 奶猫:
			q.奶猫++
		case 抱枕:
			q.抱枕++
		case 小可爱:
			q.小可爱++
		case 大可爱:
			q.大可爱++
		case 幼年猫娘, 猫娘萝莉, 猫娘少女, 成年猫娘:
			q.猫娘++
		case 小老虎, 大老虎:
			q.老虎++
		}
		if i == 0 {
			w[i] = m.Weight
			continue
		}
		w[i] = w[i-1] + m.Weight
	}
	var (
		a  = w[c-1] // 猫猫总重量
		wi int      // 重量积分
	)
	for i, v := range w {
		wi += i * v
	}
	s := (*d)[c-10:] // 叠猫猫排行
	_ = sendTextf(handler, r[0], true, `【叠猫猫排行】
你的当前体重为 %.1f kg
在 %d 只猫猫中排行第 %d 名
所有猫猫当前的总重量为 %.1f kg%s
猫猫体重的基尼系数为 %.3f
猫车以上：	%d	只
老虎：　　	%d	只
猫娘：　　	%d	只
大可爱：　	%d	只
小可爱：　	%d	只
抱枕：　　	%d	只
奶猫：　　	%d	只
绒布球：　	%d	只`,
		i2f((*d)[ra].Weight),
		c, c-ra,
		i2f(a),
		func() string {
			if !zero.UserOrGrpAdmin(handler.Ctx) {
				return ``
			}
			return fmt.Sprintf(`
————%s`, &s)
		}(),
		1-2*float64(wi)/math.Pow(float64(c), 2)/float64(a),
		c-q.绒布球-q.奶猫-q.抱枕-q.小可爱-q.大可爱-q.猫娘-q.老虎,
		q.老虎,
		q.猫娘,
		q.大可爱,
		q.小可爱,
		q.抱枕,
		q.奶猫,
		q.绒布球,
	)
	select {
	case <-times.RandDelayRange(time.Second, 2*time.Second):
		d.rankImage(handler, q, r...)
	case <-handler.Done():
		handler.SendWithImageFail(handler.Err())
	}
}

// 叠猫猫排行图片
func (d *data) rankImage(handler *msg.Handler, q quantity, r ...replacer) message.ID {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	p, err := setRankChart([]float64{
		float64(q.绒布球),
		float64(q.奶猫),
		float64(q.抱枕),
		float64(q.小可爱),
		float64(q.大可爱),
		float64(q.猫娘),
		float64(q.老虎),
		float64(len(*d) - q.绒布球 - q.奶猫 - q.抱枕 - q.小可爱 - q.大可爱 - q.猫娘 - q.老虎),
	}, r...)
	if err != nil {
		return sendWithImageFail(handler, r[0], err)
	}
	return sendImage(handler, p, r...)
}

// 设置排行图表
func setRankChart(v []float64, r ...replacer) (*charts.Painter, error) {
	if len(r) == 0 {
		r = []replacer{l10n(cat)}
	}
	return charts.PieRender(
		v,
		charts.FontFamilyOptionFunc(hookcharts.FontName),
		charts.WidthOptionFunc(1280),
		charts.HeightOptionFunc(720),
		charts.PaddingOptionFunc(charts.Box{
			Top:    40,
			Right:  40,
			Bottom: 40,
			Left:   40,
		}),
		charts.PNGTypeOption(),
		charts.TitleOptionFunc(charts.TitleOption{
			Text:    r[0].Replace(`叠猫猫排行`),
			Subtext: `数量`,
			Left:    charts.PositionCenter,
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Orient: charts.OrientVertical,
			Data: []string{
				r[0].Replace(`绒布球`),
				r[0].Replace(`奶猫`),
				r[0].Replace(`抱枕`),
				r[0].Replace(`小可爱`),
				r[0].Replace(`大可爱`),
				r[0].Replace(`猫娘`),
				r[0].Replace(`老虎`),
				r[0].Replace(`猫车以上`),
			},
			Left: charts.PositionLeft,
		}),
		charts.PieSeriesShowLabel(),
	)
}

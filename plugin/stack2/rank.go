package stack2

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"github.com/vicanso/go-charts/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 统计组
type quantity struct {
	绒布球, 奶猫, 抱枕, 小可爱, 大可爱, 猫娘, 老虎 int
}

// 叠猫猫排行榜，不修改原数据
func (d *data) rank(msgr *kitten.Messager) {
	// 不活跃的猫猫不参与排行，但数据保留
	d.clear(msgr, true)
	// 将猫猫按体重顺序排序
	slices.SortFunc(*d, func(m, n meow) int {
		if m.Weight < n.Weight {
			return -1
		}
		if m.Weight > n.Weight {
			return 1
		}
		return cmp.Compare(n.Int(), m.Int())
	})
	var (
		c = len(*d) // 猫猫总数
		r = slices.IndexFunc(*d, func(m meow) bool {
			return msgr.Event.UserID == m.Int()
		}) // 发起查询的猫猫排名
	)
	if c == 0 {
		sendWithImageFail(msgr, `还没有猫猫喵！`)
	}
	if r == -1 {
		sendWithImageFail(msgr, `你太久或者从来没有加入过喵！`)
	}
	var (
		w = make([]int, c) // 保存累计重量的切片
		q = quantity{}
	)
	for i, m := range *d {
		switch m.getTypeID(msgr) {
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
	_ = sendTextOf(msgr, `【叠猫猫排行】
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
		itof((*d)[r].Weight),
		c, c-r,
		itof(a),
		func() string {
			if !zero.UserOrGrpAdmin(msgr.Ctx) {
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
	times.RandomDelayRange(time.Second, 2*time.Second)
	d.rankImage(msgr, q)
}

// 叠猫猫排行图片
func (d *data) rankImage(msgr *kitten.Messager, q quantity) message.ID {
	values := []float64{
		float64(q.绒布球),
		float64(q.奶猫),
		float64(q.抱枕),
		float64(q.小可爱),
		float64(q.大可爱),
		float64(q.猫娘),
		float64(q.老虎),
		float64(len(*d) - q.绒布球 - q.奶猫 - q.抱枕 - q.小可爱 - q.大可爱 - q.猫娘 - q.老虎),
	}
	p, err := setRankChart(values)
	if err != nil {
		return sendWithImageFail(msgr, err)
	}
	return sendImage(msgr, p)
}

// 设置排行图表
func setRankChart(v []float64) (*charts.Painter, error) {
	charts.SetDefaultWidth(1280)
	charts.SetDefaultHeight(720)
	return charts.PieRender(
		v,
		charts.TitleOptionFunc(charts.TitleOption{
			Text:    l10nReplacer().Replace(`叠猫猫排行`),
			Subtext: `数量`,
			Left:    charts.PositionCenter,
		}),
		charts.PaddingOptionFunc(charts.Box{
			Top:    20,
			Right:  20,
			Bottom: 20,
			Left:   20,
		}),
		charts.LegendOptionFunc(charts.LegendOption{
			Orient: charts.OrientVertical,
			Data: []string{
				l10nReplacer().Replace(`绒布球`),
				l10nReplacer().Replace(`奶猫`),
				l10nReplacer().Replace(`抱枕`),
				l10nReplacer().Replace(`小可爱`),
				l10nReplacer().Replace(`大可爱`),
				l10nReplacer().Replace(`猫娘`),
				l10nReplacer().Replace(`老虎`),
				l10nReplacer().Replace(`猫车以上`),
			},
			Left: charts.PositionLeft,
		}),
		charts.PieSeriesShowLabel(),
	)
}

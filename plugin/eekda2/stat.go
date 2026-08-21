package eekda2

import (
	"cmp"
	"context"
	"maps"
	"slices"

	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 查询被吃次数
func getStat(ctx *zero.Ctx) {
	statPath.RLock()
	defer statPath.RUnlock()
	var (
		co, cancel = context.WithTimeout(context.Background(), core.Timeout)
		handler    = msg.NewWithContext(co, ctx)
		s, err     = statPath.LoadWithContext[stat](handler, fio.Empty)
	)
	defer cancel()
	if err != nil {
		handler.SendWithImageFail(err)
	}
	i := slices.IndexFunc(s, func(f food) bool {
		return ctx.Event.UserID == f.Int()
	})
	if i == -1 {
		handler.DoNotKnow()
		return
	}
	c, err := todayPath.LoadWithContext[config](handler, fio.Empty)
	if err != nil {
		handler.SendWithImageFail(err)
	}
	for _, t := range c {
		if slices.Contains(t.Group, usr.NewQQGroup(ctx.Event.GroupID)) {
			// 如果当前角色在本群已注册，跳过
			continue
		}
		// 如果当前角色在本群未注册，移除
		maps.DeleteFunc(s[i].Stat, func(k string, _ [mealsPerDay]int) bool {
			return k == t.ID
		})
	}
	if len(s[i].Stat) == 0 {
		handler.DoNotKnow()
		return
	}
	handler.Quote().AtLf().Text(&s[i]).Send()
}

// 统计被吃次数
func doStat(handler *msg.Handler, td today) {
	s, err := statPath.LoadWithContext[stat](handler, fio.Empty)
	if err != nil {
		handler.SendWithImageFail(err)
	}
	var ok [mealsPerDay]bool
	// 查询 QQ
	for k, v := range s {
		// 用餐类型
		m := slices.Index(td.Meal[:], v.QQ)
		if m >= 0 {
			// 用餐类型有效
			a := s[k].Stat[td.ID]
			a[m]++
			s[k].Stat[td.ID] = a
			ok[m] = true
		}
	}
	// 未查询到的进行写入
	for m, v := range ok {
		if v {
			continue
		}
		var a [mealsPerDay]int
		a[m] = 1
		s = append(s, food{
			QQ: td.Meal[m],
			Stat: map[string][mealsPerDay]int{
				td.ID: a,
			},
		})
	}
	// 排序
	s.sort()
	// 写入文件
	if err := statPath.SaveWithContext(handler, s); err != nil {
		handler.SendWithImageFail(err)
	}
}

// 统计数据排序
func (s *stat) sort() {
	// 统计数据按总被吃次数排序
	slices.SortStableFunc(*s, func(i, j food) int {
		return cmp.Or(
			cmp.Compare(i.cmpStat().sum, j.cmpStat().sum), // 比较总数
			cmp.Compare(i.cmpStat().min, j.cmpStat().min), // 如果总数相等，比较集齐五餐的数量
			cmp.Compare(i.cmpStat().max, j.cmpStat().max), // 如果集齐五餐的数量相等，比较单次最高
		)
	})
}

// 比较
func (fd *food) cmpStat() (c struct{ sum, max, min int }) {
	for _, v := range fd.Stat {
		for _, n := range v {
			c.sum += n
			c.max = max(c.max, n)
			c.min = min(c.min, n)
		}
	}
	return
}

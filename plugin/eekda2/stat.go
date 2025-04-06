package eekda2

import (
	"cmp"
	"maps"
	"slices"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 查询被吃次数
func getStat(ctx *zero.Ctx) {
	statPath.RLock()
	defer statPath.RUnlock()
	var (
		s, err = fio.Load[stat](statPath.Path, fio.Empty)
		msgr   = kitten.New(ctx)
	)
	if err != nil {
		msgr.SendWithImageFail(err)
	}
	i := slices.IndexFunc(s, func(f food) bool {
		return ctx.Event.UserID == f.ID.Int()
	})
	if i == -1 {
		msgr.DoNotKnow()
		return
	}
	c, err := fio.Load[config](todayPath.Path, fio.Empty)
	if err != nil {
		msgr.SendWithImageFail(err)
	}
	for _, t := range c {
		if slices.Contains(t.Group, *kitten.NewQQGroup(ctx.Event.GroupID)) {
			// 如果当前角色在本群已注册，跳过
			continue
		}
		// 如果当前角色在本群未注册，移除
		maps.DeleteFunc(s[i].Stat, func(k string, _ [mealsPerDay]int) bool {
			return k == t.ID
		})
	}
	if len(s[i].Stat) == 0 {
		msgr.DoNotKnow()
		return
	}
	msgr.Reply().AtLf().Text(&s[i]).Send()
}

// 统计被吃次数
func doStat(msgr *kitten.Messager, td today) {
	s, err := fio.Load[stat](statPath.Path, fio.Empty)
	if err != nil {
		msgr.SendWithImageFail(err)
	}
	var ok [mealsPerDay]bool
	// 查询 QQ
	for k, v := range s {
		// 用餐类型
		m := slices.Index(td.Meal[:], v.ID)
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
			ID: td.Meal[m],
			Stat: map[string][mealsPerDay]int{
				td.ID: a,
			},
		})
	}
	// 排序
	s.sort()
	// 写入文件
	if err := fio.Save(statPath.Path, s); err != nil {
		msgr.SendWithImageFail(err)
	}
}

// 统计数据排序
func (s *stat) sort() {
	// 统计数据按总被吃次数排序
	slices.SortStableFunc(*s, func(i, j food) int {
		var (
			ic = i.cmpStat().sum
			jc = j.cmpStat().sum
		)
		if ic < jc {
			return -1
		}
		if ic > jc {
			return 1
		}
		// 如果总数相等，比较集齐五餐的数量
		if c := cmp.Compare(i.cmpStat().min, j.cmpStat().min); c != 0 {
			return c
		}
		// 如果集齐五餐的数量相等，比较单次最高
		return cmp.Compare(i.cmpStat().max, j.cmpStat().max)
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

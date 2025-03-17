package eekda2

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
)

const (
	breakfast mealType = iota // 早餐
	lunch                     // 午餐
	lowtea                    // 下午茶
	dinner                    // 晚餐
	supper                    // 夜宵
)

type (
	// 用餐类型
	mealType byte

	// 配置文件
	config []today

	// 今天吃什么
	today struct {
		*kitten.Messager `yaml:"-"`       // 待发送的消息
		Time             time.Time        // 更新时间
		ID               string           // 角色名
		Group            []kitten.QQ      // 该角色对应的群号
		Meal             [count]kitten.QQ // 今天的每一餐
	}

	// 统计数据切片
	stat []food

	// 食物数据
	food struct {
		ID   kitten.QQ             // QQ
		Stat map[string][count]int // 每个角色的个人统计数据
	}
)

// String 实现 fmt.Stringer，播报今天吃什么
func (td *today) String() string {
	return `【` + td.ID + `今天吃什么】
早餐：　	` + line(td, td.Meal[breakfast]) + `
午餐：　	` + line(td, td.Meal[lunch]) + `
下午茶：	` + line(td, td.Meal[lowtea]) + `
晚餐：　	` + line(td, td.Meal[dinner]) + `
夜宵：　	` + line(td, td.Meal[supper])
}

// String 实现 fmt.Stringer，播报今天吃什么
func (fd *food) String() string {
	var (
		r  strings.Builder
		lf bool
	)
	for id, v := range fd.Stat {
		if lf {
			r.WriteByte('\n')
		} else {
			lf = true
		}
		fmt.Fprint(&r, `【`, id, "】\n")
		fmt.Fprintf(&r, "早餐：　	%d 次\n", v[breakfast])
		fmt.Fprintf(&r, "午餐：　	%d 次\n", v[lunch])
		fmt.Fprintf(&r, "下午茶：	%d 次\n", v[lowtea])
		fmt.Fprintf(&r, "晚餐：　	%d 次\n", v[dinner])
		fmt.Fprintf(&r, `夜宵：　	%d 次`, v[supper])
	}
	return r.String()
}

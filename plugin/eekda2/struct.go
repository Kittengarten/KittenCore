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

// 用餐类型
type mealType byte

type (
	// 配置文件
	config []today
	// 今天吃什么
	today struct {
		*kitten.Messager `yaml:"-"`             // 待发送的消息
		Time             time.Time              // 更新时间
		ID               string                 // 角色名
		Group            []kitten.QQ            // 该角色对应的群号
		Meal             [mealsPerDay]kitten.QQ // 今天的每一餐
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

type (
	// 统计数据切片
	stat []food
	// 食物数据
	food struct {
		Stat      map[string][mealsPerDay]int // 每个角色的个人统计数据
		kitten.QQ `yaml:"id"`                 // QQ
	}
)

// String 实现 fmt.Stringer，播报今天吃什么
func (fd *food) String() string {
	var (
		s  strings.Builder
		lf bool
	)
	for id, v := range fd.Stat {
		if lf {
			s.WriteByte('\n')
		} else {
			lf = true
		}
		fmt.Fprint(&s, `【`, id, "】\n")
		fmt.Fprintf(&s, "早餐：　	%d 次\n", v[breakfast])
		fmt.Fprintf(&s, "午餐：　	%d 次\n", v[lunch])
		fmt.Fprintf(&s, "下午茶：	%d 次\n", v[lowtea])
		fmt.Fprintf(&s, "晚餐：　	%d 次\n", v[dinner])
		fmt.Fprintf(&s, `夜宵：　	%d 次`, v[supper])
	}
	return s.String()
}

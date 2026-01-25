package eekda2

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/tidwall/gjson"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName  = `eekda2`     // 插件名
	todayFile         = `today.yaml` // 保存今天吃什么的文件
	statFile          = `stat.yaml`  // 保存统计数据的文件
	mealsPerDay       = 5            // 每天餐数
	cEEKDA            = `今天吃什么`
	cRegister         = `注册`
	cUnregister       = `注销`
	registerSuccess   = cRegister + `成功喵！`
	unregisterSuccess = cUnregister + `成功喵！`
	isNotAdmin        = `不是管理员，无法操作喵！`
	xx                = `XX`
	thisGroup         = ` // 在本群`
	help              = cRegister + xx + cEEKDA + thisGroup + cRegister + xx + `今天吃什么（管理员可用）
` + cUnregister + xx + cEEKDA + thisGroup + cUnregister + xx + `今天吃什么（管理员可用）
` + xx + cEEKDA + ` // 获取` + xx + `今日食谱
查询被吃次数 // 查询本人被吃次数`
)

// 注册插件
var engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
	Brief:             xx + cEEKDA,
	Help:              help,
	PrivateDataFolder: replyServiceName,
}).ApplySingle(ctxext.DefaultSingle)

var (
	// 今日文件路径
	todayPath = fio.NewPath(engine.DataFolder(), todayFile).WithMutex()
	// 统计文件路径
	statPath = fio.NewPath(engine.DataFolder(), statFile).WithRWMutex()
)

func init() {
	// XX 今天吃什么
	engine.OnSuffix(cEEKDA, zero.OnlyGroup).SetBlock(true).
		Limit(rate.GroupNormal.Get()).Handle(todayMeal)

	// 查询被吃次数
	engine.OnFullMatchGroup([]string{`查询被吃次数`, `查看被吃次数`}, zero.OnlyGroup).SetBlock(true).
		Limit(rate.User.Get()).Limit(rate.GroupNormal.Get()).Handle(getStat)
}

// XX 今天吃什么
func todayMeal(ctx *zero.Ctx) {
	todayPath.Lock()
	defer todayPath.Unlock()
	var (
		co, cancel = context.WithTimeout(context.Background(), core.Timeout)
		c, err     = fio.LoadWithContext[config](co, todayPath.Path, fio.Empty)
		handler       = msg.NewWithContext(co, ctx)
	)
	defer cancel()
	if err != nil {
		handler.SendWithImageFail(err)
	}
	name := str.Mid(``, cEEKDA, str.Clean(handler.Event().RawMessage, false))
	name, needRegister := strings.CutPrefix(name, cRegister)
	name, needUnegister := strings.CutPrefix(name, cUnregister)
	if name == `` {
		// 角色名为空
		handler.SendWithImageFail(`角色名为空喵！`)
		return
	}
	if needRegister && needUnegister {
		// 指令冲突
		handler.SendWithImageFail(`指令冲突喵！`)
		return
	}
	ci := slices.IndexFunc(c, func(t today) bool {
		return name == t.ID
	})
	g := usr.NewQQGroup(handler.Event().GroupID)
	// 角色是否存在
	if ci == -1 {
		// 该角色不存在
		if !needRegister {
			// 不执行注册指令
			handler.SendWithImageFail(name, `未在任何群注册喵！`)
			return
		}
		// 执行注册指令
		if !zero.AdminPermission(handler.Ctx) {
			// 没有权限
			handler.SendWithImageFail(isNotAdmin)
			return
		}
		// 注册
		c = append(c, today{
			ID:    name,
			Group: []usr.QQ{g},
		})
		// 写入文件
		if err := fio.SaveWithContext(co, todayPath.Path, c); err != nil {
			handler.SendWithImageFail(err)
			return
		}
		handler.Quote().AtLf().Text(name, registerSuccess).Send()
		return
	}
	// 该角色存在
	if !slices.Contains(c[ci].Group, g) {
		// 该角色未在本群注册
		if !needRegister {
			// 不执行注册指令
			handler.SendWithImageFail(name, `未在本群注册喵！`)
			return
		}
		// 执行注册指令
		if !zero.AdminPermission(handler.Ctx) {
			// 没有权限
			handler.SendWithImageFail(isNotAdmin)
			return
		}
		// 注册
		c[ci].Group = append(c[ci].Group, g)
		// 写入文件
		if err := fio.SaveWithContext(co, todayPath.Path, c); err != nil {
			handler.SendWithImageFail(err)
			return
		}
		handler.Quote().AtLf().Text(name, registerSuccess).Send()
		return
	}
	// 该角色已在本群注册
	switch {
	case needRegister:
		// 执行注册指令
		handler.SendWithImageFail(name, `已在本群注册，无需重复注册喵！`)
	case needUnegister:
		// 执行注销指令
		if !zero.AdminPermission(handler.Ctx) {
			// 没有权限
			handler.SendWithImageFail(isNotAdmin)
			return
		}
		// 注销
		c[ci].Group = slices.DeleteFunc(c[ci].Group, func(g_ usr.QQ) bool {
			return g == g_
		})
		if len(c[ci].Group) == 0 {
			// 如果该角色已经在所有群注销，删除该角色
			c = slices.DeleteFunc(c, func(t today) bool {
				return name == t.ID
			})
		}
		// 写入文件
		if err := fio.SaveWithContext(co, todayPath.Path, c); err != nil {
			handler.SendWithImageFail(err)
			return
		}
		handler.Quote().AtLf().Text(name, unregisterSuccess).Send()
	default:
		// 执行通常指令，写入上下文
		c[ci].Handler = handler
		if equal.IsSameDate4AM(c[ci].Time, time.Unix(handler.Event().Time, 0)) {
			// 今天已经生成了，直接播报
			handler.Quote().AtLf().Text(&c[ci]).Send()
			return
		}
		// 今天没有生成，执行生成
		var (
			// 群员列表
			list = make([]gjson.Result, 0, 128)
			// 时钟
			t = time.NewTicker(time.Second)
		)
		// 获取该角色注册的所有群的群员列表
		for _, g := range c[ci].Group {
			<-t.C
			list = append(list, g.MemberList(handler).List...)
		}
		t.Stop()
		// 只保留昨天一天的群员
		list = slices.DeleteFunc(list, func(v gjson.Result) bool {
			return !equal.IsSameDate4AM(time.Unix(v.Get(`last_sent_time`).Int(), 0),
				time.Unix(handler.Event().Time, 0).AddDate(0, 0, -1))
		})
		// 在其中取足够人的索引
		nums, err := utils.GenerateRandomNumber(0, len(list), mealsPerDay)
		if err != nil {
			handler.SendWithImageFail(`没有足够的食物喵！`, err)
			return
		}
		// 传入足够人的 QQ
		i := 0
		for v := range nums {
			c[ci].Meal[i] = usr.NewQQ(list[v].Get(`user_id`).Int())
			i++
		}
		// 写入时间
		c[ci].Time = time.Unix(handler.Event().Time, 0)
		// 写入文件
		if err := fio.SaveWithContext(co, todayPath.Path, c); err != nil {
			handler.SendWithImageFail(err)
			return
		}
		// 播报今天吃什么
		handler.Quote().AtLf().Text(&c[ci]).Send()
		// 统计
		doStat(handler, c[ci])
	}
}

// 生成每一餐的内容
func line(td *today, u usr.QQ) string {
	return u.TitleCardOrNickName(td.Handler) + `	❤	` + u.String()
}

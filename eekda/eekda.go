package eekda

import (
	"cmp"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/Kittengarten/KittenCore/kitten"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName             = `eekda`
	namePath         kitten.Path = `eekda/name.txt` // 保存名字的文件
	todayFile        kitten.Path = `today.yaml`     // 保存今天吃什么的文件
	statFile         kitten.Path = `stat.yaml`      // 保存统计数据的文件
)

func init() {
	var (
		name  = getName()
		brief = fmt.Sprintf("%s今天吃什么", name)
		help  = strings.Join([]string{`发送`,
			fmt.Sprintf(`%s今天吃什么`, name),
			fmt.Sprintf(`获取%s今日食谱`, name),
			`查询被吃次数`,
			`查询本人被吃次数`,
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  true,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: `eekda`,
		}).ApplySingle(ctxext.DefaultSingle)
	)

	engine.OnFullMatch(fmt.Sprintf(`%s今天吃什么`, name), zero.OnlyGroup).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			tf      = kitten.FilePath(kitten.Path(engine.DataFolder()), todayFile) // 保存今天吃什么的文件路径
			isExist bool                                                           // 保存今天吃什么的文件是否存在
			err     error
		)
		if isExist, err = tf.Exists(); !kitten.Check(err) {
			// 如果不确定文件存在
			kitten.DoNotKnow(ctx)
			zap.S().Warnf("不确定 %s 存在喵！\n%v", tf, err)
			return
		}
		if !isExist {
			// 如果文件不存在，创建文件
			if fp, err := os.Create(tf.String()); kitten.Check(err) {
				// 如果文件创建成功
				if n, err := fp.WriteString(`[]`); !kitten.Check(err) {
					// 如果文件写入字符串失败
					zap.S().Errorf("写入 %d 字符串时失败了喵！\n%v", n, err)
					kitten.DoNotKnow(ctx)
				}
				defer fp.Close()
				return
			}
			zap.S().Errorf("创建 %s 文件时失败了喵！\n%v", tf.String())
			kitten.DoNotKnow(ctx)
			return
		}
		var (
			today     = tf.Read()
			todayData Today
			nums      []int
		)
		if err := yaml.Unmarshal(today, &todayData); !kitten.Check(err) {
			zap.S().Errorf("转换 %v 失败了喵！\n%v", today, err)
			return
		}
		if kitten.IsSameDate(time.Now(), todayData.Time) {
			report(todayData, name, ctx)
			return
		}
		// 生成今天吃什么
		list := ctx.GetThisGroupMemberListNoCache().Array()
		slices.SortStableFunc(list, func(i, j gjson.Result) int {
			return cmp.Compare(i.Get(`last_sent_time`).Int(), j.Get(`last_sent_time`).Int())
		})
		list = list[math.Max(0, len(list)-50):]
		nums = kitten.GenerateRandomNumber(0, len(list), 5)
		if kitten.Check(nums) {
			ctx.Send(`没有足够的食物喵！`)
			return
		}
		todayData = Today{
			Time:      time.Now(),
			Breakfast: kitten.QQ{QQ: list[nums[0]].Get(`user_id`).Int()},
			Lunch:     kitten.QQ{QQ: list[nums[1]].Get(`user_id`).Int()},
			LowTea:    kitten.QQ{QQ: list[nums[2]].Get(`user_id`).Int()},
			Dinner:    kitten.QQ{QQ: list[nums[3]].Get(`user_id`).Int()},
			Supper:    kitten.QQ{QQ: list[nums[4]].Get(`user_id`).Int()},
		}
		today, err = yaml.Marshal(todayData)
		if !kitten.Check(err) {
			zap.S().Errorf("待写入饮食统计数据有错误：\n%v", err)
			return
		}
		tf.Write(today)
		report(todayData, name, ctx)
		// 存储饮食统计数据
		var (
			sf       = kitten.FilePath(kitten.Path(engine.DataFolder()), statFile)
			stat     = sf.Read()
			statData Stat
			statMap  = make(map[kitten.QQ]Kitten) // QQ:猫猫集合
		)
		if !kitten.Check(yaml.Unmarshal(stat, &statData)) {
			zap.S().Error(`饮食统计数据损坏了喵！`)
			kitten.DoNotKnow(ctx)
			return
		}
		// 加载数据
		for k := range statData {
			statMap[statData[k].ID] = statData[k]
		}
		bf, bfok := statMap[todayData.Breakfast]
		l, lok := statMap[todayData.Lunch]
		lt, ltok := statMap[todayData.LowTea]
		d, dok := statMap[todayData.Dinner]
		s, sok := statMap[todayData.Supper]
		isNew := map[string]bool{
			`Breakfast`: !bfok,
			`Lunch`:     !lok,
			`LowTea`:    !ltok,
			`Dinner`:    !dok,
			`Supper`:    !sok,
		}
		// 修改数据
		switch true {
		case bfok:
			bf.Breakfast++
			statMap[todayData.Breakfast] = bf
			fallthrough
		case lok:
			l.Lunch++
			statMap[todayData.Lunch] = l
			fallthrough
		case ltok:
			lt.LowTea++
			statMap[todayData.LowTea] = lt
			fallthrough
		case dok:
			d.Dinner++
			statMap[todayData.Dinner] = d
			fallthrough
		case sok:
			s.Supper++
			statMap[todayData.Supper] = s
		}
		// 回写修改的数据
		for k := range statData {
			statData[k] = statMap[statData[k].ID]
		}
		// 新增新猫猫的统计数据
		for k := range isNew {
			var new Kitten
			switch k {
			case `Breakfast`:
				new = Kitten{
					ID:   todayData.Breakfast,
					Name: getLine(todayData.Breakfast, ctx),
				}
				new.Breakfast++
			case `Lunch`:
				new = Kitten{
					ID:   todayData.Lunch,
					Name: getLine(todayData.Lunch, ctx),
				}
				new.Lunch++
			case `LowTea`:
				new = Kitten{
					ID:   todayData.LowTea,
					Name: getLine(todayData.LowTea, ctx),
				}
				new.LowTea++
			case `Dinner`:
				new = Kitten{
					ID:   todayData.Dinner,
					Name: getLine(todayData.Dinner, ctx),
				}
				new.Dinner++
			case `Supper`:
				new = Kitten{
					ID:   todayData.Supper,
					Name: getLine(todayData.Supper, ctx),
				}
				new.Supper++
			}
			if isNew[k] {
				// 是新猫猫才新增统计数据
				statData = append(statData, new)
			}
		}
		// 统计数据按总被吃次数排序
		slices.SortStableFunc(statData, func(i, j Kitten) int {
			ic := i.Breakfast + i.Lunch + i.LowTea + i.Dinner + i.Supper
			jc := j.Breakfast + j.Lunch + j.LowTea + j.Dinner + j.Supper
			if ic < jc {
				return -1
			} else if ic > jc {
				return 1
			}
			// 如果总数相等，比较类型是否齐全
			c := cmp.Compare(min(i.Breakfast, i.Lunch, i.LowTea, i.Dinner, i.Supper), min(j.Breakfast, j.Lunch, j.LowTea, j.Dinner, j.Supper))
			if 0 == c {
				// 如果类型齐全，比较单次最高
				return cmp.Compare(max(i.Breakfast, i.Lunch, i.LowTea, i.Dinner, i.Supper), max(j.Breakfast, j.Lunch, j.LowTea, j.Dinner, j.Supper))
			}
			return c
		})
		stat, err = yaml.Marshal(statData)
		sf.Write(stat)
		if !kitten.Check(err) {
			zap.S().Errorf("写入饮食统计数据发生错误：\n%v", err)
		}
	})

	engine.OnFullMatchGroup([]string{`查询被吃次数`, `查看被吃次数`}, zero.OnlyGroup).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour, 2).LimitByUser).Handle(func(ctx *zero.Ctx) {
		var (
			sf      = kitten.FilePath(kitten.Path(engine.DataFolder()), statFile)
			isExist bool  // 数据文件是否存在
			err     error // 错误
		)
		if isExist, err = sf.Exists(); !kitten.Check(err) {
			// 如果不确定文件存在
			zap.S().Warnf("不确定 %s 存在喵！\n%v", sf, err)
			kitten.DoNotKnow(ctx)
			return
		}
		if !isExist {
			// 如果文件不存在，创建文件
			if fp, err := os.Create(sf.String()); kitten.Check(err) {
				if n, err := fp.WriteString(`[]`); !kitten.Check(err) {
					// 如果文件写入字符串失败
					zap.S().Errorf("写入 %d 字符串时失败了喵！\n%v", n, err)
					kitten.DoNotKnow(ctx)
				}
				defer fp.Close()
				return
			}
			zap.S().Errorf("创建 %s 文件时失败了喵！\n%v", sf.String())
			kitten.DoNotKnow(ctx)
			return
		}
		var (
			stat     = sf.Read()
			statData Stat
			isGet    bool
		)
		if !kitten.Check(yaml.Unmarshal(stat, &statData)) {
			zap.S().Error(`饮食统计数据损坏了喵！`)
			kitten.DoNotKnow(ctx)
			return
		}
		for i := range statData {
			if ctx.Event.UserID == statData[i].ID.QQ {
				report := strings.Join([]string{fmt.Sprintf(`%s的被吃次数`, getLine(kitten.QQ{QQ: ctx.Event.UserID}, ctx)),
					fmt.Sprintf(`早餐：%d 次`, statData[i].Breakfast),
					fmt.Sprintf(`午餐：%d 次`, statData[i].Lunch),
					fmt.Sprintf(`下午茶：%d 次`, statData[i].LowTea),
					fmt.Sprintf(`晚餐：%d 次`, statData[i].Dinner),
					fmt.Sprintf(`夜宵：%d 次`, statData[i].Supper),
				}, "\n")
				kitten.SendText(ctx, true, report)
				isGet = true
			}
		}
		if !isGet {
			kitten.DoNotKnow(ctx)
		}
	})
}

// 获取名字
func getName() string {
	kitten.InitFile(namePath, `翼翼`) // 创建默认名字
	data, err := os.ReadFile(namePath.String())
	if kitten.Check(err) {
		return string(data)
	}
	zap.S().Warnf("打开文件 %s 失败了喵！\n%v", namePath, err)
	return `翼翼`
}

// 获取条目，u 为 QQ
func getLine(u kitten.QQ, ctx *zero.Ctx) string {
	info := Kitten{
		ID:   u,
		Name: u.GetTitle(ctx) + ctx.CardOrNickName(u.QQ),
	}
	return fmt.Sprintf(`%s（%d）`, info.Name, info.ID)
}

// 播报今天吃什么，t 为今日数据
func report(t Today, name string, ctx *zero.Ctx) {
	report := strings.Join([]string{fmt.Sprintf(`【%s今天吃什么】`, name),
		fmt.Sprintf(`早餐：%s`, getLine(t.Breakfast, ctx)),
		fmt.Sprintf(`午餐：%s`, getLine(t.Lunch, ctx)),
		fmt.Sprintf(`下午茶：%s`, getLine(t.LowTea, ctx)),
		fmt.Sprintf(`晚餐：%s`, getLine(t.Dinner, ctx)),
		fmt.Sprintf(`夜宵：%s`, getLine(t.Supper, ctx)),
	}, "\n")
	ctx.Send(report)
}

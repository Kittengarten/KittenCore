// Package sfacg SF 轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/zap"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName     = `SFACG`
	brief                = `SF 轻小说报更`
	configFile           = `config.yaml` // 配置文件名
	ag                   = `参数`
	commandNovel         = `小说`
	commandUpdateTest    = `更新测试`
	commandUpdatePreview = `更新预览`
	commandAddUpadte     = `添加报更`
	commandCancelUpadte  = `取消报更`
	commandQueryUpadte   = `查询报更`
)

func init() {
	var (
		cpf  = kitten.Configs.CommandPrefix
		help = strings.Join([]string{`发送`,
			fmt.Sprintf(`%s%s [%s]，可获取信息`, cpf, commandNovel, ag),
			fmt.Sprintf(`%s%s [%s]，可测试报更功能`, cpf, commandUpdateTest, ag),
			fmt.Sprintf(`%s%s [%s]，可预览更新内容`, cpf, commandUpdatePreview, ag),
			fmt.Sprintf(`%s%s [%s]，可添加小说自动报更`, cpf, commandAddUpadte, ag),
			fmt.Sprintf(`%s%s [%s]，可取消小说自动报更`, cpf, commandCancelUpadte, ag),
			fmt.Sprintf(`%s%s，可查询当前小说自动报更`, cpf, commandQueryUpadte),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: `sfacg`,
		}).ApplySingle(ctxext.DefaultSingle)
	)

	go track(engine)

	// 测试小说报更功能
	engine.OnCommand(`更新测试`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			novel     = getNovel(ctx)
			report, _ = novel.update()
		)
		ctx.SendChain(message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report))
	})

	// 预览小说更新功能
	engine.OnCommand(`更新预览`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		var (
			novel  = getNovel(ctx)
			report = novel.Preview
		)
		if `` == report {
			kitten.SendText(ctx, true, `不存在的喵！`)
			return
		}
		kitten.SendText(ctx, true, report)
	})

	// 小说信息功能
	engine.OnCommand(`小说`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		kitten.SendMessage(ctx, true, message.Image(novel.CoverURL), message.Text(novel.information()))
	})

	// 设置报更
	engine.OnCommand(`添加报更`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)                                                             // 小说实例
			c        = loadConfig(kitten.FilePath(kitten.Path(engine.DataFolder()), configFile)) // 报更配置
			ci       = make([]any, len(c))                                                       // 书号接口数组
			o        int64                                                                       // 发送对象
			groupSet = make(map[string]mapset.Set)                                               // 书号:群号集合
			track    = make(Config, 1)                                                           // 报更实例
		)
		if o = getO(ctx); 0 == o || -1 == o {
			return
		}
		for k := range c {
			ci[k] = c[k].BookID
			gi := make([]any, len(c[k].GroupID)) // 群号接口数组
			for i := range c[k].GroupID {
				gi[i] = c[k].GroupID[i]
			}
			groupSet[c[k].BookID] = mapset.NewSetFromSlice(gi) // 群号集合
		}
		cs := mapset.NewSetFromSlice(ci) // 书号集合
		if cs.Contains(novel.ID) {
			// 已经有该小说
			if groupSet[novel.ID].Add(o) {
				// 尚无该群号，添加成功
				for k := range c {
					gi := groupSet[c[k].BookID].ToSlice()
					gs := make([]int64, len(gi))
					for i := range gi {
						gs[i] = gi[i].(int64)
					}
					c[k].GroupID = gs
				}
			} else {
				// 已有该群号，添加失败
				kitten.SendTextOf(ctx, false, `《%s》已经添加报更了喵！`, novel.Name)
				return
			}
		} else {
			// 没有该小说，新建并添加
			track[0].BookID = novel.ID
			track[0].BookName = novel.Name
			track[0].GroupID = []int64{o}
			c = append(c, track[0])
		}
		if saveConfig(c, engine) {
			kitten.SendTextOf(ctx, false, `添加《%s》报更成功喵！`, novel.Name)
			return
		}
		kitten.SendTextOf(ctx, false, `添加《%s》报更失败喵！`, novel.Name)
	})

	// 移除报更
	engine.OnCommand(`取消报更`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)                                                             // 小说实例
			c        = loadConfig(kitten.FilePath(kitten.Path(engine.DataFolder()), configFile)) // 报更配置
			ci       = make([]any, len(c))                                                       // 书号接口数组
			o        int64                                                                       // 发送对象
			groupSet = make(map[string]mapset.Set)                                               // 书号:群号集合
		)
		if o = getO(ctx); 0 == o || -1 == o {
			return
		}
		if 0 < len(c) {
			for k := range c {
				ci[k] = c[k].BookID
				gi := make([]any, len(c[k].GroupID)) // 群号接口数组
				for i := range c[k].GroupID {
					gi[i] = c[k].GroupID[i]
				}
				groupSet[c[k].BookID] = mapset.NewSetFromSlice(gi) // 群号集合
				n := groupSet[c[k].BookID].Cardinality()           // 群的数量
				if novel.ID == c[k].BookID {
					groupSet[c[k].BookID].Remove(o)
					if 0 < groupSet[c[k].BookID].Cardinality() {
						// 如果有群
						if n == groupSet[c[k].BookID].Cardinality() {
							// 没有该群号可移除
							kitten.SendText(ctx, false, `本书不存在或不在追更列表，也许有其它错误喵～`)
						}
						// 移除成功
						gi := groupSet[c[k].BookID].ToSlice()
						gs := make([]int64, len(gi))
						for i := range gi {
							gs[i] = gi[i].(int64)
						}
						c[k].GroupID = gs
					} else {
						// 如果没有群，则移除该小说
						c[k].GroupID = nil
						c = append(c[:k], c[k+1:]...)
					}
				}
			}
		}
		if saveConfig(c, engine) {
			kitten.SendTextOf(ctx, false, `取消《%s》报更成功喵！`, novel.Name)
			return
		}
		kitten.SendTextOf(ctx, false, `取消《%s》报更失败喵！`, novel.Name)
	})

	// 查询报更
	engine.OnCommand(`查询报更`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				r = `【报更列表】`
				c = loadConfig(kitten.FilePath(kitten.Path(engine.DataFolder()), configFile)) // 报更配置
				o int64                                                                       // 发送对象
			)
			if o = getO(ctx); 0 == o || -1 == o {
				return
			}
			if 0 < len(c) {
				for k := range c {
					var t string
					if `` == c[k].UpdateTime {
						t = `未知`
					} else {
						t = c[k].UpdateTime
					}
					for i := range c[k].GroupID {
						if o == c[k].GroupID[i] {
							r = strings.Join([]string{r,
								fmt.Sprintf(`《%s》，书号 %s`, c[k].BookName, c[k].BookID),
								fmt.Sprintf(`上次更新：%s`, t),
							}, "\n")
						}
					}
				}
			}
			if `【报更列表】` == r {
				r += "\n这里没有添加小说报更喵～"
			}
			ctx.Send(r)
		})
}

/*
获取小说

如果传入值不为书号，则先获取书号
*/
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag := ctx.State[`args`].(string)
	if ag, err := keyWord(ag).findBookID(); isInt(ag) || kitten.Check(err) {
		// 如果传入值是书号，或者成功找到了书号
		nv = *novelPool.Get().(*Novel)
		nv.init(ag)
		defer novelPool.Put(&nv)
	}
	return
}

// 报更
func track(e *control.Engine) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			zap.Errorf("%s 报更协程出现错误喵！\n%v", ReplyServiceName, err)
			debug.PrintStack()
		}
	}()
	kitten.InitFile(kitten.FilePath(kitten.Path(e.DataFolder()), configFile), `[]`)
	var (
		novel   Novel
		bot     *zero.Ctx
		name    = kitten.Configs.NickName[0]
		line    = `======================[` + name + `]======================`
		data    = loadConfig(kitten.FilePath(kitten.Path(e.DataFolder()), configFile))
		content = strings.Join([]string{
			line,
			`* OneBot + ZeroBot + Go`,
			fmt.Sprintf(`一共有 %d 本小说`, len(data)),
			`=======================================================`,
		}, "\n")
		t = time.NewTicker(5 * time.Second) // 每 5 秒检测一次
	)
	fmt.Println(content)
	for nil == bot {
		zap.Debugf(`尝试获取 Bot 实例：%d`, kitten.Configs.SelfID)
		bot = zero.GetBot(kitten.Configs.SelfID)
		zap.Debugf("获取的 Bot 实例：\n%v", bot)
		<-t.C // 阻塞协程，收到定时器信号则释放
	}
	// 报更
	for {
		// 配置池初始化
		configPool = sync.Pool{
			New: func() interface{} {
				c := loadConfig(kitten.FilePath(kitten.Path(e.DataFolder()), configFile))
				return &c
			},
		}
		var (
			// 从配置池初始化配置
			data = *configPool.Get().(*Config)
			done bool
		)
		for i := range data {
			novel.init(data[i].BookID)
			// 更新判定，并防止误报
			if novel.NewChapter.URL == data[i].RecordURL ||
				!novel.IsGet || !novel.NewChapter.IsGet {
				continue
			}
			// 距上次更新时间小于 1 秒则不报更，防止异常信息发送
			if report, d := novel.update(); time.Second < d {
				for k := range data[i].GroupID {
					if 0 < data[i].GroupID[k] {
						bot.SendGroupMessage(data[i].GroupID[k], message.Message{
							message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					} else {
						bot.SendPrivateMessage(-data[i].GroupID[k], message.Message{
							message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					}
				}
				data[i].BookName = novel.Name
				data[i].RecordURL = novel.NewChapter.URL
				data[i].UpdateTime = novel.NewChapter.Time.Format(`2006年01月02日 15时04分05秒`)
				done = true
			}
		}
		// 如果并没有报更，直接进入下一次循环
		if done {
			updateConfig, err := yaml.Marshal(data)
			configPool.Put(&data)
			kitten.FilePath(kitten.Path(e.DataFolder()), configFile).Write(updateConfig)
			if kitten.Check(err) {
				zap.Infof(`记录 %s 成功喵！`, e.DataFolder()+configFile)
			} else {
				zap.Warnf(`记录 %s 失败喵！`, e.DataFolder()+configFile)
			}
		}
		<-t.C // 阻塞协程，收到定时器信号则释放
	}
}

/*
发送对象判断

返回默认值 0 代表不支持的对象（目前是频道）

返回 -1 代表在群中无权限
*/
func getO(ctx *zero.Ctx) (o int64) {
	switch ctx.Event.DetailType {
	case "private":
		o = -ctx.Event.UserID
	case "group":
		if !zero.AdminPermission(ctx) {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n你没有管理员权限喵！"))
			return -1
		}
		o = ctx.Event.GroupID
	case "guild":
		ctx.Send(`暂不支持频道喵！`)
	}
	return
}

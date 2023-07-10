// Package sfacg SF 轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
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
		} else {
			kitten.SendText(ctx, true, report)
		}
	})

	// 小说信息功能
	engine.OnCommand(`小说`).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByGroup).Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		ctx.SendChain(message.Image(novel.CoverURL), message.Text(novel.information()))
	})

	// 设置报更
	engine.OnCommand(`添加报更`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)                                                             // 小说实例
			c        = loadConfig(kitten.FilePath(kitten.Path(engine.DataFolder()), configFile)) // 报更配置
			ci       = make([]any, len(c))                                                       // 书号接口数组
			o        int64                                                                       // 发送对象
			hasg     bool                                                                        // 是否有发送对象
			groupSet = make(map[string]mapset.Set)                                               // 书号:群号集合
			track    = make(Config, 1)                                                           // 报更实例
		)
		if o = getO(ctx); 0 == o || -1 == o {
			return
		}
		for k, v := range c {
			ci[k] = v.BookID
			gi := make([]any, len(v.GroupID)) // 群号接口数组
			for i, g := range v.GroupID {
				gi[i] = g
			}
			groupSet[v.BookID] = mapset.NewSetFromSlice(gi) // 群号集合
		}
		cs := mapset.NewSetFromSlice(ci) // 书号集合
		if cs.Contains(novel.ID) {
			if groupSet[novel.ID].Add(o) {
				for k, v := range c {
					gi := groupSet[v.BookID].ToSlice()
					gs := make([]int64, len(gi))
					for i, g := range gi {
						gs[i] = g.(int64)
					}
					c[k].GroupID = gs
				}
			} else {
				hasg = true // 已经存在此发送对象
			}
		} else {
			track[0].BookID = novel.ID
			track[0].BookName = novel.Name
			track[0].GroupID = []int64{o}
			c = append(c, track[0])
		}
		if hasg {
			kitten.SendTextOf(ctx, false, `《%s》已经添加报更了喵！`, novel.Name)
			return
		}
		if saveConfig(c, engine) {
			kitten.SendTextOf(ctx, false, `添加《%s》报更成功喵！`, novel.Name)
			return
		}
		kitten.SendTextOf(ctx, false, `添加《%s》报更失败喵！`, novel.Name)
		return
	})

	// 移除报更
	engine.OnCommand(`取消报更`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)                                                             // 小说实例
			c        = loadConfig(kitten.FilePath(kitten.Path(engine.DataFolder()), configFile)) // 报更配置
			ci       = make([]any, len(c))                                                       // 书号接口数组
			o        int64                                                                       // 发送对象
			groupSet = make(map[string]mapset.Set)                                               // 书号:群号集合
			ok       bool                                                                        // 取消成功
		)
		if o = getO(ctx); 0 == o || -1 == o {
			return
		}
		if 0 < len(c) {
			for k, v := range c {
				ci[k] = v.BookID
				gi := make([]any, len(v.GroupID)) // 群号接口数组
				for i, g := range v.GroupID {
					gi[i] = g
				}
				groupSet[v.BookID] = mapset.NewSetFromSlice(gi) // 群号集合
				n := groupSet[v.BookID].Cardinality()
				if novel.ID == v.BookID {
					groupSet[v.BookID].Remove(o)
					if 0 < groupSet[v.BookID].Cardinality() {
						// 如果群号集合不为空集
						ok = n != groupSet[v.BookID].Cardinality()
						gi := groupSet[v.BookID].ToSlice()
						gs := make([]int64, len(gi))
						for i, g := range gi {
							gs[i] = g.(int64)
						}
						c[k].GroupID = gs
					} else {
						// 如果群号集合为空集
						c[k].GroupID = nil
						ok = true
					}
					if 0 >= len(c[k].GroupID) {
						c = append(c[:k], c[k+1:]...)
					}
				}
			}
		}
		if ok {
			if saveConfig(c, engine) {
				kitten.SendTextOf(ctx, false, `取消《%s》报更成功喵！`, novel.Name)
				return
			}
			kitten.SendTextOf(ctx, false, `取消《%s》报更失败喵！`, novel.Name)
			return
		}
		kitten.SendText(ctx, false, `本书不存在或不在追更列表，也许有其它错误喵～`)
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
				for _, v := range c {
					var t string
					if `` == v.UpdateTime {
						t = `未知`
					} else {
						t = v.UpdateTime
					}
					for _, g := range v.GroupID {
						if o == g {
							r = strings.Join([]string{r,
								fmt.Sprintf(`《%s》，书号 %s`, v.BookName, v.BookID),
								fmt.Sprintf(`上次更新：%s`, t),
							}, "\n")
						}
					}
				}
			} else {
				r += "\n这里没有添加小说报更喵～"
			}
			ctx.Send(r)
		})
}

// 获取小说（如果传入值不为书号，则先获取书号）
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag, chk := ctx.State[`args`].(string), true
	if !isInt(ag) {
		ag, chk = keyWord(ag).findBookID()
		if !chk {
			kitten.SendText(ctx, false, ag)
			return
		}
	}
	nv.init(ag)
	return nv
}

// 报更
func track(e *control.Engine) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("%s 报更协程出现错误喵！\n%v", ReplyServiceName, err)
		}
	}()
	kitten.InitFile(kitten.FilePath(kitten.Path(e.DataFolder()), configFile), `[]`)
	var (
		novel   Novel
		bot     = <-kitten.BotSFACGchan
		name    = zero.BotConfig.NickName[0]
		line    = `======================[` + name + `]======================`
		data    = loadConfig(kitten.FilePath(kitten.Path(e.DataFolder()), configFile))
		content = strings.Join([]string{
			line,
			`* OneBot + ZeroBot + Go`,
			fmt.Sprintf(`一共有 %d 本小说`, len(data)),
			`=======================================================`,
		}, "\n")
		t = time.Tick(5 * time.Second) // 每 5 秒检测一次
	)
	fmt.Println(content)
	if nil == bot {
		log.Error(`报更没有获取到实例喵！`)
	} else {
		log.Info(`报更已经获取到实例了喵！`)
	}
	// 报更
	for {
		var (
			data    = loadConfig(kitten.FilePath(kitten.Path(e.DataFolder()), configFile))
			dataNew = data
			done    bool
		)
		for i := range data {
			novel.init(data[i].BookID)
			// 更新判定，并防止误报
			if novel.NewChapter.URL == data[i].RecordURL ||
				!novel.IsGet || !novel.NewChapter.IsGet {
				continue
			}
			// 防止更新异常信息发送
			if report, ok := novel.update(); ok {
				for _, groupID := range data[i].GroupID {
					if groupID > 0 {
						bot.SendGroupMessage(groupID, message.Message{
							message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					} else {
						bot.SendPrivateMessage(-groupID, message.Message{
							message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					}
				}
				dataNew[i].BookName = novel.Name
				dataNew[i].RecordURL = novel.NewChapter.URL
				dataNew[i].UpdateTime = novel.NewChapter.Time.Format(`2006年01月02日 15时04分05秒`)
				done = true
			}
		}
		// 如果并没有报更，直接进入下一次循环
		if done {
			updateConfig, err := yaml.Marshal(dataNew)
			kitten.FilePath(kitten.Path(e.DataFolder()), configFile).Write(updateConfig)
			if kitten.Check(err) {
				log.Infof(`记录 %s 成功喵！`, e.DataFolder()+configFile)
			} else {
				log.Warnf(`记录 %s 失败喵！`, e.DataFolder()+configFile)
			}
		}
		select {
		case <-t: // 阻塞协程，收到定时器信号则释放
		}
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

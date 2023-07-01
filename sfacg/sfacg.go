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
		if report == `` {
			ctx.Send(`不存在的喵！`)
		} else {
			ctx.Send(report)
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
			novel    = getNovel(ctx)
			c, err   = loadConfig(engine)
			track    = make(Config, 1)
			g        int64                         // 发送对象
			hasg     bool                          // 是否有发送对象
			groupSet = make(map[string]mapset.Set) // 书号:群号集合
		)
		if g = getg(ctx); g == 0 {
			return
		}
		if kitten.Check(err) {
			ci := make([]any, len(c)) // 书号接口数组
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
				if groupSet[novel.ID].Add(g) {
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
				track[0].GroupID = []int64{g}
				c = append(c, track[0])
			}
			if hasg {
				ctx.Send(fmt.Sprintf(`《%s》已经添加报更了喵！`, novel.Name))
			} else if saveConfig(c, engine) {
				ctx.Send(fmt.Sprintf(`添加《%s》报更成功喵！`, novel.Name))
			} else {
				ctx.Send(fmt.Sprintf(`添加《%s》报更失败喵！`, novel.Name))
			}
		} else {
			ctx.Send(fmt.Sprintf("添加《%s》报更出现错误喵！\n错误：%v", novel.Name, err))
		}
	})

	// 移除报更
	engine.OnCommand(`取消报更`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				novel    = getNovel(ctx)
				c, err   = loadConfig(engine)
				g        int64                         // 发送对象
				groupSet = make(map[string]mapset.Set) // 书号:群号集合
				ok       bool
			)
			if g = getg(ctx); g == 0 {
				return
			}
			if kitten.Check(err) && 0 < len(c) {
				ci := make([]any, len(c)) // 书号接口数组
				for k, v := range c {
					ci[k] = v.BookID
					gi := make([]any, len(v.GroupID)) // 群号接口数组
					for i, g := range v.GroupID {
						gi[i] = g
					}
					groupSet[v.BookID] = mapset.NewSetFromSlice(gi) // 群号集合
					l := groupSet[v.BookID].Cardinality()
					if novel.ID == v.BookID {
						groupSet[v.BookID].Remove(g)
						// 如果群号集合为空集
						if 0 >= groupSet[v.BookID].Cardinality() {
							c[k].GroupID = nil
							ok = true
						} else {
							ok = l != groupSet[v.BookID].Cardinality()
							gi := groupSet[v.BookID].ToSlice()
							gs := make([]int64, len(gi))
							for i, g := range gi {
								gs[i] = g.(int64)
							}
							c[k].GroupID = gs
						}
						if 0 >= len(c[k].GroupID) {
							c = append(c[:k], c[k+1:]...)
						}
					}
				}
			}
			if ok {
				if saveConfig(c, engine) {
					ctx.Send(fmt.Sprintf(`取消《%s》报更成功喵！`, novel.Name))
				} else {
					ctx.Send(fmt.Sprintf(`取消《%s》报更失败喵！`, novel.Name))
				}
			} else {
				ctx.Send(`本书不存在或不在追更列表，也许有其它错误喵～`)
			}
		})
}

// 获取小说（如果传入值不为书号，则先获取书号）
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag, chk := ctx.State[`args`].(string), true
	if !isInt(ag) {
		ag, chk = keyWord(ag).findBookID()
		if !chk {
			ctx.Send(ag)
			return
		}
	}
	nv.init(string(ag))
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
	var (
		novel     Novel
		bot       = <-kitten.BotSFACGchan
		name      = zero.BotConfig.NickName[0]
		line      = `======================[` + name + `]======================`
		data, err = loadConfig(e)
		content   = strings.Join([]string{
			line,
			`* OneBot + ZeroBot + Go`,
			fmt.Sprintf(`一共有 %d 本小说`, len(data)),
			`=======================================================`,
		}, "\n")
		t = time.Tick(5 * time.Second) // 每 5 秒检测一次
	)

	if !kitten.Check(err) {
		log.Errorf("%s 载入配置文件出现错误喵！\n%v", ReplyServiceName, err)
	}
	fmt.Println(content)
	if bot == nil {
		log.Error(`报更没有获取到实例喵！`)
	} else {
		log.Info(`报更已经获取到实例了喵！`)
	}
	// 报更
	for {
		var (
			data, err = loadConfig(e)
			dataNew   = data
			done      bool
		)
		// 如果加载文件出错，则不进行后续步骤，直接重试
		if !kitten.Check(err) {
			select {
			case <-t: // 阻塞协程，收到定时器信号则释放
			}
			continue
		}
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
		if !done {
			continue
		}
		var (
			updateConfig, err1 = yaml.Marshal(dataNew)
			err2               = kitten.Path(e.DataFolder() + configFile).Write(updateConfig)
		)
		if kitten.Check(err1) && kitten.Check(err2) {
			log.Infof(`记录 %s 成功喵！`, e.DataFolder()+configFile)
		} else {
			log.Warnf(`记录 %s 失败喵！`, e.DataFolder()+configFile)
		}
		select {
		case <-t: // 阻塞协程，收到定时器信号则释放
		}
	}
}

// 发送对象判断，返回默认值 0 代表不能使用
func getg(ctx *zero.Ctx) (g int64) {
	switch ctx.Event.DetailType {
	case "private":
		g = -ctx.Event.UserID
	case "group":
		if zero.AdminPermission(ctx) {
			g = ctx.Event.GroupID
		} else {
			ctx.Send("你没有管理员权限喵！")
		}
	case "guild":
		ctx.Send("暂不支持频道喵！")
	}
	return
}

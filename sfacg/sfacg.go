// Package sfacg SF轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName     = "SFACG"
	brief                = "获取SF小说信息"
	configFile           = "config.yaml" // 配置文件名
	ag                   = "参数"
	commandNovel         = "小说"
	commandUpdateTest    = "更新测试"
	commandUpdatePreview = "更新预览"
	commandAddUpadte     = "添加报更"
	commandCancelUpadte  = "取消报更"
)

func init() {
	var (
		cpf  = kitten.Configs.CommandPrefix
		help = strings.Join([]string{"发送",
			fmt.Sprintf("%s%s [%s]，可获取信息", cpf, commandNovel, ag),
			fmt.Sprintf("%s%s [%s]，可测试报更功能", cpf, commandUpdateTest, ag),
			fmt.Sprintf("%s%s [%s]，可预览更新内容", cpf, commandUpdatePreview, ag),
			fmt.Sprintf("%s%s [%s]，可添加小说自动报更", cpf, commandAddUpadte, ag),
			fmt.Sprintf("%s%s [%s]，可取消小说自动报更", cpf, commandAddUpadte, ag),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: "sfacg",
		})
	)

	go track(engine)

	// 测试小说报更功能
	engine.OnCommand("更新测试").Handle(func(ctx *zero.Ctx) {
		var (
			novel  = getNovel(ctx)
			report = novel.update()
		)
		ctx.SendChain(message.Image(novel.HeadURL), message.Text(report))
	})

	// 预览小说更新功能
	engine.OnCommand("更新预览").Handle(func(ctx *zero.Ctx) {
		var (
			novel  = getNovel(ctx)
			report = novel.Preview
		)
		if report == "" {
			ctx.Send("不存在的喵！")
		} else {
			ctx.Send(report)
		}
	})

	// 小说信息功能
	engine.OnCommand("小说").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		ctx.SendChain(message.Image(novel.CoverURL), message.Text(novel.information()))
	})

	// 设置报更
	engine.OnCommand("添加报更", zero.AdminPermission).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)
			cf, err  = loadConfig(engine)
			track    = make(Config, 1)
			hasGroup bool
			groupSet = make(map[string]mapset.Set) // 书号:群号集合
		)
		if kitten.Check(err) {
			cfi := make([]interface{}, len(cf)) // 书号接口数组
			for k, v := range cf {
				cfi[k] = v.BookID
				gi := make([]interface{}, len(v.GroupID))
				for i, g := range v.GroupID {
					gi[i] = g
				}
				groupSet[v.BookID] = mapset.NewSetFromSlice(gi) // 群号集合
			}
			cfs := mapset.NewSetFromSlice(cfi) // 书号集合
			if cfs.Contains(novel.ID) {
				if groupSet[novel.ID].Add(ctx.Event.GroupID) {
					for k, v := range cf {
						gi := groupSet[v.BookID].ToSlice()
						gs := make([]int64, len(v.GroupID)+1)
						for i, g := range gi {
							gs[i] = g.(int64)
						}
						cf[k].GroupID = gs
					}
				} else {
					hasGroup = true // 已经存在此群
				}
			} else {
				track[0].BookID = novel.ID
				track[0].BookName = novel.Name
				track[0].GroupID = []int64{ctx.Event.GroupID}
				cf = append(cf, track[0])
			}
			if saveConfig(cf, engine) {
				ctx.Send(fmt.Sprintf("添加《%s》报更成功喵！", novel.Name))
			} else if hasGroup {
				ctx.Send(fmt.Sprintf("《%s》已经添加报更了喵！", novel.Name))
			} else {
				ctx.Send(fmt.Sprintf("添加《%s》报更失败喵！", novel.Name))
			}
		} else {
			ctx.Send(fmt.Sprintf("添加《%s》报更出现错误喵！\n错误：%s", novel.Name, err))
		}

	})

	// 移除报更
	engine.OnCommand("取消报更", zero.AdminPermission).Handle(func(ctx *zero.Ctx) {
		var (
			novel    = getNovel(ctx)
			cf, err  = loadConfig(engine)
			groupSet = make(map[string]mapset.Set) // 书号:群号集合
			ok       bool
		)
		if kitten.Check(err) && 0 < len(cf) {
			cfi := make([]interface{}, len(cf)) // 书号接口数组
			for k, v := range cf {
				cfi[k] = v.BookID
				gi := make([]interface{}, len(v.GroupID))
				for i, g := range v.GroupID {
					gi[i] = g
				}
				groupSet[v.BookID] = mapset.NewSetFromSlice(gi) // 群号集合
				l := groupSet[v.BookID].Cardinality()
				groupSet[v.BookID].Remove(ctx.Event.GroupID)
				ok = l != groupSet[v.BookID].Cardinality()
				for k, v := range cf {
					gi := groupSet[v.BookID].ToSlice()
					gs := make([]int64, len(v.GroupID)-1)
					for i, g := range gi {
						gs[i] = g.(int64)
					}
					cf[k].GroupID = gs
				}
				if 0 >= len(cf[k].GroupID) {
					cf = append(cf[:k], cf[k+1:]...)
				}
			}
		}
		if ok {
			if saveConfig(cf, engine) {
				ctx.Send(fmt.Sprintf("取消《%s》报更成功喵！", novel.Name))
			} else {
				ctx.Send(fmt.Sprintf("取消《%s》报更成功喵！", novel.Name))
			}
		} else {
			ctx.Send("本书不存在或不在追更列表，也许有其它错误喵～")
		}
	})
}

// 获取小说（如果传入值不为书号，则先获取书号）
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag := ctx.State["args"].(string)
	if !isInt(ag) {
		ag, chk := keyWord(ag).findBookID()
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
			log.Errorf("报更协程出现错误喵！\n%v", ReplyServiceName, err)
		}
	}()

	var (
		novel     Novel
		bot       = <-kitten.Botchan
		name      = zero.BotConfig.NickName[0]
		line      = "======================[" + name + "]======================"
		data, err = loadConfig(e)
		content   = strings.Join([]string{
			line,
			"* OneBot + ZeroBot + Go",
			fmt.Sprintf("一共有 %d 本小说", len(data)),
			"=======================================================",
		}, "\n")
		t = time.Tick(5 * time.Second) // 每 5 秒检测一次
	)

	if !kitten.Check(err) {
		log.Errorf("%s 载入配置文件出现错误喵！\n%v", err)
	}

	fmt.Println(content)

	if bot == nil {
		log.Error("报更没有获取到实例喵！")
	} else {
		log.Info("报更已经获取到实例了喵！")
	}

	// 报更
	for {
		var (
			data, err           = loadConfig(e)
			dataNew             = data
			updateError, update bool
		)
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
				!novel.IsGet ||
				!novel.NewChapter.IsGet {
				continue
			}

			// 防止更新异常信息发到群里
			if report := novel.update(); report == "更新异常喵！" {
				updateError = true
				log.Warn(novel.NewChapter.BookURL + report)
			} else {
				for _, groupID := range data[i].GroupID {
					bot.SendGroupMessage(groupID, message.Message{
						message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					update = true
				}
				dataNew[i].BookName = novel.Name
				dataNew[i].RecordURL = novel.NewChapter.URL
				dataNew[i].UpdateTime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
			}
		}

		// 将本轮获取到的更新链接和更新时间记录至文件，如果没有获取到，或者报更出错，则不写入
		if !updateError && update {
			var (
				updateConfig, err1 = yaml.Marshal(dataNew)
				err2               = kitten.Path(e.DataFolder() + configFile).Write(updateConfig)
			)
			if kitten.Check(err1) && kitten.Check(err2) {
				log.Infof("记录 %s 成功喵！", e.DataFolder()+configFile)
			} else {
				log.Warnf("记录 %s 失败喵！", e.DataFolder()+configFile)
			}
		}
		select {
		case <-t: // 阻塞协程，收到定时器信号则释放
		}
	}
}

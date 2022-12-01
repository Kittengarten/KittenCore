// Package sfacg SF轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"fmt"
	"strings"
	"time"

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
	path                 = "sfacg/config.yaml" // 配置文件路径
	ag                   = "参数"
	commandNovel         = "小说"
	commandUpdateTest    = "更新测试"
	commandUpdatePreview = "更新预览"
)

func init() {
	var (
		cpf  = kitten.Configs.CommandPrefix
		help = strings.Join([]string{"发送",
			fmt.Sprintf("%s%s [%s]，可获取信息", cpf, commandNovel, ag),
			fmt.Sprintf("%s%s [%s]，可测试报更功能", cpf, commandUpdateTest, ag),
			fmt.Sprintf("%s%s [%s]，可预览更新内容", cpf, commandUpdatePreview, ag),
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Brief:            brief,
			Help:             help,
		})
	)

	go track()

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
			ctx.SendGroupMessage(ctx.Event.GroupID, "不存在的喵！")
		} else {
			ctx.SendGroupMessage(ctx.Event.GroupID, report)
		}

	})

	// 小说信息功能
	engine.OnCommand("小说").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		ctx.SendChain(message.Image(novel.CoverURL), message.Text(novel.information()))
	})
}

// 获取小说（如果传入值不为书号，则先获取书号）
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag := ctx.State["args"].(string)
	if !isInt(ag) {
		var chk bool
		ag, chk = findBookID(ag)
		if !chk {
			ctx.SendGroupMessage(ctx.Event.GroupID, ag)
			return
		}
	}
	nv.init(ag)
	return nv
}

// 报更
func track() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("报更协程出现了错误：%s", err)
		}
	}()

	var (
		novel   Novel
		bot     = <-kitten.Botchan
		name    = zero.BotConfig.NickName[0]
		line    = "======================[" + name + "]======================"
		content = strings.Join([]string{
			line,
			"* OneBot + ZeroBot + Golang",
			fmt.Sprintf("一共有 %d 本小说", len(loadConfig())),
			"=======================================================",
		}, "\n")
		t = time.Tick(5 * time.Second) // 每 5 秒检测一次
	)

	fmt.Println(content)

	if bot != nil {
		log.Info("报更已经获取到实例了喵！")
	} else {
		log.Error("报更没有获取到实例了喵！")
	}

	// 报更
	for {
		var (
			data                = loadConfig()
			dataNew             = data
			updateError, update bool
		)
		for idx := range data {
			novel.init(data[idx].BookID)

			// 更新判定，并防止误报
			if novel.NewChapter.URL == data[idx].RecordURL ||
				!novel.IsGet ||
				!novel.NewChapter.IsGet {
				continue
			}

			// 防止更新异常信息发到群里
			if report := novel.update(); report == "更新异常喵！" {
				updateError = true
				log.Warn(novel.NewChapter.BookURL + report)
			} else {
				for _, groupID := range data[idx].GroupID {
					bot.SendGroupMessage(groupID, message.Message{
						message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)})
					update = true
				}
				dataNew[idx].BookName = novel.Name
				dataNew[idx].RecordURL = novel.NewChapter.URL
				dataNew[idx].UpdateTime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
			}
		}

		// 将本轮获取到的更新链接和更新时间记录至文件，如果没有获取到，或者报更出错，则不写入
		if !updateError && update {
			var (
				updateConfig, err1 = yaml.Marshal(dataNew)
				err2               = kitten.FileWrite(path, updateConfig)
			)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				log.Warn(fmt.Sprintf("记录 %s 失败喵！", path))
			} else {
				log.Info(fmt.Sprintf("记录 %s 成功喵！", path))
			}
		}
		select {
		case <-t: // 阻塞协程，收到定时器信号则释放
		}
	}
}

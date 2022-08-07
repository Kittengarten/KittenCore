// Package sfacg SF轻小说更新播报、小说信息查询、小说更新查询
package sfacg

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/FloatTech/zbputils/control"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName = "SFACG"
	path             = "sfacg/config.yaml" // 配置文件路径
)

var (
	kittenConfig = kitten.LoadConfig()
	// Alive 报更协程是否存活
	Alive bool
)

func init() {
	// 注册插件
	const (
		ag                   = "参数"
		commandNovel         = "小说"
		commandUpdateTest    = "更新测试"
		commandUpdatePreview = "更新预览"
	)

	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s%s [%s]，可获取信息", kittenConfig.CommandPrefix, commandNovel, ag),
		fmt.Sprintf("%s%s [%s]，可测试报更功能", kittenConfig.CommandPrefix, commandUpdateTest, ag),
		fmt.Sprintf("%s%s [%s]，可预览更新内容", kittenConfig.CommandPrefix, commandUpdatePreview, ag),
	}, "\n")
	engine := control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	go track()

	// 测试小说报更功能
	engine.OnCommand("更新测试").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		report := novel.update()
		ctx.SendChain(message.Image(novel.HeadURL), message.Text(report))
	})

	// 预览小说更新功能
	engine.OnCommand("更新预览").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		report := novel.Preview
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

func track() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("报更协程出现了错误：%s", err)
		}
	}()

	name := kitten.LoadConfig().NickName[0]
	line := "======================[" + name + "]======================"
	content := strings.Join([]string{
		line,
		"* OneBot + ZeroBot + Golang",
		fmt.Sprintf("一共有 %d 本小说", len(loadConfig())),
		"=======================================================",
	}, "\n")
	fmt.Println(content)

	var (
		novel Novel
		bot   = <-kitten.Bot
	)

	if bot != nil {
		log.Info("报更已经获取到实例了喵！")
	} else {
		log.Error("报更没有获取到实例了喵！")
	}

	// 报更
	for {
		data := loadConfig()
		dataNew := data
		var updateError, update bool
		for idx := range data {
			id := data[idx].BookID
			novel.init(id)
			chapterURL := novel.NewChapter.URL

			// 更新判定，并防止误报
			if chapterURL == data[idx].RecordURL ||
				!novel.IsGet ||
				!novel.NewChapter.IsGet {
				continue
			}

			report := novel.update()

			// 防止更新异常信息发到群里
			if report == "更新异常喵！" {
				updateError = true
				log.Warn(novel.NewChapter.BookURL + report)
			} else {
				for _, groupID := range data[idx].GroupID {
					messageReport := message.Message{message.Image(novel.CoverURL), message.Image(novel.HeadURL), message.Text(report)}
					bot.SendGroupMessage(groupID, messageReport)
					update = true
				}
				dataNew[idx].BookName = novel.Name
				dataNew[idx].RecordURL = novel.NewChapter.URL
				dataNew[idx].UpdateTime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
			}
		}

		// 将本轮获取到的更新链接和更新时间记录至文件，如果没有获取到，或者报更出错，则不写入
		if !updateError && update {
			updateConfig, err1 := yaml.Marshal(dataNew)
			err2 := kitten.FileWrite(path, updateConfig)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				log.Warn(fmt.Sprintf("记录 %s 失败喵！", path))
			} else {
				log.Info(fmt.Sprintf("记录 %s 成功喵！", path))
			}
		}
		time.Sleep(5 * time.Second) // 每 5 秒检测一次
		Alive = true                // 报告协程存活
	}
}

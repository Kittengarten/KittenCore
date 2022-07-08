package sfacg

import (
	"fmt"
	"kitten/kitten"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/control"
	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"

	log "github.com/sirupsen/logrus"
)

var api SFAPI

const (
	replyServiceName = "Kitten_SFACGBP"    // 插件名
	path             = "sfacg/config.yaml" // 配置文件路径
)

func init() {
	// 注册插件
	config := kitten.LoadConfig()
	help := "发送\n" + config.CommandPrefix +
		"小说 [搜索关键词]|[书号]，可获取信息\n" +
		"更新测试 [书号]，可测试报更功能\n" +
		"更新预览 [书号]，可预览更新内容\n"
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	go sfacgTrack()

	// 测试小说报更功能
	engine.OnCommand("更新测试").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		report := novel.Update()
		ctx.SendGroupMessage(ctx.Event.GroupID, report)
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
		ctx.SendGroupMessage(ctx.Event.GroupID, novel.Information())
	})
}

// 获取小说（如果传入值不为书号，则先获取书号）
func getNovel(ctx *zero.Ctx) (nv Novel) {
	ag := ctx.State["args"].(string)
	if !IsInt(ag) {
		var chk bool
		ag, chk = api.FindBookID(ag)
		if !chk {
			ctx.SendGroupMessage(ctx.Event.GroupID, ag)
			return
		}
	}
	nv.Init(ag)
	return nv
}

func sfacgTrack() {
	// 处理panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Error(err)
		}
	}()

	var bot *zero.Ctx
	var novel Novel
	config := LoadConfig()
	name := kitten.LoadConfig().NickName[0]
	line := "======================[" + name + "]======================"
	content := strings.Join([]string{
		line,
		"* OneBot + ZeroBot + Golang",
		fmt.Sprintf("一共有%d本小说", len(config)),
		"=======================================================",
	}, "\n")

	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		bot = ctx
		fmt.Println(content)
		return false
	})

	// 报更
	var report string
	for {
		data := LoadConfig()
		dataNew := data
		updateError := false
		update := false
		for idx := range config {
			id := data[idx].BookId
			chapterUrl := api.FindChapterUrl(id)
			chapterUpdateTime := api.FindChapterUpdateTime(id)
			novel.Init(id)

			// 更新判定，并防止误报
			if chapterUrl == "" ||
				chapterUrl == data[idx].RecordUrl ||
				chapterUpdateTime == data[idx].UpdateTime ||
				!novel.IsGet {
				continue
			}

			report = novel.Update()

			// 防止更新异常信息发到群里
			if report == "更新异常喵！" {
				updateError = true
				log.Warn(novel.NewChapter.BookUrl + report)
			} else {
				for _, groupID := range data[idx].GroupID {
					bot.SendGroupMessage(groupID, report)
				}
				update = true
				dataNew[idx].BookName = novel.Name
				dataNew[idx].RecordUrl = novel.NewChapter.Url
				dataNew[idx].UpdateTime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
			}
		}

		// 将本轮获取到的更新链接和更新时间记录至文件，如果没有获取到，或者报更出错，则不写入
		if !updateError && update {
			updateConfig, err1 := yaml.Marshal(dataNew)
			err2 := kitten.FileWrite(path, updateConfig)
			if !kitten.Check(err1) || !kitten.Check(err2) {
				log.Warn("定时记录" + path + "失败喵！")
			} else {
				log.Info("定时记录" + path + "成功喵！")
			}

		}
		time.Sleep(5 * time.Second) // 每 5 秒检测一次
	}
}

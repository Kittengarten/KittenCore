package sfacg

import (
	"fmt"
	"kitten/kitten"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/control"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"

	log "github.com/sirupsen/logrus"
)

var api SFAPI

const (
	replyServiceName = "Kitten_SFACGBP" // 插件名
)

func init() {
	config := kitten.LoadConfig()
	help := "发送\n" + config.CommandPrefix +
		"小说 [搜索关键词]|[书号]，可获取信息\n" +
		"更新测试 [书号]，可测试报更功能\n" +
		"更新预览 [书号]，可预览更新内容\n"
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	}) // 注册插件

	log.Debug(replyServiceName + "正在启动喵……")
	go sfacgTrack()

	engine.OnCommand("更新测试").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		report := novel.Update()
		ctx.SendGroupMessage(ctx.Event.GroupID, report)
	}) // 测试小说报更功能

	engine.OnCommand("更新预览").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		report := novel.Preview
		if report == "" {
			ctx.SendGroupMessage(ctx.Event.GroupID, "不存在的喵！")
			log.Info("更新预览不存在的喵！")
		} else {
			ctx.SendGroupMessage(ctx.Event.GroupID, report)
		}

	}) // 预览小说更新功能

	engine.OnCommand("小说").Handle(func(ctx *zero.Ctx) {
		novel := getNovel(ctx)
		ctx.SendGroupMessage(ctx.Event.GroupID, novel.Information())
	}) // 小说信息功能
}

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
} // 获取小说（如果传入值不为书号，则先获取书号）

func sfacgTrack() {
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

	var report string
	for {
		for idx := 0; idx < len(config); idx++ {

			chapterUrl := api.FindChapterUrl(config[idx].BookId)
			if chapterUrl == "" {
				continue
			} // 防止误报
			if config[idx].RecordUrl == chapterUrl &&
				config[idx].Updatetime == api.FindChapterUpdateTime(config[idx].BookId) {
				continue
			} // 更新判定
			novel.Init(config[idx].BookId)
			config[idx].RecordUrl = novel.NewChapter.Url
			config[idx].Updatetime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")

			report = novel.Update()

			for _, groupID := range config[idx].GroupID {
				bot.SendGroupMessage(groupID, report)
			}
		}
		time.Sleep(5 * time.Second)
	} // 报更
}

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

	var report string
	for {
		for idx := 0; idx < len(config); idx++ {
			chapterUrl := api.FindChapterUrl(config[idx].BookId)
			chapterUpdateTime := api.FindChapterUpdateTime(config[idx].BookId)
			if chapterUrl == "" ||
				config[idx].RecordUrl == chapterUrl ||
				config[idx].Updatetime == chapterUpdateTime {
				continue
			} // 更新判定，并防止误报
			novel.Init(config[idx].BookId)
			config[idx].RecordUrl = novel.NewChapter.Url
			config[idx].Updatetime = novel.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
			if novel.Update() == "更新异常喵！" {
				continue
			} // 防止更新异常信息发到群里
			report = novel.Update()
			for _, groupID := range config[idx].GroupID {
				bot.SendGroupMessage(groupID, report)
			}
		}
		time.Sleep(5 * time.Second)
	} // 报更
}

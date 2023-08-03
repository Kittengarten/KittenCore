// KittenCore 的主函数所在包
package main

import (
	// WebUI，不需要使用可以注释
	webctrl "github.com/FloatTech/zbputils/control/web"
	"go.uber.org/zap"

	// 以下为外部插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager"
	// _ "github.com/Kittengarten/KittenCore/plugin/kokomi"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aipaint"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/alipayvoice"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dress"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drift_bottle"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jiami"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kfccrazythursday"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moegoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu_calendar"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativesetu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativewife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nsfw"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wantquotes"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenben"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinAI"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/word_count"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus"

	// 以下为内部插件
	// _ "github.com/Kittengarten/KittenCore/draw"
	_ "github.com/Kittengarten/KittenCore/eekda" // XX 今天吃什么
	// _ "github.com/Kittengarten/KittenCore/essence"
	_ "github.com/Kittengarten/KittenCore/perf"  // 查看 XX
	_ "github.com/Kittengarten/KittenCore/sfacg" // SF 轻小说报更
	_ "github.com/Kittengarten/KittenCore/stack" // 叠猫猫

	// 以下为核心依赖
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	// KittenCore 的核心库
	"github.com/Kittengarten/KittenCore/kitten"

	// 官方库
	"runtime/debug"
)

func init() {
	// 启用 WebUI
	go webctrl.RunGui(kitten.Configs.WebUI.URL)
}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			zap.S().Errorf("主函数有 Bug 喵！\n%v", err)
			debug.PrintStack()
		}
	}()
	zero.RunAndBlock(&zero.Config{
		NickName:      kitten.Configs.NickName,
		CommandPrefix: kitten.Configs.CommandPrefix,
		SuperUsers:    kitten.Configs.SuperUsers,
		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向 WS 默认使用 6700 端口
				Url:         kitten.Configs.WebSocket.URL,
				AccessToken: kitten.Configs.WebSocket.AccessToken,
			},
		},
	}, process.GlobalInitMutex.Unlock)
}

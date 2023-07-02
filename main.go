// KittenCore 的主函数所在包
package main

import (
	// 标准库
	"math/rand"
	"runtime"
	"strings"
	"time"

	// KittenCore 的核心库
	"github.com/Kittengarten/KittenCore/kitten"

	// 以下为内部插件
	// _ "github.com/Kittengarten/KittenCore/draw"
	_ "github.com/Kittengarten/KittenCore/eekda" // XX 今天吃什么
	// _ "github.com/Kittengarten/KittenCore/essence"
	_ "github.com/Kittengarten/KittenCore/perf"  // 查看 XX
	_ "github.com/Kittengarten/KittenCore/sfacg" // SF 轻小说报更
	_ "github.com/Kittengarten/KittenCore/stack" // 叠猫猫

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

	// 以下为核心依赖
	"github.com/FloatTech/floatbox/process"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	// WebUI，不需要使用可以注释
	webctrl "github.com/FloatTech/zbputils/control/web"
)

var config = kitten.LoadMainConfig()

func init() {
	logConfig()
	// 启用 WebUI
	go webctrl.RunGui(string(config.WebUI.URL))
}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("主函数有 Bug 喵！\n%v", err)
		}
	}()
	// Go 1.20 之前版本需要全局重置随机数种子，插件无须再次使用
	if !strings.Contains(runtime.Version(), "go1.2") {
		rand.Seed(time.Now().UnixNano())
	}
	zero.RunAndBlock(&zero.Config{
		NickName:      config.NickName,
		CommandPrefix: config.CommandPrefix,
		SuperUsers:    config.SuperUsers,
		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向 WS 默认使用 6700 端口
				Url:         string(config.WebSocket.URL),
				AccessToken: config.WebSocket.AccessToken,
			},
		},
	}, process.GlobalInitMutex.Unlock)
}

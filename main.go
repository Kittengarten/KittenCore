// KittenCore 的主函数所在包
package main

import (
	// 标准库
	"fmt"
	"runtime/debug"

	// KittenCore 的核心库
	"github.com/Kittengarten/KittenCore/internal/protocol"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"

	// 内部插件
	// _ "github.com/Kittengarten/KittenCore/internal/auth" // 内置黑名单控制插件
	// _ "github.com/Kittengarten/KittenCore/plugin/draw"   // 牌堆
	_ "github.com/Kittengarten/KittenCore/plugin/eekda2" // XX 今天吃什么
	_ "github.com/Kittengarten/KittenCore/plugin/rcon"   // RCON
	_ "github.com/Kittengarten/KittenCore/plugin/repeat" // 喵类的本质
	_ "github.com/Kittengarten/KittenCore/plugin/stack2" // 叠猫猫
	_ "github.com/Kittengarten/KittenCore/plugin/track"  // 小说报更
	_ "github.com/Kittengarten/KittenCore/plugin/view"   // 查看 XX

	// _ "github.com/Kittengarten/KittenCore/plugin/weather" // 查看天气

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/sleepmanage" // 统计睡眠时间

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager" // 群管

	_ "github.com/FloatTech/zbputils/job" // 定时指令触发器

	// 外部插件
	_ "github.com/FloatTech/ZeroBot-Plugin-Playground/plugin/vote"   // 实时投票
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"             // ahsai tts
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aifalse"           // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"            // 随机老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/animetrace"        // AnimeTrace 动画/Galgame识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base16384"         // base16384加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"         // base64卦加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"       // base天城文加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"          // b站相关
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chatcount"         // 聊天时长统计
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"            // 选择困难症帮手
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"      // 说抽象话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"             // 英文字符翻转
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"             // 三次元小姐姐
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"           // cp短打
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dailynews"         // 今日早报
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"          // DeepDanbooru二次元图标签识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"              // 程序员做饭指南
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"          // 多功能抽签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/driftbottle"       // 漂流瓶
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"          // 合成emoji
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/event"             // 好友申请群聊邀请事件处理
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"              // 渲染任意文字到图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"             // 笑话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"               // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"            // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"          // 一言
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"            // 注入指令
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jandan"            // 煎蛋网无聊图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"           // lolicon 随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolimi"            // 桑帛云 API
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/magicprompt"       // magicprompt吟唱提示
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"        // 简易midi音乐制作
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/minecraftobserver" // Minecraft服务器监控&订阅
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/movies"            // 电影插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"              // 摸鱼
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyucalendar"      // 摸鱼人日历
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"             // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"           // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/omikuji"           // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/poker"             // 抽扑克
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"             // qq空间表白墙
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/realcugan"         // realcugan清晰术
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"            // 投胎
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/robbery"           // 打劫群友的ATRI币
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"           // 在线运行代码
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"          // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"          // 来份涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"           // 沙雕app
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"           // 测定
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"             // 抽塔罗牌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"           // 舔狗日记
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"          // 搜番
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"       // 翻译
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wallet"            // 钱包
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wantquotes"        // 据意查句
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wife"              // 抽老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordcount"         // 聊天热词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"             // 月幕galgame
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/yujn"              // 遇见API

	// _ "github.com/Kittengarten/KittenCore/plugin/aireply"    // 人工智能回复

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus" // 词典匹配回复

	// WebUI，不需要使用可以注释
	webctrl "github.com/FloatTech/zbputils/control/web"
)

func init() {
	// 启用 WebUI，不需要使用可以注释
	go webctrl.RunGui(kitten.MainConfig().Host)
}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); err != nil {
			kitten.Errorf("main() 从 panic 中恢复\n%v\n%s", err, debug.Stack())
			if err := fio.NewPath(kitten.MainConfig().Log.Crash).WriteString(
				fmt.Sprintf("%v\n%s\n", err, debug.Stack()),
			); err != nil {
				kitten.Errorln(`写入`, kitten.MainConfig().Crash, err)
			}
		}
	}()
	protocol.RunBot(protocol.Forward)
}

// Package view 查看服务器运行状况
package view

import (
	"fmt"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/plugin/view/utils"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `perf` // 插件名
	brief            = `查看运行状况`
	filePath         = `file.txt` // 保存微星小飞机温度配置文件路径的文件，非 Windows 系统或不使用可以忽略
	cView            = `查看`
	poke             = `notice/notify/poke`
	logFile          = `C:\Program Files (x86)\MSI Afterburner\HardwareMonitoring.hml`
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            brief,
		Help: func() string {
			var s strings.Builder // 字符串构建器
			s.Grow(32 * len(kitten.MainConfig().NickName))
			for _, n := range kitten.MainConfig().NickName {
				fmt.Fprintln(&s, kitten.MainConfig().CommandPrefix+cView, n, `// 可获取服务器运行状况`)
			}
			fmt.Fprint(&s, `戳一戳`, kitten.MainConfig().NickName[0], ` // 可得到响应`)
			return s.String()
		}(),
	}).ApplySingle(ctxext.DefaultSingle)
	// bot 自身 ID
	sid = kitten.Self()
	// 日志文件
	logPath = io.FilePath(engine.DataFolder(), `logPath.txt`)
	// 日志文件路径
	logFilePath io.Path
)

func init() {
	if err := logPath.InitFile(logFile); err != nil {
		kitten.Error(err)
	}
	logFilePath = logPath.Get(logFile)

	// 查看功能
	engine.OnCommand(cView).SetBlock(true).
		Limit(rate.Get(rate.User)).
		Limit(rate.Get(rate.GroupNormal)).
		Handle(func(ctx *zero.Ctx) {
			utils.View(kitten.New(ctx), replyServiceName, logFilePath)
		})

	// 支付宝到账语音
	engine.OnPrefix(`支付宝到账`).SetBlock(true).
		Limit(rate.Get(rate.User)).
		Limit(rate.Get(rate.GroupNormal)).
		Handle(func(ctx *zero.Ctx) {
			utils.SendAlipayVoice(kitten.New(ctx))
		})

	// Ping 功能
	engine.OnCommandGroup([]string{`Ping`, `ping`}, zero.SuperUserPermission).
		SetBlock(true).Limit(rate.Get(rate.GroupFast)).Handle(func(ctx *zero.Ctx) {
		utils.Ping(kitten.New(ctx))
	})

	// 戳一戳
	engine.On(poke, sid.IsTarget()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.Poke(kitten.New(ctx))
	})

	// 通过链接让 Bot 发送图片，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.SendImage(kitten.New(ctx), zero.SuperUserPermission(ctx))
	})

	// 通过链接、图片等让 Bot 扫描二维码，为防止滥用，仅管理员可用
	zero.OnCommandGroup([]string{`扫码`, `扫描`}, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.Scan(kitten.New(ctx),
			zero.SuperUserPermission(ctx),
			zero.HasPicture(ctx),
			func() bool { return zero.MustProvidePicture(ctx) },
		)
	})
}

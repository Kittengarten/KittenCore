// Package view 查看服务器运行状况
package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/kitten/usr"
	"github.com/Kittengarten/KittenCore/plugin/view/img"
	"github.com/Kittengarten/KittenCore/plugin/view/perf"
	"github.com/Kittengarten/KittenCore/plugin/view/views"
	"github.com/Kittengarten/KittenCore/plugin/view/voice"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `perf` // 插件名
	brief            = `查看运行状况`
	cView            = `查看`
	poke             = `notice/notify/poke`
)

// 注册插件
var engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
	DisableOnDefault: false,
	Brief:            brief,
	Help: func() string {
		s := new(strings.Builder)
		s.Grow(32 * len(config.Name()))
		for _, n := range config.Name() {
			fmt.Fprintln(s, config.CommandPrefix()+cView, n, `// 可获取服务器运行状况`)
		}
		fmt.Fprint(s, `戳一戳`, config.DefaultName(), ` // 可得到响应`)
		return s.String()
	}(),
}).ApplySingle(ctxext.NewGroupSingle(`等一下喵！`))

var (
	// 日志文件
	logFile = perf.LogFile
	// 日志文件配置
	logPath = fio.NewPath(engine.DataFolder(), `logPath.txt`)
	// 日志文件路径
	logFilePath fio.Path
)

func init() {
	if err := logPath.InitFileText(string(logFile)); err != nil {
		log.Error(err)
	}
	logFilePath = logPath.Get(logFile)
	perf.LogFile = logFilePath

	// 查看功能
	engine.OnCommand(cView).SetBlock(true).
		Limit(rate.User.Get()).
		Limit(rate.GroupNormal.Get()).
		Handle(func(ctx *zero.Ctx) {
			c, cancel := context.WithTimeout(context.Background(), core.Timeout)
			defer cancel()
			views.View(msg.NewWithContext(c, ctx), replyServiceName, logFilePath)
		})

	// 支付宝到账语音
	engine.OnPrefix(`支付宝到账`).SetBlock(true).
		Limit(rate.User.Get()).
		Limit(rate.GroupNormal.Get()).
		Handle(func(ctx *zero.Ctx) {
			c, cancel := context.WithTimeout(context.Background(), core.Timeout)
			defer cancel()
			voice.SendAlipayVoice(msg.NewWithContext(c, ctx))
		})

	// Ping 功能
	engine.OnCommandGroup([]string{`Ping`, `ping`}, zero.SuperUserPermission).
		SetBlock(true).Limit(rate.GroupFast.Get()).Handle(func(ctx *zero.Ctx) {
		c, cancel := context.WithTimeout(context.Background(), core.Timeout)
		defer cancel()
		perf.Ping(msg.NewWithContext(c, ctx))
	})

	// 戳一戳
	engine.On(poke, usr.Self().IsTarget()).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, cancel := context.WithTimeout(context.Background(), core.Timeout)
		defer cancel()
		views.Poke(msg.NewWithContext(c, ctx))
	})

	// 通过链接让 Bot 发送图片，为防止滥用，仅管理员可用
	zero.OnCommand(`图片`, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, cancel := context.WithTimeout(context.Background(), core.Timeout)
		defer cancel()
		img.SendImage(msg.NewWithContext(c, ctx), zero.SuperUserPermission(ctx))
	})

	// 通过链接、图片等让 Bot 扫描二维码，为防止滥用，仅管理员可用
	zero.OnCommandGroup([]string{`扫码`, `扫描`}, zero.AdminPermission).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, cancel := context.WithTimeout(context.Background(), core.Timeout)
			defer cancel()
			img.Scan(msg.NewWithContext(c, ctx),
				zero.SuperUserPermission,
				zero.MustProvidePicture,
			)
		})
}

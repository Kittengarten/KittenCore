package rcon

import (
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/plugin/rcon/utils"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `rcon` // 插件名
	brief            = `RCON 命令执行`
	configFile       = `config.yaml` // 配置文件名
	help             = `RCON [命令]
————
私聊可用：
设置 RCON 主机 [主机]
设置 RCON 密码 [密码]`
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 配置文件路径
	configPath = io.NewPath(engine.DataFolder(), configFile)
)

func init() {
	// RCON
	engine.OnPrefixGroup([]string{`RCON`, `rcon`}, zero.SuperUserPermission).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.Command(kitten.New(ctx), configPath)
	})

	// 设置 RCON 主机
	engine.OnRegex(`^设置\s*(?i)RCON\s*主机\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.Set(kitten.New(ctx), utils.Host, configPath)
	})

	// 设置 RCON 密码
	engine.OnRegex(`^设置\s*(?i)RCON\s*密码\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		utils.Set(kitten.New(ctx), utils.Password, configPath)
	})
}

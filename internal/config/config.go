// Package config 配置文件
package config

import (
	_ "embed"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	corelog "github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/mode"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
)

const (
	file         = `config.yaml`   // 配置文件名
	imageFolder  = `image`         // 图片文件夹名
	dataInternal = `data_internal` // 内置数据文件夹名
)

var (
	//go:embed config_example.yaml
	defaultConfig string
	botConfig     config // 来自 Bot 的配置文件
	cfg           = slog.String(`文件`, file)
	imagePath     mio.Path // 图片路径
)

func init() {
	if mode.Test() {
		slog.Info(`测试中，不初始化资源喵！`)
		return
	}
	var err error
	// 配置文件初始化
	if botConfig, err = fio.Load[config](fio.NewPath(file), defaultConfig); err != nil {
		log.Fatalln(err, `请按 YAML 格式配置`, file, `后重新启动喵！`)
	}
	defer corelog.ZapInit(botConfig.Log, mode.Test())
	if len(botConfig.SuperUsers) == 0 {
		slog.Error(`superusers 未配置喵！`, cfg)
	}
	if len(botConfig.NickName) == 0 {
		slog.Warn(`没有配置昵称，使用默认昵称喵！`)
		botConfig.NickName = []string{`喵喵`}
	}
	if _, err = url.Parse(botConfig.URL); err != nil {
		log.Fatalln(err, `请正确配置`, file, `中的 websocket.url 喵！`)
	}
	// 重定向崩溃日志
	crashLog()
	slog.Info(`当前配置`, slog.Any(`config`, botConfig))
	// 运行 pprof
	RunPProf()
	// 初始化资源文件
	if err = initResource(); err != nil {
		slog.Error(`path 配置失败喵！`, cfg, slog.Any(`错误`, err))
	}
	// 图片路径
	imagePath = mio.NewPath(botConfig.Path, imageFolder)
	slog.Info(`图片库配置完成`, slog.Any(`路径`, imagePath))
}

// 重定向崩溃日志
func crashLog() {
	crash, err := os.Open(botConfig.Crash)
	if err != nil {
		crash, err = os.Create(botConfig.Crash)
		if err != nil {
			slog.Error(`log.crash 配置失败喵！`, cfg, slog.Any(`错误`, err))
			return
		}
	}
	if err := debug.SetCrashOutput(crash, debug.CrashOptions{}); err != nil {
		slog.Error(`设置崩溃日志输出失败喵！`, slog.Any(`错误`, err))
	}
}

// 初始化资源文件
func initResource() error {
	if botConfig.Exists() {
		return nil
	}
	slog.Info(`没有找到资源文件，正在初始化喵！`, slog.Any(`目标`, botConfig.Path))
	return os.CopyFS(botConfig.Path.String(), os.DirFS(dataInternal))
}

// Get 获取主配置
func Get() config {
	return botConfig
}

// Name 获取昵称
func Name() []string {
	return botConfig.NickName
}

// DefaultName 获取默认昵称
func DefaultName() string {
	if len(botConfig.NickName) == 0 {
		return `喵喵`
	}
	return botConfig.NickName[0]
}

// ID 获取 SelfID
func ID() int64 {
	return botConfig.SelfID
}

// SuperUsers 获取超级用户
func SuperUsers() []int64 {
	return botConfig.SuperUsers
}

// CommandPrefix 获取命令前缀
func CommandPrefix() string {
	return botConfig.CommandPrefix
}

// AddSpaceAfterAt 获取是否添加空格
func AddSpaceAfterAt() bool {
	return botConfig.AddSpaceAfterAt
}

// ImagePath 获取图片路径
func ImagePath() mio.Path {
	return imagePath
}

// WebUIURL 获取 WebUI 路径
func WebUIURL() string {
	return botConfig.WebUI.getURL()
}

// RunPProf 运行 pprof
func RunPProf() {
	if botConfig.PProf.Enable {
		utils.Go(`pprof`, func() {
			slog.Info(`请访问：http://` + PProfURL() + `/debug/pprof/`)
			if err := http.ListenAndServe(PProfURL(), nil); err != nil {
				slog.Error(`pprof 服务端启动失败喵！`, slog.Any(`错误`, err))
			}
		})
	}
}

// PProfURL 获取 pprof 路径
func PProfURL() string {
	return botConfig.PProf.getURL()
}

func (s Server) getURL() string {
	return net.JoinHostPort(s.Host, strconv.FormatUint(s.Port, 10))
}

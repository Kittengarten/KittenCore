// Package kitten KittenCore 核心依赖
package kitten

import (
	"embed"
	"io/fs"
	"log"
	"net/url"
	"os"
	"runtime/debug"

	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/kitten/core/msg/mio"
)

const (
	configFile  = `config.yaml` // 配置文件名
	imageFolder = `image`       // 图片文件夹名
)

var (
	//go:embed config_example.yaml
	defaultConfig string
	botConfig     config // 来自 Bot 的配置文件
	//go:embed data_internal
	data      embed.FS
	imagePath mio.Path // 图片路径
	Weight    int      // 自身叠猫猫体重（0.1 kg 数）
)

func init() {
	var err error
	// 配置文件初始化
	if botConfig, err = io.Load[config](io.NewPath(configFile), defaultConfig); err != nil {
		log.Fatalln(err, `请按 YAML 格式配置`, configFile, `后重新启动喵！`)
	}
	// 启用 zap 日志格式
	zapInit()
	if len(botConfig.SuperUsers) == 0 {
		log.Println(`请在`, configFile, `中配置 superusers 喵！`)
	}
	if len(botConfig.NickName) == 0 {
		log.Println(`没有配置昵称，使用默认昵称喵！`)
		botConfig.NickName = []string{`喵喵`}
	}
	if _, err = url.Parse(botConfig.WebSocket.URL); err != nil {
		log.Fatalln(err, `请正确配置 `, configFile, ` 中的 websocket.url 喵！`)
	}
	if _, err = url.Parse(`http://` + botConfig.WebUI.Host); err != nil {
		log.Fatalln(err, `请正确配置`, configFile, `中的 webui.url 喵！`)
	}
	// 重定向崩溃日志
	crashLog()
	log.Printf("当前配置：%#v\n", botConfig)
	// 初始化资源文件
	if err = initResource(); err != nil {
		log.Println(err, `请正确配置`, configFile, `中的 path 喵！`)
	}
	// 图片路径
	imagePath = mio.New(io.NewPath(botConfig.Path, imageFolder))
	log.Println(`图片路径：`, imagePath)
}

// 重定向崩溃日志
func crashLog() {
	crash, err := os.Open(botConfig.Log.Crash)
	if err != nil {
		crash, err = os.Create(botConfig.Log.Crash)
		if err != nil {
			log.Println(err, `请配置`, configFile, `中的 log.crash 喵！`)
			return
		}
	}
	if err := debug.SetCrashOutput(crash, debug.CrashOptions{}); err != nil {
		log.Println(err)
	}
}

// 初始化资源文件
func initResource() error {
	if botConfig.Path.Exists() {
		return nil
	}
	log.Println(`没有找到资源文件，正在初始化喵！目标：`, botConfig.Path)
	data, err := fs.Sub(data, `data_internal`)
	if err != nil {
		return err
	}
	if err := os.CopyFS(botConfig.Path.String(), data); err != nil {
		return err
	}
	return nil
}

// MainConfig 获取主配置
func MainConfig() config {
	return botConfig
}

// ImagePath 获取图片路径
func ImagePath() mio.Path {
	return imagePath
}

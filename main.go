// KittenCore 的主函数所在包
package main

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/process"
	webctrl "github.com/FloatTech/zbputils/control/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	"github.com/Kittengarten/KittenCore/kitten"

	_ "github.com/Kittengarten/KittenCore/abuse"
	_ "github.com/Kittengarten/KittenCore/eekda"
	_ "github.com/Kittengarten/KittenCore/essence"
	_ "github.com/Kittengarten/KittenCore/perf"
	_ "github.com/Kittengarten/KittenCore/sfacg"
	_ "github.com/Kittengarten/KittenCore/stack"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aipaint"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/alipayvoice"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dress"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jiami"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moegoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu_calendar"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativewife"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/quan"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenben"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinAI"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"
	_ "github.com/Kittengarten/KittenCore/plugin/kokomi"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply"

	logf "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
)

// 颜色代码常量
const (
	colorCodePanic = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
	colorCodeFatal = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
	colorCodeError = "\x1b[31m"   // color.Style{color.Red}.String()
	colorCodeWarn  = "\x1b[33m"   // color.Style{color.Yellow}.String()
	colorCodeInfo  = "\x1b[37m"   // color.Style{color.White}.String()
	colorCodeDebug = "\x1b[32m"   // color.Style{color.Green}.String()
	colorCodeTrace = "\x1b[36m"   // color.Style{color.Cyan}.String()
	colorReset     = "\x1b[0m"
)

var config = kitten.LoadConfig()

// 获取日志等级对应色彩代码
func getLogLevelColorCode(level log.Level) string {
	switch level {
	case log.PanicLevel:
		return colorCodePanic
	case log.FatalLevel:
		return colorCodeFatal
	case log.ErrorLevel:
		return colorCodeError
	case log.WarnLevel:
		return colorCodeWarn
	case log.InfoLevel:
		return colorCodeInfo
	case log.DebugLevel:
		return colorCodeDebug
	case log.TraceLevel:
		return colorCodeTrace
	default:
		return colorCodeInfo
	}
}

// LogFormat 日志输出样式
type LogFormat struct{}

// Format 设置该日志输出样式的具体样式
func (f LogFormat) Format(entry *log.Entry) ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(getLogLevelColorCode(entry.Level))
	buf.WriteByte('[')
	buf.WriteString(entry.Time.Format(`2006-01-02 15:04:05`))
	buf.WriteString(`] `)
	buf.WriteByte('[')
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteString(`]: `)
	buf.WriteString(entry.Message)
	buf.WriteString(" \n")
	buf.WriteString(colorReset)
	return buf.Bytes(), nil
}

func init() {
	var (
		logName = config.Log.Path // 日志文件路径
		logF    = config.Log.Days // 单段分割文件记录的天数
		// 配置分割日志文件
		writer, err = logf.New(
			// 分割日志文件命名规则
			kitten.GetMidText(``, `.txt`, logName)+`-%Y-%m-%d.txt`,
			// 与最新的日志文件建立软链接
			logf.WithLinkName(logName),
			// 分割日志文件间隔
			logf.WithRotationTime(24*time.Duration(int(time.Hour)*logF)),
			// 禁用清理
			logf.WithMaxAge(-1),
		)
	)
	if !kitten.Check(err) {
		log.Errorf("主函数配置分割日志文件失败喵！\n%v", err)
		return
	}
	// 设置日志输出样式
	log.SetFormatter(&LogFormat{})
	if mw := io.MultiWriter(os.Stdout, writer); kitten.Check(err) {
		log.SetOutput(mw)
	} else {
		log.Warn(`主函数写入日志失败了喵！`)
	}
	// 设置最低日志等级
	log.SetLevel(config.Log.GetLogLevel())
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

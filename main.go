// KittenCore 的主函数所在包
package main

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/process"
	_ "github.com/Kittengarten/KittenCore/abuse"
	_ "github.com/Kittengarten/KittenCore/essence"
	"github.com/Kittengarten/KittenCore/kitten"
	_ "github.com/Kittengarten/KittenCore/perf"
	"github.com/Kittengarten/KittenCore/sfacg"
	_ "github.com/Kittengarten/KittenCore/sfacg"
	_ "github.com/Kittengarten/KittenCore/stack"

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

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
	buf.WriteString(entry.Time.Format("2006-01-02 15:04:05"))
	buf.WriteString("] ")
	buf.WriteByte('[')
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteString("]: ")
	buf.WriteString(entry.Message)
	buf.WriteString(" \n")
	buf.WriteString(colorReset)
	return buf.Bytes(), nil
}

func init() {
	var (
		config  = kitten.LoadConfig()
		logName = config.Log.Path // 日志文件路径
		logF    = config.Log.Days // 单段分割文件记录的天数
	)
	// 配置分割日志文件
	writer, err := logf.New(
		// 分割日志文件命名规则
		kitten.GetMidText("", ".txt", logName)+"-%Y-%m-%d.txt",
		// 与最新的日志文件建立软链接
		logf.WithLinkName(logName),
		// 分割日志文件间隔
		logf.WithRotationTime(time.Duration(int(time.Hour)*24*logF)),
		// 禁用清理
		logf.WithMaxAge(-1),
	)
	if !kitten.Check(err) {
		log.Errorf("配置分割日志文件失败：%v", err)
	}

	log.SetFormatter(&LogFormat{}) // 设置日志输出样式
	mw := io.MultiWriter(os.Stdout, writer)
	if kitten.Check(err) {
		log.SetOutput(mw)
	} else {
		log.Warn("写入日志失败了喵！")
	}
	log.SetLevel(log.TraceLevel) // 设置最低日志等级

}

func main() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("main 函数有 Bug：%s，喵！", err)
		}
	}()

	go checkAlive(&sfacg.Alive, "sfacg 报更") // 检查 sfacg 报更协程是否存活

	config := kitten.LoadConfig()
	log.Info("已经载入配置了喵！")
	rand.Seed(time.Now().UnixNano()) // 全局重置随机数种子，插件无须再次使用

	zero.RunAndBlock(zero.Config{
		NickName:      config.NickName,
		CommandPrefix: config.CommandPrefix,
		SuperUsers:    config.SuperUsers,
		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向WS 默认使用 6700 端口
				Url:         config.WebSocket.URL,
				AccessToken: config.WebSocket.AccessToken,
			},
		},
	}, process.GlobalInitMutex.Unlock)
}

// 检查协程是否存活
func checkAlive(ok *bool, name string) {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); !kitten.Check(err) {
			log.Errorf("checkAlive 函数有 Bug：%s，喵！", err)
		}
	}()

	config := kitten.LoadConfig()

	for {
		*ok = false
		time.Sleep(time.Minute) // 每分钟检测一次
		if !*ok {
			zero.GetBot(config.SelfID).SendPrivateMessage(config.SuperUsers[0], name+"协程挂掉了喵！")
			break // 停止检测
		}
	}
}

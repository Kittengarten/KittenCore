package kitten

import (
	"bytes"
	"io"
	"os"
	"strings"
	"time"

	logf "github.com/lestrrat-go/file-rotatelogs"
	log "github.com/sirupsen/logrus"
)

// LogFormat 日志输出样式
type LogFormat struct{}

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

// LogConfigInit 日志配置初始化
func LogConfigInit(config Config) {
	var (
		logName   = config.Log.Path   // 日志文件路径
		logRotate = config.Log.Days   // 单段分割文件记录的天数
		logExpire = config.Log.Expire // 日志文件的过期天数
		// 配置分割日志文件
		writer, err = logf.New(
			// 分割日志文件命名规则
			GetMidText(``, `.txt`, logName)+`-%Y-%m-%d.txt`,
			// 与最新的日志文件建立软链接
			logf.WithLinkName(logName),
			// 分割日志文件间隔
			logf.WithRotationTime(24*time.Hour*time.Duration(logRotate)),
			// 日志文件过期天数
			logf.WithMaxAge(24*time.Hour*time.Duration(logExpire)),
		)
	)
	if !Check(err) {
		log.Errorf("主函数配置分割日志文件失败喵！\n%v", err)
		return
	}
	// 设置日志输出样式
	log.SetFormatter(&LogFormat{})
	if mw := io.MultiWriter(os.Stdout, writer); Check(err) {
		log.SetOutput(mw)
	} else {
		log.Error(`主函数写入日志失败了喵！`)
	}
	// 设置最低日志等级
	log.SetLevel(config.Log.GetLogLevel())
}

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

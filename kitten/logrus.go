package kitten

// import (
// 	"bytes"
// 	"os"
// 	"strings"

// 	"github.com/sirupsen/logrus"
// )

// // LogFormat 日志输出样式
// type LogFormat struct{}

// // 颜色代码常量
// const (
// 	colorCodePanic = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
// 	colorCodeFatal = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
// 	colorCodeError = "\x1b[31m"   // color.Style{color.Red}.String()
// 	colorCodeWarn  = "\x1b[33m"   // color.Style{color.Yellow}.String()
// 	colorCodeInfo  = "\x1b[37m"   // color.Style{color.White}.String()
// 	colorCodeDebug = "\x1b[32m"   // color.Style{color.Green}.String()
// 	colorCodeTrace = "\x1b[36m"   // color.Style{color.Cyan}.String()
// 	colorReset     = "\x1b[0m"
// )

// func init() {
// 	// 启用日志格式
// 	LogrusConfigInit(Configs)
// }

// // LogrusConfigInit 日志配置初始化
// func LogrusConfigInit(config Config) {
// 	// 设置日志输出样式
// 	logrus.SetFormatter(&LogFormat{})
// 	logrus.SetOutput(os.Stdout)
// 	// 设置最低日志等级
// 	logrus.SetLevel(config.Log.GetLogrusLevel())
// }

// // GetLogrusLevel 从日志配置获取日志等级
// func (lc LogConfig) GetLogrusLevel() logrus.Level {
// 	switch lc.Level {
// 	case logrus.PanicLevel.String():
// 		return logrus.PanicLevel
// 	case logrus.FatalLevel.String():
// 		return logrus.FatalLevel
// 	case logrus.ErrorLevel.String():
// 		return logrus.ErrorLevel
// 	case logrus.WarnLevel.String():
// 		return logrus.WarnLevel
// 	case logrus.InfoLevel.String():
// 		return logrus.InfoLevel
// 	case logrus.DebugLevel.String():
// 		return logrus.DebugLevel
// 	case logrus.TraceLevel.String():
// 		return logrus.TraceLevel
// 	default:
// 		return logrus.WarnLevel
// 	}
// }

// // 获取日志等级对应色彩代码
// func getLogLevelColorCode(level logrus.Level) string {
// 	switch level {
// 	case logrus.PanicLevel:
// 		return colorCodePanic
// 	case logrus.FatalLevel:
// 		return colorCodeFatal
// 	case logrus.ErrorLevel:
// 		return colorCodeError
// 	case logrus.WarnLevel:
// 		return colorCodeWarn
// 	case logrus.InfoLevel:
// 		return colorCodeInfo
// 	case logrus.DebugLevel:
// 		return colorCodeDebug
// 	case logrus.TraceLevel:
// 		return colorCodeTrace
// 	default:
// 		return colorCodeInfo
// 	}
// }

// // Format 设置该日志输出样式的具体样式
// func (f LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
// 	buf := new(bytes.Buffer)
// 	buf.WriteString(getLogLevelColorCode(entry.Level))
// 	buf.WriteByte('[')
// 	buf.WriteString(entry.Time.Format(`2006.1.2 15:04:05`))
// 	buf.WriteString(`] `)
// 	buf.WriteByte('[')
// 	buf.WriteString(strings.ToUpper(entry.Level.String()))
// 	buf.WriteString(`]: `)
// 	buf.WriteString(entry.Message)
// 	buf.WriteString(" \n")
// 	buf.WriteString(colorReset)
// 	return buf.Bytes(), nil
// }

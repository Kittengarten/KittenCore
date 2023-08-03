package kitten

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	// 启用日志格式
	LogConfigInit(Configs)
}

// LogConfigInit 日志配置初始化
func LogConfigInit(config Config) {
	var (
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:       `time`,
			LevelKey:      `level`,
			NameKey:       `logger`,
			CallerKey:     `caller`,
			MessageKey:    `msg`,
			StacktraceKey: `stacktrace`,
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalColorLevelEncoder, // 指定颜色
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString("[" + t.Format(`2006.1.2 15:04:05`) + "]")
			}, // 时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString("[" + caller.TrimmedPath() + "]")
			}, // 路径编码器
			EncodeName: zapcore.FullNameEncoder,
		}
		logWriteSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.Log.Path,       // 日志文件存放目录，如果文件夹不存在会自动创建
			MaxSize:    config.Log.MaxSize,    // 文件大小限制，单位 MB
			MaxBackups: config.Log.MaxBackups, // 最大保留日志文件数量
			MaxAge:     config.Log.MaxAge,     // 日志文件保留天数
			LocalTime:  true,                  // 采用本地时间
			Compress:   false,                 // 是否压缩处理
		})
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // 获取编码器，NewJSONEncoder() 输出 json 格式，NewConsoleEncoder() 输出普通文本格式
		core    = zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.Lock(WrappedWriteSyncer{os.Stdout}), logWriteSyncer), config.Log.GetLogLevel())
		log_    = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2)) // 配置日志记录器
	)
	defer func() {
		if err := log_.Sync(); !Check(err) {
			log_.Sugar().Errorf("日志刷新失败喵！\n%v", err)
		} else {
			log_.Info(`日志刷新成功喵！`)
		}
	}()
	zap.ReplaceGlobals(log_)
	zap.RedirectStdLog(log_)
}

// GetLogLevel 从日志配置获取日志等级
func (lc LogConfig) GetLogLevel() zapcore.Level {
	switch lc.Level {
	case zap.FatalLevel.String():
		return zap.FatalLevel
	case zap.PanicLevel.String():
		return zap.PanicLevel
	case zap.DPanicLevel.String():
		return zap.DPanicLevel
	case zap.ErrorLevel.String():
		return zap.ErrorLevel
	case zap.WarnLevel.String():
		return zap.WarnLevel
	case zap.InfoLevel.String():
		return zap.InfoLevel
	case zap.DebugLevel.String():
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}

type WrappedWriteSyncer struct {
	file *os.File
}

func (mws WrappedWriteSyncer) Write(p []byte) (n int, err error) {
	return mws.file.Write(p)
}
func (mws WrappedWriteSyncer) Sync() error {
	return nil
}

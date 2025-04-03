package kitten

import (
	"os"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WrappedWriteSyncer 包装写入同步结构
type WrappedWriteSyncer struct {
	file *os.File
}

// 日志编码器配置
var encoderConfig = zapcore.EncoderConfig{
	TimeKey:       `time`,
	LevelKey:      `level`,
	NameKey:       `logger`,
	CallerKey:     `caller`,
	MessageKey:    `msg`,
	StacktraceKey: `stacktrace`,
	LineEnding:    zapcore.DefaultLineEnding,
	EncodeLevel:   zapcore.CapitalColorLevelEncoder, // 指定颜色
	EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(`[` + t.Format(times.Layout) + `]`)
	}, // 时间格式
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(`[` + caller.TrimmedPath() + `]`)
	}, // 路径编码器
	EncodeName: zapcore.FullNameEncoder,
}

// zap 日志配置初始化
func zapInit() {
	if err := fio.NewPath(botConfig.Log.Path).InitFile(``); err != nil {
		zap.Error(err)
	}
	// 日志记录器配置
	log := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(
				zapcore.Lock(WrappedWriteSyncer{os.Stdout}),
				zapcore.AddSync(rotate(botConfig)),
			),
			level(botConfig.Log),
		),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	)
	defer func() {
		if err := log.Sync(); err != nil {
			log.Sugar().Error(`日志刷新失败喵！`, err)
			return
		}
		log.Info(`日志刷新成功喵！`)
	}()
	zap.ReplaceGlobals(log)
	zap.RedirectStdLog(log)
	zap.AddStacktrace(zap.WarnLevel)
}

// 获取 zap 日志等级
func level(lc LogConfig) zapcore.Level {
	level, err := zap.ParseAtomicLevel(lc.Level)
	if err != nil {
		zap.Error(err)
		return zap.InfoLevel
	}
	return level.Level()
}

// Write WrappedWriteSyncer 实现 Writer 接口
func (mws WrappedWriteSyncer) Write(p []byte) (int, error) {
	return mws.file.Write(p)
}

// Sync 同步
func (mws WrappedWriteSyncer) Sync() error {
	return nil
}

// Debug 在 Debug 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Debug(args ...any) {
	zap.S().Debug(args...)
}

// Info 在 Info 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Info(args ...any) {
	zap.S().Info(args...)
}

// Warn 在 Warn 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Warn(args ...any) {
	zap.S().Warn(args...)
}

// Error 在 Error 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Error(args ...any) {
	zap.S().Error(args...)
}

// Panic 在 Panic 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Panic(args ...any) {
	zap.S().Panic(args...)
}

// Fatal 在 Fatal 等级记录提供的参数。当参数都不是字符串时，会在参数之间添加空格。
func Fatal(args ...any) {
	zap.S().Fatal(args...)
}

// Debugf 根据格式说明符设置消息的格式，并将其记录在 Debug 等级中。
func Debugf(format string, args ...any) {
	zap.S().Debugf(format, args...)
}

// Infof 根据格式说明符设置消息的格式，并将其记录在 Info 等级中。
func Infof(format string, args ...any) {
	zap.S().Infof(format, args...)
}

// Warnf 根据格式说明符设置消息的格式，并将其记录在 Warn 等级中。
func Warnf(format string, args ...any) {
	zap.S().Warnf(format, args...)
}

// Errorf 根据格式说明符设置消息的格式，并将其记录在 Error 等级中。
func Errorf(format string, args ...any) {
	zap.S().Errorf(format, args...)
}

// Panicf 根据格式说明符设置消息的格式，并将其记录在 Panic 等级中。
func Panicf(format string, args ...any) {
	zap.S().Panicf(format, args...)
}

// Fatalf 根据格式说明符设置消息的格式，并将其记录在 Fatal 等级中。
func Fatalf(format string, args ...any) {
	zap.S().Fatalf(format, args...)
}

// Debugln 在 Debug 等级记录一条消息。参数之间始终添加空格。
func Debugln(args ...any) {
	zap.S().Debugln(args...)
}

// Infoln 在 Info 等级记录一条消息。参数之间始终添加空格。
func Infoln(args ...any) {
	zap.S().Infoln(args...)
}

// Warnln 在 Warn 等级记录一条消息。参数之间始终添加空格。
func Warnln(args ...any) {
	zap.S().Warnln(args...)
}

// Errorln 在 Error 等级记录一条消息。参数之间始终添加空格。
func Errorln(args ...any) {
	zap.S().Errorln(args...)
}

// Panicln 在 Panic 等级记录一条消息。参数之间始终添加空格。
func Panicln(args ...any) {
	zap.S().Panicln(args...)
}

// Fatalln 在 Fatal 等级记录一条消息。参数之间始终添加空格。
func Fatalln(args ...any) {
	zap.S().Fatalln(args...)
}

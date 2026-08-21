package log

import (
	"os"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapInit zap 日志配置初始化
func ZapInit(cfg Log, test bool) {
	if !test {
		if err := fio.NewPath(cfg.Path).InitFile(); err != nil {
			zap.Error(err)
		}
	}
	// 日志记录器配置
	log := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
				MessageKey:    `msg`,
				LevelKey:      `level`,
				TimeKey:       `time`,
				NameKey:       `logger`,
				CallerKey:     `caller`,
				FunctionKey:   `func`,
				StacktraceKey: `trace`,
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
			}), // 日志编码器配置
			func() zapcore.WriteSyncer {
				stdout := zapcore.Lock(zapcore.AddSync(os.Stdout))
				if test {
					return stdout
				}
				return zapcore.NewMultiWriteSyncer(
					stdout,
					zapcore.Lock(zapcore.AddSync(rotate(cfg))),
				)
			}(),
			func() zapcore.Level {
				if test {
					return zap.DebugLevel
				}
				return level(cfg)
			}(), // 日志等级
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.WarnLevel),
	)
	zap.ReplaceGlobals(log)
	zap.RedirectStdLog(log)
}

// 获取 zap 日志等级
func level(cfg Log) zapcore.Level {
	level, err := zap.ParseAtomicLevel(strings.ToLower(cfg.Level))
	if err != nil {
		zap.Error(err)
		return zap.InfoLevel
	}
	return level.Level()
}

// Skip 创建配置了 AddCallerSkip 的新 *zap.SugaredLogger
//
// 默认已有 1 层，配置时会向上追加
func Skip(n int) *zap.SugaredLogger {
	return zap.S().WithOptions(zap.AddCallerSkip(n))
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

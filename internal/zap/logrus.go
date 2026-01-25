package logrus

import (
	"context"
	"io"
	"log"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func (l Logger) Warningf(format string, args ...any) {
	l.Warnf(format, args...)
}

// WithContext ctx 会被忽略
func WithContext(_ context.Context) Logger {
	return Logger{SugaredLogger: zap.S()}
}

func WithError(err error) Logger {
	return Logger{SugaredLogger: zap.S().With(zap.Error(err))}
}

func Debug(args ...any) {
	zap.S().Debug(args...)
}

func Info(args ...any) {
	zap.S().Info(args...)
}

func Warn(args ...any) {
	zap.S().Warn(args...)
}

func Warning(args ...any) {
	zap.S().Warn(args...)
}

func Error(args ...any) {
	zap.S().Error(args...)
}

func Panic(args ...any) {
	zap.S().Panic(args...)
}

func Fatal(args ...any) {
	zap.S().Fatal(args...)
}

func Debugf(format string, args ...any) {
	zap.S().Debugf(format, args...)
}

func Infof(format string, args ...any) {
	zap.S().Infof(format, args...)
}

func Warnf(format string, args ...any) {
	zap.S().Warnf(format, args...)
}

func Warningf(format string, args ...any) {
	zap.S().Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	zap.S().Errorf(format, args...)
}

func Panicf(format string, args ...any) {
	zap.S().Panicf(format, args...)
}

func Fatalf(format string, args ...any) {
	zap.S().Fatalf(format, args...)
}

func Debugln(args ...any) {
	zap.S().Debugln(args...)
}

func Infoln(args ...any) {
	zap.S().Infoln(args...)
}

func Warnln(args ...any) {
	zap.S().Warnln(args...)
}

func Warningln(args ...any) {
	zap.S().Warnln(args...)
}

func Errorln(args ...any) {
	zap.S().Errorln(args...)
}

func Panicln(args ...any) {
	zap.S().Panicln(args...)
}

func Fatalln(args ...any) {
	zap.S().Fatalln(args...)
}

func Printf(format string, args ...any) {
	log.Printf(format, args...)
}

func Println(args ...any) {
	log.Println(args...)
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

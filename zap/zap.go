package zap

import "github.com/Kittengarten/KittenCore/kitten"

// Debug kitten.Logs a message at level Debug on the standard kitten.Logger.
func Debug(args ...interface{}) {
	kitten.Log.Debug(args...)
}

// Info kitten.Logs a message at level Info on the standard kitten.Logger.
func Info(args ...interface{}) {
	kitten.Log.Info(args...)
}

// Warn kitten.Logs a message at level Warn on the standard kitten.Logger.
func Warn(args ...interface{}) {
	kitten.Log.Warn(args...)
}

// Error kitten.Logs a message at level Error on the standard kitten.Logger.
func Error(args ...interface{}) {
	kitten.Log.Error(args...)
}

// Panic kitten.Logs a message at level Panic on the standard kitten.Logger.
func Panic(args ...interface{}) {
	kitten.Log.Panic(args...)
}

// Fatal kitten.Logs a message at level Fatal on the standard kitten.Logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	kitten.Log.Fatal(args...)
}

// Debugf kitten.Logs a message at level Debug on the standard kitten.Logger.
func Debugf(format string, args ...interface{}) {
	kitten.Log.Debugf(format, args...)
}

// Infof kitten.Logs a message at level Info on the standard kitten.Logger.
func Infof(format string, args ...interface{}) {
	kitten.Log.Infof(format, args...)
}

// Warnf kitten.Logs a message at level Warn on the standard kitten.Logger.
func Warnf(format string, args ...interface{}) {
	kitten.Log.Warnf(format, args...)
}

// Errorf kitten.Logs a message at level Error on the standard kitten.Logger.
func Errorf(format string, args ...interface{}) {
	kitten.Log.Errorf(format, args...)
}

// Panicf kitten.Logs a message at level Panic on the standard kitten.Logger.
func Panicf(format string, args ...interface{}) {
	kitten.Log.Panicf(format, args...)
}

// Fatalf kitten.Logs a message at level Fatal on the standard kitten.Logger then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	kitten.Log.Fatalf(format, args...)
}

// Debugln kitten.Logs a message at level Debug on the standard kitten.Logger.
func Debugln(args ...interface{}) {
	kitten.Log.Debugln(args...)
}

// Infoln kitten.Logs a message at level Info on the standard kitten.Logger.
func Infoln(args ...interface{}) {
	kitten.Log.Infoln(args...)
}

// Warnln kitten.Logs a message at level Warn on the standard kitten.Logger.
func Warnln(args ...interface{}) {
	kitten.Log.Warnln(args...)
}

// Errorln kitten.Logs a message at level Error on the standard kitten.Logger.
func Errorln(args ...interface{}) {
	kitten.Log.Errorln(args...)
}

// Panicln kitten.Logs a message at level Panic on the standard kitten.Logger.
func Panicln(args ...interface{}) {
	kitten.Log.Panicln(args...)
}

// Fatalln kitten.Logs a message at level Fatal on the standard kitten.Logger then the process will exit with status set to 1.
func Fatalln(args ...interface{}) {
	kitten.Log.Fatalln(args...)
}

package hook

import "log/slog"

func DebugLog(name string) {
	slog.Debug(`执行函数`, slog.String(`hook`, name))
}

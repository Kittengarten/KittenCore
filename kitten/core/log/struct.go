package log

// Log 是一个日志的配置
type Log struct {
	Level      string // 日志等级
	Path       string // 日志路径
	Crash      string // 崩溃日志路径
	MaxMB      int    `yaml:"max_mb"` // 文件大小限制
	Expire     int    // 日志文件的过期天数，大于该天数前的日志文件会被清理。设置为 -1 可以禁用。
	MaxReserve int    `yaml:"max_reserve"` // 最大保留日志文件数量
}

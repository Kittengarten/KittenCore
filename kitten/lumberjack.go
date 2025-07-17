package kitten

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 返回旋转日志配置
func rotate(c config) io.Writer {
	return &lumberjack.Logger{
		Filename:   c.Log.Path,   // 日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    c.MaxMB,      // 文件大小限制
		MaxBackups: c.MaxReserve, // 最大保留日志文件数量
		MaxAge:     c.Expire,     // 日志文件保留天数
		LocalTime:  true,         // 采用本地时间
		Compress:   false,        // 是否压缩处理
	}
}

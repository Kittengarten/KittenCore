package log

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

// 返回旋转日志配置
func rotate(cfg Log) io.Writer {
	return &lumberjack.Logger{
		Filename:   cfg.Path,       // 日志文件存放目录，如果文件夹不存在会自动创建
		MaxSize:    cfg.MaxMB,      // 文件大小限制
		MaxAge:     cfg.Expire,     // 日志文件保留天数
		MaxBackups: cfg.MaxReserve, // 最大保留日志文件数量
		LocalTime:  true,           // 采用本地时间
		Compress:   false,          // 是否压缩处理
	}
}

package kitten

import "github.com/Kittengarten/KittenCore/kitten/core/io"

type (
	// 来自 Bot 的配置文件的数据集
	config struct {
		NickName      []string        // 昵称
		SelfID        QQ              // Bot 自身 ID
		SuperUsers    []QQ            // 亲妈账号
		CommandPrefix string          // 指令前缀
		Path          io.Path         // 资源文件路径
		WebSocket     WebSocketConfig // WebSocket 配置
		Log           LogConfig       // 日志配置
		WebUI         WebUIConfig     // WebUI 配置
	}

	// WebSocketConfig 是一个 WebSocket 链接的配置
	WebSocketConfig struct {
		URL         string // WebSocket 链接
		AccessToken string // WebSocket 密钥
	}

	// WebUIConfig 是一个 WebUI 的配置
	WebUIConfig struct {
		Host string // WebUI 链接
	}

	// LogConfig 是一个日志的配置
	LogConfig struct {
		Level      string // 日志等级
		Path       string // 日志路径
		Crash      string // 崩溃日志路径
		MaxSize    int    // 文件大小限制，单位 MB
		MaxBackups int    // 最大保留日志文件数量
		Expire     int    // 日志文件的过期天数，大于该天数前的日志文件会被清理。设置为 -1 可以禁用。
	}
)

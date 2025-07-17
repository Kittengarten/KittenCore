package kitten

import "github.com/Kittengarten/KittenCore/kitten/core/fio"

type (
	// 来自 Bot 的配置文件的数据集
	config struct {
		WebSocket     `yaml:"web_socket"` // WebSocket 配置
		CommandPrefix string              `yaml:"command_prefix"` // 指令前缀
		fio.Path                          // 资源文件路径
		WebUI         `yaml:"web_ui"`     // WebUI 配置
		NickName      []string            `yaml:"nick_name"`   // 昵称
		SuperUsers    []QQ                `yaml:"super_users"` // 亲妈账号
		Log                               // 日志配置
		QQ            `yaml:"self_id"`    // Bot 自身 ID
	}

	// WebSocket 是一个 WebSocket 链接的配置
	WebSocket struct {
		URL         string // WebSocket 链接
		AccessToken string `yaml:"access_token"` // WebSocket 密钥
	}

	// WebUI 是一个 WebUI 的配置
	WebUI struct {
		Host string // WebUI 链接
	}

	// Log 是一个日志的配置
	Log struct {
		Level      string // 日志等级
		Path       string // 日志路径
		Crash      string // 崩溃日志路径
		MaxMB      int    `yaml:"max_mb"`      // 文件大小限制
		MaxReserve int    `yaml:"max_reserve"` // 最大保留日志文件数量
		Expire     int    // 日志文件的过期天数，大于该天数前的日志文件会被清理。设置为 -1 可以禁用。
	}
)

package kitten

type (
	KittenConfig struct {
		NickName      []string        // 昵称
		SuperUsers    []int64         // 管理员账号
		CommandPrefix string          // 指令前缀
		WebSocket     WebSocketConfig // WebSocket 配置
		Log           LogConfig       // 日志配置
	}

	WebSocketConfig struct {
		Url         string // WebSocket 链接
		AccessToken string // WebSocket 密钥
	}

	LogConfig struct {
		Level string // 日志等级
		Path  string // 日志路径
	}
)

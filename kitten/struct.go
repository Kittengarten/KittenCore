package kitten

type (
	// Config 是来自 Bot 的配置文件的数据集
	Config struct {
		NickName      []string        // 昵称
		SelfID        int64           // Bot自身ID
		SuperUsers    []int64         // 管理员账号
		CommandPrefix string          // 指令前缀
		WebSocket     WebSocketConfig // WebSocket 配置
		Log           LogConfig
	}

	// WebSocketConfig 是一个 WebSocket 链接的配置
	WebSocketConfig struct {
		URL         string // WebSocket 链接
		AccessToken string // WebSocket 密钥
	}

	// LogConfig 是一个日志的配置
	LogConfig struct {
		Level string // 日志等级
		Path  string // 日志路径
	}

	// Choice 是一个随机项目的抽象接口
	Choice interface {
		GetInformation() string // 该项目的信息
		GetChance() int         // 该项目的权重
	}
)

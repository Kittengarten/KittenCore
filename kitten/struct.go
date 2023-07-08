package kitten

// ImagePath 图片路径
var ImagePath = FilePath(Path(Configs.Path), `image`)

type (
	// Config 是来自 Bot 的配置文件的数据集
	Config struct {
		NickName      []string        `yaml:"nickname"`      // 昵称
		SelfID        int64           `yaml:"selfid"`        // Bot 自身 ID
		SuperUsers    []int64         `yaml:"superusers"`    // 亲妈账号
		CommandPrefix string          `yaml:"commandprefix"` // 指令前缀
		Path          string          `yaml:"path"`          // 资源文件路径
		WebSocket     WebSocketConfig `yaml:"websocket"`     // WebSocket 配置
		Log           LogConfig       `yaml:"log"`           // 日志配置
		WebUI         WebUIConfig     `yaml:"webui"`         // WebUI 配置
	}

	// WebSocketConfig 是一个 WebSocket 链接的配置
	WebSocketConfig struct {
		URL         string // WebSocket 链接
		AccessToken string // WebSocket 密钥
	}

	// // URL 代表 URL 的字符串
	// URL string

	// WebUIConfig 是一个 WebUI 的配置
	WebUIConfig struct {
		URL string // WebUI 链接
	}

	// LogConfig 是一个日志的配置
	LogConfig struct {
		Level  string // 日志等级
		Path   string // 日志路径
		Days   int    // 单段分割文件记录的天数
		Expire int    // 日志文件的过期天数，大于该天数前的日志文件会被清理。设置为 -1 可以禁用。
	}

	// Path 是一个表示文件路径的字符串
	Path string

	// QQ 是一个表示 QQ 的 int64
	QQ int64

	// Choices 是由随机项目的抽象接口组成的数组
	Choices []interface {
		GetInformation() string // 该项目的信息
		GetChance() int         // 该项目的权重
	}

	// Pingstr 是延迟毫秒数对应的语言描述
	Pingstr struct {
		Min    string // 最小延迟
		Avg    string // 平均延迟
		Max    string // 最大延迟
		StdDev string // 延迟抖动
		Loss   string // 丢包率
	}
)

package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
)

type (
	// 来自 Bot 的配置文件的数据集
	config struct {
		Protocol                 // 协议配置
		CommandPrefix   string   `yaml:"command_prefix"` // 指令前缀
		fio.Path                 // 资源文件路径
		WebUI           Server   `yaml:"web_ui"` // WebUI 配置
		PProf           Server   // pprof 配置
		NickName        []string `yaml:"nick_name"`   // 昵称
		SuperUsers      []int64  `yaml:"super_users"` // 亲妈账号
		log.Log                  // 日志配置
		SelfID          int64    `yaml:"self_id"`            // Bot 自身 ID
		AddSpaceAfterAt bool     `yaml:"add_space_after_at"` // 是否在 At 消息后没有空格时自动添加空格
	}

	// Protocol 是一个协议的配置
	Protocol struct {
		Type        string // 协议类型
		URL         string // 链接
		AccessToken string `yaml:"access_token"` // 密钥
		CallerURL   string `yaml:"caller_url"`   // HTTP 调用链接
		CallerToken string `yaml:"caller_token"` // HTTP 调用密钥
	}

	// Server 是一个 HTTP 服务端的监听配置
	Server struct {
		Host   string // HTTP 主机
		Port   uint64 // HTTP 端口（使用 uint64 避免二次转换）
		Enable bool   // 是否启用
	}
)

// Format 实现 fmt.Formatter
func (c config) Format(f fmt.State, _ rune) {
	fmt.Fprint(f, c.String())
}

// String 实现 fmt.Stringer
func (c config) String() string {
	s := new(strings.Builder)
	if err := json.NewEncoder(s).Encode(c); err != nil {
		return err.Error()
	}
	return s.String()
}

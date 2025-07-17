package book

import (
	"errors"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
)

type (
	// Books 多项小说报更项目的数据集组成的切片
	Books []Book
	// Book 小说报更项目的数据集
	Book struct {
		UpdateTime   time.Time   `yaml:",omitempty"` // UpdateTime 上次更新时间
		Platform     string      // Platform 报更平台
		BookID       string      // BookID 报更书号（为了未来兼容性，不使用整数）
		BookName     string      // BookName 报更书名
		Writer       string      // Writer 小说作者
		RecordURL    string      `yaml:",omitempty"` // RecordURL 上次更新链接
		Protagonists []string    `yaml:",omitempty"` // Protagonists 主角
		Users        []kitten.QQ // Users 用户，正数代表 QQ 号，负数代表群号
	}
)

// 报更配置文件错误喵！
const errConfig = `报更配置文件错误喵！`

var (
	ErrLoad      = errors.New(`加载` + errConfig)
	ErrSave      = errors.New(`保存` + errConfig)
	ErrNotConfig = errors.New(`这里没有添加小说报更喵！`)
)

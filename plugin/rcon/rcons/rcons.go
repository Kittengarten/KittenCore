package rcons

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	rcon struct {
		HOST     string // RCON 主机
		Password string // RCON 密码
	}

	item byte
)

const (
	Unknown  item = iota // Unknown 未知
	Host                 // Host 主机
	Password             // Password 密码
)

var sec = regexp.MustCompile(`§.`)

// RCON
func Command(handler *msg.Handler, cp fio.PathRWMutex) message.ID {
	cp.RLock()
	defer cp.RUnlock()
	config, err := fio.LoadWithContext[rcon](handler, cp.Path, fio.Empty)
	if err != nil {
		return handler.SendWithImageFail(`RCON 配置文件错误喵！`, err)
	}
	conn := new(MCConn)
	if err = conn.Open(handler, config.HOST, config.Password); err != nil {
		return handler.SendWithImageFail(`连接 RCON 服务器错误喵！`, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			handler.SendWithImageFail(`关闭 RCON 连接错误喵！`, err)
		}
	}()
	if err = conn.Authenticate(); err != nil {
		return handler.SendWithImageFail(`RCON 密码验证错误喵！`, err)
	}
	resp, err := conn.SendCommand(handler.Args())
	if err != nil {
		return handler.SendWithImageFail(`发送 RCON 命令错误喵！`, err)
	}
	if resp == `` {
		return handler.Quote().At().Text(`命令响应为空喵！`).Send()
	}
	return handler.Quote().AtLf().Text(
		sec.ReplaceAllString(
			strings.TrimRight(
				strings.ReplaceAll(resp, ` ms`, " ms\n"),
				"\n\r",
			),
			``,
		),
	).Send()
}

// 设置 RCON
func Set(handler *msg.Handler, i item, cp fio.PathRWMutex) message.ID {
	s, err := func() (string, error) {
		rm := msg.State[[]string](handler, `regex_matched`)
		if len(rm) == 0 {
			return ``, fmt.Errorf(`设置 RCON 失败：%w`, utils.ErrNoMatch)
		}
		return rm[1], nil
	}()
	if err != nil {
		return handler.SendWithImageFail(err)
	}
	cp.Lock()
	defer cp.Unlock()
	config, err := fio.LoadWithContext[rcon](handler, cp.Path, fio.Empty)
	if err != nil {
		return handler.SendWithImageFail(`RCON 配置文件错误喵！`, err)
	}
	switch i {
	case Host:
		config.HOST = s
	case Password:
		config.Password = s
	}
	if err = fio.SaveWithContext(handler, cp.Path, config); err != nil {
		return handler.SendWithImageFail(`保存 RCON 配置文件错误喵！`, err)
	}
	return handler.Quote().At().Text(`RCON `, &i, `设置成功喵！`).Send()
}

// String 实现 fmt.Stringer
func (i item) String() string {
	return map[item]string{
		Host:     `主机`,
		Password: `密码`,
	}[i]
}

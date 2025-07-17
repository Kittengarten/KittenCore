package rcons

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

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

// 设置类型
var setItem = map[item]string{
	Host:     `主机`,
	Password: `密码`,
}

// RCON
func Command(msgr *kitten.Messager, cp fio.PathRWMutex) message.ID {
	cp.RLock()
	defer cp.RUnlock()
	config, err := fio.Load[rcon](cp.Path, fio.Empty)
	if err != nil {
		return msgr.SendWithImageFail(`RCON 配置文件错误喵！`, err)
	}
	conn := new(MCConn)
	if err = conn.Open(config.HOST, config.Password); err != nil {
		return msgr.SendWithImageFail(`连接 RCON 服务器错误喵！`, err)
	}
	defer conn.Close()
	if err = conn.Authenticate(); err != nil {
		return msgr.SendWithImageFail(`RCON 密码验证错误喵！`, err)
	}
	resp, err := conn.SendCommand(msgr.Args())
	if err != nil {
		return msgr.SendWithImageFail(`发送 RCON 命令错误喵！`, err)
	}
	if resp == `` {
		return msgr.Quote().At().Text(`命令响应为空喵！`).Send()
	}
	resp = strings.ReplaceAll(resp, ` ms`, " ms\n")
	resp = strings.TrimRight(resp, "\n\r")
	return msgr.Quote().AtLf().Text(regexp.MustCompile(`§.`).ReplaceAllString(resp, ``)).Send()
}

// 设置 RCON
func Set(msgr *kitten.Messager, i item, cp fio.PathRWMutex) message.ID {
	s, err := func() (string, error) {
		rm := kitten.State[[]string](msgr, `regex_matched`)
		if len(rm) == 0 {
			return ``, fmt.Errorf(`设置 RCON 失败：%w`, utils.ErrNoMatch)
		}
		return rm[1], nil
	}()
	if err != nil {
		return msgr.SendWithImageFail(err)
	}
	cp.Lock()
	defer cp.Unlock()
	config, err := fio.Load[rcon](cp.Path, fio.Empty)
	if err != nil {
		return msgr.SendWithImageFail(`RCON 配置文件错误喵！`, err)
	}
	switch i {
	case Host:
		config.HOST = s
	case Password:
		config.Password = s
	}
	if err = fio.Save(cp.Path, config); err != nil {
		return msgr.SendWithImageFail(`保存 RCON 配置文件错误喵！`, err)
	}
	return msgr.Quote().At().Text(`RCON `, &i, `设置成功喵！`).Send()
}

// String 实现 fmt.Stringer
func (i item) String() string {
	return setItem[i]
}

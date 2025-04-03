package rcons

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

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
	Host     item = iota // 主机
	Password             // 密码
)

var (
	// 读写锁
	mu sync.RWMutex
	// 设置类型
	setItem = map[item]string{
		Host:     `主机`,
		Password: `密码`,
	}
)

// RCON
func Command(msgr *kitten.Messager, cp fio.Path) message.ID {
	mu.RLock()
	defer mu.RUnlock()
	config, err := fio.Load[rcon](cp, fio.Empty)
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
		return msgr.Reply().At().Text(`命令响应为空喵！`).Send()
	}
	resp = strings.ReplaceAll(resp, ` ms`, " ms\n")
	resp = strings.TrimRight(resp, "\n\r")
	return msgr.Reply().AtLf().Text(regexp.MustCompile(`§.`).ReplaceAllString(resp, ``)).Send()
}

// 设置 RCON
func Set(msgr *kitten.Messager, i item, cp fio.Path) message.ID {
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
	mu.Lock()
	defer mu.Unlock()
	config, err := fio.Load[rcon](cp, fio.Empty)
	if err != nil {
		return msgr.SendWithImageFail(`RCON 配置文件错误喵！`, err)
	}
	switch i {
	case Host:
		config.HOST = s
	case Password:
		config.Password = s
	}
	if err = fio.Save(cp, config); err != nil {
		return msgr.SendWithImageFail(`保存 RCON 配置文件错误喵！`, err)
	}
	return msgr.Reply().At().Text(`RCON `, &i, `设置成功喵！`).Send()
}

// String 实现 fmt.Stringer
func (i item) String() string {
	return setItem[i]
}

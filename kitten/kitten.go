// Package kitten 包含了 KittenCore 以及各插件的核心依赖结构体、方法和函数
package kitten

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"

	log "github.com/sirupsen/logrus"
)

const path Path = `config.yaml` // 配置文件名

var (
	// Configs 来自 Bot 的配置文件
	Configs = LoadMainConfig()
	// Bot 实例
	Bot *zero.Ctx
	// BotSFACGchan 用于传送 Bot 实例的通道
	BotSFACGchan = make(chan *zero.Ctx)
	t            = time.Tick(5 * time.Second) // 每 5 秒尝试获取一次
)

func init() {
	// 向 SFACG 插件传入 Bot 实例
	go func() {
		for nil == Bot {
			select {
			case <-t: // 阻塞协程，收到定时器信号则释放
				log.Trace(`尝试获取 Bot 实例`)
			}
			Bot = zero.GetBot(Configs.SelfID)
		}
		BotSFACGchan <- Bot
	}()
}

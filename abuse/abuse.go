// Package abuse 在线挨骂，如果担心被冒犯到，请勿使用，否则后果自负
package abuse

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName = "挨骂"
	filePath         = "abuse/config.yaml" // 配置文件路径
	imagePath        = "abuse/path.txt"    // 保存图片路径的文件
	randMax          = 100                 // 随机数上限（不包含）
)

var (
	abuseResponses []kitten.Choice
	abuseConfig    []Response
)

func init() {
	go load()
	config := kitten.LoadConfig()
	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s骂我 或 %s挨骂", config.CommandPrefix, config.CommandPrefix),
		"在线挨骂，如果担心被冒犯到，请勿使用，否则后果自负",
	}, "\n")
	// 注册插件
	engine := control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	command := []string{"骂我", "挨骂"}
	engine.OnCommandGroup(command).Handle(func(ctx *zero.Ctx) {
		idx := kitten.Choose(abuseResponses)
		var messageToSend message.MessageSegment
		if abuseConfig[idx].String == "" {
			if abuseConfig[idx].Image == "" {
				log.Warn("获取不到 abuse 信息喵！")
				messageToSend = kitten.TextOf("喵喵不想理你（好感 - %d）", rand.Intn(randMax)+1)
			} else {
				messageToSend = kitten.GetImage(imagePath, abuseConfig[idx].GetInformation())
			}
		} else {
			messageToSend = message.Text(abuseConfig[idx].GetInformation())
		}
		ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
	})
}

// 加载配置
func loadConfig() (config []Response) {
	yaml.Unmarshal(kitten.FileRead(filePath), &config)
	return config
}

// 将配置转换为 []kitten.Choice 接口
func load() {
	yaml.Unmarshal(kitten.FileRead(filePath), &abuseConfig)
	values := make([]kitten.Choice, len(abuseConfig))
	for idx, value := range abuseConfig {
		values[idx] = value
	}
	abuseResponses = values[:len(loadConfig())]
}

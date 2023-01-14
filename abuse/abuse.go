// Package abuse 在线挨骂，如果担心被冒犯到，请勿使用，否则后果自负
package abuse

import (
	"math/rand"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName             = "挨骂"
	brief                        = "在线挨骂"
	filePath         kitten.Path = "abuse/config.yaml" // 配置文件路径
	imagePath        kitten.Path = "abuse/path.txt"    // 保存图片路径的文件
	randMax                      = 100                 // 随机数上限（不包含）
)

var (
	abuseResponses kitten.Choices
	abuseConfig    []Response
)

func init() {
	go load()
	var (
		help = strings.Join([]string{"发送",
			"骂我 或 挨骂",
			"在线挨骂，如果担心被冒犯到，请勿使用，否则后果自负",
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Brief:            brief,
			Help:             help,
		})
	)

	engine.OnFullMatchGroup([]string{"骂我", "挨骂"}).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Minute, 1).LimitByUser).Handle(func(ctx *zero.Ctx) {
		var (
			i             = abuseResponses.Choose()
			messageToSend message.MessageSegment
		)
		if abuseConfig[i].String == "" {
			if abuseConfig[i].Image == "" {
				log.Warnf("获取不到 %s 信息喵！", ReplyServiceName)
				messageToSend = kitten.TextOf("喵喵不想理你（好感 - %d）", rand.Intn(randMax)+1)
			} else {
				messageToSend = imagePath.GetImage(abuseConfig[i].GetInformation())
			}
		} else {
			messageToSend = message.Text(abuseConfig[i].GetInformation())
		}
		ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
	})
}

// 加载配置
func loadConfig() (cf []Response) {
	b, err := filePath.Read()
	if kitten.Check(err) {
		yaml.Unmarshal(b, &cf)
	} else {
		log.Errorf("加载 %s 配置失败喵！", ReplyServiceName)
	}
	return
}

// 将配置转换为 []kitten.Choice 接口
func load() {
	b, err := filePath.Read()
	if kitten.Check(err) {
		yaml.Unmarshal(b, &abuseConfig)
		values := make(kitten.Choices, len(abuseConfig))
		for i, value := range abuseConfig {
			values[i] = value
		}
		abuseResponses = values[:len(loadConfig())]
	}
}

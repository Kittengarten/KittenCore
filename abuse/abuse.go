package abuse

import (
	"io/ioutil"
	"kitten/kitten"
	"os"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	replyServiceName = "Kitten_AbuseBP" // 插件名
)

var abuseResponses []kitten.Choice
var abuseConfig []Response

func init() {
	go Load()
	config := kitten.LoadConfig()
	help := "发送\n" + config.CommandPrefix + "骂我或" + config.CommandPrefix + "挨骂" +
		"\n在线挨骂，如果担心被冒犯到，请勿使用，否则后果自负"
	// 注册插件
	engine := control.Register(replyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		PrivateDataFolder: "image",
		Help:              help,
	})

	command := []string{"骂我", "挨骂"}
	engine.OnCommandGroup(command).Handle(func(ctx *zero.Ctx) {
		idx := kitten.Choose(abuseResponses)
		var messageToSend message.MessageSegment
		if abuseConfig[idx].String == "" {
			if abuseConfig[idx].Image == "" {
				log.Warn("获取不到abuse信息喵！")
				messageToSend = message.Text("喵喵不想理你（好感-1）")
			} else {
				messageToSend = message.Image(LoadImagePath() + abuseConfig[idx].GetInformation())
			}
		} else {
			messageToSend = message.Text(abuseConfig[idx].GetInformation())
		}
		ctx.SendChain(message.At(ctx.Event.UserID), messageToSend)
	})
}

// 加载图片路径
func LoadImagePath() string {
	path := "abuse/path.txt"
	res, err := os.Open(path)
	if !kitten.Check(err) {
		log.Warn("打开文件" + path + "失败了喵！")
	} else {
		defer res.Close()
	}
	data, _ := ioutil.ReadAll(res)
	return string(data)
}

// 加载配置
func LoadConfig() (config []Response) {
	yaml.Unmarshal(kitten.FileRead("abuse/config.yaml"), &config)
	return config
}

// 加载配置
func Load() {
	yaml.Unmarshal(kitten.FileRead("abuse/config.yaml"), &abuseConfig)
	values := make([]kitten.Choice, len(abuseConfig))
	for idx, value := range abuseConfig {
		values[idx] = value
	}
	abuseResponses = values[:len(LoadConfig())]
}

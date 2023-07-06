package sfacg

import (
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/FloatTech/zbputils/control"
	"github.com/Kittengarten/KittenCore/kitten"
	log "github.com/sirupsen/logrus"
)

// 加载配置
func loadConfig(configFile kitten.Path) (c Config) {
	if err := yaml.Unmarshal(configFile.Read(), &c); !kitten.Check(err) {
		log.Errorf("%s 载入配置文件出现错误喵！\n%v", ReplyServiceName, err)
	}
	return
}

// 保存配置，成功则返回 True
func saveConfig(c Config, e *control.Engine) (ok bool) {
	data, err := yaml.Marshal(c)
	kitten.FilePath(kitten.Path(e.DataFolder()), configFile).Write(data)
	ok = kitten.Check(err)
	if !ok {
		log.Errorf("配置文件写入错误喵！\n%v", err)
		reciver := kitten.Configs.SuperUsers[0]
		if kitten.Check(kitten.Bot) {
			kitten.Bot.SendPrivateMessage(reciver, `追更配置文件写入错误，请检查日志喵！`)
		}
	}
	return
}

// 判断字符串是否为整数（可用于判断是书号还是关键词）
func isInt(str string) bool {
	match, err := regexp.MatchString(`^[0-9]+$`, str)
	if kitten.Check(err) {
		return match
	}
	log.Error(`判断字符串是否为整数时，字符串正则匹配错误喵！`)
	return false
}

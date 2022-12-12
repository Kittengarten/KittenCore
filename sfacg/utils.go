package sfacg

import (
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/FloatTech/zbputils/control"
	"github.com/Kittengarten/KittenCore/kitten"
	log "github.com/sirupsen/logrus"
)

// 加载配置
func loadConfig(e *control.Engine) (cf Config, err error) {
	f, err := kitten.FileRead(e.DataFolder() + configFile)
	if !kitten.Check(err) {
		return
	}
	yaml.Unmarshal(f, &cf)
	return
}

// 保存配置，成功则返回 True
func saveConfig(cf Config, e *control.Engine) (ok bool) {
	var (
		data, err1 = yaml.Marshal(cf)
		err2       = kitten.FileWrite(e.DataFolder()+configFile, data)
	)
	ok = kitten.Check(err1) && kitten.Check(err2)
	if !ok {
		log.Errorf("配置文件写入错误喵！\n%v\n%v", err1, err2)
		reciver := kitten.Configs.SuperUsers[0]
		if kitten.Bot != nil {
			kitten.Bot.SendPrivateMessage(reciver, "追更配置文件写入错误，请检查日志喵！")
		}
	}
	return
}

// 判断字符串是否为整数（可用于判断是书号还是搜索关键词）
func isInt(str string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", str)
	return match
}

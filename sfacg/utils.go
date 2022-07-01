package sfacg

import (
	"kitten/kitten"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

func LoadConfig() (config Config) {
	var sf SFAPI
	yaml.Unmarshal(kitten.FileRead("sfacg/config.yaml"), &config)
	for idx := range config {
		chapterUrl, chk := sf.FindChapterUrl(config[idx].BookId)
		if !chk {
			log.Warn(chapterUrl)
		}
		config[idx].RecordUrl = chapterUrl
		config[idx].Updatetime = sf.FindChapterUpdateTime(config[idx].BookId)
	}
	return config
} //加载配置

func IsInt(str string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", str)
	return match
} //判断字符串是否为整数（可用于判断是书号还是搜索关键词）

func GetMidText(pre string, suf string, str string) string {
	n := strings.Index(str, pre)
	if n == -1 {
		n = 0
	} else {
		n = n + len(pre)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, suf)
	if m == -1 {
		m = len(str)
	}
	return string([]byte(str)[:m])
} //获取中间字符串

package sfacg

import (
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/Kittengarten/KittenCore/kitten"
)

// 加载配置
func loadConfig() (cf Config) {
	yaml.Unmarshal(kitten.FileRead("sfacg/config.yaml"), &cf)
	return cf
}

// 判断字符串是否为整数（可用于判断是书号还是搜索关键词）
func isInt(str string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", str)
	return match
}

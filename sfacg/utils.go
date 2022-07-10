package sfacg

import (
	"regexp"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten"

	"gopkg.in/yaml.v3"
)

// 加载配置
func LoadConfig() (cf Config) {
	yaml.Unmarshal(kitten.FileRead("sfacg/config.yaml"), &cf)
	return cf
}

// 判断字符串是否为整数（可用于判断是书号还是搜索关键词）
func IsInt(str string) bool {
	match, _ := regexp.MatchString("^[0-9]+$", str)
	return match
}

// 获取中间字符串
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
}

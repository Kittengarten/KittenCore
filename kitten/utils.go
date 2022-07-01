package kitten

import (
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

func Atoi(str string) int {
	num, _ := strconv.Atoi(str)
	return num
} //字符串转换为数字

func FileRead(path string) []byte {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn("打开文件" + path + "失败了喵！")
	} else {
		defer res.Close()
	}
	data, _ := ioutil.ReadAll(res)
	return data
} //YAML文件读取

func LoadConfig() (config KittenConfig) {
	err := yaml.Unmarshal(FileRead("config.yaml"), &config)
	if !Check(err) {
		log.Fatal("打开配置文件失败了喵！", err)
		return
	}
	return config
} //加载配置

func Check(err interface{}) bool {
	if err != nil {
		return false
	} else {
		return true
	}
} //处理错误

package kitten

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
)

// 字符串转换为数字
func Atoi(str string) int {
	num, _ := strconv.Atoi(str)
	return num
}

// 文件读取
func FileRead(path string) []byte {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn("读取文件" + path + "失败了喵！")
	} else {
		defer res.Close()
	}
	data, _ := ioutil.ReadAll(res)
	return data
}

// 文件写入
func FileWrite(path string, data []byte) (err error) {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn("写入文件" + path + "失败了喵！")
	} else {
		defer res.Close()
	}
	err = ioutil.WriteFile(path, data, 0666)
	return err
}

// 加载配置
func LoadConfig() (config KittenConfig) {
	err := yaml.Unmarshal(FileRead("config.yaml"), &config)
	if !Check(err) {
		log.Fatal("打开配置文件失败了喵！", err)
		return
	}
	return config
}

// 处理错误
func Check(err interface{}) bool {
	if err != nil {
		return false
	} else {
		return true
	}
}

// 按权重抽取一个项目的idx
func Choose(choices []Choice) int {
	choiceAll := 0
	for idx := range choices {
		choiceAll += choices[idx].GetChance()
	}
	choiceNum := rand.Intn(choiceAll)
	for idx := range choices {
		choiceNum -= choices[idx].GetChance()
		if choiceNum < 0 {
			return idx
		}
	}
	return len(choices) - 1
}

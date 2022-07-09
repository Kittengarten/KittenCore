package kitten

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

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
		log.Warn(fmt.Sprintf("读取文件%s失败了喵！", path))
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
		log.Warn(fmt.Sprintf("写入文件%s失败了喵！", path))
	} else {
		defer res.Close()
	}
	err = ioutil.WriteFile(path, data, 0666)
	return err
}

// 加载配置
func LoadConfig() (config KittenConfig) {
	const path = "config.yaml"
	err := yaml.Unmarshal(FileRead(path), &config)
	if !Check(err) {
		log.Fatal(fmt.Sprintf("打开%s失败了喵！", path), err)
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

// 按权重抽取一个项目的idx，有可能返回-1（这种情况代表项目列表为空，需要处理以免报错）
func Choose(choices []Choice) int {
	choiceAll := 0
	choiceNum := 0
	for idx := range choices {
		choiceAll += choices[idx].GetChance()
	}
	if choiceAll > 0 {
		choiceNum = rand.Intn(choiceAll)
	}
	for idx := range choices {
		choiceNum -= choices[idx].GetChance()
		if choiceNum < 0 {
			return idx
		}
	}
	return len(choices) - 1
}

// 判断两个时间是否是同一天
func IsSameDate(t1 time.Time, t2 time.Time) bool {
	year1, month1, day1 := t1.Date()
	year2, month2, day2 := t2.Date()
	if year1 == year2 && month1 == month2 && day1 == day2 {
		return true
	} else {
		return false
	}
}

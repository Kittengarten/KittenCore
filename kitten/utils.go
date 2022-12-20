package kitten

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenAnno/wta"

	log "github.com/sirupsen/logrus"
)

// Atoi 将字符串转换为数字
func Atoi(str string) (num int) {
	num, _ = strconv.Atoi(str)
	return
}

// FileReadDirect 文件读取
func FileReadDirect(path string) (data []byte) {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn(fmt.Sprintf("读取文件 %s 失败了喵！", path))
	} else {
		defer res.Close()
	}
	data, _ = ioutil.ReadAll(res)
	return
}

// FileRead 文件读取
func FileRead(path string) (data []byte, err error) {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn(fmt.Sprintf("读取文件 %s 失败了喵！", path))
	} else {
		defer res.Close()
	}
	data, _ = ioutil.ReadAll(res)
	return
}

// FileWrite 文件写入
func FileWrite(path string, data []byte) (err error) {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn(fmt.Sprintf("写入文件 %s 失败了喵！", path))
	} else {
		defer res.Close()
	}
	err = ioutil.WriteFile(path, data, 0666)
	return
}

// LoadConfig 加载配置
func LoadConfig() (config Config) {
	if err := yaml.Unmarshal(FileReadDirect(path), &config); !Check(err) {
		log.Fatal(fmt.Sprintf("打开 %s 失败了喵！", path), err)
		return
	}
	return
}

// PathExists 判断文件是否存在，path 为要判断的文件路径，不确定存在的情况下报错
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	// 当为空文件或文件夹存在
	if Check(err) {
		return true, nil
	}
	// os.IsNotExist(err)为true，文件或文件夹不存在
	if os.IsNotExist(err) {
		return false, nil
	}
	// 其它类型，不确定是否存在
	return false, err
}

// （私有）判断路径是否文件夹
func isDir(path string) bool {
	s, err := os.Stat(path)
	if !Check(err) {
		return Check(err)
	}
	return s.IsDir()
}

// （私有）加载图片路径
func loadImagePath(path string) string {
	res, err := os.Open(path)
	if !Check(err) {
		log.Warn(fmt.Sprintf("打开文件 %s 失败了喵！", path))
	} else {
		defer res.Close()
	}
	data, _ := ioutil.ReadAll(res)
	return string(data)
}

// GetImage 加载图片，path 参数可以是保存路径的文件，也可以是路径本身（绝对路径）
func GetImage(path, name string) message.MessageSegment {
	if isDir(path) {
		return message.Image(path + name)
	}
	return message.Image(loadImagePath(path) + name)
}

// Check 处理错误，没有错误则返回 True
func Check(err interface{}) bool {
	if err != nil {
		return false
	}
	return true
}

// Choose 按权重抽取一个项目的序号 i，有可能返回-1（这种情况代表项目列表为空，需要处理以免报错）
func Choose(choices []Choice) int {
	var choiceAll, choiceNum = 0, 0
	for i := range choices {
		choiceAll += choices[i].GetChance()
	}
	if 0 < choiceAll {
		choiceNum = rand.Intn(choiceAll)
	}
	for i := range choices {
		if choiceNum -= choices[i].GetChance(); 0 > choiceNum {
			return i
		}
	}
	return len(choices) - 1
}

// IsSameDate 判断两个时间是否是同一天
func IsSameDate(t1 time.Time, t2 time.Time) bool {
	var (
		year1, month1, day1 = t1.Date()
		year2, month2, day2 = t2.Date()
	)
	if year1 == year2 && month1 == month2 && day1 == day2 {
		return true
	}
	return false
}

// GetMidText 获取中间字符串，pre 为获取字符串的前缀（不包含），suf 为获取字符串的后缀（不包含），str 为整个字符串
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

// TextOf 格式化构建 message.Text 文本，格式同 fmt.Sprintf
func TextOf(format string, a ...any) message.MessageSegment {
	return message.Text(fmt.Sprintf(format, a...))
}

// GetTitle 从 UID 获取【头衔】
func GetTitle(ctx zero.Ctx, uid int64) (title string) {
	gmi := ctx.GetGroupMemberInfo(ctx.Event.GroupID, uid, true)
	if titleStr := gjson.Get(gmi.Raw, "title").Str; titleStr == "" {
		title = titleStr
	} else {
		title = fmt.Sprintf("【%s】", gjson.Get(gmi.Raw, "title").Str)
	}
	return
}

// GetWTAAnno 获取世界树纪元的完整字符串和额外信息
func GetWTAAnno() (str string, flower string, elemental string, imagery string, err error) {
	anno, err := wta.GetAnno()
	str = anno.YearStr + anno.MonthStr + anno.DayStr
	str = fmt.Sprintf("%s　%d:%0*d:%0*d", str, anno.Hour, 2, anno.Minute, 2, anno.Second)
	flower, elemental, imagery = anno.Flower, anno.Elemental, anno.Imagery
	return
}

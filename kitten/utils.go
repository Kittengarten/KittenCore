package kitten

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ping/ping"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenAnno/wta"

	log "github.com/sirupsen/logrus"
)

// Int 将字符串转换为数字
func (str IntString) Int() (num int) {
	num, _ = strconv.Atoi(string(str))
	return
}

// GetLogLevel 从日志配置获取日志等级
func (lc LogConfig) GetLogLevel() log.Level {
	switch lc.Level {
	case log.PanicLevel.String():
		return log.PanicLevel
	case log.FatalLevel.String():
		return log.FatalLevel
	case log.ErrorLevel.String():
		return log.ErrorLevel
	case log.WarnLevel.String():
		return log.WarnLevel
	case log.InfoLevel.String():
		return log.InfoLevel
	case log.DebugLevel.String():
		return log.DebugLevel
	case log.TraceLevel.String():
		return log.TraceLevel
	default:
		return log.WarnLevel
	}
}

// Read 文件读取
func (path Path) Read() (data []byte, err error) {
	res, err := os.Open(string(path))
	if !Check(err) {
		log.Warnf("读取文件 %s 失败了喵！", path)
	} else {
		defer res.Close()
	}
	data, err = io.ReadAll(res)
	if Check(err) {
		return
	}
	log.Warnf("打开文件 %s 失败了喵！\n%v", path, err)
	return
}

// Write 文件写入
func (path Path) Write(data []byte) (err error) {
	res, err := os.Open(string(path))
	if !Check(err) {
		log.Warnf("写入文件 %s 失败了喵！", path)
	} else {
		defer res.Close()
	}
	err = os.WriteFile(string(path), data, 0666)
	return
}

// Exists 判断文件是否存在，不确定存在的情况下报错
func (path Path) Exists() (bool, error) {
	_, err := os.Stat(string(path))
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
func (path Path) isDir() bool {
	s, err := os.Stat(string(path))
	if !Check(err) {
		return Check(err)
	}
	return s.IsDir()
}

// （私有）加载文件中保存的路径
func (path Path) loadPath() Path {
	res, err := os.Open(string(path))
	if Check(err) {
		defer res.Close()
	} else {
		log.Warnf("打开文件 %s 失败了喵！", path)
	}
	data, _ := io.ReadAll(res)
	return Path(data)
}

// GetImage 从保存图片路径的文件，或图片的绝对路径加载图片
func (path Path) GetImage(name string) message.MessageSegment {
	if path.isDir() {
		return message.Image(string(path) + name)
	}
	return message.Image(string(path.loadPath()) + name)
}

// Check 处理错误，没有错误则返回 True
func Check(err interface{}) bool {
	if err != nil {
		return false
	}
	return true
}

// Choose 按权重抽取一个项目的序号 i，有可能返回-1（这种情况代表项目列表为空，需要处理以免报错）
func (cs Choices) Choose() int {
	var cAll, cNum = 0, 0
	for i := range cs {
		cAll += cs[i].GetChance()
	}
	if 0 < cAll {
		cNum = rand.Intn(cAll)
	}
	for i := range cs {
		if cNum -= cs[i].GetChance(); 0 > cNum {
			return i
		}
	}
	return len(cs) - 1
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

// GetTitle 从 QQ 获取【头衔】
func (u QQ) GetTitle(ctx zero.Ctx) (title string) {
	gmi := ctx.GetGroupMemberInfo(ctx.Event.GroupID, int64(u), true)
	if titleStr := gjson.Get(gmi.Raw, "title").Str; titleStr == "" {
		title = titleStr
	} else {
		title = fmt.Sprintf("【%s】", gjson.Get(gmi.Raw, "title").Str)
	}
	return
}

// LoadConfig 加载配置
func LoadConfig() (config Config) {
	d, err1 := path.Read()
	if !Check(err1) {
		log.Fatalf("加载配置失败喵！\n%v", err1)
	} else if err2 := yaml.Unmarshal(d, &config); !Check(err2) {
		log.Fatalf("打开 %s 失败了喵！\n%v", path, err2)
		return
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

// CheckServer 检查连接状况，错误则返回 -1，正常则返回延迟的毫秒数
func (s URL) CheckServer() int64 {
	s1 := GetMidText("//", ":", string(s))
	log.Tracef("正在 Ping %s 喵……", s1)
	pinger, err := ping.NewPinger(s1)
	if Check(err) {
		pinger.Count = 1             // 检测 1 次
		pinger.Timeout = time.Second // 超时为 1 秒
		pinger.SetPrivileged(true)
		pinger.Run() // 直到完成之前，阻塞
		return pinger.Statistics().AvgRtt.Milliseconds()
	}
	log.Warnf("Ping 出现错误了喵！\n%v", err)
	return -1
}

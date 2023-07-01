package kitten

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/control"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenAnno/wta"

	log "github.com/sirupsen/logrus"
)

// Int 将字符串转换为数字，转换失败则返回 0（建议不要用于转换 0）
func (str IntString) Int() int {
	if num, err := strconv.Atoi(string(str)); Check(err) {
		return num
	}
	return 0
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
		log.Warnf(`读取文件 %s 失败了喵！`, path)
	} else {
		defer res.Close()
	}
	data, err = io.ReadAll(res)
	if Check(err) {
		return
	}
	log.Warnf(`打开文件 %s 失败了喵！\n%v`, path, err)
	return
}

// Write 文件写入
func (path Path) Write(data []byte) (err error) {
	res, err := os.Open(string(path))
	if !Check(err) {
		log.Warnf(`写入文件 %s 失败了喵！`, path)
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
	// os.IsNotExist(err)为 true，文件或文件夹不存在
	if os.IsNotExist(err) {
		return false, nil
	}
	// 其它类型，不确定是否存在
	return false, err
}

// （私有）判断路径是否文件夹
func (path Path) isDir() bool {
	s, err := os.Stat(fmt.Sprintf("%s", bytes.TrimPrefix([]byte(string(path)), []byte(`file://`))))
	if !Check(err) {
		return false
	}
	return s.IsDir()
}

// LoadPath 加载文件中保存的路径
func (path Path) LoadPath() Path {
	res, err := os.Open(string(path))
	if Check(err) {
		defer res.Close()
	} else {
		log.Warnf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	data, err := io.ReadAll(res)
	if !Check(err) {
		log.Warnf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	return Path(data)
}

// GetImage 从保存图片路径的文件，或图片的绝对路径加载图片
func (path Path) GetImage(name Path) message.MessageSegment {
	if path.isDir() {
		return message.Image(string(path + name))
	}
	return message.Image(string(path.LoadPath() + name))
}

// InitFile 初始化文本文件
func InitFile(name Path, text string) (err error) {
	e, _ := name.Exists()
	if !e {
		err = name.Write([]byte(text))
	}
	return
}

// LoadMainConfig 加载主配置
func LoadMainConfig() (config Config) {
	d, err1 := path.Read()
	if !Check(err1) {
		log.Fatalf("加载配置失败喵！\n%v", err1)
	} else if err2 := yaml.Unmarshal(d, &config); !Check(err2) {
		log.Fatalf("打开 %s 失败了喵！\n%v", path, err2)
		return
	}
	return
}

// LoadConfig 加载配置
func LoadConfig(e *control.Engine, configFile Path, ReplyServiceName string) (c any, err error) {
	if d, err := (Path(e.DataFolder()) + configFile).Read(); Check(err) {
		yaml.Unmarshal(d, &c)
	} else {
		log.Fatalf("%s 配置文件加载失败喵！\n%v", ReplyServiceName, err)
	}
	return
}

// Check 处理错误，没有错误则返回 True
func Check(err any) bool {
	if err != nil {
		return false
	}
	return true
}

// Choose 按权重抽取一个项目的序号 i，有可能返回 -1（这种情况代表项目列表为空，需要处理以免报错）
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
func (u QQ) GetTitle(ctx *zero.Ctx) (title string) {
	gmi := ctx.GetGroupMemberInfo(ctx.Event.GroupID, int64(u), true)
	if titleStr := gjson.Get(gmi.Raw, `title`).Str; titleStr == `` {
		title = titleStr
	} else {
		title = fmt.Sprintf(`【%s】`, gjson.Get(gmi.Raw, `title`).Str)
	}
	return
}

// GetWTAAnno 获取世界树纪元的完整字符串和额外信息
func GetWTAAnno() (str string, flower string, elemental string, imagery string, err error) {
	anno, err := wta.GetAnno()
	str = anno.YearStr + anno.MonthStr + anno.DayStr
	str = fmt.Sprintf(`%s　%d:%0*d:%0*d`, str, anno.Hour, 2, anno.Minute, 2, anno.Second)
	flower, elemental, imagery = anno.Flower, anno.Elemental, anno.Imagery
	return
}

// CheckServer 检查连接状况，错误则返回 -1，正常则返回延迟的毫秒数
func (s URL) CheckServer() *probing.Statistics {
	s1 := GetMidText(`//`, `:`, string(s))
	log.Tracef(`正在 Ping %s 喵……`, s1)
	pinger, err := probing.NewPinger(s1)
	if Check(err) {
		pinger.Timeout = 120 * time.Second // 超时为 120 秒
		pinger.Count = 100                 // 检测 100 次
		pinger.OnRecv = func(pkt *probing.Packet) {
			log.Tracef("%d 字节来自 %s：icmp 顺序：%d 延迟：%v\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		}
		pinger.OnDuplicateRecv = func(pkt *probing.Packet) {
			log.Tracef("%d 字节来自 %s：icmp 顺序：%d 延迟：%v TTL：%v (DUP!)\n",
				pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.TTL)
		}
		pinger.OnFinish = func(stats *probing.Statistics) {
			log.Tracef("\n--- %s Ping 统计 ---\n", stats.Addr)
			log.Tracef("发出了 %d 个包，接收了 %d 个包，丢包率 %v%%\n",
				stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
			log.Tracef("延迟 最小值 / 平均值 / 最大值 / 抖动 毫秒数 = %v / %v / %v / %v \n",
				stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		}
		pinger.Run() // 直到完成之前，阻塞
		return pinger.Statistics()
	}
	log.Warnf("Ping 出现错误了喵！\n%v", err)
	return nil
}

// CheckPing 检查延迟，返回延迟毫秒数对应的语言描述
func CheckPing(p *probing.Statistics) (ps Pingstr) {
	if 0 > p.MinRtt {
		ps.Min = `连接超时喵！`
		return
	} else if time.Microsecond > p.MinRtt {
		ps.Min = `最小延迟：< 1 μs`
	} else {
		ps.Min = fmt.Sprintf(`最小延迟：%v`, p.MinRtt)
	}
	if 0 > p.AvgRtt {
		ps.Avg = `连接超时喵！`
		return
	} else if time.Microsecond > p.AvgRtt {
		ps.Avg = `平均延迟：< 1 μs`
	} else {
		ps.Avg = fmt.Sprintf(`平均延迟：%v`, p.AvgRtt)
	}
	if 0 > p.MaxRtt {
		ps.Max = `连接超时喵！`
	} else if time.Microsecond > p.MaxRtt {
		ps.Max = `最大延迟：< 1 μs`
	} else {
		ps.Max = fmt.Sprintf(`最大延迟：%v`, p.MaxRtt)
	}
	ps.StdDev = fmt.Sprintf(`延迟抖动：%v`, p.StdDevRtt)
	ps.Loss = fmt.Sprintf(`丢包率：%.0f%%`, p.PacketLoss)
	return
}

// DoNotKnow 喵喵不知道哦
func DoNotKnow(ctx *zero.Ctx) {
	ctx.Send(fmt.Sprintf(`%s不知道哦`, zero.BotConfig.NickName[0]))
}

// 获取信息
func (u QQ) getInfo(ctx *zero.Ctx) gjson.Result {
	return ctx.GetStrangerInfo(int64(u), true)
}

// IsAdult 是成年人
func (u QQ) IsAdult(ctx *zero.Ctx) bool {
	if age := gjson.Get(u.getInfo(ctx).Raw, `age`).Int(); 18 < age {
		return true
	}
	return false
}

// IsFemale 是女性
func (u QQ) IsFemale(ctx *zero.Ctx) bool {
	if sex := gjson.Get(u.getInfo(ctx).Raw, `sex`).String(); `female` == sex {
		return true
	}
	return false
}

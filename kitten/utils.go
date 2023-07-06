package kitten

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// FilePath 文件路径构建
func FilePath(elem ...Path) Path {
	var s = make([]string, len(elem))
	for k, v := range elem {
		s[k] = string(v)
	}
	return Path(filepath.Join([]string(s)...))
}

// Read 文件读取
func (path Path) Read() (data []byte) {
	res, err := os.Open(string(path))
	if Check(err) {
		defer res.Close()
	} else {
		log.Errorf("读取：文件 %s 无法读取或不存在喵！\n%v", path, err)
	}
	data, err = io.ReadAll(res)
	if !Check(err) {
		log.Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	return
}

// Write 文件写入，如文件不存在会尝试新建
func (path Path) Write(data []byte) {
	e, err := path.Exists()
	if !e {
		// 如果文件或文件夹不存在，或不确定是否存在
		if Check(err) {
			// 如果文件不存在，新建该文件所在的文件夹；如果文件夹不存在，新建该文件夹本身
			os.MkdirAll(filepath.Dir(string(path)), os.ModeDir)
		} else {
			// 文件或文件夹不确定是否存在
			log.Warnf("不确定 %s 存在喵！\n%v", path, err)
		}
	}
	res, err := os.Open(string(path))
	if Check(err) {
		defer res.Close()
	} else {
		log.Infof("写入：文件 %s 无法读取或不存在喵，尝试新建。\n%v", path, err)
	}
	err = os.WriteFile(string(path), data, 0666)
	if !Check(err) {
		log.Errorf("写入文件 %s 失败了喵！\n%v", path, err)
	}
	return
}

// Exists 判断文件是否存在，不确定存在的情况下报错
func (path Path) Exists() (bool, error) {
	_, err := os.Stat(string(path))
	// 当 err 为空，文件或文件夹存在
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
	s, err := os.Stat(string(path))
	if Check(err) {
		return s.IsDir()
	}
	log.Errorf("识别 %s 失败了喵！\n%v", path, err)
	return false
}

// LoadPath 加载文件中保存的相对路径或绝对路径
func (path Path) LoadPath() Path {
	res, err := os.Open(string(path))
	if Check(err) {
		defer res.Close()
	} else {
		log.Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	data, err := io.ReadAll(res)
	if !Check(err) {
		log.Errorf("打开文件 %s 失败了喵！\n%v", path, err)
	}
	if filepath.IsAbs(string(data)) {
		return Path(`file://`) + FilePath(Path(data))
	}
	return FilePath(Path(data))
}

// GetImage 从图片的相对路径或绝对路径，或相对路径或绝对路径文件中保存的相对路径或绝对路径加载图片
func (path Path) GetImage(name Path) message.MessageSegment {
	if filepath.IsAbs(string(FilePath(path))) {
		if path.isDir() {
			return message.Image(`file://` + string(FilePath(path, name)))
		}
		return message.Image(`file://` + string(FilePath(path.LoadPath(), name)))
	}
	if path.isDir() {
		return message.Image(string(FilePath(path, name)))
	}
	return message.Image(string(FilePath(path.LoadPath(), name)))
}

// InitFile 初始化文本文件
func InitFile(name Path, text string) {
	e, err := name.Exists()
	if Check(err) {
		log.Warnf("不确定 %s 存在喵！\n%v", path, err)
		return
	}
	// 如果文件不存在，初始化
	if !e {
		name.Write([]byte(text))
	}
	return
}

// LoadMainConfig 加载主配置
func LoadMainConfig() (config Config) {
	if err := yaml.Unmarshal(path.Read(), &config); !Check(err) {
		log.Fatalf("打开 %s 失败了喵！\n%v", path, err)
	}
	return
}

// Check 处理错误，没有错误则返回 true
func Check(err any) bool {
	if nil == err {
		return true
	}
	return false
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
	if -1 == n {
		n = 0
	} else {
		n = n + len(pre)
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, suf)
	if -1 == m {
		m = len(str)
	}
	return string([]byte(str)[:m])
}

// TextOf 格式化构建 message.Text 文本，格式同 fmt.Sprintf
func TextOf(format string, a ...any) message.MessageSegment {
	return message.Text(fmt.Sprintf(format, a...))
}

// SendTextOf 发送格式化文本，lf 控制群聊的 @ 后是否换行
func SendTextOf(ctx *zero.Ctx, lf bool, format string, a ...any) {
	switch ctx.Event.DetailType {
	case `private`:
		ctx.Send(TextOf(format, a...))
	case `group`, `guild`:
		if lf {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n"), TextOf(format, a...))
			return
		}
		ctx.SendChain(message.At(ctx.Event.UserID), TextOf(format, a...))
	}
}

// SendMessage 发送消息，lf 控制群聊的 @ 后是否换行
func SendMessage(ctx *zero.Ctx, lf bool, m ...message.MessageSegment) {
	switch ctx.Event.DetailType {
	case `private`:
		ctx.Send(m)
	case `group`, `guild`:
		var n []message.MessageSegment
		if lf {
			ctx.SendChain(append(append(append(n, message.At(ctx.Event.UserID)), message.Text("\n")), m...)...)
			return
		}
		ctx.SendChain(append(append(n, message.At(ctx.Event.UserID)), m...)...)
	}
}

// DoNotKnow 喵喵不知道哦
func DoNotKnow(ctx *zero.Ctx) {
	SendMessage(ctx, false, ImagePath.GetImage(`哈——？.png`), TextOf(`%s不知道哦`, zero.BotConfig.NickName[0]))
}

// GetTitle 从 QQ 获取【头衔】
func (u QQ) GetTitle(ctx *zero.Ctx) (title string) {
	gmi := ctx.GetGroupMemberInfo(ctx.Event.GroupID, int64(u), true)
	if titleStr := gjson.Get(gmi.Raw, `title`).Str; `` == titleStr {
		title = titleStr
		return
	}
	title = fmt.Sprintf(`【%s】`, gjson.Get(gmi.Raw, `title`).Str)
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
	if 0 >= p.MinRtt {
		ps.Min = `连接超时喵！`
		return
	} else if time.Microsecond > p.MinRtt {
		ps.Min = `最小延迟：< 1 μs`
	} else {
		ps.Min = fmt.Sprintf(`最小延迟：%v`, p.MinRtt)
	}
	if 0 >= p.AvgRtt {
		ps.Avg = `连接超时喵！`
		return
	} else if time.Microsecond > p.AvgRtt {
		ps.Avg = `平均延迟：< 1 μs`
	} else {
		ps.Avg = fmt.Sprintf(`平均延迟：%v`, p.AvgRtt)
	}
	if 0 >= p.MaxRtt {
		ps.Max = `最大延迟：连接超时喵！`
	} else if time.Microsecond > p.MaxRtt {
		ps.Max = `最大延迟：< 1 μs`
	} else {
		ps.Max = fmt.Sprintf(`最大延迟：%v`, p.MaxRtt)
	}
	ps.StdDev = fmt.Sprintf(`延迟抖动：%v`, p.StdDevRtt)
	ps.Loss = fmt.Sprintf(`丢包率：%.0f%%`, p.PacketLoss)
	return
}

// （私有）获取信息
func (u QQ) getInfo(ctx *zero.Ctx) gjson.Result {
	return ctx.GetStrangerInfo(int64(u), true)
}

// IsAdult 是成年人
func (u QQ) IsAdult(ctx *zero.Ctx) bool {
	if age := gjson.Get(u.getInfo(ctx).Raw, `age`).Int(); 18 <= age {
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

// GenerateRandomNumber 生成 count 个 [start, end) 范围的不重复的随机数
func GenerateRandomNumber(start int, end int, count int) []int {
	// 范围检查
	if end < start || (end-start) < count {
		return nil
	}
	var (
		// 存放结果的集合（不重复）
		set = make(map[int]bool)
		// 存放结果的数组
		nums []int = make([]int, count)
		// 数组下标
		i int
	)
	// 随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(set) < count {
		// 生成随机数
		set[r.Intn(end-start)+start] = false
	}
	// 集合转换为数组
	for k := range set {
		nums[i] = k
		i++
	}
	return nums
}

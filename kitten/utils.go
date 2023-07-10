package kitten

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/Kittengarten/KittenAnno/wta"
)

/*
Check 处理错误，没有错误则返回 true

也可用于检查传入参数是否存在，如果不存在则返回 true
*/
func Check(err any) bool {
	if nil == err {
		return true
	}
	return false
}

// Choose 按权重抽取一个项目的序号
func (c Choices) Choose() (result int, err error) {
	var (
		a int // 总权重
		n int // 抽取的权重
	)
	// 计算总权重
	for i := range c {
		a += c[i].GetChance()
	}
	// 抽取权重
	if 0 < a {
		n = rand.Intn(a)
	}
	// 计算抽取的权重所在的序号
	for i := range c {
		if n -= c[i].GetChance(); 0 > n {
			return i, nil
		}
	}
	// 如果没有项目，则返回 -1 并报错
	return len(c) - 1, errors.New(`没有项目喵！`)
}

// IsSameDate 判断两个时间是否在同一天
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

/*
GetMidText 获取中间字符串

pre 为获取字符串的前缀（不包含）

suf 为获取字符串的后缀（不包含）

str 为整个字符串
*/
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

// GetWTAAnno 获取世界树纪元的完整字符串和额外信息
func GetWTAAnno() (str string, flower string, elemental string, imagery string, err error) {
	anno, err := wta.GetAnno()
	str = anno.YearStr + anno.MonthStr + anno.DayStr
	str = fmt.Sprintf(`%s　%d:%0*d:%0*d`, str, anno.Hour, 2, anno.Minute, 2, anno.Second)
	flower, elemental, imagery = anno.Flower, anno.Elemental, anno.Imagery
	return
}

/*
CheckPing 检查延迟

返回延迟毫秒数对应的语言描述
*/
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

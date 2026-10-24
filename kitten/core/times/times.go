// Package times 处理时间
package times

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

const (
	Layout       = `2006.1.2 15:04:05`     // Layout 日期时间格式
	LayoutHeart  = `2006.1.2	❤	15:04:05`   // LayoutHeart 带❤的日期时间格式
	LocationName = `Asia/Shanghai`         // LocationName 时区名称
	HoursPerDay  = 24                      // HoursPerDay 每天小时数
	Day          = HoursPerDay * time.Hour // Day 天
	DaysPerWeek  = 7                       // DaysPerWeek 每周天数
	Week         = DaysPerWeek * Day       // Week 周
	Year         = (365*4 + 1) * Day / 4   // Year 年
	Month        = Year / 12               // Month 月
)

// Location 固定时区
var Location *time.Location

func init() {
	var err error
	Location, err = time.LoadLocation(LocationName)
	if err != nil {
		panic(err)
	}
}

// TimeDuration 表示时间间隔的结构体
type TimeDuration struct {
	d int // 天
	h int // 小时
	m int // 分钟
	s int // 秒
}

// RandDelay 随机阻塞等待
func RandDelay(t time.Duration) <-chan time.Time {
	return RandDelayRange(0, t)
}

// RandDelayRange 带上下限的阻塞等待
func RandDelayRange(minDelay, maxDelay time.Duration) <-chan time.Time {
	return time.After(RandDurationRange(minDelay, maxDelay))
}

// RandDurationRange 随机时间间隔
func RandDurationRange(minDelay, maxDelay time.Duration) time.Duration {
	if minDelay > maxDelay {
		minDelay, maxDelay = maxDelay, minDelay
	}
	//nolint:gosec
	return minDelay + rand.N(maxDelay-minDelay)
}

// ConvertTimeDuration 转换时间间隔
func ConvertTimeDuration(d time.Duration) TimeDuration {
	if d < 0 {
		// 避免出现不正确的负数显示
		return TimeDuration{}
	}
	return TimeDuration{
		d: int(d / Day),
		h: int(d % Day / time.Hour),
		m: int(d % time.Hour / time.Minute),
		s: int(d % time.Minute / time.Second),
	}
}

// String 实现 fmt.Stringer
func (t TimeDuration) String() string {
	var (
		s   []string
		add = func(d int, e string) {
			if d != 0 {
				s = append(s, strconv.Itoa(d)+e)
			}
		}
	)
	add(t.d, ` 天`)
	add(t.h, ` 小时`)
	add(t.m, ` 分钟`)
	add(t.s, ` 秒`)
	return strings.Join(s, ` `)
}

// DailyDeadline 距离每日截止时刻的剩余时间
//
//nolint:predeclared
func DailyDeadline(hour, min, sec, nsec int) time.Duration {
	// 获取当前时间
	var (
		now = time.Now()
		// 创建今天的截止时刻
		target = time.Date(
			now.Year(), now.Month(), now.Day(),
			hour, min, sec, nsec,
			now.Location())
	)
	// 如果已经过了今天的截止时刻，就切换到明天的
	if now.After(target) {
		target = target.Add(Day)
	}
	// 计算剩余时间
	return target.Sub(now)
}

var standardDayNames = [...]string{"日", "一", "二", "三", "四", "五", "六"}

var youbiDayNames = [...]string{"日", "月", "水", "火", "木", "金", "土"}

// ParseWeekday 将 time.Weelday 转换为中文
func ParseWeekday(d time.Weekday, youbi, abbr bool) string {
	if youbi {
		s := youbiDayNames[int(d)]
		if abbr {
			return `(` + s + `)`
		}
		return s + `曜日`
	}
	s := standardDayNames[int(d)]
	if abbr {
		return `周` + s
	}
	return `星期` + s
}

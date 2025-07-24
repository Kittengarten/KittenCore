// Package times 处理时间
package times

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"time"
)

// TimeDuration 表示时间间隔的结构体
type TimeDuration struct {
	d int // 天
	h int // 小时
	m int // 分钟
	s int // 秒
}

const (
	Layout      = `2006.1.2 15:04:05`   // Layout 日期时间格式
	LayoutHeart = `2006.1.2	❤	15:04:05` // LayoutHeart 带❤的日期时间格式
	HoursPerDay = 24                    // HoursPerDay 每天小时数
)

// RandomDelay 随机阻塞等待
func RandomDelay(t time.Duration) {
	RandomDelayRange(0, t)
}

// RandomDelayRange 带上下限的阻塞等待
func RandomDelayRange(minDelay, maxDelay time.Duration) {
	if minDelay > maxDelay {
		minDelay, maxDelay = maxDelay, minDelay
	}
	//nolint:gosec
	if t := minDelay + rand.N(maxDelay-minDelay); t > 0 {
		time.Sleep(t)
	}
}

// ConvertTimeDuration 转换时间间隔
func ConvertTimeDuration(d time.Duration) TimeDuration {
	return TimeDuration{
		d: int(d / HoursPerDay / time.Hour),
		h: int(d % (HoursPerDay * time.Hour) / time.Hour),
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
		target = target.Add(HoursPerDay * time.Hour)
	}
	// 计算剩余时间
	return target.Sub(now)
}

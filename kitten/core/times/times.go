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
	d time.Duration // 天
	h time.Duration // 小时
	m time.Duration // 分钟
	s time.Duration // 秒
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
		<-time.NewTimer(t).C
	}
}

// ConvertTimeDuration 转换时间间隔
func ConvertTimeDuration(d time.Duration) TimeDuration {
	return TimeDuration{
		d: d / HoursPerDay / time.Hour,
		h: d % (HoursPerDay * time.Hour) / time.Hour,
		m: d % time.Hour / time.Minute,
		s: d % time.Minute / time.Second,
	}
}

// String 实现 fmt.Stringer
func (t TimeDuration) String() string {
	var (
		s []string
		c = func(t time.Duration) string {
			return strconv.FormatInt(int64(t), 10)
		}
		add = func(d time.Duration, e string) {
			if d != 0 {
				s = append(s, c(d)+e)
			}
		}
	)
	add(t.d, ` 天`)
	add(t.h, ` 小时`)
	add(t.m, ` 分钟`)
	add(t.s, ` 秒`)
	return strings.Join(s, ` `)
}

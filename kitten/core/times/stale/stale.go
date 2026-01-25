package stale

import (
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
)

// Level 停更等级
type Level byte

const (
	Day Level = iota
	Week
	Month
	Year
	Life
)

// String 实现 fmt.Stringer
func (l Level) String() string {
	return map[Level]string{
		Day:   `💖`,
		Week:  `✨`,
		Month: `🥀`,
		Year:  `⚰`,
		Life:  `🪦`,
	}[l]
}

// Check 检查停更等级
func Check(last time.Time) Level {
	now := time.Now()
	if d := equal.CmpDay(last, now); d <= 1 {
		return Day
	}
	if w := equal.CmpWeek(last, now); w <= 1 {
		return Week
	}
	if m := equal.CmpMonth(last, now); m <= 1 {
		return Month
	}
	if y := equal.CmpYear(last, now); y <= 1 {
		return Year
	}
	return Life
}

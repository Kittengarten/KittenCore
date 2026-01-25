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
	if diffDays := equal.CmpDay(last, now); diffDays <= 1 {
		return Day
	}
	if diffWeeks := equal.CmpWeek(last, now); diffWeeks <= 1 {
		return Week
	}
	if diffMonths := equal.CmpMonth(last, now); diffMonths <= 1 {
		return Month
	}
	if diffYears := equal.CmpYear(last, now); diffYears <= 1 {
		return Year
	}
	return Life
}

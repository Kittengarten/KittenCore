package track

import (
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
)

// 时间布局
var timeLayout = map[Platform]string{
	CWM: time.DateTime,
	FQ:  fqDateTime,
	SF:  sfDateTime,
}

// 保存报更
func (c *books) saveConfig() error {
	c.sortByUpdate()
	if err := io.Save(configPath, *c); err != nil {
		return err
	}
	cu <- *c
	return nil
}

// 按更新时间倒序排列小说
func (c *books) sortByUpdate() {
	slices.SortFunc(*c, func(j, i book) int {
		return i.UpdateTime.Compare(j.UpdateTime)
	})
}

// 时间解析，匹配不到支持的平台时使用默认时间格式
func (p Platform) ParseTime(str string) (time.Time, error) {
	if str == `` {
		return time.Time{}, nil
	}
	layout, ok := timeLayout[p]
	if !ok {
		return time.Parse(times.Layout, str)
	}
	return time.Parse(layout, str)
}

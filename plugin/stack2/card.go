package stack2

import (
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"github.com/RomiChan/syncx"
)

var active syncx.Map[kitten.QQ, time.Time] // 各群的上次活跃时间

// 设置群名片
func setCard(msgr *kitten.Messager, h int) {
	if g := kitten.NewQQGroup(msgr.Event.GroupID); g.IsGroup() {
		// 保存本群的活跃时间
		active.Store(*g, time.Unix(msgr.Event.Time, 0))
	}
	active.Range(func(g kitten.QQ, t time.Time) bool {
		msgr.SetCard(func() int {
			if time.Since(t) <= times.HoursPerDay*time.Hour {
				return h
			}
			// 如果群距上次活跃时间大于一天，则删除
			active.Delete(g)
			return -1
		}(), g)
		return true
	})
}

package stack2

import (
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/RomiChan/syncx"
)

var active syncx.Map[usr.QQ, time.Time] // 各群的上次活跃时间

// 设置群名片
func setCard(handler *msg.Handler, h int) {
	if g := usr.NewQQGroup(handler.Event().GroupID); g.IsGroup() {
		// 保存本群的活跃时间
		active.Store(g, time.Unix(handler.Event().Time, 0))
	}
	active.Range(func(g usr.QQ, t time.Time) bool {
		handler.SetCard(func() int {
			if equal.CmpDay4AM(t, time.Now()) <= 1 {
				return h
			}
			// 如果群距上次活跃时间大于一天，则删除
			active.Delete(g)
			return -1
		}(), g)
		return true
	})
}

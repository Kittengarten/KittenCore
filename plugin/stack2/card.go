package stack2

import (
	"context"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/lru"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"
)

var (
	// 最近 5 个活跃的群
	active = lru.New[usr.QQ, utils.Object](5)
	// 局部 Handler
	localHandler usr.Handler
)

func init() {
	active.OnEvicted = func(g usr.QQ, _ utils.Object) {
		if localHandler == nil {
			return
		}
		if err := msg.Handle(localHandler.(*msg.Handler).SetCard(-1, g)); err != nil {
			log.Error(err)
		}
	}
}

// 设置群名片
func setCard(handler *msg.Handler, h int) {
	// 独立的超时控制，不继承上游，以免上游提前完成导致本函数执行超时
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	localHandler = handler.SetContext(ctx)
	if g := usr.NewQQGroup(localHandler.Event().GroupID); g.IsGroup() {
		// 保存本群的活跃时间
		active.Add(g, utils.Object{})
	}
	active.Range(func(g usr.QQ, _ utils.Object) bool {
		if err := msg.Handle(localHandler.(*msg.Handler).SetCard(h, g)); err != nil {
			log.Error(err)
		}
		return true
	})
}

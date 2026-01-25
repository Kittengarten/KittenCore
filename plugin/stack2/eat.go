package stack2

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 吃猫猫执行逻辑
func eatExe(handler *msg.Handler) {
	if !setGlobalLocation(handler.Args()) {
		// 设置全局地区标记位，如当前活动未开放则返回
		asyncSendEmoji(handler, `辣眼睛`)
		handler.SendWithImageFail(`当前活动未开放喵！`)
		return
	}
	GlobalMessager.Ctx = handler.Ctx
	d, err := fio.LoadWithContext[data](handler, dataPath, fio.Empty)
	if err != nil {
		sendWithImageFail(handler, `加载叠猫猫数据文件时发生错误喵！`, err)
		return
	}
	stackStatus.refresh(handler, &d)
	_ = d.eat(handler)
	self(handler, d)
}

// 吃猫猫
func (d *data) eat(handler *msg.Handler) message.ID {
	var (
		// 初始化自身
		m, err = d.pre(handler)
		// 未在叠猫猫的队列
		dn data
		// 取消恢复数据状态
		cancel bool
		// 恢复数据状态
		restore = func() {
			if cancel {
				return
			}
			// 如果没有取消，则下次进行取消
			*d, cancel = slices.Concat(dn, *d, data{m}), true
		}
	)
	// 延迟恢复数据状态
	defer restore()
	if err != nil {
		// 初始化错误（需要休息或已经加入），不能吃猫猫，依靠延迟函数恢复数据状态
		return message.ID{}
	}
	if m.getTypeID(handler) < 小老虎 {
		// 不是老虎，不能吃猫猫，依靠延迟函数恢复数据状态
		asyncSendEmoji(handler, `NO`)
		return sendWithImageFail(handler, `老虎以上才可以吃猫猫——`)
	}
	// 未在叠猫猫的队列
	dn = d.getNoStack()
	// 执行吃猫猫
	if !d.doEat(handler, &m) {
		// 如果不能吃猫猫，依靠延迟函数恢复数据状态
		return message.ID{}
	}
	// 合并当前未叠猫猫与叠猫猫的队列，将老虎追加入切片中
	restore()
	// 清理过期玩家
	d.clear(handler, false)
	// 存储叠猫猫数据
	if err := fio.SaveWithContext(handler, dataPath, d); err != nil {
		return sendWithImageFail(handler, `存储叠猫猫数据时发生错误喵！`, err)
	}
	return message.ID{}
}

// 执行吃猫猫
func (d *data) doEat(handler *msg.Handler, m *meow) bool {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if l == 0 {
		// 没有猫猫
		asyncSendEmoji(handler, `哦`)
		sendWithImageFail(handler, `猫堆中没有猫猫可以吃——`)
		return false
	}
	if t := (*d)[l-1].getTypeID(handler); t >= 小老虎 {
		// 老虎以上无法被吃
		asyncSendEmoji(handler, `😁 呲牙`)
		sendWithImageFail(handler, `不可以吃`, &t, `——`)
		return false
	}
	if t := (*d)[0].getTypeID(handler); t >= 猫猫巴士 {
		// 底座是猫猫巴士以上无法被吃
		asyncSendEmoji(handler, `✨ 闪光`)
		sendWithImageFail(handler, `不可以吃`, &t, `载的猫猫——`)
		return false
	}
	var (
		mv   = m
		w, c int // 老虎吃到的体重（0.1 kg 数）和猫猫数
	)
	// 从队列的最上部开始遍历（后来居上）
	for i := range *d {
		// 下方的猫猫
		n := &(*d)[l-i-1]
		if !mv.checkEat(handler, *n) {
			// 这只猫猫没有被吃，直接结束遍历
			break
		}
		mv = n
		c++
		// 去除被吃的猫猫
		exit(handler, n, eaten, 0 /* 此参数无效 */)
		// 老虎增加被吃的猫猫的体重
		w += n.Weight
	}
	utils.Go(`吃猫猫设置群昵称`, func() { setCard(handler, l-c) })
	// 老虎进入休息
	exit(handler, m, eat, w)
	s := new(strings.Builder)
	s.Grow(256)
	if w == 0 {
		// 吃猫猫失败
		asyncSendEmoji(handler, `调皮`)
		fmt.Fprintf(s, `吃猫猫失败，杂鱼～杂鱼❤需要休息 %s。`,
			times.ConvertTimeDuration(m.Time.Sub(time.Unix(handler.Event().Time, 0))))
		doClear(handler, l, c, m.Weight, m, s)
		s.WriteRune('🐅')
		sendWithZako(handler, s)
		return true
	}
	// 吃猫猫成功
	asyncSendEmoji(handler, `😰 紧张`)
	fmt.Fprintf(s, `吃猫猫成功，你吃掉了 %d 只猫猫！需要休息 %s。`,
		c, times.ConvertTimeDuration(m.Time.Sub(time.Unix(handler.Event().Time, 0))))
	doClear(handler, l, c, m.Weight-w, m, s)
	s.WriteRune('🐯')
	for range c {
		s.WriteRune('😿')
	}
	e := dr[l-c:]
	sendText(handler, true, s, &e)
	return true
}

// 检查是否成功吃掉，m 为老虎，n 为猫猫
func (m meow) checkEat(handler *msg.Handler, n meow) bool {
	if n.getTypeID(handler) >= 小老虎 {
		// 老虎不能被吃
		return false
	}
	//nolint:gosec
	return rand.Float64() < m.chanceFall(n)
}

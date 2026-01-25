package stack2

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const lorryImage = `lorry`

// 撞大运执行逻辑
func lorryExe(handler *msg.Handler) {
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
	_ = d.lorry(handler)
	self(handler, d)
}

// 撞大运
func (d *data) lorry(handler *msg.Handler) message.ID {
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
		// 初始化错误（需要休息或已经加入）
		return message.ID{}
	}
	if m.getTypeID(handler) < 猫车 {
		// 不是猫车，不能撞大运
		asyncSendEmoji(handler, `NO`)
		return sendWithImageFail(handler, `猫车以上才可以撞大运——`)
	}
	// 未在叠猫猫的队列
	dn = d.getNoStack()
	// 执行撞大运
	if !d.doLorry(handler, &m) {
		// 如果不能撞大运，依靠延迟函数恢复数据状态
		return message.ID{}
	}
	// 合并当前未叠猫猫与叠猫猫的队列，将大运追加入切片中
	restore()
	// 清理过期玩家
	d.clear(handler, false)
	// 存储叠猫猫数据
	if err := fio.SaveWithContext(handler, dataPath, d); err != nil {
		return sendWithImageFail(handler, `存储叠猫猫数据时发生错误喵！`, err)
	}
	return message.ID{}
}

// 执行撞大运
func (d *data) doLorry(handler *msg.Handler, m *meow) bool {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if l == 0 {
		// 没有猫猫
		asyncSendEmoji(handler, `哦`)
		sendWithImageFail(handler, `猫堆中没有猫猫可以撞——`)
		return false
	}
	s := new(strings.Builder)
	s.Grow(256)
	if !d.checkLorry(*m) {
		// 撞大运失败
		// 猫车进入休息
		exit(handler, m, lorry, 0)
		asyncSendEmoji(handler, `调皮`)
		fmt.Fprintf(s, `撞大运失败，杂鱼～杂鱼❤需要休息 %s。`,
			times.ConvertTimeDuration(m.Time.Sub(time.Unix(handler.Event().Time, 0))))
		doClear(handler, l, 0, m.Weight, m, s)
		s.WriteRune('🚚')
		sendWithZako(handler, s)
		return true
	}
	// 撞大运成功
	for i := range *d {
		// 去除被撞飞的猫猫
		exit(handler, &(*d)[i], fly, l)
	}
	p := m.Weight
	// 猫车增加体重、进入休息
	exit(handler, m, lorry, l)
	asyncSendEmoji(handler, `😰 紧张`)
	fmt.Fprintf(s, `撞大运成功，你撞飞了 %d 只猫猫！需要休息 %s。`,
		l, times.ConvertTimeDuration(m.Time.Sub(time.Unix(handler.Event().Time, 0))))
	doClear(handler, l, l, p, m, s)
	s.WriteRune('🚛')
	for range l {
		s.WriteRune('😿')
	}
	sendWithImageLorry(handler, s, &dr)
	return true
}

// 检查撞大运是否成功
func (d *data) checkLorry(m meow) bool {
	//nolint:gosec
	return rand.Float64() < d.chanceLorry(m)
}

// 获取撞大运成功的概率
func (d *data) chanceLorry(m meow) float64 {
	return float64(m.Weight) / float64(d.totalWeight()+m.Weight)
}

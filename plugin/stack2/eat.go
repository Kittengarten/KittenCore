package stack2

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 吃猫猫执行逻辑
func eatExe(msgr *kitten.Messager) {
	if !setGlobalLocation(msgr.Args()) {
		// 设置全局地区标记位，如当前活动未开放则返回
		if err := msgr.SendEmojiLike(`辣眼睛`); err != nil {
			kitten.Error(err)
		}
		msgr.SendWithImageFail(`当前活动未开放喵！`)
		return
	}
	GlobalMessager = msgr
	d, err := fio.Load[data](dataPath, fio.Empty)
	if err != nil {
		sendWithImageFail(msgr, `加载叠猫猫数据文件时发生错误喵！`, err)
		return
	}
	stackBuffer.refresh(msgr, &d)
	_ = d.eat(msgr)
	if !selfEat(msgr, d) {
		times.RandomDelayRange(time.Second, 2*time.Second)
		selfIn(msgr, d)
	}
}

// 吃猫猫
func (d *data) eat(msgr *kitten.Messager) message.ID {
	var (
		// 初始化自身
		m, err = d.pre(msgr)
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
		// 如果初始化错误（需要休息或已经加入），不能吃猫猫，依靠延迟函数恢复数据状态
		return message.ID{}
	}
	if m.getTypeID(msgr) < 小老虎 {
		// 如果不是老虎，不能吃猫猫，依靠延迟函数恢复数据状态
		if err := msgr.SendEmojiLike(`NO`); err != nil {
			kitten.Error(err)
		}
		return sendWithImageFail(msgr, `老虎以上才可以吃猫猫——`)
	}
	// 未在叠猫猫的队列
	dn = d.getNoStack()
	// 执行吃猫猫
	if !d.doEat(msgr, &m) {
		// 如果不能吃猫猫，依靠延迟函数恢复数据状态
		return message.ID{}
	}
	// 合并当前未叠猫猫与叠猫猫的队列，将老虎追加入切片中
	restore()
	// 清理过期玩家
	d.clear(msgr, false)
	// 存储叠猫猫数据
	if err := fio.Save(dataPath, d); err != nil {
		return sendWithImageFail(msgr, `存储叠猫猫数据时发生错误喵！`, err)
	}
	return message.ID{}
}

// 执行吃猫猫
func (d *data) doEat(msgr *kitten.Messager, m *meow) bool {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if l == 0 {
		// 如果没有猫猫
		if err := msgr.SendEmojiLike(`哦`); err != nil {
			kitten.Error(err)
		}
		sendWithImageFail(msgr, `猫堆中没有猫猫可以吃——`)
		return false
	}
	if t := (*d)[l-1].getTypeID(msgr); t >= 小老虎 {
		// 老虎以上无法被吃
		if err := msgr.SendEmojiLike(`😁 呲牙`); err != nil {
			kitten.Error(err)
		}
		sendWithImageFail(msgr, `不可以吃`, &t, `——`)
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
		if !mv.checkEat(msgr, *n) {
			// 这只猫猫没有被吃，直接结束遍历
			break
		}
		mv = n
		c++
		// 去除被吃的猫猫
		exit(msgr, n, eaten, 0 /* 此参数无效 */)
		// 老虎增加被吃的猫猫的体重
		w += n.Weight
	}
	go setCard(msgr, l-c)
	// 老虎进入休息
	exit(msgr, m, eat, w)
	var r strings.Builder
	if w == 0 {
		if err := msgr.SendEmojiLike(`调皮`); err != nil {
			kitten.Error(err)
		}
		fmt.Fprintf(&r, `吃猫猫失败，杂鱼～杂鱼❤需要休息 %s。`,
			times.ConvertTimeDuration(m.Time.Sub(time.Unix(msgr.Event.Time, 0))))
		doClear(msgr, l, c, m.Weight, m, &r)
		r.WriteRune('🐅')
		sendWithZako(msgr, &r)
		return true
	}
	if err := msgr.SendEmojiLike(`😰 紧张`); err != nil {
		kitten.Error(err)
	}
	fmt.Fprintf(&r, `吃猫猫成功，你吃掉了 %d 只猫猫！需要休息 %s。`,
		c, times.ConvertTimeDuration(m.Time.Sub(time.Unix(msgr.Event.Time, 0))))
	doClear(msgr, l, c, m.Weight-w, m, &r)
	r.WriteRune('🐯')
	for range c {
		r.WriteRune('😿')
	}
	e := dr[l-c:]
	sendText(msgr, &r, &e)
	return true
}

// 检查是否成功吃掉，m 为老虎，n 为猫猫
func (m meow) checkEat(msgr *kitten.Messager, n meow) bool {
	if n.getTypeID(msgr) >= 小老虎 {
		// 老虎不能被吃
		return false
	}
	return rand.Float64() < m.chanceFall(n)
}

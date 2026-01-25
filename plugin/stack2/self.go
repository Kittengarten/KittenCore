package stack2

import (
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/usr"
)

// 自动过程
func self(handler *msg.Handler, d data) {
	if selfDaily(handler, d) {
		return
	}
	if selfLorry(handler, d) {
		return
	}
	if selfEat(handler, d) {
		return
	}
	if selfIn(handler, d) {
		return
	}
	_ = selfOC(handler, d)
}

// 自动日常任务
func selfDaily(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateDaily(handler) {
		// 以 评估 的概率，触发喵喵使用 /叠猫猫 日常
		return false
	}
	return selfDo(handler, cStack+cMeow+` `+cDaily, func() { d.daily(handler) })
}

// 评估日常任务，返回日常任务的权重作为概率
func (d *data) evaluateDaily(handler *msg.Handler) float64 {
	var (
		dr      = slices.Clone(*d)                   // 克隆切片，防止对后续调用造成影响
		m, err  = dr.pre(handler)                       // 初始化自身
		nre, ok = errors.AsType[*needRestError](err) // 错误是否需要休息
	)
	if !ok {
		// 当前不在休息，不需要日常任务，什么也不做
		return 0
	}
	if nre.Duration <= time.Hour {
		// 剩余休息时间过短，不能进行日常任务，什么也不做
		return 0
	}
	if equal.IsSameDate4AM(m.Daily, time.Unix(handler.Event().Time, 0)) {
		// 已经完成日常任务，不能进行日常任务，直接返回
		return 0
	}
	// 加速的权重为
	// 当前猫堆高度 ÷ max(100, 剩余休息小时数) ÷ min(1, 距离下一个 4:00 剩余 4 小时数)
	return float64(len(d.getStack())) /
		max(100, float64(nre.Duration)) /
		min(1, float64(times.DailyDeadline(4, 0, 0, 0))/
			float64(time.Hour)/4)
}

// 自动撞大运
func selfLorry(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateLorry(handler) {
		// 以评估的概率，触发喵喵使用 /你要撞大运了
		return false
	}
	return selfDo(handler, `你要撞大运了`, func() { _ = d.lorry(handler) })
}

// 评估撞大运，返回撞大运的权重作为概率
func (d *data) evaluateLorry(handler *msg.Handler) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		m, err = dr.pre(handler)     // 初始化自身
	)
	if err != nil {
		// 不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if l == 0 {
		// 空队列，什么也不做
		return 0
	}
	// 非空队列，撞大运的概率为期望比例 - 0.2
	return s.chanceLorry(m)*float64(l)/100 - 0.2
}

// 自动吃猫猫
func selfEat(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateEat(handler) {
		// 以评估的概率，触发喵喵使用 /吃猫猫
		return false
	}
	return selfDo(handler, cEat+cMeow, func() { _ = d.eat(handler) })
}

// 评估吃猫猫，返回吃的权重作为概率
func (d *data) evaluateEat(handler *msg.Handler) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		m, err = dr.pre(handler)     // 初始化自身
	)
	if err != nil {
		// 不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if l == 0 {
		// 空队列，什么也不做
		return 0
	}
	if t := s[l-1].getTypeID(handler); t >= 小老虎 {
		// 老虎以上无法被吃
		return 0
	}
	if t := s[0].getTypeID(handler); t >= 猫猫巴士 {
		// 底座是猫猫巴士以上无法被吃
		return 0
	}
	// 非空队列，吃猫猫的概率为期望占自己和小老虎体重中较高者的比例 - 0.5
	return float64(s[l-1].Weight)*m.chanceFall(s[l-1])/
		i2f(max(m.Weight, mapMeow[猫娘少女].weight)) - 0.5
}

// 自动加入
func selfIn(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateIn(handler) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 加入
		return false
	}
	return selfDo(handler, cStack+cMeow+` `+cIn, func() { _ = d.in(handler) })
}

// 自动分析
func selfAnalysis(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateIn(handler) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 分析
		return false
	}
	return selfDo(handler, cStack+cMeow+` `+cAnalysis, func() { d.analysis(handler) })
}

// 评估叠猫猫，返回叠入的权重作为概率
func (d *data) evaluateIn(handler *msg.Handler) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		m, err = dr.pre(handler)     // 初始化自身
	)
	if err != nil {
		// 不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if l == 0 {
		// 空队列，直接叠入，尝试平地摔或载猫猫
		return 1
	}
	// 非空队列
	if float64(m.Weight)*chanceFlat(m)*chanceClear(handler, s, m)*(math.E-1) >= float64(l) {
		// 清空特效导致体重增加的期望不少于当前的猫堆高度，直接叠入
		return 1
	}
	var (
		sn = append(s, m)           // 用于压坏判定的队列
		cp = sn.chancePressed(handler) // 压坏概率
		gp = func() float64 {
			if m.getTypeID(handler) <= 抱枕 || s[l-1].getTypeID(handler) >= 幼年猫娘 {
				// 抱枕及以下的猫猫不会导致猫猫摔下去，直接在猫娘身上叠猫猫不会摔下去
				return 0
			}
			return m.chanceFall(m)
		}() // 不压坏的情况下，摔下去的概率
		cf = (1 - cp) * gp // 摔下概率

	)
	if cf >= 0.5 {
		// 摔下概率达到 50% 以上，不叠入
		return 0
	}
	if s[0].Weight < mapMeow[大老虎].weight &&
		m.Weight >= mapMeow[大老虎].weight {
		// 底座不是猫车以上，自己是猫车以上
		return 0.1 - cf
	}
	if s[0].Weight >= mapMeow[幼年猫娘].weight && cp > 0 {
		// 底座是猫娘萝莉以上，只要可能压坏，就不叠入
		return 0
	}
	if cp >= 0.5 {
		// 压猫猫！
		// 压坏概率达到 50% 以上，队列中没有猫娘萝莉以上
		// 按照（压坏概率 - 摔下概率）× 自身体重与平均体重 e 倍的比值叠入
		for _, m := range s {
			if mapMeow[幼年猫娘].weight <= m.Weight {
				// 有猫娘萝莉以上，快跑！
				return 0
			}
		}
		return (cp - cf) * float64(m.Weight) / (math.E * float64(s.totalWeight()) / float64(l))
	}
	// 傍大猫
	// 底座越重且层数越低，越应该叠入
	return (1.0-float64(m.Weight)/float64(s[0].Weight))/float64(l) - cf
}

// 自动加速
func selfOC(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateOC(handler) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 锻炼
		return false
	}
	return selfDo(handler, cStack+cMeow+` `+cOC, func() { d.oc(handler) })
}

// 评估加速，返回加速的权重作为概率
func (d *data) evaluateOC(handler *msg.Handler) float64 {
	var (
		dr      = slices.Clone(*d)                   // 克隆切片，防止对后续调用造成影响
		_, err  = dr.pre(handler)                       // 初始化自身
		nre, ok = errors.AsType[*needRestError](err) // 错误是否需要休息
	)
	if !ok {
		// 当前不在休息，不需要加速，什么也不做
		return 0
	}
	if (*d)[nre.i].getTypeID(handler) < 大老虎 {
		// 不是大老虎，不能加速，什么也不做
		return 0
	}
	var (
		omrt  = time.Hour * time.Duration(stackConfig.OCMinRestHours)                  // 最小休息时间
		hours = int(math.RoundToEven(float64(nre.Duration-omrt) / float64(time.Hour))) // 加速的小时数
	)
	if hours <= 0 {
		// 加速的小时数不大于 0，则不能加速，什么也不做
		return 0
	}
	if (*d)[nre.i].Weight-hours < 1 {
		// 体重不足，则不能加速，什么也不做
		return 0
	}
	// 加速的权重为 (ln(当前体重（0.1 kg 数） ÷ 加速的小时数) - e) ÷ e^e
	return math.Log(float64((*d)[nre.i].Weight)/float64(hours)-math.E) / math.Pow(math.E, math.E)
}

// 自动排行
func selfRank(handler *msg.Handler, d data) bool {
	if handler.Event().UserID == 0 {
		return false
	}
	handler.Event().UserID = usr.Self().Int()
	//nolint:gosec
	if rand.Float64() >= 0.1 {
		// 以 0.1 的概率，触发喵喵使用 /叠猫猫 排行
		return false
	}
	return selfDo(handler, cStack+cMeow+` `+cRank, func() { d.rank(handler) })
}

// 自动过程
func selfDo(handler *msg.Handler, command string, f func()) bool {
	const (
		minDelay = time.Second
		maxDelay = 2 * minDelay
	)
	select {
	case <-times.RandDelayRange(minDelay, maxDelay):
	case <-handler.Done():
		handler.SendWithImageFail(handler.Err())
		return true
	}
	handler.ID = handler.SendWithContext(p + command)
	select {
	case <-times.RandDelayRange(minDelay, maxDelay):
		f()
	case <-handler.Done():
		handler.SendWithImageFail(handler.Err())
	}
	return true
}

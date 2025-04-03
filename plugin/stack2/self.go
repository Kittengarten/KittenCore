package stack2

import (
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
)

// 评估叠猫猫，返回叠入的权重作为概率
func (d *data) evaluate(msgr *kitten.Messager) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		m, err = dr.pre(msgr)     // 初始化自身
	)
	if err != nil {
		// 如果不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if l == 0 {
		// 如果是空队列，直接叠入，尝试平地摔
		return 1
	}
	// 如果是非空队列
	if float64(m.Weight)*chanceFlat(m)*chanceClear(msgr, s, m)*(math.E-1) >= float64(l) {
		// 如果清空特效导致体重增加的期望不少于当前的猫堆高度，直接叠入
		return 1
	}
	var (
		sn = append(s, m)           // 用于压坏判定的队列
		cp = sn.chancePressed(msgr) // 压坏概率
		gp = func() float64 {
			if m.getTypeID(msgr) <= 抱枕 || s[l-1].getTypeID(msgr) >= 幼年猫娘 {
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
	if s[0].Weight >= mapMeow[幼年猫娘].weight && cp > 0 {
		// 如果底座是猫娘萝莉以上，只要可能压坏，就不叠入
		return 0
	}
	if cp >= 0.5 {
		// 压猫猫！
		// 如果压坏概率达到 50% 以上，只要队列中没有猫娘萝莉以上，就按照（压坏概率 - 摔下概率）× 自身体重与平均体重 e 倍的比值叠入
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

// 自动加入
func selfIn(msgr *kitten.Messager, d data) bool {
	if msgr.Event.UserID == 0 {
		return false
	}
	msgr.Event.UserID = sid.Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluate(msgr) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 加入
		return false
	}
	times.RandomDelayRange(time.Second, 2*time.Second)
	msgr.ID = msgr.Ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cIn)
	times.RandomDelayRange(time.Second, 2*time.Second)
	_ = d.in(msgr)
	return true
}

// 自动分析
func selfAnalysis(msgr *kitten.Messager, d data) {
	if msgr.Event.UserID == 0 {
		return
	}
	msgr.Event.UserID = sid.Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluate(msgr) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 分析
		return
	}
	times.RandomDelayRange(time.Second, 2*time.Second)
	msgr.ID = msgr.Ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cAnalysis)
	times.RandomDelayRange(time.Second, 2*time.Second)
	d.analysis(msgr)
}

// 自动排行
func selfRank(msgr *kitten.Messager, d data) {
	if msgr.Event.UserID == 0 {
		return
	}
	msgr.Event.UserID = sid.Int()
	//nolint:gosec
	if rand.Float64() >= 0.1 {
		// 以 0.1 的概率，触发喵喵使用 /叠猫猫 排行
		return
	}
	times.RandomDelayRange(time.Second, 2*time.Second)
	msgr.ID = msgr.Ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cRank)
	times.RandomDelayRange(time.Second, 2*time.Second)
	d.rank(msgr)
}

// 评估吃猫猫，返回吃的权重作为概率
func (d *data) evaluateEat(msgr *kitten.Messager) float64 {
	var (
		dr     = slices.Clone(*d) // 克隆切片，防止对后续调用造成影响
		m, err = dr.pre(msgr)     // 初始化自身
	)
	if err != nil {
		// 如果不能活动，什么也不做
		return 0
	}
	var (
		s = dr.getStack() // 获取叠猫猫队列
		l = len(s)        // 叠猫猫队列长度
	)
	if l == 0 {
		// 如果是空队列，什么也不做
		return 0
	}
	// 如果是非空队列，吃猫猫的概率为期望占小老虎体重的比例 - 0.5
	return float64(s[l-1].Weight)*m.chanceFall(s[l-1])/itof(mapMeow[猫娘少女].weight) - 0.5
}

// 自动吃猫猫
func selfEat(msgr *kitten.Messager, d data) bool {
	if msgr.Event.UserID == 0 {
		return false
	}
	msgr.Event.UserID = sid.Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateEat(msgr) {
		// 以评估的概率，触发喵喵使用 /吃猫猫
		return false
	}
	times.RandomDelayRange(time.Second, 2*time.Second)
	msgr.ID = msgr.Ctx.Send(botConfig.CommandPrefix + cEat + cMeow)
	times.RandomDelayRange(time.Second, 2*time.Second)
	_ = d.eat(msgr)
	return true
}

// 评估加速，返回加速的权重作为概率
func (d *data) evaluateOC(msgr *kitten.Messager) float64 {
	var (
		dr     = slices.Clone(*d)  // 克隆切片，防止对后续调用造成影响
		_, err = dr.pre(msgr)      // 初始化自身
		nre    = needRest(0, 0, 0) // 默认错误：需要休息
	)
	if !errors.As(err, &nre) {
		// 如果当前不在休息，不需要加速，什么也不做
		return 0
	}
	if (*d)[nre.i].getTypeID(msgr) < 大老虎 {
		// 如果不是大老虎，不能加速，什么也不做
		return 0
	}
	var (
		omrt  = time.Hour * time.Duration(stackConfig.OCMinRestHours)                  // 最小休息时间
		hours = int(math.RoundToEven(float64(nre.Duration-omrt) / float64(time.Hour))) // 加速的小时数
	)
	if hours <= 0 {
		// 如果加速的小时数不大于 0，则不能加速，什么也不做
		return 0
	}
	if (*d)[nre.i].Weight-hours < 1 {
		// 如果体重不足，则不能加速，什么也不做
		return 0
	}
	// 加速的权重为 (ln(当前体重（0.1 kg 数） ÷ 加速的小时数) - e) ÷ e^e
	return math.Log(float64((*d)[nre.i].Weight)/float64(hours)-math.E) / math.Pow(math.E, math.E)
}

// 自动加速
func selfOC(msgr *kitten.Messager, d data) bool {
	if msgr.Event.UserID == 0 {
		return false
	}
	msgr.Event.UserID = sid.Int()
	//nolint:gosec
	if rand.Float64() >= d.evaluateOC(msgr) {
		// 以评估的概率，触发喵喵使用 /叠猫猫 锻炼
		return false
	}
	times.RandomDelayRange(time.Second, 2*time.Second)
	msgr.ID = msgr.Ctx.Send(botConfig.CommandPrefix + cStack + cMeow + ` ` + cOCCat)
	times.RandomDelayRange(time.Second, 2*time.Second)
	d.oc(msgr)
	return true
}

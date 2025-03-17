package stack2

import (
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 缓存刷新
func (b *buffer) refresh(msgr *kitten.Messager, d *data) {
	stackBuffer, err = core.Load[buffer](bufferPath, `medianweight: 0
maxresttime: 0`)
	if err != nil {
		sendWithImageFail(msgr, `读取叠猫猫缓存时发生错误喵！`, err)
		return
	}
	// 计算猫池中位数重量
	d.median(msgr)
	// 计算最大休息时间
	stackBuffer.MaxRestTime = core.HoursPerDay * time.Hour *
		time.Duration(stackConfig.MinRestHours*stackBuffer.MedianWeight)
	if core.Save(bufferPath, stackBuffer) != nil {
		sendWithImageFail(msgr, `叠猫猫缓存时发生错误喵！`, err)
	}
}

// 获取平地摔或特效的概率
func chanceFlat(m meow) float64 {
	return min(float64(mapMeow[抱枕].weight)/float64(m.Weight), 1)
}

// String 实现 fmt.Stringer
func (m meow) String() string {
	if globalLocation == cockroach {
		return fmt.Sprintf(`【%s】	翼展 %.1f cm`, m.getType(GlobalMessager).String(), itof(m.Weight))
	}
	return fmt.Sprintf(l10nReplacer().Replace(`%s	❤	%d	❤	%.1f kg	%s`),
		m.TitleCardOrNickName(GlobalMessager), m.Int(), itof(m.Weight), m.getType(GlobalMessager).String())
}

// 获取 m 摔坏 n 的概率
func (m meow) chanceFall(n meow) float64 {
	return math.Pow(float64(m.Weight)/float64(m.Weight+n.Weight),
		math.Ln2/(math.Log(math.E+1.0)-1))
}

/*
检查是否因为叠猫猫失败摔下去

m 为上方的猫猫，n 为下方的猫猫

如果没有摔下去则返回 true
*/
func (m meow) checkFall(n meow) bool {
	return rand.Float64() >= m.chanceFall(n)
}

// 获取猫猫类型
func (m meow) getTypeID(msgr *kitten.Messager) meowTypeID {
	for i := range unknown {
		if m.Weight < mapMeow[i].weight {
			if i == 猫娘少女 && m.IsAdult(msgr) {
				continue
			}
			return i
		}
	}
	return unknown
}

// 获取猫猫类型
func (m meow) getType(msgr *kitten.Messager) meowType {
	return mapMeow[m.getTypeID(msgr)]
}

// String 实现 fmt.Stringer
func (m meowTypeID) String() string {
	return mapMeow[m].str
}

// String 实现 fmt.Stringer
func (m meowType) String() string {
	return m.str
}

// 整数体重转换为浮点（千克数）
func itof(w int) float64 {
	return float64(w) / 10
}

// 浮点体重（千克数）转换为整数
func ftoi(w float64) int {
	return int(min(10*w, math.MaxInt))
}

// 返回服从正态分布 N(0, σ²) 的随机数的绝对值，相当于此分布的右半边
func normal(σ float64) float64 {
	return σ * math.Abs(rand.NormFloat64())
}

// 遍历断言
func rangeAssertion(a []any) []any {
	for k, v := range a {
		switch v := v.(type) {
		case error:
			a[k] = l10nReplacer().Replace(v.Error())
		case fmt.Stringer:
			a[k] = l10nReplacer().Replace(v.String())
		case string:
			a[k] = l10nReplacer().Replace(v)
		}
	}
	return a
}

// 发送本地化文本
func sendText(msgr *kitten.Messager, text ...any) message.ID {
	return msgr.Reply().AtLf().Text(rangeAssertion(text)...).Send()
}

// 发送本地化格式化文本
func sendTextOf(msgr *kitten.Messager, format string, a ...any) message.ID {
	return msgr.Reply().AtLf().TextOf(
		l10nReplacer().Replace(format),
		rangeAssertion(a)...,
	).Send()
}

// 发送带有失败图片的本地化文字消息
func sendWithImageFail(msgr *kitten.Messager, text ...any) message.ID {
	return msgr.SendWithImageFail(rangeAssertion(text)...)
}

// 发送带有杂鱼图片的本地化文字消息
func sendWithZako(msgr *kitten.Messager, text ...any) message.ID {
	return msgr.Reply().AtLf().Image(core.Path(zako)).Text(rangeAssertion(text)...).Send()
}

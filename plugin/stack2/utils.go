package stack2

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"

	"github.com/wdvxdr1123/ZeroBot/message"
)

// 缓存刷新
func (b *status) refresh(handler *msg.Handler, d *data) {
	stackStatus, err = fio.LoadWithContext[status](handler, bufferPath, fio.Blank)
	if err != nil {
		sendWithImageFail(handler, `读取叠猫猫缓存时发生错误喵！`, err)
		return
	}
	// 计算猫池中位数重量
	d.median(handler)
	// 计算最大休息时间
	stackStatus.MaxRestTime = min(math.MaxInt64/dailyRatio, times.Day*
		time.Duration(stackConfig.MinRestHours*stackStatus.MedianWeight))
	if err := fio.SaveWithContext(handler, bufferPath, stackStatus); err != nil {
		sendWithImageFail(handler, `叠猫猫缓存时发生错误喵！`, err)
	}
}

// 获取平地摔或特效的概率
func chanceFlat(m meow) float64 {
	return min(float64(mapMeow[抱枕].weight)/float64(m.Weight), 1)
}

// 检查是否因为叠猫猫失败摔下去
//
//	m 为上方的猫猫
//	n 为下方的猫猫
//	如果没有摔下去则返回 true
func (m meow) checkFall(n meow) bool {
	//nolint:gosec
	return rand.Float64() >= m.chanceFall(n)
}

// 获取 m 摔坏 n 的概率
func (m meow) chanceFall(n meow) float64 {
	return math.Pow(float64(m.Weight)/float64(m.Weight+n.Weight),
		math.Ln2/(math.Log(math.E+1.0)-1))
}

// 获取猫猫类型
func (m meow) getType(handler *msg.Handler) meowType {
	return mapMeow[m.getTypeID(handler)]
}

// 获取猫猫类型 ID
func (m meow) getTypeID(handler *msg.Handler) meowTypeID {
	for i := range unknown {
		if m.Weight < mapMeow[i].weight {
			if i == 猫娘少女 && m.IsAdult(handler) {
				continue
			}
			return i
		}
	}
	return unknown
}

// String 实现 fmt.Stringer
func (m meowType) String() string {
	return m.str
}

// String 实现 fmt.Stringer
func (m meowTypeID) String() string {
	return mapMeow[m].str
}

// 返回服从正态分布 N(0, σ²) 的随机数的绝对值，相当于此分布的右半边
func normal(σ float64) float64 {
	//nolint:gosec
	return σ * math.Abs(rand.NormFloat64())
}

// 发送本地化文本
func sendText(handler *msg.Handler, lf bool, text ...any) message.ID {
	if lf {
		return handler.Quote().AtLf().Text(rangeAssertion(text)...).Send()
	}
	return handler.Quote().At().Text(rangeAssertion(text)...).Send()
}

// 发送本地化格式化文本
func sendTextf(handler *msg.Handler, lf bool, format string, a ...any) message.ID {
	if lf {
		return handler.Quote().AtLf().Textf(
			l10n.Replace(format),
			rangeAssertion(a)...,
		).Send()
	}
	return handler.Quote().At().Textf(
		l10n.Replace(format),
		rangeAssertion(a)...,
	).Send()
}

// 发送带有撞大运图片的本地化文字消息
func sendWithImageLorry(handler *msg.Handler, text ...any) message.ID {
	return handler.Quote().AtLf().Image(fio.NewPath(imagePath, lorryImage)).
		Text(rangeAssertion(text)...).Send()
}

// 发送带有失败图片的本地化文字消息
func sendWithImageFail(handler *msg.Handler, text ...any) message.ID {
	return handler.SendWithImageFail(rangeAssertion(text)...)
}

// 发送带有杂鱼图片的本地化文字消息
func sendWithZako(handler *msg.Handler, text ...any) message.ID {
	return handler.Quote().AtLf().Image(fio.NewPath(zako)).
		Text(rangeAssertion(text)...).Send()
}

// 发送带有压扁图片的本地化文字消息
func sendWithPressed(handler *msg.Handler, text ...any) message.ID {
	return handler.Quote().AtLf().Image(fio.NewPath(imagePath, `压扁.gif`)).
		Text(rangeAssertion(text)...).Send()
}

// 发送带有没压扁图片的本地化格式文字消息
func sendWithNoPressedf(handler *msg.Handler, format string, text ...any) message.ID {
	return handler.Quote().AtLf().Image(fio.NewPath(imagePath, `没压扁.png`)).
		Textf(format, rangeAssertion(text)...).Send()
}

// 发送带有跳跃图片的本地化格式文字消息
func sendWithJump(handler *msg.Handler, text ...any) message.ID {
	return handler.Quote().AtLf().Image(fio.NewPath(imagePath, `跳.gif`)).
		Text(rangeAssertion(text)...).Send()
}

// 异步发送表情回复
func asyncSendEmoji(handler *msg.Handler, emojiName string) {
	utils.Go(`叠猫猫发送表情回复`, func() {
		if err := handler.SendEmojiLike(emojiName); err != nil {
			log.Warn(err)
		}
	})
}

// 遍历断言
func rangeAssertion(a []any) []any {
	for k, v := range a {
		switch v := v.(type) {
		case error:
			a[k] = l10n.Replace(v.Error())
		case fmt.Stringer:
			a[k] = l10n.Replace(v.String())
		case string:
			a[k] = l10n.Replace(v)
		}
	}
	return a
}

// Format 实现 fmt.Formatter
func (m meow) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v': // 需要 %v 以屏蔽底层的 usr.QQ 的格式
		fmt.Fprint(f, m.String())
	default:
		type raw meow
		fmt.Fprintf(f, fmt.FormatString(f, verb), raw(m))
	}
}

// String 实现 fmt.Stringer
func (m meow) String() string {
	ctx, cancel := context.WithTimeout(context.Background(), shttp.Timeout)
	defer cancel()
	GlobalMessager.Context = ctx // 设置上下文
	defer func() { GlobalMessager.Context = context.Background() }()
	if globalLocation == cockroach {
		return fmt.Sprintf(`【%s】	翼展 %.1f cm`, m.getType(GlobalMessager), i2f(m.Weight))
	}
	return fmt.Sprintf(
		l10n.Replace(`%s	❤	%d	❤	%.1f kg	%s`),
		cmp.Or(m.TitleCardOrNickName(
			GlobalMessager,
		), m.Name),
		m.Int(),
		i2f(m.Weight),
		m.getType(GlobalMessager),
	)
}

// 整数体重转换为浮点（千克数）
func i2f(w int) float64 {
	return float64(w) / 10
}

// 浮点体重（千克数）转换为整数
func f2i(w float64) int {
	return int(min(10*w, math.MaxInt))
}

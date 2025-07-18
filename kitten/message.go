package kitten

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/core/msg/seg"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 待发送的消息
type Messager struct {
	err             error // 错误（不使用内嵌，避免 Messager 实现 error）
	*zero.Ctx             // 上下文
	message.ID            // 引用消息 ID
	message.Message       // 消息
	record          bool  // 是否有语音
}

// New 创建待发送的消息
func New(ctx *zero.Ctx) *Messager {
	return &Messager{Ctx: ctx}
}

// 比较含有的两个消息段切片是否相等
func equalContainedMessage(a, b *Messager) bool {
	return equalMessage(a.Message, b.Message)
}

// 比较待发送的消息是否相等
func hasSameMessage(m ...*Messager) bool {
	return equal.IsSameFunc(equalContainedMessage, m...)
}

// 设置回复消息 ID
func (m *Messager) id(id message.ID) *Messager {
	m.ID = id
	return m
}

/*
Quote 引用消息，id 为引用的消息 ID（仅限一个）

如引用消息为空则引用本消息的触发来源消息

已经设置过引用消息 ID 时，会被参数覆盖，不提供参数时无效
*/
func (m *Messager) Quote(id ...message.ID) *Messager {
	if len(id) > 0 {
		return m.id(id[0])
	}
	if m.ID != (message.ID{}) {
		return m
	}
	switch id := m.Event.MessageID.(type) {
	case int64:
		return m.id(message.NewMessageIDFromInteger(id))
	case string:
		return m.id(message.NewMessageIDFromString(id))
	default:
		if m.Event.PostType == `message` {
			Infof(`引用消息 ID 断言不成功：%+v`, m.Event)
		}
		return m
	}
}

// At 附带 @，u 为 @ 对象，如 @ 对象为空则 @ Messager 的来源
func (m *Messager) At(u ...QQ) *Messager {
	if m.Event.DetailType == Private {
		// 私聊中的 @ 无效
		return m
	}
	if len(u) == 0 {
		// 如果@ 对象为空，则 @ Messager 的来源
		return m.Seg(message.At(m.Event.UserID)).Text(` `)
	}
	for _, i := range u {
		// @ 对象
		m.Seg(i.At()).Text(` `)
	}
	return m
}

// AtAll 附带 @ 全体成员
func (m *Messager) AtAll(g ...QQ) *Messager {
	if seg := atAll(m, g...); seg.Type != `` {
		return m.Seg(seg).Text(` `)
	}
	return m
}

// Text 附带文本
func (m *Messager) Text(text ...any) *Messager {
	m.Seg(Text(text...))
	return m
}

// Lf 附带换行
func (m *Messager) Lf(n ...int) *Messager {
	if len(n) == 0 {
		return m.Text("\n")
	}
	return m.Text(strings.Repeat("\n", n[0]))
}

// AtLf 附带 @ 并换行
func (m *Messager) AtLf(qq ...QQ) *Messager {
	if n := *m; !hasSameMessage(m, n.At(qq...)) {
		return n.Lf()
	}
	return m
}

// AtAllLf 附带 @ 全体成员 并换行
func (m *Messager) AtAllLf(g ...QQ) *Messager {
	if n := *m; !hasSameMessage(m, n.AtAll(g...)) {
		return n.Lf()
	}
	return m
}

// Textf 附带格式化文本
func (m *Messager) Textf(format string, a ...any) *Messager {
	m.Seg(Textf(format, a...))
	return m
}

/*
Image 从图片的相对 | 绝对路径（文件夹），

或相对 | 绝对路径文件中保存的相对 | 绝对路径，

或网络路径中附带图片
*/
func (m *Messager) Image(name ...fio.Path) *Messager {
	for _, n := range name {
		img, err := imagePath.Image(n)
		if err != nil {
			m.err = errors.Join(m.err, fmt.Errorf(`附带图片错误：%w`, err))
			img, err = imagePath.Image(fio.NewPath(`error.jpg`))
			if err != nil {
				m.err = errors.Join(m.err, fmt.Errorf(`附带图片错误：%w`, err))
			}
		}
		m.Seg(img)
	}
	return m
}

/*
Record 附带语音，支持网络路径

只支持附带一条语音
*/
func (m *Messager) Record(name ...string) *Messager {
	if m.record || len(name) == 0 {
		// 已经有语音，或未附带语音
		return m
	}
	if n := cmp.Or(name...); n != `` {
		return m.Seg(message.Record(n))
	}
	return m
}

// Message 附带消息段
func (m *Messager) Seg(seg ...message.Segment) *Messager {
	m.Message = append(m.Message, seg...)
	return m
}

// SendMulti 发送多条消息
func (m *Messager) SendMulti(u ...QQ) (id []message.ID) {
	defer m.Reset()
	if m.err != nil {
		// 有错误，将其打包进消息
		m = m.Text("\n", m.err)
	}
	if len(m.Message) == 0 {
		// 没有消息段，无法发送
		return nil
	}
	if len(u) != 0 {
		// 发送对象不为空，向发送对象发送
		for _, o := range u {
			switch {
			case o.IsGroup(), o.IsQQ():
				id = append(id, o.Send(m))
				times.RandomDelayRange(time.Second, 2*time.Second)
			}
		}
		return id
	}
	// 发送对象为空，向 Messager 的来源发送
	if !m.Check(Caller, Event) {
		// 没有 APICaller 或 Event ，无法发送
		Warn(m)
		return nil
	}
	if m.Event.PostType != `message` || m.ID.ID() == 0 {
		// 不是消息引发的发送或没有回复，不予回复
		return []message.ID{m.Ctx.Send(m.Message)}
	}
	for _, e := range m.Message {
		switch e.Type {
		case seg.Text, seg.Face, seg.Image, seg.At:
			// 消息段兼容回复，不执行操作
		default:
			// 消息段不兼容回复，或未经验证，跳过回复程序
			return []message.ID{m.Ctx.Send(m.Message)}
		}
	}
	// 有回复
	return []message.ID{m.Ctx.Send(message.ReplyWithMessage(m.ID, m.Message...))}
}

// Send 发送消息，u 为可选的发送对象，如 u 为空则发送给上下文的来源
func (m *Messager) Send(u ...QQ) message.ID {
	if ids := m.SendMulti(u...); len(ids) != 0 {
		// 返回第一个 ID
		return ids[0]
	}
	return message.ID{}
}

// Reset 重置 Messager，保留上下文
func (m *Messager) Reset() *Messager {
	m.Message = nil
	m.ID = message.ID{}
	m.err = nil
	m.record = false
	return m
}

// 比较多个消息段是否相等
func IsSameSegment(s ...message.Segment) bool {
	return equal.IsSameFunc(equalSegment, s...)
}

// 比较多个消息段切片是否相等
func IsSameMessage(m ...message.Message) bool {
	return equal.IsSameFunc(equalMessage, m...)
}

// 比较两个消息段切片是否相等
func equalMessage(a, b message.Message) bool {
	if len(a) != len(b) {
		// 如果两个消息段切片的长度不同，则不相等
		return false
	}
	for i, seg := range a {
		if !equalSegment(seg, b[i]) {
			return false
		}
	}
	return true
}

// 比较两个消息段是否相等
func equalSegment(a, b message.Segment) bool {
	if a.Type != b.Type {
		// 如果两个消息段类型不同，则不相等
		return false
	}
	// 按类型的特殊比较路径
	switch a.Type {
	case seg.Image:
		// 图片，比较文件或路径
		return mio.GetImagePath(a) != `` && mio.GetImagePath(a) == mio.GetImagePath(b) ||
			mio.GetImageURL(a) != `` && mio.GetImageURL(a) == mio.GetImageURL(b)
	case seg.Record, seg.Video, seg.Anonymous, seg.Share, seg.Contact,
		seg.Location, seg.Music, seg.Forward, seg.Node, seg.XML, seg.JSON:
		// 忽略的类型，视为不相等
		return false
	default:
		// 直接比较数据
		return equal.IsSameMap(a.Data, b.Data)
	}
}

// CallAction 调用 cqhttp API
func (m *Messager) CallAction(action string, params zero.H) zero.APIResponse {
	return m.Ctx.CallAction(action, params)
}

// Textf 格式化构建 message.Segment 文本，格式同 fmt.Sprintf
func Textf(format string, a ...any) message.Segment {
	return Text(fmt.Sprintf(format, a...))
}

// Text 构建 message.Segment 文本，格式同 fmt.Sprint
func Text(text ...any) message.Segment {
	checkErr(text)
	return message.Text(text...)
}

// 检查切片的每个元素是否为错误，如果为非空错误则记录日志
func checkErr(v []any) {
	for _, i := range v {
		if err, ok := i.(error); ok && err != nil {
			Error(err)
		}
	}
}

// Image 将收到的图片文件名 | 绝对路径 | 网络 URL | Base64 编码转换为图片消息
func Image(file string, summary ...any) message.Segment {
	return message.Image(file, summary...)
}

package kitten

import (
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
	*zero.Ctx             // 上下文
	message.Message       // 消息
	message.ID            // 回复
	err             error // 错误
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
	return equal.IsSameByFunc(equalContainedMessage, m...)
}

// 设置回复消息 ID
func (m *Messager) id(id message.ID) *Messager {
	m.ID = id
	return m
}

// Reply 附带回复，id 为回复的消息 ID（仅限一个），如回复对象为空则回复消息的来源
func (m *Messager) Reply(id ...message.ID) *Messager {
	if len(id) > 0 {
		return m.id(id[0])
	}
	switch id := m.Event.MessageID.(type) {
	case int64:
		return m.id(message.NewMessageIDFromInteger(id))
	case string:
		return m.id(message.NewMessageIDFromString(id))
	default:
		if m.Event.PostType == `message` {
			Infof(`附带回复消息 ID 断言不成功：%+v`, m.Event)
		}
		return m
	}
}

// At 附带 @，u 为 @ 对象，如 @ 对象为空则 @ @ 的来源
func (m *Messager) At(u ...QQ) *Messager {
	if m.Event.DetailType == Private {
		// 私聊中的 @ 无效
		return m
	}
	if len(u) == 0 {
		// 如果@ 对象为空，则 @ @ 的来源
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
func (m *Messager) AtAllLf(qq ...QQ) *Messager {
	if n := *m; !hasSameMessage(m, n.AtAll(qq...)) {
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

// Record 附带语音，支持网络路径
func (m *Messager) Record(name ...string) *Messager {
	if m.record || len(name) == 0 {
		return m
	}
	for _, n := range name {
		if len(n) == 0 {
			continue
		}
		m.record = true
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
	if nil != m.err {
		// 有错误，将其打包进消息
		m = m.Text(m.err)
	}
	if len(m.Message) == 0 {
		// 没有消息段，无法发送
		return nil
	}
	if len(u) != 0 {
		// 如果发送对象不为空，则向发送对象发送
		for _, o := range u {
			switch {
			case o.IsGroup(), o.IsQQ():
				id = append(id, o.Send(m))
				times.RandomDelayRange(time.Second, 2*time.Second)
			}
		}
		return id
	}
	if !m.Check(Event) || !m.Check(Caller) {
		// 没有事件或 APICaller ，无法发送
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
func (m *Messager) Send(u ...QQ) (id message.ID) {
	return m.SendMulti(u...)[0]
}

// Reset 重置 Messager
func (m *Messager) Reset() *Messager {
	m.Message = nil
	m.ID = message.ID{}
	m.err = nil
	m.record = false
	return m
}

// 比较两个消息段是否相等
func equalSegment(a, b message.Segment) bool {
	if a.Type != b.Type {
		// 如果两个消息段类型不同，则不相等
		return false
	}
	switch a.Type {
	// 按类型的特殊比较路径
	case seg.Image:
		if mio.GetImagePath(a) == mio.GetImagePath(b) {
			// 如果图片文件相同，则相等，继续遍历比较
			return true
		}
	case seg.Record, seg.Video, seg.Anonymous, seg.Share, seg.Contact,
		seg.Location, seg.Music, seg.Forward, seg.Node, seg.XML, seg.JSON:
		// 忽略的类型，将导致停止比较，视为不相等
	default:
		if equal.IsSameMap(a.Data, b.Data) {
			// 相等，继续遍历比较
			return true
		}
	}
	return false
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

// 比较多个消息段是否相等
func IsSameSegment(s ...message.Segment) bool {
	return equal.IsSameByFunc(equalSegment, s...)
}

// 比较多个消息段切片是否相等
func IsSameMessage(m ...message.Message) bool {
	return equal.IsSameByFunc(equalMessage, m...)
}

// CallAction 调用 cqhttp API
func (m *Messager) CallAction(action string, params map[string]any) zero.APIResponse {
	return m.Ctx.CallAction(action, params)
}

// 检查接口切片的每个元素中是否为错误，如果有则记录日志
func checkErr(v []any) {
	for _, i := range v {
		if err, ok := i.(error); ok {
			Error(err)
		}
	}
}

// Text 构建 message.Segment 文本，格式同 fmt.Sprint
func Text(text ...any) message.Segment {
	checkErr(text)
	return message.Text(text...)
}

// Textf 格式化构建 message.Segment 文本，格式同 fmt.Sprintf
func Textf(format string, a ...any) message.Segment {
	return Text(fmt.Sprintf(format, a...))
}

// Image 将收到的图片文件名 | 绝对路径 | 网络 URL | Base64 编码转换为图片消息
func Image(file string, summary ...any) message.Segment {
	return message.Image(file, summary...)
}

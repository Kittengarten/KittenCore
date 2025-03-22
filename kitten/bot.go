package kitten

import (
	"errors"
	"fmt"
	"image"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/io"
	ms "github.com/Kittengarten/KittenCore/kitten/core/msg/seg"
	"github.com/Kittengarten/KittenCore/kitten/qqemoji"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type Item byte // 对上下文的检查类型

const (
	Caller Item = iota // APICaller
	Event              // *Event
)

const noEvent = `上下文无 *Event，不可使用`

const (
	Private = `private`
	Group   = `group`
)

/*
Send 向 QQ（负数代表群聊）发送消息

发送完毕后不会重置 *Messager，可能需要调用 msgr.Reset() 手动重置
*/
func (u *QQ) Send(msgr *Messager) message.ID {
	if !msgr.Check(Caller) {
		// 没有 APICaller ，无法发送
		Warn(msgr)
		return message.NewMessageIDFromInteger(0)
	}
	// 是否需要回复
	if func() bool {
		if msgr.ID.ID() == 0 {
			return false
		}
		for _, seg := range msgr.Message {
			switch seg.Type {
			case ms.Text, ms.Face, ms.Image, ms.At:
				// 消息段兼容回复，不执行操作
			default:
				// 消息段不兼容回复，或未经验证，跳过回复程序
				return false
			}
		}
		return true
	}() {
		msgr.Message = message.ReplyWithMessage(msgr.ID, msgr.Message...)
	}
	var id int64
	switch {
	case u.IsQQ():
		id = msgr.SendPrivateMessage(u.Int(), msgr.Message)
	case u.IsGroup():
		id = msgr.SendGroupMessage(u.Int(), msgr.Message)
	default:
		Error(`无效的发送对象：`, u)
	}
	return message.NewMessageIDFromInteger(id)
}

// SendWithImageFail 发送带有失败图片的文字消息
func (m *Messager) SendWithImageFail(text ...any) message.ID {
	return m.Reply().AtLf().Image(`no.png`).Text(text...).Send()
}

// SendWithImageFailOf 发送带有失败图片的文字消息
func (m *Messager) SendWithImageFailOf(format string, a ...any) message.ID {
	return m.Reply().AtLf().Image(`no.png`).TextOf(format, a...).Send()
}

// DoNotKnow 喵喵不知道哦
func (m *Messager) DoNotKnow() message.ID {
	handleErr := func(err error) message.ID {
		Error(err)
		return m.Reply().AtLf().Image(`哈.png`).Text(botConfig.NickName[0], `不知道哦`).Send()
	}
	o, err := m.Object()
	if err != nil {
		return handleErr(err)
	}
	n, err := o.Name()
	if err != nil {
		return handleErr(err)
	}
	return m.Reply().AtLf().Image(`哈.png`).Text(n, `不知道哦`).Send()
}

// @ 全体成员，带有检查功能
func atAll(msgr *Messager, g ...QQ) message.Segment {
	if !msgr.Check(Caller) {
		Error(`没有 APICaller ，无法获取`)
		return message.Segment{}
	}
	if !msgr.Check(Event) {
		if len(g) == 0 {
			return message.Segment{}
		}
		msgr.Event = &zero.Event{GroupID: g[0].Int()}
	}
	g = []QQ{*NewQQGroup(msgr.Event.GroupID)}
	if !g[0].IsGroup() {
		Error(`无效的群聊：`, g)
		return message.Segment{}
	}
	r := msgr.GetGroupAtAllRemain(g[0].Int())
	if !r.Get(`can_at_all`).Bool() {
		Info(`在该群聊不能 @ 全体成员：`, g)
		return message.Segment{}
	}
	if r.Get(`remain_at_all_count_for_uin`).Int() <= 0 {
		Info(`剩余 @ 全体成员次数不足`)
		return message.Segment{}
	}
	return message.AtAll()
}

// SendEmojiLike 发送表情回复，emoji 为空则发送随机表情
func (m *Messager) SendEmojiLike(emoji ...string) error {
	if u := NewQQGroup(m.Event.GroupID); !u.IsGroup() {
		Info(`无效的群聊：`, u)
		return errors.New(`无效的群聊`)
	}
	return m.SetMessageEmojiLike(m.Event.MessageID, func() rune {
		if len(emoji) == 0 {
			return qqemoji.Random()
		}
		return qqemoji.New(emoji[0])
	}())
}

// Poke 戳一戳
func (m *Messager) Poke() {
	if !m.Check(Caller) || !m.Check(Event) {
		// 没有 APICaller 或 Event ，无法使用
		return
	}
	if u, g := NewQQ(m.Event.UserID), NewQQGroup(m.Event.GroupID); u.IsQQ() {
		if g.IsGroup() {
			m.CallAction("group_poke", zero.H{
				"group_id": g.Int(),
				"user_id":  u.Int(),
			})
			return
		}
		m.CallAction("friend_poke", zero.H{
			"user_id": u.Int(),
		})
	}
}

// Object 获取发送对象
func (m *Messager) Object() (*QQ, error) {
	if !m.Check(Event) {
		// 没有事件，无法获取
		return nil, errors.New(noEvent)
	}
	switch m.Event.DetailType {
	case Private:
		// 私聊
		return NewQQ(m.Event.UserID), nil
	case Group:
		// 群聊
		return NewQQGroup(m.Event.GroupID), nil
	default:
		if g := NewQQGroup(m.Event.GroupID); g.IsGroup() {
			return g, nil
		}
		if u := NewQQ(m.Event.UserID); u.IsQQ() {
			return u, nil
		}
		return nil, errors.New(`不支持的对象喵！`)
	}
}

// Check 检查上下文的某个项目是否有效且不为空
func (m *Messager) Check(i Item) bool {
	if m.Ctx == nil {
		// 没有上下文，无法获取
		return false
	}
	switch i {
	case Caller:
		c := reflect.ValueOf(m.Ctx).Elem().FieldByName(`caller`)
		return c.IsValid() && !c.IsNil()
	case Event:
		if m.Event == nil {
			// 非消息的上下文，直接返回
			Info(noEvent)
			return false
		}
	default:
		// 检查了错误的项目
		return false
	}
	return true
}

// State 获取上下文中的字段
func State[T any](msgr *Messager, name string) (t T) {
	f, ok := msgr.State[name]
	if !ok {
		return
	}
	t, _ = f.(T)
	return
}

// Matched 获取上下文中的匹配项
func (m *Messager) Matched() string {
	return State[string](m, `matched`)
}

// RegexMatched 获取上下文中的正则匹配项
func (m *Messager) RegexMatched() []string {
	return State[[]string](m, `regex_matched`)
}

// Prefix 获取上下文中的前缀
func (m *Messager) Prefix() string {
	return State[string](m, `prefix`)
}

// Suffix 获取上下文中的后缀
func (m *Messager) Suffix() string {
	return State[string](m, `suffix`)
}

// Keyword 获取上下文中的关键词
func (m *Messager) Keyword() string {
	return State[string](m, `keyword`)
}

// Command 获取上下文中的命令
func (m *Messager) Command() string {
	return State[string](m, `command`)
}

// Args 获取上下文中的参数
func (m *Messager) Args() string {
	return State[string](m, `args`)
}

// ArgsSlice 获取上下文中的参数切片
func (m *Messager) ArgsSlice() []string {
	return slices.DeleteFunc(strings.Split(m.Args(), ` `),
		func(s string) bool {
			return s == ``
		})
}

// ImageURL 获取上下文中的图片链接
func (m *Messager) ImageURL() []string {
	return State[[]string](m, `image_url`)
}

// （私有）扫描二维码
func scanQRCode(imgfile *os.File) (fmt.Stringer, error) {
	defer imgfile.Close()
	img, _, err := image.Decode(imgfile)
	if err != nil {
		return nil, err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return bmp, err
	}
	return qrcode.NewQRCodeReader().DecodeWithoutHints(bmp)
}

// ScanQRCode 扫描二维码
func ScanQRCode(name string) (fmt.Stringer, error) {
	var (
		msg = Image(name)
		n   = io.NewPath(`data`, `zbp`, `code.png`)
	)
	bytes, err := n.DownloadImage(msg.Data[`file`])
	if err != nil {
		return io.NewPath(msg.Data[`file`]), err
	}
	Info(`正在扫描二维码喵！字节数：`, bytes)
	imgfile, err := os.Open(n.String())
	if err != nil {
		return io.NewPath(msg.Data[`file`]), err
	}
	return scanQRCode(imgfile)
}

// ScanQRCodeInQQ 扫描 QQ 消息中的二维码图片
func (m *Messager) ScanQRCodeInQQ(file string) (fmt.Stringer, error) {
	imgfile, err := os.Open(m.GetImage(file).Get(`file`).String())
	if err != nil {
		return io.NewPath(file), err
	}
	return scanQRCode(imgfile)
}

// 对象
type obj bool

const (
	// ForSelf 为自己
	ForSelf obj = true
	// ForUser 为对方
	ForUser obj = false
)

// Age 获取年龄，self 为 true 则获取自己的年龄，否则获取对方的年龄
func (m *Messager) Age(self obj) int64 {
	if m == nil {
		return 0
	}
	if self {
		return botConfig.SelfID.Age(m)
	}
	return NewQQ(m.Event.UserID).Age(m)
}

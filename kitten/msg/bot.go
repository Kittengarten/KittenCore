package msg

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/qqemoji"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	Private = `private` // Private 私聊
	Group   = `group`   // Group 群聊
)

var (
	// ErrNoEvent 上下文无 *Event，不可使用喵！
	ErrNoEvent = errors.New(`上下文无 *Event，不可使用喵！`)
	// ErrGroupInvalid 无效的群聊喵！
	ErrGroupInvalid = errors.New(`无效的群聊喵！`)
)

// SendEmojiLike 发送表情回复（只接受首个参数），参数为空则发送随机表情
func (handler *Handler) SendEmojiLike(emoji ...string) error {
	if u := usr.NewQQGroup(handler.Event().GroupID); !u.IsGroup() {
		if u == 0 {
			// 忽略私聊
			return nil
		}
		return fmt.Errorf(`%w%s`, ErrGroupInvalid, u)
	}
	return handler.SetMessageEmojiLike(handler.Event().MessageID, func() rune {
		if len(emoji) == 0 {
			return qqemoji.Random()
		}
		return qqemoji.Get(emoji[0])
	}())
}

// Poke 戳一戳
func (handler *Handler) Poke() {
	if !handler.Check(kitten.Caller, kitten.Event) {
		// 没有 APICaller 或 kitten.Event ，无法发送
		return
	}
	if u, g := usr.NewQQ(handler.Event().UserID), usr.NewQQGroup(handler.Event().GroupID); u.IsQQ() {
		if g.IsGroup() {
			handler.CallAction(`group_poke`, zero.H{
				`group_id`: g.Int(),
				`user_id`:  u.Int(),
			})
			return
		}
		handler.CallAction(`friend_poke`, zero.H{
			`user_id`: u.Int(),
		})
	}
}

// Object 获取发送对象
func (handler *Handler) Object() (usr.QQ, error) {
	if !handler.Check(kitten.Event) {
		// 没有事件，无法获取
		return 0, ErrNoEvent
	}
	switch event := handler.Event(); event.DetailType {
	case Private:
		// 私聊
		return usr.NewQQ(event.UserID), nil
	case Group:
		// 群聊
		return usr.NewQQGroup(event.GroupID), nil
	default:
		if g := usr.NewQQGroup(event.GroupID); g.IsGroup() {
			return g, nil
		}
		if u := usr.NewQQ(event.UserID); u.IsQQ() {
			return u, nil
		}
		return 0, fmt.Errorf(`发送对象 %s 不支持喵！%w`, event.DetailType, errors.ErrUnsupported)
	}
}

// QQ 获取发送者
func (handler *Handler) QQ() (usr.QQ, error) {
	if !handler.Check(kitten.Event) {
		// 没有事件，无法获取
		return 0, ErrNoEvent
	}
	u := usr.NewQQ(handler.Event().UserID)
	if u.IsQQ() {
		return u, nil
	}
	return 0, fmt.Errorf(`发送者 %s 不支持喵！%w`, u, errors.ErrUnsupported)
}

// Check 检查上下文的项目是否均有效且不为空
func (handler *Handler) Check(i ...kitten.Item) bool {
	if handler.Ctx == nil {
		// 没有上下文，无法获取
		return false
	}
	for _, v := range i {
		switch v {
		case kitten.Caller:
			c := reflect.ValueOf(handler.Ctx).Elem().FieldByName(`caller`)
			if !c.IsValid() || c.IsNil() {
				return false
			}
		case kitten.Event:
			if handler.Event() == nil {
				// 非消息的上下文，直接返回
				// 不需要 Error 等级，以免污染日志
				log.Info(ErrNoEvent)
				return false
			}
		default:
			// 检查了错误的项目
			return false
		}
	}
	return true
}

// Matched 获取上下文中的匹配项
func (handler *Handler) Matched() string {
	return State[string](handler, `matched`)
}

// RegexMatched 获取上下文中的正则匹配项
func (handler *Handler) RegexMatched() []string {
	return State[[]string](handler, `regex_matched`)
}

// Prefix 获取上下文中的前缀
func (handler *Handler) Prefix() string {
	return State[string](handler, `prefix`)
}

// Suffix 获取上下文中的后缀
func (handler *Handler) Suffix() string {
	return State[string](handler, `suffix`)
}

// Keyword 获取上下文中的关键词
func (handler *Handler) Keyword() string {
	return State[string](handler, `keyword`)
}

// Command 获取上下文中的命令
func (handler *Handler) Command() string {
	return State[string](handler, `command`)
}

// Args 获取上下文中的参数
func (handler *Handler) Args() string {
	return State[string](handler, `args`)
}

// ArgsSlice 获取上下文中的参数切片
func (handler *Handler) ArgsSlice() []string {
	return slices.DeleteFunc(strings.Split(handler.Args(), ` `),
		func(s string) bool {
			return s == ``
		})
}

// ImageURL 获取上下文中的图片链接
func (handler *Handler) ImageURL() []string {
	return State[[]string](handler, `image_url`)
}

const (
	no = `no.png`
	ha = `哈.png`
)

// SendWithImageFail 发送带有失败图片的文字消息
func (handler *Handler) SendWithImageFail(text ...any) message.ID {
	return handler.Quote().AtLf().Image(no).Text(text...).Send()
}

// SendWithImageFailf 发送带有失败图片的格式化文字消息
func (handler *Handler) SendWithImageFailf(format string, a ...any) message.ID {
	return handler.Quote().AtLf().Image(no).Textf(format, a...).Send()
}

// DoNotKnow 喵喵不知道哦
func (handler *Handler) DoNotKnow() message.ID {
	var (
		handleErr = func(err error) message.ID {
			log.Error(err)
			return handler.Quote().AtLf().Image(ha).Text(config.DefaultName(), `不知道哦`).Send()
		}
		o, err = handler.Object()
	)
	if err != nil {
		return handleErr(err)
	}
	n, err := o.Name(handler)
	if err != nil {
		return handleErr(err)
	}
	return handler.Quote().AtLf().Image(ha).Text(n, `不知道哦`).Send()
}

// @ 全体成员，带有检查功能
func atAll(handler *Handler, g ...usr.QQ) message.Segment {
	const n = 1 // 封装 1 层
	if !handler.Check(kitten.Caller) {
		log.Skip(n).Error(`没有 APICaller ，无法获取`)
		return message.Segment{}
	}
	if !handler.Check(kitten.Event) {
		if len(g) == 0 {
			return message.Segment{}
		}
		handler.Ctx.Event = &zero.Event{GroupID: g[0].Int()}
	}
	g = []usr.QQ{usr.NewQQGroup(handler.Event().GroupID)}
	if !g[0].IsGroup() {
		log.Skip(n).Error(`无效的群聊：`, g)
		return message.Segment{}
	}
	r := handler.GetGroupAtAllRemain(g[0].Int())
	if !r.Get(`can_at_all`).Bool() {
		log.Skip(n).Info(`在该群聊不能 @ 全体成员：`, g)
		return message.Segment{}
	}
	if r.Get(`remain_at_all_count_for_uin`).Int() <= 0 {
		log.Skip(n).Info(`剩余 @ 全体成员次数不足`)
		return message.Segment{}
	}
	return message.AtAll()
}

// State 获取上下文中的字段
func State[T any](handler *Handler, name string) (t T) {
	f, ok := handler.State[name]
	if !ok {
		return
	}
	t, _ = f.(T)
	return
}

// ScanQRCode 扫描二维码
func ScanQRCode(ctx context.Context, name string) (fmt.Stringer, error) {
	var (
		msg = Image(name)
		n   = fio.NewPath(cmp.Or(os.TempDir(), `data/zbp`), `kitten`, `code.png`)
	)
	bytes, err := n.DownloadImage(ctx, mio.GetImageURL(msg))
	if err != nil {
		return nil, fmt.Errorf(`图片下载错误喵！%w`, err)
	}
	log.Info(`正在扫描二维码喵！字节数：`, bytes)
	imgfile, err := os.Open(n.String())
	if err != nil {
		return nil, err
	}
	return scanQRCode(imgfile)
}

// ScanQRCodeInQQ 扫描 QQ 消息中的二维码图片
func (m *Handler) ScanQRCodeInQQ(file string) (fmt.Stringer, error) {
	imgfile, err := os.Open(m.GetImage(file).Get(`file`).String())
	if err != nil {
		return nil, err
	}
	return scanQRCode(imgfile)
}

// 扫描二维码
func scanQRCode(imgfile *os.File) (s fmt.Stringer, err error) {
	defer func() {
		if e := imgfile.Close(); e != nil {
			err = fmt.Errorf(`关闭二维码文件失败喵！%w`, e)
		}
	}()
	img, _, err := image.Decode(imgfile)
	if err != nil {
		return nil, err
	}
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}
	return qrcode.NewQRCodeReader().DecodeWithoutHints(bmp)
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
func (handler *Handler) Age(self obj) int64 {
	if handler == nil {
		return 0
	}
	if self {
		return usr.Self().Age(handler)
	}
	return usr.NewQQ(handler.Event().UserID).Age(handler)
}

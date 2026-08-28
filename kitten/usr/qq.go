package usr

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg/seg"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// QQ 是一个表示 QQ 的 int64
type QQ int64

// ErrBirthMissing 生日信息缺失喵！
var ErrBirthMissing = errors.New(`生日信息缺失喵！`)

var sid = NewQQ(config.ID())

// Self 获取 bot 的 ID
func Self() QQ {
	return sid
}

// NewQQ QQ 的构造函数
func NewQQ(v ...int64) QQ {
	if len(v) > 0 {
		return QQ(v[0])
	}
	return 0
}

// NewQQGroup QQ 群的构造函数
func NewQQGroup(v ...int64) QQ {
	if len(v) > 0 {
		return QQ(-v[0])
	}
	return 0
}

// ParseQQ 从字符串获取 QQ
func ParseQQ(s string) QQ {
	if q, err := strconv.Atoi(s); err == nil {
		return QQ(q)
	}
	return 0
}

// ParseQQGroup 从字符串获取 QQ 群
func ParseQQGroup(s string) QQ {
	if q, err := strconv.Atoi(s); err == nil {
		return QQ(-q)
	}
	return 0
}

// Int 获取 QQ 的 int64 类型表示
func (u QQ) Int() int64 {
	switch {
	case u.IsQQ():
		return int64(u)
	case u.IsGroup():
		return -int64(u)
	default:
		return 0
	}
}

// Format 实现 fmt.Formatter，返回 QQ 字符串
//
//	%s 十进制数字
//	%o 原始十进制数字（群号为负）
func (u QQ) Format(state fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		_, _ = fmt.Fprint(state, u.String())
	case 'o':
		_, _ = fmt.Fprint(state, u.Int())
	default:
		type raw QQ
		_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), raw(u))
	}
}

// String 获取 QQ 的十进制 string 类型表示
func (u QQ) String() string {
	return strconv.FormatInt(u.Int(), 10)
}

// String 获取 QQ 的十进制 string 类型原始表示
//
//  群号为负值，可用于 key
func (u QQ) Str() string {
	return strconv.FormatInt(int64(u), 10)
}

// At @ 该 QQ
func (u QQ) At() message.Segment {
	if !u.IsQQ() {
		return message.Segment{}
	}
	return message.At(u.Int())
}

// IsTarget 判断是否是 notify 的目标
func (u QQ) IsTarget() func(ctx *zero.Ctx) bool {
	if !u.IsQQ() {
		return func(*zero.Ctx) bool {
			return false
		}
	}
	return func(ctx *zero.Ctx) bool {
		return u.Int() == ctx.Event.TargetID
	}
}

// Birthday 获取生日
func (u QQ) Birthday(handler Context) (time.Time, error) {
	if !u.IsQQ() {
		return time.Time{}, fmt.Errorf(`%s 不是 QQ，%w`, u, ErrBirthMissing)
	}
	// 群成员信息中没有生日，直接退化至陌生人
	var (
		info = u.info(handler)
		y    = info.Get(`birthday_year`).String()
		m    = info.Get(`birthday_month`).String()
		d    = info.Get(`birthday_day`).String()
	)
	if m == `` || m == `0` || d == `` || d == `0` {
		return time.Time{}, fmt.Errorf(`%s %w`, u, ErrBirthMissing)
	}
	if y == `` || y == `0` {
		y = `0001`
	}
	return time.Parse(`2006.1.2`, strings.Join([]string{y, m, d}, `.`))
}

// CallName 从 QQ 获取用于称呼的简单昵称（经过修剪）
func (u QQ) CallName(handler Context) (n string) {
	overLen := func(name string) bool {
		return len(name) > 16 && utf8.RuneCountInString(name) > 8
	}
	n = str.FirstText(str.Clean(u.Card(handler), false))
	if n == `` || overLen(n) {
		n = str.FirstText(str.Clean(u.NickName(handler), false))
	}
	if n == `` || overLen(n) {
		n = str.FirstText(str.Clean(u.Title(handler), false))
	}
	return
}

// Card 从 QQ 获取群昵称（事件必须是群）
func (u QQ) Card(handler Context) string {
	if !NewQQGroup(handler.Event().GroupID).IsGroup() {
		return ``
	}
	return u.memberInfo(handler).Get(`card`).Str
}

// NickName 从 QQ 获取昵称
func (u QQ) NickName(handler Context) string {
	return u.info(handler).Get(`nickname`).Str
}

// TitleCardOrNickName 从 QQ 获取【头衔】群昵称 | 昵称（经过修剪）
func (u QQ) TitleCardOrNickName(handler Context) string {
	// 头衔
	var title string
	if NewQQGroup(handler.Event().GroupID).IsGroup() {
		// 是群聊，获取该 QQ 在群内的头衔
		if title = u.Title(handler); title != `` {
			// 如果头衔存在，则添加实心方头括号
			title = `【` + title + `】	`
		}
	}
	// 返回【头衔】群昵称 | 昵称
	return title + str.Clean(u.CardOrNickName(handler), false)
}

// CardOrNickName 从 QQ 获取群昵称 | 昵称
func (u QQ) CardOrNickName(handler Context) string {
	if !handler.Check(kitten.Caller) || !u.IsQQ() {
		// 没有 APICaller 或不是 QQ，无法获取
		return ``
	}
	if NewQQGroup(handler.Event().GroupID).IsGroup() {
		// 是群聊，获取群昵称
		if card := u.Card(handler); card != `` {
			// 如果不为空，返回群昵称
			return card
		}
	}
	// 不是群聊或群昵称为空，返回昵称
	return u.NickName(handler)
}

// Title 从 QQ 获取头衔（必须是群）
func (u QQ) Title(handler Context) string {
	return u.memberInfo(handler).Get(`title`).Str
}

// InThisGroup 是否在本群
func (u QQ) InThisGroup(handler Context) bool {
	return u.memberInfo(handler).Exists()
}

// 获取群成员信息
func (u QQ) memberInfo(handler Context) qqInfo {
	if !u.IsQQ() {
		// 不是 QQ，无法获取
		return qqInfo{}
	}
	g := NewQQGroup(handler.Event().GroupID)
	if !g.IsGroup() {
		// 不是群，无法获取，退化至陌生人
		return u.info(handler)
	}
	list := g.MemberList(handler)
	if len(list.List) == 0 {
		// 本群成员列表为空，无法获取，退化至陌生人
		return u.info(handler)
	}
	// 从本群成员列表中查找
	i := slices.IndexFunc(list.List, func(i gjson.Result) bool {
		return i.Get(`user_id`).Int() == u.Int()
	})
	if i == -1 {
		// 本群成员列表中找不到，为陌生人
		return u.info(handler)
	}
	return qqInfo{
		Result: list.List[i],
		Time:   list.Time,
	}
}

// MemberList 获取特定群的成员列表
func (g QQ) MemberList(handler Context) groupList {
	if !handler.Check(kitten.Caller) || !g.IsGroup() {
		// 没有 APICaller 或不是 QQ 群，无法获取
		return groupList{}
	}
	// 从缓存获取该群成员列表
	gmi, ok := groupMemberList.Load(g)
	if !ok {
		// 如果获取不到，同步更新
		g.updateMemberList(handler)
		gmi, _ = groupMemberList.Load(g)
		return gmi
	}
	// 如果缓存已经过期，异步（后台）更新缓存的该群成员列表
	if time.Since(gmi.Time) > expire {
		utils.Go(`更新群成员列表`, func() {
			// 独立上下文，不继承上游
			ctx, cancel :=
				context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			g.updateMemberList(handler.SetContext(ctx))
		})
	}
	return gmi
}

// 更新特定群的成员列表
func (g QQ) updateMemberList(handler Context) {
	groupMemberList.Store(g,
		groupList{
			List: handler.GetGroupMemberListNoCache(g.Int()).Array(),
			Time: time.Now(),
		},
	)
	if err := saveMember(); err != nil {
		log.Error(err)
	}
}

// IsQQ 是 QQ
func (u QQ) IsQQ() bool {
	return u > 10000
}

// IsGroup 是群
func (u QQ) IsGroup() bool {
	return u < -100000
}

// IsAdult 是成年人
func (u QQ) IsAdult(handler Context) bool {
	return 18 <= u.Age(handler)
}

// IsLoli 是萝莉
func (u QQ) IsLoli(handler Context) bool {
	return u.IsFemale(handler) && 0 < u.Age(handler) && 18 > u.Age(handler)
}

// IsFemale 是女性
func (u QQ) IsFemale(handler Context) bool {
	return u.info(handler).Get(`sex`).String() == `female`
}

// Age 获取年龄
func (u QQ) Age(handler Context) int64 {
	return u.info(handler).Get(`age`).Int()
}

// 获取陌生人信息
func (u QQ) info(handler Context) qqInfo {
	if !handler.Check(kitten.Caller) || !u.IsQQ() {
		// 没有 APICaller 或不是 QQ，无法获取
		return qqInfo{}
	}
	// 从缓存获取该 QQ 的信息
	si, ok := strangerInfo.Load(u)
	if !ok {
		// 如果获取不到，同步更新
		u.UpdateInfo(handler)
		si, _ = strangerInfo.Load(u)
		return si
	}
	// 如果缓存已经过期，异步更新缓存的陌生人信息
	if time.Since(si.Time) > expire {
		utils.Go(`更新陌生人信息`, func() {
			// 独立上下文，不继承上游
			ctx, cancel :=
				context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			u.UpdateInfo(handler.SetContext(ctx))
		})
	}
	return si
}

// UpdateInfo 更新陌生人信息
func (u QQ) UpdateInfo(handler Context) {
	strangerInfo.Store(u,
		qqInfo{
			Result: handler.GetStrangerInfo(u.Int(), true),
			Time:   time.Now(),
		},
	)
	if err := saveInfo(); err != nil {
		log.Error(err)
	}
}

// Send 向 QQ（负数代表群聊）发送消息
//
//	发送完毕后不会重置 Handler，可能需要调用 handler.Reset() 手动重置
func (u QQ) Send(handler Handler) message.ID {
	if !handler.Check(kitten.Caller) {
		// 没有 APICaller ，无法发送
		log.Warn(handler)
		return message.ID{}
	}
	// 是否需要回复
	func() {
		if handler.QuoteID().ID() == 0 {
			return
		}
		for _, e := range handler.Get() {
			switch e.Type {
			case seg.Text, seg.Face, seg.Image, seg.At:
				// 消息段兼容回复，不执行操作
			default:
				// 消息段不兼容回复，或未经验证，跳过回复程序
				return
			}
		}
		handler.Set(message.ReplyWithMessage(handler.QuoteID(), handler.Get()...))
	}()
	return message.NewMessageIDFromInteger(func() int64 {
		switch {
		case u.IsQQ():
			return handler.SendPrivateMessage(u.Int())
		case u.IsGroup():
			return handler.SendGroupMessage(u.Int())
		default:
			log.Debug(`无效的发送对象：`, u)
			return 0
		}
	}())
}

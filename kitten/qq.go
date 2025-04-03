package kitten

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Kittengarten/KittenCore/kitten/core/str"

	"github.com/RomiChan/syncx"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	// QQ 是一个表示 QQ 的 int64
	QQ int64

	// QQ 信息
	qqInfo struct {
		gjson.Result // 信息
		time.Time    // 上次更新时间
	}

	// 群成员列表
	groupList struct {
		List      []gjson.Result // 每个群员的信息
		time.Time                // 上次更新时间
	}
)

// 缓存过期时间
const expire = time.Hour

var (
	strangerInfo    syncx.Map[QQ, qqInfo]    // 各陌生人信息缓存
	groupMemberList syncx.Map[QQ, groupList] // 各群的成员列表缓存
)

// ErrBirthMissing 生日信息缺失喵！
var ErrBirthMissing = errors.New(`生日信息缺失喵！`)

// Self 获取 bot 的 ID
func Self() QQ {
	return botConfig.SelfID
}

// NewQQ QQ 的构造函数
func NewQQ(v ...int64) *QQ {
	if len(v) > 0 {
		qq := QQ(v[0])
		return &qq
	}
	return new(QQ)
}

// NewQQGroup QQ 群的构造函数
func NewQQGroup(v ...int64) *QQ {
	if len(v) > 0 {
		qq := QQ(-v[0])
		return &qq
	}
	return new(QQ)
}

// Int 获取 QQ 的 int64 类型表示
func (u *QQ) Int() int64 {
	switch {
	case u.IsQQ():
		return int64(*u)
	case u.IsGroup():
		return -int64(*u)
	default:
		return 0
	}
}

// Int 获取 QQ 的 string 类型表示
func (u *QQ) String() string {
	return strconv.FormatInt(u.Int(), 10)
}

// At @ 该 QQ
func (u *QQ) At() message.Segment {
	if !u.IsQQ() {
		return message.Segment{}
	}
	return message.At(u.Int())
}

// IsTarget 判断是否是 notify 的目标
func (u *QQ) IsTarget() func(ctx *zero.Ctx) bool {
	if !u.IsQQ() {
		return func(*zero.Ctx) bool {
			return false
		}
	}
	return func(ctx *zero.Ctx) bool {
		return u.Int() == ctx.Event.TargetID
	}
}

// （私有）获取陌生人信息
func (u *QQ) info(msgr *Messager) qqInfo {
	if !msgr.Check(Caller) || !u.IsQQ() {
		// 没有 APICaller 或不是 QQ，无法获取
		return qqInfo{}
	}
	// 从缓存获取该 QQ 的信息
	si, ok := strangerInfo.Load(*u)
	if !ok {
		// 如果获取不到，同步更新
		u.updateInfo(msgr)
		si, _ = strangerInfo.Load(*u)
	}
	// 如果缓存已经过期，异步更新缓存的陌生人信息
	if time.Since(si.Time) > expire {
		go u.updateInfo(msgr)
	}
	return si
}

// （私有）更新陌生人信息
func (u *QQ) updateInfo(msgr *Messager) {
	strangerInfo.Store(*u,
		qqInfo{
			Result: msgr.GetStrangerInfo(u.Int(), true),
			Time:   time.Now(),
		},
	)
}

// （私有）获取群成员信息
func (u *QQ) memberInfo(msgr *Messager) qqInfo {
	g := NewQQGroup(msgr.Event.GroupID)
	if !g.IsGroup() {
		// 如果不是群，退化至陌生人
		return u.info(msgr)
	}
	list := g.MemberList(msgr)
	if len(list.List) == 0 {
		// 如果本群成员列表为空，退化至陌生人
		return u.info(msgr)
	}
	// 从本群成员列表中查找
	i := slices.IndexFunc(list.List, func(i gjson.Result) bool {
		return i.Get(`user_id`).Int() == u.Int()
	})
	if i == -1 {
		// 如果本群成员列表中找不到，退化至陌生人
		return u.info(msgr)
	}
	return qqInfo{
		Result: list.List[i],
		Time:   list.Time,
	}
}

// IsQQ 是 QQ
func (u *QQ) IsQQ() bool {
	return *u > 10000
}

// IsGroup 是群
func (u *QQ) IsGroup() bool {
	return *u < -100000
}

// Age 获取年龄
func (u *QQ) Age(msgr *Messager) int64 {
	return u.info(msgr).Get(`age`).Int()
}

// Birthday 获取生日
func (u *QQ) Birthday(msgr *Messager) (time.Time, error) {
	if !u.IsQQ() {
		return time.Time{}, fmt.Errorf(`%s 不是QQ，%w`, u, ErrBirthMissing)
	}
	// 群成员信息中没有生日，直接退化至陌生人
	var (
		info = u.info(msgr)
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
	return time.Parse(`2006-1-2`, strings.Join([]string{y, m, d}, `-`))
}

// IsAdult 是成年人
func (u *QQ) IsAdult(msgr *Messager) bool {
	return 18 <= u.Age(msgr)
}

// IsFemale 是女性
func (u *QQ) IsFemale(msgr *Messager) bool {
	return u.info(msgr).Get(`sex`).String() == `female`
}

// IsLoli 是萝莉
func (u *QQ) IsLoli(msgr *Messager) bool {
	return u.IsFemale(msgr) && 0 < u.Age(msgr) && 18 > u.Age(msgr)
}

// Title 从 QQ 获取头衔（必须是群）
func (u *QQ) Title(msgr *Messager) string {
	return u.memberInfo(msgr).Get(`title`).Str
}

// Card 从 QQ 获取群昵称（必须是群）
func (u *QQ) Card(msgr *Messager) string {
	return u.memberInfo(msgr).Get(`card`).Str
}

// NickName 从 QQ 获取昵称
func (u *QQ) NickName(msgr *Messager) string {
	return u.info(msgr).Get(`nickname`).Str
}

// TitleCardOrNickName 从 QQ 获取【头衔】群昵称 | 昵称（经过修剪）
func (u *QQ) TitleCardOrNickName(msgr *Messager) string {
	// 头衔
	var title string
	if NewQQGroup(msgr.Event.GroupID).IsGroup() {
		// 是群聊，获取该 QQ 在群内的头衔
		if title = u.Title(msgr); title != `` {
			// 如果头衔存在，则添加实心方头括号
			title = `【` + title + `】	`
		}
	}
	// 返回【头衔】群昵称 | 昵称
	return title + str.CleanAll(ctxCardOrNickName(msgr.Ctx, u.Int()), false)
}

// CallName 从 QQ 获取用于称呼的简单昵称（经过修剪）
func (u *QQ) CallName(msgr *Messager) (n string) {
	overLen := func(name string) bool {
		return len(name) > 1<<4 && utf8.RuneCountInString(name) > 1<<3
	}
	n = str.FirstText(str.CleanAll(u.Card(msgr), false))
	if n == `` || overLen(n) {
		n = str.FirstText(str.CleanAll(u.NickName(msgr), false))
	}
	if n == `` || overLen(n) {
		n = str.FirstText(str.CleanAll(u.Title(msgr), false))
	}
	return
}

// MemberList 获取特定群的成员列表
func (u *QQ) MemberList(msgr *Messager) groupList {
	if !msgr.Check(Caller) || !u.IsGroup() {
		// 没有 APICaller 或不是 QQ 群，无法获取
		return groupList{}
	}
	// 从缓存获取该群成员列表
	gmi, ok := groupMemberList.Load(*u)
	if !ok {
		// 如果获取不到，同步更新
		u.updateMemberList(msgr)
		gmi, _ = groupMemberList.Load(*u)
	}
	// 如果缓存已经过期，异步更新缓存的该群成员列表
	if time.Since(gmi.Time) > expire {
		go u.updateMemberList(msgr)
	}
	return gmi
}

// （私有）更新特定群的成员列表
func (u *QQ) updateMemberList(msgr *Messager) {
	groupMemberList.Store(*u,
		groupList{
			List: msgr.GetGroupMemberListNoCache(u.Int()).Array(),
			Time: time.Now(),
		},
	)
}

package msg

import (
	"fmt"
	"time"

	"golang.org/x/exp/constraints"

	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// ReplaceCard 替换当前 bot 的群昵称
func (handler *Handler) ReplaceCard(n string) error {
	o, err := handler.Object()
	if err != nil {
		return err
	}
	if err := o.SetName(handler, n); err != nil {
		return err
	}
	return Handle(handler.SetCard(-1))
}

// SetCardThisGroup 在本群设置自己的群昵称，h 为猫堆高度
func (handler *Handler) SetCardThisGroup(h ...int) zero.APIResponse {
	if len(h) == 0 {
		// 默认高度为 -1
		h = []int{-1}
	}
	if handler.Event().DetailType != Group {
		return zero.APIResponse{}
	}
	return handler.SetThisGroupCard(usr.Self(), handler.card(h[0]))
}

// SetCard 设置自己的群昵称
//
//	h 为猫堆高度
//	g 为发送对象，若无则使用当前发送对象
func (handler *Handler) SetCard(h int, g ...usr.QQ) (res zero.APIResponse) {
	if len(g) == 0 {
		// 当前群
		return handler.SetCardThisGroup(h)
	}
	for _, v := range g {
		if !v.IsGroup() {
			continue
		}
		res = handler.SetGroupCard(v, usr.Self(), handler.card(h))
	}
	return
}

// SetGroupCard 设置群名片（群备注）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#set_group_card-%E8%AE%BE%E7%BD%AE%E7%BE%A4%E5%90%8D%E7%89%87%E7%BE%A4%E5%A4%87%E6%B3%A8
func (handler *Handler) SetGroupCard(g, u usr.QQ, card string) zero.APIResponse {
	return handler.CallAction(`set_group_card`, zero.H{
		`group_id`: g.Int[int64](),
		`user_id`:  u.Int[int64](),
		`card`:     card,
	})
}

// SetThisGroupCard 设置本群名片（群备注）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#set_group_card-%E8%AE%BE%E7%BD%AE%E7%BE%A4%E5%90%8D%E7%89%87%E7%BE%A4%E5%A4%87%E6%B3%A8
func (handler *Handler) SetThisGroupCard(u usr.QQ, card string) zero.APIResponse {
	return handler.SetGroupCard(usr.NewQQGroup(handler.Event().GroupID), u, card)
}

// 根据发送对象生成群昵称
//
//	h 为猫堆高度
//	g 为发送对象，若无则使用当前发送对象
func (handler *Handler) card(h int, g ...usr.QQ) string {
	if len(g) == 0 {
		g = []usr.QQ{usr.NewQQGroup(handler.Event().GroupID)}
	}
	n, err := g[0].Name(handler)
	if err != nil {
		log.Error(err)
		return ``
	}
	return card(n, usr.Self().Age(handler), h)
}

// 生成群昵称，h 为猫堆高度
func card[T constraints.Integer](name string, age T, h int) string {
	return fmt.Sprintf(`%s（%d岁）（%s）`,
		name,
		max(int(age), equal.FullYears(time.Date(2017, time.April, 45,
			0, 0, 0, 0, time.Local), time.Now())),
		func() string {
			if h > 0 {
				return fmt.Sprintf(`猫堆高度：%d`, h)
			}
			if h == 0 {
				return `猫堆已清空`
			}
			return `内置冷却，禁止调戏`
		}(),
	)
}

// GroupInfo 获取群信息
func (handler *Handler) GroupInfo() string {
	info := handler.GetThisGroupInfo(false)
	return fmt.Sprintf(`群“%s”（%d），%d/%d`,
		info.Name, info.ID, info.MemberCount, info.MaxMemberCount)
}

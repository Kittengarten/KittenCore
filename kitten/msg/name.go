package msg

import (
	"cmp"
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
	handler.SetCard(-1)
	return nil
}

// SetCardThisGroup 在本群设置自己的群昵称，h 为猫堆高度
func (handler *Handler) SetCardThisGroup(h ...int) {
	if len(h) == 0 {
		// 默认高度为 -1
		h = []int{-1}
	}
	if handler.Event().DetailType == Group {
		handler.SetThisGroupCard(usr.Self().Int(), handler.card(h[0]))
	}
}

// SetCard 设置自己的群昵称
//
//	h 为猫堆高度
//	g 为发送对象，若无则使用当前发送对象
func (handler *Handler) SetCard(h int, g ...usr.QQ) {
	if len(g) == 0 {
		// 当前群
		handler.SetCardThisGroup(h)
		return
	}
	for _, v := range g {
		if !v.IsGroup() {
			continue
		}
		handler.SetGroupCard(v.Int(), usr.Self().Int(), handler.card(h))
	}
}

// SetGroupCard 设置群名片（群备注）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#set_group_card-%E8%AE%BE%E7%BD%AE%E7%BE%A4%E5%90%8D%E7%89%87%E7%BE%A4%E5%A4%87%E6%B3%A8
func (handler *Handler) SetGroupCard(groupID, userID int64, card string) {
	handler.CallAction(`set_group_card`, zero.H{
		`group_id`: groupID,
		`user_id`:  userID,
		`card`:     card,
	})
}

// SetThisGroupCard 设置本群名片（群备注）
// https://github.com/botuniverse/onebot-11/blob/master/api/public.md#set_group_card-%E8%AE%BE%E7%BD%AE%E7%BE%A4%E5%90%8D%E7%89%87%E7%BE%A4%E5%A4%87%E6%B3%A8
func (handler *Handler) SetThisGroupCard(userID int64, card string) {
	handler.SetGroupCard(handler.Event().GroupID, userID, card)
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
			switch cmp.Compare(0, h) {
			case -1:
				return fmt.Sprintf(`猫堆高度：%d`, h)
			case 0:
				return `猫堆已清空`
			case 1:
				return `内置冷却，禁止调戏`
			default:
				// 死码，为了防止编译器报 missing return 而保留
				return ``
			}
		}(),
	)
}

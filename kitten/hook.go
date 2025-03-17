package kitten

import (
	"log"
	_ "unsafe"

	"github.com/Kittengarten/KittenCore/kitten/core/str"

	"github.com/brahma-adshonor/gohook"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const funcCardOrNickName = `CardOrNickName`

// 备份 (*zero.Ctx).CardOrNickName
var ctxCardOrNickNameBackup = (*zero.Ctx).CardOrNickName

func init() {
	if err := gohook.HookMethod(
		&zero.Ctx{},
		funcCardOrNickName,
		ctxCardOrNickName,
		ctxCardOrNickNameBackup,
	); err != nil {
		log.Println(err)
	}
}

// hook (*zero.Ctx).CardOrNickName
func ctxCardOrNickName(ctx *zero.Ctx, uid int64) string {
	Info(`执行已 hook 的函数：`, funcCardOrNickName)
	var (
		msgr = New(ctx)
		u    = NewQQ(uid)
	)
	if !msgr.Check(Caller) || !u.IsQQ() {
		// 没有 APICaller 或不是 QQ，无法获取
		return ``
	}
	if NewQQGroup(msgr.Event.GroupID).IsGroup() {
		// 是群聊，获取修剪后的群昵称
		if card := str.CleanAll(u.memberInfo(msgr).Get(`card`).Str, false); card != `` {
			// 如果不为空，返回群昵称
			return card
		}
	}
	// 不是群聊或群昵称为空，返回修剪后的昵称
	return str.CleanAll(u.info(msgr).Get(`nickname`).Str, false)
}

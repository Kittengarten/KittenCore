package kitten

import (
	"fmt"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

/*
TextOf 格式化构建 message.Text 文本

格式同 fmt.Sprintf
*/
func TextOf(format string, a ...any) message.MessageSegment {
	return message.Text(fmt.Sprintf(format, a...))
}

/*
SendTextOf 发送格式化文本

lf 控制群聊的 @ 后是否换行
*/
func SendTextOf(ctx *zero.Ctx, lf bool, format string, a ...any) {
	switch ctx.Event.DetailType {
	case `private`:
		ctx.Send(TextOf(format, a...))
	case `group`, `guild`:
		if lf {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n"), TextOf(format, a...))
			return
		}
		ctx.SendChain(message.At(ctx.Event.UserID), TextOf(format, a...))
	}
}

/*
SendText 发送文本

lf 控制群聊的 @ 后是否换行
*/
func SendText(ctx *zero.Ctx, lf bool, text string) {
	switch ctx.Event.DetailType {
	case `private`:
		ctx.Send(text)
	case `group`, `guild`:
		if lf {
			ctx.SendChain(message.At(ctx.Event.UserID), message.Text("\n"), message.Text(text))
			return
		}
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(text))
	}
}

/*
SendMessage 发送消息

lf 控制群聊的 @ 后是否换行
*/
func SendMessage(ctx *zero.Ctx, lf bool, m ...message.MessageSegment) {
	switch ctx.Event.DetailType {
	case `private`:
		ctx.Send(m)
	case `group`, `guild`:
		var n []message.MessageSegment
		if lf {
			ctx.SendChain(append(append(append(n, message.At(ctx.Event.UserID)), message.Text("\n")), m...)...)
			return
		}
		ctx.SendChain(append(append(n, message.At(ctx.Event.UserID)), m...)...)
	}
}

// DoNotKnow 喵喵不知道哦
func DoNotKnow(ctx *zero.Ctx) {
	SendMessage(ctx, false, ImagePath.GetImage(`哈——？.png`), TextOf(`%s不知道哦`, zero.BotConfig.NickName[0]))
}

// Int64 QQ 号转换为整型
func (u QQ) Int64() int64 {
	return int64(u)
}

// GetTitle 从 QQ 获取【头衔】
func (u QQ) GetTitle(ctx *zero.Ctx) string {
	gmi := ctx.GetGroupMemberInfo(ctx.Event.GroupID, u.Int64(), true)
	if titleStr := gjson.Get(gmi.Raw, `title`).Str; `` == titleStr {
		return ``
	}
	return fmt.Sprintf(`【%s】`, gjson.Get(gmi.Raw, `title`).Str)
}

// （私有）获取信息
func (u QQ) getInfo(ctx *zero.Ctx) gjson.Result {
	return ctx.GetStrangerInfo(int64(u), true)
}

// IsAdult 是成年人
func (u QQ) IsAdult(ctx *zero.Ctx) bool {
	age := gjson.Get(u.getInfo(ctx).Raw, `age`).Int()
	return 18 <= age
}

// IsFemale 是女性
func (u QQ) IsFemale(ctx *zero.Ctx) bool {
	sex := gjson.Get(u.getInfo(ctx).Raw, `sex`).String()
	return `female` == sex
}

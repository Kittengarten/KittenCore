package essence

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/tidwall/gjson"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Kittengarten/KittenCore/kitten"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName = "精华消息"
	brief            = "获取精华消息"
	fail             = "获取精华消息失败喵！"
)

func init() {
	var (
		help = strings.Join([]string{"发送",
			fmt.Sprintf("%s精华", kitten.Configs.CommandPrefix),
			"随机抽取一条本群的精华消息",
		}, "\n")
		// 注册插件
		engine = control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Brief:            brief,
			Help:             help,
		})
	)

	engine.OnCommand("精华").Handle(func(ctx *zero.Ctx) {
		var (
			essenceList  = ctx.GetThisGroupEssenceMessageList()
			essenceCount = len(essenceList.Array())
		)
		if essenceCount == 0 {
			ctx.Send(fail)
			log.Error(fail)
		} else {
			var (
				IDx            = rand.Intn(int(essenceCount))
				essenceMessage = essenceList.Array()[IDx]
				ID             = gjson.Get(essenceMessage.Raw, "sender_id")
				nickname       = gjson.Get(essenceMessage.Raw, "sender_nick")
				msID           = gjson.Get(essenceMessage.Raw, "message_id")
				_              = ctx.GetGroupMessageHistory(ctx.Event.GroupID, msID.Int())
				ms             = ctx.GetMessage(message.NewMessageIDFromInteger(msID.Int()))
				report         = make(message.Message, len(ms.Elements))
				reportText     = kitten.TextOf("【精华消息】\n%s（%d）:\n", kitten.GetTitle(*ctx, ID.Int())+nickname.String(), ID.Int())
			)
			log.Tracef("获得了 %d 条精华消息喵！", essenceCount)
			log.Trace(essenceMessage)
			log.Trace(ms)
			report = append(report, reportText)
			report = append(report, ms.Elements...)
			ctx.Send(report)
		}
	})
}

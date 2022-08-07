package essence

import (
	"fmt"
	"math/rand"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
)

const (
	// ReplyServiceName 插件名
	ReplyServiceName = "精华消息"
	fail             = "获取精华消息失败喵！"
)

func init() {
	config := kitten.LoadConfig()
	help := strings.Join([]string{"发送",
		fmt.Sprintf("%s精华", config.CommandPrefix),
		"随机抽取一条本群的精华消息",
	}, "\n")
	// 注册插件
	engine := control.Register(ReplyServiceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             help,
	})

	engine.OnCommand("精华").Handle(func(ctx *zero.Ctx) {
		essenceList := ctx.GetThisGroupEssenceMessageList()
		essenceCount := len(essenceList.Array())
		if essenceCount == 0 {
			ctx.Send(fail)
			log.Error(fail)
		} else {
			log.Tracef("获得了 %d 条精华消息喵！", essenceCount)
			IDx := rand.Intn(int(essenceCount))
			essenceMessage := essenceList.Array()[IDx]
			log.Trace(essenceMessage)
			var (
				ID       = gjson.Get(essenceMessage.Raw, "sender_id")
				nickname = gjson.Get(essenceMessage.Raw, "sender_nick")
				msID     = gjson.Get(essenceMessage.Raw, "message_id")
			)
			ctx.GetGroupMessageHistory(ctx.Event.GroupID, msID.Int())
			ms := ctx.GetMessage(message.NewMessageIDFromInteger(msID.Int()))
			log.Trace(ms)
			reportText := kitten.TextOf("【精华消息】\n%s（%d）:\n", kitten.GetTitle(*ctx, ID.Int())+nickname.String(), ID.Int())
			report := make(message.Message, len(ms.Elements))
			report = append(report, reportText)
			report = append(report, ms.Elements...)
			ctx.Send(report)
		}
	})
}

// repeat 喵类的本质是复读姬
package repeat

import (
	"math/rand/v2"
	"strconv"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/core/msg/seg"

	"github.com/RomiChan/syncx"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName = `repeat`      // 插件名
	configFile       = `config.yaml` // 配置文件名
	cRepeat          = `复读`
	brief            = `喵类的本质是复读姬`
	maxThreshold     = 10 // MaxThreshold 最大复读阈值
)

type (
	// 消息统计
	stat struct {
		t               uint // 阈值
		message.Message      // 消息
	}

	// 复读姬配置
	config struct {
		Threshold uint    // 触发复读的阈值
		Chance    float64 // 触发复读的概率
	}
)

var (
	// 帮助
	help = kitten.MainConfig().CommandPrefix + cRepeat + ` [阈值] [概率]`
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	})
	// 配置文件路径
	configPath = fio.PathMutex{Path: fio.NewPath(engine.DataFolder(), configFile)}
	// 触发复读的阈值
	threshold uint = 2
	// 触发复读的概率
	chance = 0.5
	// 消息缓存
	m syncx.Map[kitten.QQ, stat]
)

func init() {
	repeatInit()

	// 复读设置
	engine.OnCommand(cRepeat, zero.SuperUserPermission).SetBlock(true).Handle(repeatSet)

	// 复读
	engine.OnMessage(zero.OnlyGroup).Handle(repeat)
}

func repeatInit() {
	repeatConfig, err := fio.Load[config](configPath.Path, "threshold: 2\nchance: 0.5") // 复读姬配置文件
	if err != nil {
		kitten.Error(`复读姬配置文件错误喵！`, err)
		return
	}
	threshold, chance = repeatConfig.Threshold, repeatConfig.Chance
}

// 复读设置
func repeatSet(ctx *zero.Ctx) {
	var (
		msgr = kitten.New(ctx)
		args = msgr.ArgsSlice()
	)
	if len(args) != 2 {
		msgr.SendWithImageFailf(`本命令参数数量：%d
传入的参数数量：%d`,
			2,
			len(args),
		)
		return
	}
	t, err := strconv.ParseUint(args[0], 10, core.PlatformBits)
	if err != nil {
		msgr.SendWithImageFailf(`[阈值] 错误：
%v
%s`,
			err,
			help,
		)
		return
	}
	if t > 2 || maxThreshold < t {
		msgr.SendWithImageFailf(`[阈值] 错误：最少为 2，最多为 %d 喵！`, maxThreshold)
		return
	}
	threshold = uint(t)
	if chance, err = strconv.ParseFloat(args[1], 64); err != nil {
		msgr.SendWithImageFailf(`[概率] 错误：
%v
%s`,
			err,
			help,
		)
		return
	}
	if chance < 0 {
		chance = 0
		msgr.SendWithImageFail(`[概率] 警告：不能 ＜ 0 喵！`)
		return
	}
	if chance > 1 {
		chance = 1
		msgr.SendWithImageFail(`[概率] 警告：不能 ＞ 1 喵！`)
		return
	}
	configPath.Lock()
	defer configPath.Unlock()
	err = fio.Save(configPath.Path, config{
		Threshold: threshold,
		Chance:    chance,
	})
	if err != nil {
		msgr.SendWithImageFail(`保存复读姬配置文件错误喵！`, err)
		return
	}
	o, err := msgr.Object()
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	n, err := o.Name()
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	msgr.Reply().AtLf().
		Textf(`%s将会开始以 %.2f%% 概率复读重复 %d 次的消息喵！`,
			n, 100*chance, threshold,
		).Send()
}

func repeat(ctx *zero.Ctx) {
	var (
		g     = kitten.NewQQGroup(ctx.Event.GroupID) // 群号
		c, ok = m.Load(*g)                           // 尝试获取本群的缓存
	)
	if ok && kitten.IsSameMessage(c.Message, ctx.Event.Message) {
		// 消息与缓存的内容一致
		if c.t == 0 {
			// 已经复读过，不再复读
			return
		}
		// 增加一次复读计数
		c.t++
	} else {
		// 如果没有缓存，或消息与缓存的内容不一致，初始化缓存
		c.t = 1
		c.Message = ctx.Event.Message
	}
	// 更新缓存
	m.Store(*g, c)
	//nolint:gosec
	if c.t < threshold || rand.Float64() >= chance {
		// 如果没有达到复读阈值，或者没有按概率触发复读，则返回
		return
	}
	// 处理图片
	c.handleImage(kitten.New(ctx))
	// 发送消息
	ctx.Send(c.Message)
	// 清空复读计数，避免再次复读
	c.t = 0
}

// 处理图片
func (s *stat) handleImage(msgr *kitten.Messager) {
	for i, e := range s.Message {
		if e.Type != seg.Image {
			// 不是图片，跳过
			continue
		}
		// 单张图片，进行一个图片的获取
		switch file := mio.GetImagePath(e); file {
		case ``:
			// 获取不到，跳过
			continue
		case `marketface`:
			// 市场表情
			s.Message[i] = kitten.Image(mio.GetImageURL(e))
		default:
			// 默认
			if len(e.Data[`summary`]) > 2 {
				// 表情包，跳过
				continue
			}
			s.Message[i] = kitten.Image(msgr.GetImage(file.String()).Get(`file`).String())
		}
	}
}

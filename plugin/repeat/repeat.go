// repeat 喵类的本质是复读姬
package repeat

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/msg/mio"
	"github.com/Kittengarten/KittenCore/kitten/msg/seg"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/goccy/go-yaml"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	replyServiceName = `repeat`      // 插件名
	cfgFile          = `config.yaml` // 配置文件名
	cRepeat          = `复读`
	brief            = `喵类的本质是复读姬`
	cThreshold       = `[阈值]`
	cChance          = `[概率]`
	minThreshold     = 2
	maxThresholdBits = 4
	maxThreshold     = 1 << maxThresholdBits // MaxThreshold 最大复读阈值
)

type (
	// 消息统计
	stat struct {
		message.Message        // 消息
		t               uint64 // 阈值
	}
	// 缓存
	cache = map[usr.QQ]stat

	// 复读姬配置
	cfg struct {
		Threshold uint64  // 触发复读的阈值
		Chance    float64 // 触发复读的概率
	}
)

var (
	// 帮助
	help = config.CommandPrefix() + strings.Join(
		[]string{cRepeat, cThreshold, cChance}, ` `)
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	})
)

var (
	// 配置文件路径
	cfgPath = fio.NewPath(engine.DataFolder(), cfgFile).WithMutex()
	// 触发复读的阈值
	threshold uint64 = 2
	// 触发复读的概率
	chance = 0.5
)

// 消息缓存
var m = struct {
	cache
	*sync.Mutex
}{
	cache: make(cache),
	Mutex: new(sync.Mutex),
}

func init() {
	repeatInit()

	// 复读设置
	engine.OnCommand(cRepeat, zero.SuperUserPermission).
		SetBlock(true).Handle(repeatSet)

	// 复读
	engine.OnMessage(zero.OnlyGroup,
		core.NotCommand,
		core.NotOnlyToMe,
	).Handle(repeat)
}

func repeatInit() {
	s := new(strings.Builder)
	s.Grow(24)
	if err := yaml.NewEncoder(s).Encode(cfg{
		Threshold: 2,
		Chance:    0.5,
	}); err != nil {
		log.Error(`复读姬配置文件初始化错误喵！`, err)
		return
	}
	repeatConfig, err := cfgPath.Load[cfg](s.String()) // 复读姬配置文件
	if err != nil {
		log.Error(`复读姬配置文件错误喵！`, err)
		return
	}
	threshold, chance = repeatConfig.Threshold, repeatConfig.Chance
}

// 复读设置
func repeatSet(ctx *zero.Ctx) {
	var (
		c, cancel = context.WithTimeout(context.Background(), core.Timeout)
		handler   = msg.NewWithContext(c, ctx)
		args      = handler.Args()
		threshold uint64
		chance    float64
	)
	defer cancel()
	n, err := fmt.Sscanln(args, &threshold, &chance)
	if err != nil {
		handler.SendWithImageFailf(`%s %s
参数错误（识别到 %d 个）：
%v`,
			cThreshold, cChance,
			n,
			err)
		return
	}
	if chance < minThreshold || chance > maxThreshold {
		handler.SendWithImageFailf(`[概率] 错误：必须是 [%d, %d] 的正整数喵！`,
			minThreshold, maxThreshold)
		return
	}
	if chance < 0 || chance > 1 {
		handler.SendWithImageFail(`[概率] 错误：必须位于 [0, 1] 喵！`)
		return
	}
	cfgPath.Lock()
	defer cfgPath.Unlock()
	if err = cfgPath.SaveWithContext(handler, cfg{
		Threshold: threshold,
		Chance:    chance,
	}); err != nil {
		handler.SendWithImageFail(`保存复读姬配置文件错误喵！`, err)
		return
	}
	o, err := handler.Object()
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	name, err := o.Name(c)
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	handler.Quote().AtLf().
		Textf(`%s将会开始以 %.2f%% 概率复读重复 %d 次的消息喵！`,
			name, 100*chance, threshold,
		).Send()
}

// 复读
func repeat(ctx *zero.Ctx) {
	m.Lock()
	defer m.Unlock()
	var (
		g     = usr.NewQQGroup(ctx.Event.GroupID) // 群号
		c, ok = m.cache[g]                        // 本群的缓存
	)
	if ok && msg.IsSame(c.Message, ctx.Event.Message) {
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
	defer func() { m.cache[g] = c }()
	//nolint:gosec
	if c.t < threshold || rand.Float64() >= chance {
		// 没有达到复读阈值，或者没有按概率触发复读，则返回
		return
	}
	// 处理图片
	var (
		co, cancel = context.WithTimeout(context.Background(), core.Timeout)
		handler    = msg.NewWithContext(co, ctx)
	)
	defer cancel()
	c.handleImage(handler)
	// @ 后面增加空格
	for i, m := range c.Message {
		if m.Type == seg.At {
			c.Message = slices.Insert(c.Message, i+1, message.Text(` `))
		}
	}
	// 发送消息
	handler.Set(c.Message).(*msg.Handler).Send()
	// 清空复读计数，避免再次复读
	c.t = 0
}

// 处理图片
func (s *stat) handleImage(handler *msg.Handler) {
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
			s.Message[i] = msg.Image(mio.GetImageURL(e))
		default:
			// 默认
			if len(e.Data[`summary`]) > 2 {
				// 表情，不需要处理
				continue
			}
			s.Message[i] = msg.Image(handler.GetImage(file.String()).Get(`file`).String())
		}
	}
}

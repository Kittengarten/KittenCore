// Package stack2 叠猫猫 v2
package stack2

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/kitten/usr"

	"github.com/goccy/go-yaml"

	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName                     = `stack2` // 插件名
	brief                                = `一起来玩叠猫猫 v2`
	dataFile                             = `data.yaml`   // 叠猫猫数据文件
	bufferFile                           = `buffer.yaml` // 叠猫猫缓存文件
	tipsFile                             = `tips.yaml`   // 叠猫猫小贴士文件
	cStack, cStackT0, cStackT1, cStackT2 = `叠`, `曡`, `疊`, `堆`
	cMeow                                = `猫猫`
	cIn                                  = `加入`
	cView                                = `查看`
	cAnalysis                            = `分析`
	cRank                                = `排行`
	cOC, cOCFox, cOCGPU, cOCCockroach    = `锻炼`, `化功`, `加速`, `起飞`
	cDaily, cDaily1                      = `日常`, `每日`
	cLorry                               = `大运`
	cEat                                 = `吃`
	cEatGPU                              = `抢`
	zako                                 = `zako`
)

// GlobalMessager 全局上下文，仅用于获取猫猫信息
var GlobalMessager = &msg.Handler{Context: context.Background()}

// 叠猫猫状态
var stackStatus status

func init() {
	if err != nil {
		log.Error(`叠猫猫配置文件错误喵！`, err)
		return
	}
	// 初始化小贴士
	if err := yaml.Unmarshal(tipYAML, &tipSlice); err != nil {
		log.Error(`小贴士初始化错误喵！`, err)
	}

	// 叠猫猫、吃猫猫
	engine.OnCommandGroup([]string{
		cStack, cStackT0, cStackT1, cStackT2,
		cEat, cEatGPU,
	}).SetBlock(true).
		Limit(rate.User.Get()).
		Limit(rate.GroupFast.Get()).
		Handle(func(ctx *zero.Ctx) {
			Lock()
			defer Unlock()
			c, cancel := context.WithTimeout(context.Background(), core.Timeout)
			defer cancel()
			switch handler := msg.NewWithContext(c, ctx); handler.Command() {
			case cStack, cStackT0, cStackT1:
				// 叠猫猫
				stackExe(handler)
			case cEat, cEatGPU:
				// 吃猫猫
				eatExe(handler)
			}
		})

	// 撞大运
	engine.OnKeyword(cLorry).SetBlock(true).Limit(rate.User.Get()).
		Limit(rate.GroupFast.Get()).
		Handle(func(ctx *zero.Ctx) {
			if !strings.HasPrefix(ctx.Event.RawMessage, p) {
				return
			}
			Lock()
			defer Unlock()
			c, cancel := context.WithTimeout(context.Background(), core.Timeout)
			defer cancel()
			lorryExe(msg.NewWithContext(c, ctx))
		})
}

// 叠猫猫执行逻辑
func stackExe(handler *msg.Handler) {
	var (
		args    = handler.Args()
		loc, op string
	)
	n, err := fmt.Sscanln(args, &loc, &op)
	if err != nil {
		asyncSendEmoji(handler, `❔ 问号`)
		handler.SendWithImageFailf(`%s%s%s %s|%s|%s|%s|%s|%s
参数错误（识别到 %d 个）：
%v`,
			p, cStack, cMeow, cIn, cView, cAnalysis, cRank, cOC, cDaily,
			n,
			err)
		return
	}
	if !setGlobalLocation(loc) {
		// 设置全局地区标记位，如当前活动未开放则返回
		asyncSendEmoji(handler, `辣眼睛`)
		handler.SendWithImageFail(`当前活动未开放喵！`)
		return
	}
	GlobalMessager.Ctx = handler.Ctx
	d, err := fio.LoadWithContext[data](handler, dataPath, fio.Empty)
	if err != nil {
		sendWithImageFail(handler, `加载叠猫猫数据文件时发生错误喵！`, err)
		return
	}
	stackStatus.refresh(handler, &d)
	switch op {
	case cIn:
		_ = d.in(handler)
		self(handler, d)
	case cView:
		asyncSendEmoji(handler, `暗中观察`)
		d.view(handler, zero.UserOrGrpAdmin(handler.Ctx))
		select {
		case <-times.RandDelayRange(time.Second, 2*time.Second):
			d.viewImage(handler)
			self(handler, d)
		case <-handler.Done():
			handler.SendWithImageFail(handler.Err())
		}
	case cAnalysis:
		asyncSendEmoji(handler, `仔细分析`)
		d.analysis(handler)
		_ = selfAnalysis(handler, d)
	case cRank:
		asyncSendEmoji(handler, `🎉 庆祝`)
		d.rank(handler)
		_ = selfRank(handler, d)
	case cOC, cOCFox, cOCGPU, cOCCockroach:
		asyncSendEmoji(handler, `奋斗`)
		d.oc(handler)
		self(handler, d)
	case cDaily, cDaily1:
		asyncSendEmoji(handler, `💝 礼物`)
		d.daily(handler)
		self(handler, d)
	default:
		asyncSendEmoji(handler, `吃糖`)
		var (
			u    = usr.NewQQ(handler.Event().UserID)
			m, i = d.getMeow(u)
			w    = func() int {
				if i != -1 {
					return m.Weight
				}
				return len(u.TitleCardOrNickName(handler))
			}()
		)
		helpText := []string{help}
		if m.getTypeID(handler) >= 小老虎 {
			// 老虎以上，发送吃猫猫帮助文本
			helpText = append(helpText, helpEat)
		}
		if m.getTypeID(handler) >= 猫车 {
			// 猫车以上，发送撞大运帮助文本
			helpText = append(helpText, helpLorry)
		}
		sendText(handler, true, strings.NewReplacer(
			`(抱枕突破所需体重/当前体重)`,
			fmt.Sprintf(` %.2f%% `, 100*chanceFlat(m)),
			`[当前猫堆高度]`,
			strconv.Itoa(len(d.getStack())),
			`N(0, 体重²)`,
			fmt.Sprintf(`N(0, (%s)²)`, times.ConvertTimeDuration(
				time.Hour*time.Duration(stackConfig.RestHoursPerKG*w)/10,
			)),
			`N(0, (e*体重)²)`,
			fmt.Sprintf(`N(0, (%s)²)`, times.ConvertTimeDuration(
				time.Duration(
					float64(stackConfig.RestHoursPerKG)*float64(time.Hour)*math.E*i2f(w),
				))),
			`[最大休息时间]`,
			times.ConvertTimeDuration(stackStatus.MaxRestTime).String(),
		).Replace(strings.Join(helpText, "\n\n")))
	}
}

// 设置全局地区标记位
func setGlobalLocation(loc string) bool {
	switch {
	case strings.ContainsAny(loc, `狐狸`):
		globalLocation = fox // 狐狐
		return true
	case strings.Contains(loc, `显卡`):
		globalLocation = gpu // 显卡
		return true
	case strings.ContainsAny(loc, `蟑螂`),
		strings.ContainsAny(loc, `蜚蠊`),
		strings.Contains(loc, `小强`):
		globalLocation = cockroach // 蟑螂
		return checkCockroachDate()
	case strings.ContainsAny(loc, `猫虎喵貓`):
		fallthrough // 猫猫
	default:
		globalLocation = cat // 默认叠猫猫
		return true
	}
}

// 加入叠猫猫，当且仅当叠猫猫失败时返回的是 *stackErr
//
//	错误已经打印，无需重复打印
//	会修改原数据
func (d *data) in(handler *msg.Handler) error {
	// 初始化自身
	k, err := d.pre(handler)
	if err != nil {
		return err
	}
	// 未在叠猫猫的队列
	dn := d.getNoStack()
	// 执行叠猫猫
	e := d.doStack(handler, &k)
	// 合并当前未叠猫猫与叠猫猫的队列，将叠入的猫猫追加入切片中
	*d = slices.Concat(dn, *d, data{k})
	// 清理过期玩家
	d.clear(handler, false)
	// 存储叠猫猫数据
	if err = fio.SaveWithContext(handler, dataPath, d); err != nil {
		sendWithImageFail(handler, `存储叠猫猫数据时发生错误喵！`, err)
		return err
	}
	return e
}

// 叠猫猫尝试加入前的初始化，返回叠入的猫猫
//
//	如果不用于叠入，则需要克隆切片
//	错误已经打印，无需重复打印
func (d *data) pre(handler *msg.Handler) (meow, error) {
	var (
		u = usr.NewQQ(handler.Event().UserID) // 叠入猫猫的 QQ
		w int                              // 叠入猫猫的体重
		r time.Duration                    // 剩余的休息时间
		b bool                             // 日常任务是否完成
	)
	if i := slices.IndexFunc(*d, func(m meow) bool {
		now := time.Unix(handler.Event().Time, 0)
		r = m.Time.Sub(now)
		w = m.Weight
		b = equal.IsSameDate4AM(m.Daily, now)
		return u.Int() == m.Int() && !m.Status && 0 < r
	}); i >= 0 {
		err := needRest(r, w, i, b)
		if usr.Self() == u {
			stat.Weight = w
			return meow{}, err
		}
		sendWithImageFail(handler, err)
		return meow{}, err
	}
	if slices.ContainsFunc(*d, func(m meow) bool {
		return u.Int() == m.Int() && m.Status
	}) {
		err := alreadyJoined()
		if usr.Self() == u {
			stat.Weight = w
			return meow{}, err
		}
		sendWithImageFail(handler, err)
		return meow{}, err
	}
	var (
		name = u.TitleCardOrNickName(handler) // 叠入猫猫的名称
		m, i = d.getMeow(u)                // 获取叠入的猫猫及其索引，如果不用于叠入，则需要克隆切片
	)
	if i == -1 {
		// 如果是首次叠猫猫
		m = meow{
			QQ:     u,
			Name:   name,
			Weight: max(1, len(name)),
			Time:   time.Unix(handler.Event().Time, 0),
		}
		return m, nil
	}
	// 如果是已经存在的猫猫，更新其名称
	m.Name = name
	return m, nil
}

// 执行叠猫猫，k 为叠入的猫猫
//
//	错误已经打印，无需重复打印
func (d *data) doStack(handler *msg.Handler, m *meow) error {
	*d = d.getStack() // 正在叠猫猫的队列
	var (
		dr = slices.Clone(*d) // 叠猫猫队列的克隆
		l  = len(dr)          // 叠猫猫队列高度
	)
	if d.checkFlat(*m) {
		// 如果平地摔
		err := stack(handler, m, l, 0, flat)
		asyncSendEmoji(handler, `糗大了`)
		sendWithJump(handler, err)
		return err
	}
	if p := d.pressResult(handler, *m); p != 0 {
		// 压坏了别的猫猫
		var (
			err = stack(handler, m, l, p, press)
			e   = dr[:p]
		)
		asyncSendEmoji(handler, `晕`)
		sendWithPressed(handler, err, &e)
		return err
	}
	// 如果没有猫猫被压坏，叠猫猫初步成功
	if f := d.fallResult(handler, *m); f != 0 {
		// 摔坏了别的猫猫
		var (
			err = stack(handler, m, l, f, fall)
			e   = dr[l-f:]
		)
		asyncSendEmoji(handler, `😓 汗`)
		sendWithZako(handler, err, &e)
		return err
	}
	// 如果没有摔坏猫猫，叠猫猫成功
	m.Status = true
	asyncSendEmoji(handler, `爱心`)
	_ = sendWithNoPressedf(handler, `叠猫猫成功，目前处于队列中第 %d 位喵～
你的当前体重为 %.1f kg。`,
		l+1,
		i2f(m.Weight))
	utils.Go(`叠猫猫成功设置群昵称`, func() { setCard(handler, l+1) })
	return nil
}

// 获取并返回叠猫猫队列，不影响原切片
func (d *data) getStack() data {
	// 删除未在叠猫猫中的猫猫，得到叠猫猫队列
	return slices.DeleteFunc(slices.Clone(*d), func(m meow) bool { return !m.Status })
}

// 获取并返回未在叠猫猫的队列，不影响原切片
func (d *data) getNoStack() data {
	// 删除叠猫猫中的猫猫，得到未在叠猫猫的队列
	return slices.DeleteFunc(slices.Clone(*d), func(m meow) bool { return m.Status })
}

// 清空猫堆特效
//
//	根据是否清空猫堆，添加提示语
//	l 为队列高度
//	n 为结果
//	w 为叠猫猫前的体重
func doClear(handler *msg.Handler, l, n int, w int, m *meow, s *strings.Builder) {
	s.Grow(256)
	s.WriteByte('\n')
	if n == l {
		// 清空了猫堆
		switch hasClear(m) {
		case true:
			// 触发了特效
			asyncSendEmoji(handler, `👏 鼓掌`)
			fmt.Fprintln(s, `你触发了清空猫堆的特效！`)
		case false:
			// 没有触发特效
			asyncSendEmoji(handler, `哈欠`)
			fmt.Fprintln(s, `你清空了猫堆，但没有发生特别的事情。`)
		}
	}
	// 体重变化
	if m.Weight == w {
		fmt.Fprintf(s, "你的体重为 %.1f kg 不变。\n", i2f(w))
		return
	}
	fmt.Fprintf(s, "你的体重由 %.1f kg 变为 %.1f kg。\n", i2f(w), i2f(m.Weight))
}

// 提取猫猫及其索引，会从切片中删除提取的猫猫
//
//	无此猫猫则返回空结构体及 -1
func (d *data) getMeow(u usr.QQ) (meow, int) {
	i := slices.IndexFunc(*d, func(m meow) bool {
		return u.Int() == m.Int()
	})
	if i == -1 {
		return meow{}, i
	}
	m := (*d)[i]
	*d = slices.Delete(*d, i, i+1)
	return m, i
}

// Format 实现 fmt.Formatter 返回叠猫猫字符串
//
//	%s 完整字符串
//	默认 省略过的字符串
func (d *data) Format(f fmt.State, verb rune) {
	switch verb {
	case 's':
		fmt.Fprint(f, d.String())
	case '略':
		fmt.Fprint(f, d.Str())
	default:
		type raw data
		fmt.Fprintf(f, fmt.FormatString(f, verb), raw(*d))
	}
}

// String 实现 fmt.Stringer
//
//	从叠猫猫队列生成完整字符串（开头有一次换行）
func (d *data) String() string {
	// 克隆一份防止修改源数据
	dr := slices.Clone(*d)
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dr)
	s := new(strings.Builder)
	s.Grow(32 * len(dr))
	for _, k := range dr {
		fmt.Fprint(s, "\n", k)
	}
	return s.String()
}

// 从叠猫猫队列生成省略过的字符串
//
//	队列高度不超过 20 时，无需省略
func (d *data) Str() string {
	var (
		dr = slices.Clone(*d) // 克隆一份防止修改源数据
		l  = len(dr)          // 叠猫猫队列高度
		s  = new(strings.Builder)
		ok bool
	)
	s.Grow(32 * min(l, 20))
	// 按“后来居上”排列叠猫猫队列
	slices.Reverse(dr)
	for i, k := range dr {
		if l > 20 && 5 <= i && i < l-5 {
			// 当高度 > 20 时，跳过中间的猫猫，只取上下 5 只
			if ok {
				continue
			}
			s.WriteString("\n…………\n")
			for range l - 10 {
				s.WriteRune('🐱')
			}
			s.WriteString("\n…………")
			ok = true
			continue
		}
		s.WriteByte('\n')
		fmt.Fprint(s, k)
	}
	return s.String()
}

// 获取被压坏猫猫的数量，并将被压坏的猫猫标记为未在叠猫猫
//
//	不含叠入的猫猫
//	正在叠猫猫的队列才能调用
func (d *data) pressResult(handler *msg.Handler, m meow) int {
	var (
		s = append(*d, m) // 将叠入的猫猫纳入队列重量计算
		l = len(*d)       // 原队列高度
	)
	for i := range *d {
		n := &(*d)[i]
		if a := s[i:]; a.checkPress(handler) {
			// 没有被压坏，则直接返回
			return i
		}
		if i == 0 && n.getTypeID(handler) >= 猫车 {
			// 猫车为底座且被压坏，清空整个猫堆（载猫猫被动）
			for j := range *d {
				exit(handler, &(*d)[j], pressed, l-j)
			}
			return l
		}
		// 去除压坏的猫猫
		exit(handler, n, pressed, l-i)
		// 如果压坏的是猫娘萝莉以上，则不会继续压坏上方的猫猫
		if n.getTypeID(handler) >= 猫娘萝莉 {
			return i + 1
		}
	}
	return l
}

// 检查最下方的猫猫是否被压坏
//
//	如果没有被压坏则返回 true
func (d *data) checkPress(handler *msg.Handler) bool {
	//nolint:gosec
	return rand.Float64() >= d.chancePressed(handler)
}

// 获取最下方的猫猫被压坏的概率
func (d *data) chancePressed(handler *msg.Handler) float64 {
	if len(*d) <= 1 {
		// 如果只有一只猫猫或者没有猫猫，直接返回，避免索引越界
		return 0
	}
	a := (*d)[1:]
	switch (*d)[0].getTypeID(handler) {
	case 小老虎, 大老虎:
		// 如果是老虎，压坏的概率不同
		return min(1, float64(a.totalWeight())/math.Pow(math.E, math.E)/
			float64((*d)[0].Weight))
	}
	// 常规压坏概率
	return max(0, (float64(a.totalWeight())-math.E*float64((*d)[0].Weight))/
		float64(d.totalWeight()))
}

// 获取全队列的总重量（0.1 kg 数）
func (d *data) totalWeight() (w int) {
	for _, m := range *d {
		if math.MaxInt-m.Weight < w {
			// 防止溢出
			return math.MaxInt
		}
		w += m.Weight
	}
	return
}

// 检查是否平地摔，正在叠猫猫的队列才能调用
func (d *data) checkFlat(m meow) bool {
	// 当叠猫猫队列为空， 抱枕突破所需体重/当前体重的概率平地摔
	//nolint:gosec
	return len(*d) == 0 && rand.Float64() < chanceFlat(m)
}

// 获取叠猫猫失败摔下去猫猫的数量，并将摔下去的猫猫标记为未在叠猫猫
//
//	不含叠入的猫猫
//	正在叠猫猫的队列才能调用
func (d *data) fallResult(handler *msg.Handler, m meow) int {
	// 初始猫猫数量
	l := len(*d)
	if l == 0 || m.getTypeID(handler) <= 抱枕 ||
		(*d)[l-1].getTypeID(handler) >= 幼年猫娘 ||
		(*d)[0].getTypeID(handler) >= 猫车 {
		// 抱枕及以下的猫猫不会导致猫猫摔下去
		// 直接在猫娘以上级别的身上叠猫猫不会摔下去
		// 底座为猫车以上时，不会摔下去
		return 0
	}
	// 从队列的最上部开始遍历（后来居上）
	for i := range *d {
		// 下方的猫猫
		n := &(*d)[l-i-1]
		if m.checkFall(*n) {
			// 这只猫猫没有摔下去，直接返回
			return i
		}
		m = *n
		// 去除摔下去的猫猫
		exit(handler, n, fall, l-i)
		if n.getTypeID(handler) >= 猫娘少女 {
			// 如果摔下去的是猫娘少女以上级别，则下方的猫猫不会继续摔下去
			return i + 1
		}
	}
	return l
}

// 去除退出的猫猫 k，并使其进入休息，然后调整体重
//
//	r 为退出原因
//	h 为 摔下去的高度 | 压坏的猫猫总数 | 上方的猫猫总数 | 吃掉的猫猫总重量（0.1 kg 数）
func exit(handler *msg.Handler, m *meow, r result, h int) {
	// 去除
	m.Status = false
	// 计算休息时间（纳秒）
	rest := float64(time.Hour) * float64(stackConfig.RestHoursPerKG) * normal(i2f(m.Weight))
	// 体重变化
	switch r {
	case flat:
		// 平地摔，体重变为 e 倍
		w := int(math.RoundToEven(math.E * float64(m.Weight)))
		m.Weight = max(w, -(w + 1))
	case fall:
		// 摔下去，体重 - 100g × 当前高度
		m.Weight -= h
		if m.Weight < 1 {
			// 如果体重应归零或为负，休息时间增加至补偿量（0.1 kg 数）的 e^e 倍
			rest *= float64(1-m.Weight) * math.Pow(math.E, math.E)
			m.Weight = 1
		}
	case pressed, lorry, fly:
		if m.getTypeID(handler) < 猫车 {
			// 被压坏、撞飞，体重 + 100g × 上方的猫猫总数
			m.Weight = min(m.Weight, math.MaxInt-h) + h
			break
		}
		// 猫车以上，休息时间为 e 倍，体重增加更多
		h = f2i(normal(i2f(m.Weight))/100) * h
		fallthrough
	case eat:
		// 吃猫猫的休息时间为 e 倍
		rest *= math.E
		// 吃猫猫，体重 + 吃掉的猫猫总重量（0.1 kg 数）
		fallthrough
	case press:
		// 压坏了猫猫，体重 + 100g × 压坏的猫猫总数
		m.Weight = min(m.Weight, math.MaxInt-h) + h
	case eaten:
		// 被吃掉，休息时间和体重不变
	default:
		// 未知原因，直接返回
		return
	}
	// 被老虎吃掉，体重不变
	// 进入休息
	m.Time = time.Unix(handler.Event().Time, 0).
		Add(min(stackStatus.MaxRestTime, max(
			time.Hour*time.Duration(stackConfig.MinRestHours),
			time.Duration(rest),
		)))
}

// 清空猫堆的体重调整
func hasClear(m *meow) bool {
	//nolint:gosec
	if rand.Float64() >= float64(mapMeow[抱枕].weight)/float64(m.Weight) {
		return false
	}
	// 以抱枕突破所需体重/当前体重的概率，体重变为 e 倍
	w := int(math.RoundToEven(math.E * float64(m.Weight)))
	m.Weight = max(w, -(w + 1))
	return true
}

// 计算猫池中位数重量
func (d *data) median(handler *msg.Handler) {
	// 克隆一份防止修改原始数据
	dr := slices.Clone(*d)
	// 不活跃的猫猫不参与计算，但数据保留
	dr.clear(handler, true)
	// 猫池容量
	l := len(dr)
	if l == 0 {
		stackStatus.MedianWeight = 0
		return
	}
	// 按猫猫重量排序
	slices.SortStableFunc(dr, func(m, n meow) int {
		return cmp.Compare(m.Weight, n.Weight)
	})
	if l%2 != 0 {
		stackStatus.MedianWeight = dr[l/2].Weight
		return
	}
	stackStatus.MedianWeight = (dr[l/2-1].Weight + dr[l/2].Weight) / 2
}

// 清理 || 忽略过期玩家，范围为休息完毕的绒布球，以及超期的奶猫
//
//	行为取决于是否回写
//	如果 ignore 为 true，则忽略不活跃玩家（禁止回写）
func (d *data) clear(handler *msg.Handler, ignore bool) {
	var del int // 删除的猫猫数量
	for i := range *d {
		(*d)[i-del] = (*d)[i] // 移动猫猫以填充删除后的空隙
		if (*d)[i].Status {
			// 如果在叠猫猫中，不处理
			continue
		}
		if (*d)[i].getTypeID(handler) == 绒布球 && /* 如果是绒布球，且不在休息 */
			(*d)[i].Time.Before(time.Unix(handler.Event().Time, 0)) ||
			(ignore || 奶猫 == (*d)[i].getTypeID(handler)) && /* 如果忽略不活跃玩家或是奶猫，且已经超期 */
				(*d)[i].Time.Add(stackStatus.MaxRestTime).Before(time.Unix(handler.Event().Time, 0)) {
			del++ // 执行清理
			continue
		}
		if (*d)[i].Time.Sub(time.Unix(handler.Event().Time, 0)) > stackStatus.MaxRestTime {
			// 如果猫猫剩余的休息时间大于当前上限，缩短至上限
			(*d)[i].Time = time.Unix(handler.Event().Time, 0).Add(stackStatus.MaxRestTime)
		}
	}
	*d = (*d)[:len(*d)-del] // 清除掉经过移动后失效的猫猫
}

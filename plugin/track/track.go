// Package track 小说更新播报、小说信息查询、小说更新查询
package track

import (
	"cmp"
	"fmt"
	"log"
	stdhttp "net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/http"
	"github.com/Kittengarten/KittenCore/kitten/core/io"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/rate"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	replyServiceName = `track` // 插件名
	brief            = `小说报更`
	configFile       = `config.yaml` // 配置文件名
	pf               = `[平台]`        // [#平台] 不使用 API
	ag               = `[关键词|书号]`
	cNovel           = `小说`
	cNovelComment    = `点评小说`
	cUpdateTest      = `更新测试`
	cUpdatePreview   = `更新预览`
	cAddUpdate       = `添加报更`
	cCancelUpdate    = `取消报更`
	cQueryUpdate     = `查询报更`
	without          = `这里没有添加小说报更喵～`
	errConfig        = `报更配置文件错误喵！`
	errLoad          = `加载` + errConfig
	errSave          = `保存` + errConfig
	unknown          = `未知`
)

var (
	// 指令前缀
	p = kitten.MainConfig().CommandPrefix
	// 帮助
	help = p + cNovel + ` ` + pf + ` ` + ag + ` // 可获取信息
` + p + cNovelComment + ` ` + pf + ` ` + ag + ` // 可获取带点评的信息
` + p + cUpdatePreview + ` ` + pf + ` ` + ag + ` // 可预览更新内容
` + p + cQueryUpdate + ` // 可查询当前小说自动报更
————
管理员或私聊可用：
` + p + cAddUpdate + ` ` + pf + ` ` + ag + ` // 可添加小说自动报更
` + p + cCancelUpdate + ` ` + pf + ` ` + ag + ` // 可取消小说自动报更
` + p + cUpdateTest + ` ` + pf + ` ` + ag + ` // 可测试报更功能`
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             brief,
		Help:              help,
		PrivateDataFolder: replyServiceName,
	}).ApplySingle(ctxext.DefaultSingle)
	// 配置文件路径
	configPath = io.NewPath(engine.DataFolder(), configFile)
	// 报更更新的信号
	cu = make(chan books)
	// 读写锁
	mu sync.RWMutex
)

func init() {
	go track()

	// 更新测试
	engine.OnCommand(cUpdateTest, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			updateTest(kitten.New(ctx))
		})

	engine.OnCommandGroup([]string{
		cNovel,
		cNovelComment,
		cUpdatePreview,
	}).SetBlock(true).
		Limit(rate.Get(rate.User)).
		Limit(rate.Get(rate.GroupNormal)).
		Handle(func(ctx *zero.Ctx) {
			switch msgr := kitten.New(ctx); msgr.Command() {
			case cNovel:
				// 小说信息
				novelInfo(msgr, false)
			case cNovelComment:
				// 点评小说
				novelInfo(msgr, true)
			case cUpdatePreview:
				// 更新预览
				updatePreview(msgr)
			}
		})

	engine.OnCommandGroup([]string{
		cAddUpdate,
		cCancelUpdate,
	}, zero.UserOrGrpAdmin).SetBlock(true).
		Limit(rate.Get(rate.GroupFast)).
		Handle(func(ctx *zero.Ctx) {
			switch msgr := kitten.New(ctx); msgr.Command() {
			case cAddUpdate:
				// 添加报更
				add(msgr)
			case cCancelUpdate:
				// 取消报更
				cancel(msgr)
			}
		})

	// 查询报更
	engine.OnCommand(cQueryUpdate).SetBlock(true).
		Limit(rate.Get(rate.GroupSlow)).Handle(func(ctx *zero.Ctx) {
		query(kitten.New(ctx))
	})
}

// 更新测试
func updateTest(msgr *kitten.Messager) {
	nv, err := getNovel(msgr)
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	o, err := msgr.Object()
	if err != nil {
		kitten.Error(err)
	}
	go tryCommentUpdate(
		msgr,
		msgr.Reply().AtLf().
			Image(
				io.NewPath(nv.coverURL),
				io.NewPath(nv.headURL),
			).
			Text(nv.update()).
			SendMulti(),
		[]kitten.QQ{*o},
		nv,
		getChapterID(nv.Platform, nv.newChapter.url),
	)
}

// 更新预览
func updatePreview(msgr *kitten.Messager) {
	n, err := getNovel(msgr)
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	if r := n.preview; r != `` {
		msgr.Reply().AtLf().Text(`《`, n.name, `》
`, &n.newChapter, `
`, r).Send()
		return
	}
	msgr.SendWithImageFail(`不存在的喵！`)
}

// Comment 评论
var Comment = func(nv fmt.Stringer) string {
	// 默认为空实现
	return ``
}

// 小说信息
func novelInfo(msgr *kitten.Messager, comment bool) {
	nv, err := getNovel(msgr)
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	msgr = msgr.Reply().AtLf().
		Image(io.NewPath(nv.coverURL)).
		Text(&nv)
	if comment {
		msgr.Text(Comment(&nv))
	}
	msgr.Send()
}

// 添加报更
func add(msgr *kitten.Messager) {
	o, err := msgr.Object() // 发送对象
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	c, err := io.Load[books](configPath, io.Empty) // 报更配置
	if err != nil {
		msgr.SendWithImageFail(errLoad, err)
		return
	}
	nv, err := getNovel(msgr) // 小说实例
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	if i := slices.IndexFunc(c, func(b book) bool {
		return b.Platform == nv.Platform && b.BookID == nv.id
	}); i == -1 {
		// 没有该小说，新建并添加
		c = append(c, book{
			Platform: func() Platform {
				if nv.Platform == FQAPI {
					return FQ
				}
				return nv.Platform
			}(),
			BookID:   nv.id,
			BookName: nv.name,
			Writer:   nv.writer,
			Users:    []kitten.QQ{*o},
		})
	} else {
		// 已经有该小说
		if slices.Contains(c[i].Users, *o) {
			// 已有该用户，无需添加
			msgr.SendWithImageFail(`《`, nv.name, `》已经添加报更了喵！`)
			return
		}
		// 尚无该用户，需要添加
		c[i].Users = append(c[i].Users, *o)
		slices.Sort(c[i].Users)
	}
	if err := c.saveConfig(); err != nil {
		msgr.SendWithImageFail(`添加《`, nv.name, `》时`, errSave, err)
		return
	}
	msgr.Reply().AtLf().Text(`添加《`, nv.name, `》报更成功喵！`).Send()
}

// 取消报更
func cancel(msgr *kitten.Messager) {
	o, err := msgr.Object() // 发送对象
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	c, err := io.Load[books](configPath, io.Empty) // 报更配置
	if err != nil {
		msgr.SendWithImageFail(errLoad, err)
		return
	}
	if len(c) == 0 {
		msgr.Reply().AtLf().Text(without).Send()
		return
	}
	nv, err := getNovel(msgr) // 小说实例
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	// 本书下标
	i := slices.IndexFunc(c, func(b book) bool {
		return b.Platform == nv.Platform && b.BookID == nv.id
	})
	if i == -1 {
		msgr.SendWithImageFail(`未在追更《`, nv.name, `》喵！`)
		return
	}
	// 用户下标
	uid := slices.Index(c[i].Users, *o)
	if uid == -1 {
		msgr.SendWithImageFail(`未在追更《`, nv.name, `》喵！`)
		return
	}
	// 移除在当前发送对象的报更
	if len(slices.Delete(c[i].Users, uid, uid+1)) == 0 {
		// 如果移除后，不再有报更对象（也可能本来就没有报更对象），则整体移除该小说
		c = slices.Delete(c, i, i+1)
	}
	if err := c.saveConfig(); err != nil {
		msgr.SendWithImageFail(`取消《`, nv.name, `》时`, errSave, err)
		return
	}
	msgr.Reply().AtLf().Text(`取消《`, nv.name, `》报更成功喵！`).Send()
}

// 查询报更
func query(msgr *kitten.Messager) {
	o, err := msgr.Object() // 发送对象
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	mu.RLock()
	c, err := io.Load[books](configPath, io.Empty) // 报更配置
	mu.RUnlock()
	if err != nil {
		msgr.SendWithImageFail(errLoad, err)
		return
	}
	if len(c) == 0 {
		msgr.Reply().AtLf().Text(without).Send()
		return
	}
	const h = `【报更列表】`
	var r strings.Builder
	r.Grow(64 * len(c))
	r.WriteString(h)
	for _, b := range c {
		if !slices.Contains(b.Users, *o) {
			// 如果本书不在这里报更，则直接遍历至下一本书
			continue
		}
		r.WriteByte('\n')
		fmt.Fprint(&r, b)
	}
	msgr.Reply().AtLf().Text(&r).Send()
}

// 平台匹配器
func getPlatform(keyword string) Platform {
	switch {
	case strings.ContainsAny(keyword, `刺猬猫客`),
		strings.Contains(strings.ToUpper(keyword), `CWM`),
		strings.Contains(strings.ToLower(keyword), `ciweimao`):
		return CWM
	case strings.ContainsAny(keyword, `菠萝包`),
		strings.Contains(strings.ToUpper(keyword), `SF`),
		strings.Contains(strings.ToLower(keyword), `blb`):
		return SF
	case strings.ContainsAny(keyword, `番茄柿`),
		strings.Contains(strings.ToUpper(keyword), `FQ`):
		if strings.HasPrefix(keyword, `#`) {
			return FQ
		}
		return FQAPI
	default:
		return Platform(keyword)
	}
}

/*
获取小说

如果传入值不为书号，则先获取书号
*/
func getNovel(msgr *kitten.Messager) (novel, error) {
	args := msgr.ArgsSlice()
	if len(args) != 2 {
		return novel{}, fmt.Errorf(`本命令参数数量：2
%s %s
传入的参数数量：%d
参数数量错误喵！`,
			pf, ag,
			len(args))
	}
	var (
		p      = getPlatform(args[0])
		bookID = args[1]
	)
	if _, err := strconv.Atoi(bookID); err != nil {
		// 获取小说时，参数字符串无法转换为书号，尝试作为搜索关键词
		if bookID, err = keyword(args[1]).findBookID(p); err != nil {
			return novel{}, fmt.Errorf(`关键词 %s 搜索时发生错误：%w`, args[1], err)
		}
	}
	nv := *novelPool.Get().(*novel)
	defer novelPool.Put(&nv)
	return nv, nv.init(p, bookID)
}

// 报更
func track() {
	// 处理 panic，防止程序崩溃
	defer func() {
		if err := recover(); err != nil {
			kitten.Errorln(replyServiceName, `协程出现错误喵！`, err, string(debug.Stack()))
		}
	}()
	// 初始化报更配置文件
	if err := configPath.InitFile(io.Empty); err != nil {
		kitten.Error(`初始化报更配置文件时发生错误喵！`, err)
		return
	}
	mu.RLock()
	data, err := io.Load[books](configPath, io.Empty)
	mu.RUnlock()
	if err != nil {
		kitten.Error(errLoad, err)
		return
	}
	log.Printf(`======================[%s]======================
* OneBot + ZeroBot + Go
一共有 %d 本小说
=======================================================
`,
		kitten.MainConfig().NickName[0],
		len(data))
	func() {
		process.GlobalInitMutex.Lock()
		defer process.GlobalInitMutex.Unlock()
	}()
	var (
		t   = time.NewTicker(http.TimeOutSeconds * time.Second) // 定期检测，间隔为超时时间
		st  = time.NewTicker(http.TimeOutSeconds * time.Minute) // 专用慢速时钟
		sid = kitten.Self()
		bot = kitten.New(zero.GetBot(sid.Int()))
	)
	if !bot.Check(kitten.Caller) {
		kitten.Fatal(`获取 Bot 实例失败喵！`, bot)
	}
	// 报更
	for {
		select {
		case data = <-cu: // 接收到更新配置则使用
		case <-t.C: // 接收到时钟信号则释放
		}
		data.report(bot, st) // 执行报更
	}
}

// 执行报更
func (c *books) report(msgr *kitten.Messager, st *time.Ticker) {
	// 从小说池初始化小说
	nv := *novelPool.Get().(*novel)
	// 将小说重新收回小说池
	defer novelPool.Put(&nv)
	for i, b := range *c {
		times.RandomDelayRange(http.TimeOutSeconds*time.Second,
			2*http.TimeOutSeconds*time.Second)
		switch b.Platform {
		case FQ:
			res, err := stdhttp.Get(fqAPIHOST)
			if err != nil {
				// API 无法访问，使用网页模式
				// 接收到专用的慢速定时器信号才释放
				<-st.C
				break
			}
			if res != nil && res.StatusCode == 200 {
				res.Body.Close()
				// API 可以访问，切换为 API 模式
				b.Platform = FQAPI
			}
		}
		nv = novel{}
		err := nv.init(b.Platform, b.BookID)
		if err != nil {
			kitten.Error(err)
			continue
		}
		switch b.Platform {
		case FQ, FQAPI:
			if cmp.Or(
				getChapterID(FQAPI, nv.newChapter.url),
				getChapterID(FQ, nv.newChapter.url),
			) == cmp.Or(
				getChapterID(FQAPI, b.RecordURL),
				getChapterID(FQ, b.RecordURL),
			) {
				// 如果番茄（API 和 网页不同途径之间的比较）没有更新，则跳过
				continue
			}
		}
		if nv.newChapter.url == b.RecordURL {
			// 如果没有更新，则跳过
			continue
		}
		// 发送更新消息
		go tryCommentUpdate(
			msgr,
			msgr.Image(
				io.NewPath(nv.coverURL),
				io.NewPath(nv.headURL),
			).Text(nv.update()).SendMulti(b.Users...),
			b.Users,
			nv,
			getChapterID(nv.Platform, nv.newChapter.url))
		// 写入小说更新数据
		(*c)[i].BookName = nv.name
		(*c)[i].Writer = nv.writer
		(*c)[i].RecordURL = nv.newChapter.url
		(*c)[i].UpdateTime = nv.newChapter.Time
		// 按更新时间倒序排列
		c.sortByUpdate()
		// 异步保存配置
		go func() {
			err = c.saveConfig()
		}()
		if err != nil {
			kitten.Error(errSave, err)
			continue
		}
		kitten.Info(`更新《`, nv.name, `》成功喵！`)
	}
}

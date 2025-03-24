// Package track 小说更新播报、小说信息查询、小说更新查询
package track

import (
	"fmt"
	"log"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/plugin/track/book"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/ciweimao"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/fanqie"
	"github.com/Kittengarten/KittenCore/plugin/track/platform/sfacg"
	"github.com/Kittengarten/KittenCore/plugin/track/search"

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
	cycle            = shttp.TimeOutSeconds * time.Second // 最小循环间隔（最大可翻倍）
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
	configPath = fio.NewPath(engine.DataFolder(), configFile)
	// 报更更新的信号
	cu = make(chan book.Books)
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
	go novel.TryCommentUpdate(
		msgr,
		msgr.Reply().AtLf().
			Image(
				fio.NewPath(nv.CoverURL),
				fio.NewPath(nv.HeadURL),
			).
			Text(nv.Update()).
			SendMulti(),
		[]kitten.QQ{*o},
		nv,
		platform.Get(nv.Platform).ChapterID(nv.Chapter.URL),
	)
}

// 更新预览
func updatePreview(msgr *kitten.Messager) {
	n, err := getNovel(msgr)
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	if r := n.Preview; r != `` {
		msgr.Reply().AtLf().Text(`《`, n.Name, `》
`, &n.Chapter, `
`, r).Send()
		return
	}
	msgr.SendWithImageFail(`不存在的喵！`)
}

// 小说信息
func novelInfo(msgr *kitten.Messager, comment bool) {
	nv, err := getNovel(msgr)
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	msgr = msgr.Reply().AtLf().
		Image(fio.NewPath(nv.CoverURL)).
		Text(nv)
	if comment {
		msgr.Text(novel.Comment(nv))
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
	c, err := fio.Load[book.Books](configPath, fio.Empty) // 报更配置
	if err != nil {
		msgr.SendWithImageFail(book.ErrLoad, err)
		return
	}
	nv, err := getNovel(msgr) // 小说实例
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	// 本书下标
	if i := slices.IndexFunc(c, func(b book.Book) bool {
		return equal(nv, b)
	}); i == -1 {
		// 没有该小说，新建并添加
		c = append(c, book.Book{
			Platform: func() string {
				if nv.Platform != fanqie.API.String() {
					return nv.Platform
				}
				return fanqie.Platform.String()
			}(),
			BookID:   nv.ID,
			BookName: nv.Name,
			Writer:   nv.Writer,
			Users:    []kitten.QQ{*o},
		})
	} else {
		// 已经有该小说
		if slices.Contains(c[i].Users, *o) {
			// 已有该用户，无需添加
			msgr.SendWithImageFail(`《`, nv.Name, `》已经添加报更了喵！`)
			return
		}
		// 尚无该用户，需要添加
		c[i].Users = append(c[i].Users, *o)
		slices.Sort(c[i].Users)
	}
	if err := c.SaveConfig(cu, configPath); err != nil {
		msgr.SendWithImageFail(`添加《`, nv.Name, `》时`, book.ErrSave, err)
		return
	}
	msgr.Reply().AtLf().Text(`添加《`, nv.Name, `》报更成功喵！`).Send()
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
	c, err := fio.Load[book.Books](configPath, fio.Empty) // 报更配置
	if err != nil {
		msgr.SendWithImageFail(book.ErrLoad, err)
		return
	}
	if len(c) == 0 {
		msgr.Reply().AtLf().Text(book.Without).Send()
		return
	}
	nv, err := getNovel(msgr) // 小说实例
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	// 本书下标
	i := slices.IndexFunc(c, func(b book.Book) bool {
		return equal(nv, b)
	})
	if i == -1 {
		msgr.SendWithImageFail(`未在追更《`, nv.Name, `》喵！`)
		return
	}
	// 用户下标
	uid := slices.Index(c[i].Users, *o)
	if uid == -1 {
		msgr.SendWithImageFail(`未在追更《`, nv.Name, `》喵！`)
		return
	}
	// 移除在当前发送对象的报更
	if len(slices.Delete(c[i].Users, uid, uid+1)) == 0 {
		// 如果移除后，不再有报更对象（也可能本来就没有报更对象），则整体移除该小说
		c = slices.Delete(c, i, i+1)
	}
	if err := c.SaveConfig(cu, configPath); err != nil {
		msgr.SendWithImageFail(`取消《`, nv.Name, `》时`, book.ErrSave, err)
		return
	}
	msgr.Reply().AtLf().Text(`取消《`, nv.Name, `》报更成功喵！`).Send()
}

// 查询报更
func query(msgr *kitten.Messager) {
	o, err := msgr.Object() // 发送对象
	if err != nil {
		msgr.SendWithImageFail(err)
		return
	}
	mu.RLock()
	c, err := fio.Load[book.Books](configPath, fio.Empty) // 报更配置
	mu.RUnlock()
	if err != nil {
		msgr.SendWithImageFail(book.ErrLoad, err)
		return
	}
	if len(c) == 0 {
		msgr.Reply().AtLf().Text(book.Without).Send()
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
func getPlatform(keyword string) (platform.Platform, error) {
	switch {
	case strings.ContainsAny(keyword, `刺猬猫客`),
		strings.Contains(strings.ToUpper(keyword), `CWM`),
		strings.Contains(strings.ToLower(keyword), `ciweimao`):
		return ciweimao.Platform, nil
	case strings.ContainsAny(keyword, `菠萝包`),
		strings.Contains(strings.ToUpper(keyword), `SF`),
		strings.Contains(strings.ToLower(keyword), `blb`):
		return sfacg.Platform, nil
	case strings.ContainsAny(keyword, `番茄柿`),
		strings.Contains(strings.ToUpper(keyword), `FQ`):
		if strings.HasPrefix(keyword, `#`) {
			return fanqie.Platform, nil
		}
		return fanqie.API, nil
	default:
		return nil, platform.NotSupported(nil)
	}
}

/*
获取小说

如果传入值不为书号，则先获取书号
*/
func getNovel(msgr *kitten.Messager) (*novel.Novel, error) {
	args := msgr.ArgsSlice()
	if len(args) != 2 {
		return nil, fmt.Errorf(`本命令参数数量：2
%s %s
传入的参数数量：%d
参数数量错误喵！`,
			pf, ag,
			len(args))
	}
	var (
		p, err = getPlatform(args[0])
		bookID = args[1]
	)
	if err != nil {
		return nil, err
	}
	if _, err := strconv.Atoi(bookID); err != nil {
		// 获取小说时，参数字符串无法转换为书号，尝试作为搜索关键词
		if bookID, err = p.FindBookID(search.Keyword(args[1])); err != nil {
			return nil, fmt.Errorf(`关键词 %s 搜索时发生错误：%w`, args[1], err)
		}
	}
	return p.Init(bookID)
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
	if err := configPath.InitFile(fio.Empty); err != nil {
		kitten.Error(`初始化报更配置文件时发生错误喵！`, err)
		return
	}
	mu.RLock()
	data, err := fio.Load[book.Books](configPath, fio.Empty)
	mu.RUnlock()
	if err != nil {
		kitten.Error(book.ErrLoad, err)
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
		t   = time.NewTicker(shttp.TimeOutSeconds * time.Second) // 定期检测，间隔为超时时间
		st  = time.NewTicker(shttp.TimeOutSeconds * time.Minute) // 专用慢速时钟
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
		data.Report(bot, cu, configPath, cycle, st) // 执行报更
	}
}

// 判断是否为同一本书
func equal(nv *novel.Novel, b book.Book) bool {
	return nv.Platform == b.Platform && nv.ID == b.BookID
}

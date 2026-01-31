// Package track 小说更新播报、小说信息查询、小说更新查询
package track

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/internal/config"
	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core"
	"github.com/Kittengarten/KittenCore/kitten/core/fio"
	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/mode"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/kitten/msg"
	"github.com/Kittengarten/KittenCore/kitten/rate"
	"github.com/Kittengarten/KittenCore/kitten/usr"
	"github.com/Kittengarten/KittenCore/plugin/track/book"
	"github.com/Kittengarten/KittenCore/plugin/track/check"
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
	cSetProtagonists = `设置主角`
)

var (
	// 指令前缀
	p = config.CommandPrefix()
	// 帮助
	help = p + cNovel + ` ` + pf + ` ` + ag + ` // 可获取信息
` + p + cNovelComment + ` ` + pf + ` ` + ag + ` // 可获取带点评的信息
` + p + cUpdatePreview + ` ` + pf + ` ` + ag + ` // 可预览更新内容
` + p + cQueryUpdate + ` // 可查询当前小说自动报更
————
管理员或私聊可用：
` + p + cAddUpdate + ` ` + pf + ` ` + ag + ` // 可添加小说自动报更
` + p + cCancelUpdate + ` ` + pf + ` ` + ag + ` // 可取消小说自动报更
` + p + cUpdateTest + ` ` + pf + ` ` + ag + ` // 可测试报更功能
` + p + cSetProtagonists + ` ` + pf + ` ` + ag + ` [主角1] [主角2] ... // 可设置小说主角`
	// 注册插件
	engine = func() *control.Engine {
		if mode.Test() {
			return nil
		}
		return control.AutoRegister(&ctrl.Options[*zero.Ctx]{
			DisableOnDefault:  false,
			Brief:             brief,
			Help:              help,
			PrivateDataFolder: replyServiceName,
		}).ApplySingle(ctxext.DefaultSingle)
	}()
)

var (
	// 配置文件路径
	configPath = fio.NewPath(engine.DataFolder(), configFile).WithRWMutex()
	// 报更更新的信号
	cu = make(chan book.Books, 1)
)

func init() {
	if mode.Test() {
		slog.Info(`测试中，跳过初始化`, `插件`, replyServiceName)
		return
	}

	utils.Go(replyServiceName, track)

	engine.OnCommandGroup([]string{
		cUpdateTest,
		cSetProtagonists,
	}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, canc := context.WithTimeout(context.Background(), core.Timeout)
			defer canc()
			switch handler := msg.NewWithContext(c, ctx); handler.Command() {
			case cUpdateTest:
				// 更新测试
				updateTest(handler)
			case cSetProtagonists:
				// 设置主角
				c, canc := context.WithTimeout(context.Background(), shttp.Timeout)
				defer canc()
				setProtagonists(msg.NewWithContext(c, ctx))
			}
		})

	engine.OnCommandGroup([]string{
		cNovel,
		cNovelComment,
		cUpdatePreview,
	}).SetBlock(true).
		Limit(rate.User.Get()).
		Limit(rate.GroupNormal.Get()).
		Handle(func(ctx *zero.Ctx) {
			c, canc := context.WithTimeout(context.Background(), core.Timeout)
			defer canc()
			switch handler := msg.NewWithContext(c, ctx); handler.Command() {
			case cNovel:
				// 小说信息
				novelInfo(handler, false)
			case cNovelComment:
				// 点评小说
				novelInfo(handler, true)
			case cUpdatePreview:
				// 更新预览
				updatePreview(handler)
			}
		})

	engine.OnCommandGroup([]string{
		cAddUpdate,
		cCancelUpdate,
	}, zero.UserOrGrpAdmin).SetBlock(true).
		Limit(rate.GroupFast.Get()).
		Handle(func(ctx *zero.Ctx) {
			c, canc := context.WithTimeout(context.Background(), core.Timeout)
			defer canc()
			switch handler := msg.NewWithContext(c, ctx); handler.Command() {
			case cAddUpdate:
				// 添加报更
				add(handler)
			case cCancelUpdate:
				// 取消报更
				cancel(handler)
			}
		})

	// 查询报更
	engine.OnCommand(cQueryUpdate).SetBlock(true).
		Limit(rate.GroupSlow.Get()).Handle(func(ctx *zero.Ctx) {
		c, canc := context.WithTimeout(context.Background(), core.Timeout)
		defer canc()
		query(msg.NewWithContext(c, ctx))
	})
}

// 更新测试
func updateTest(handler *msg.Handler) {
	nv, err := getNovel(handler, false) // 小说实例
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	defer novel.Pool.Put(nv)
	o, err := handler.Object()
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	novel.TryCommentUpdate(
		handler,
		handler.Quote().AtLf().
			Image(
				fio.NewPath(nv.CoverURL),
				fio.NewPath(nv.HeadURL),
			).
			Text(nv.Update(handler)).
			SendMulti(),
		[]usr.QQ{o},
		nv,
	)
}

// 设置主角
func setProtagonists(handler *msg.Handler) {
	args := handler.ArgsSlice()
	if len(args) < 3 {
		handler.SendWithImageFail(`请输入主角名喵！`)
		return
	}
	p, err := getPlatform(args[0])
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	bookID := args[1]
	if _, err := strconv.Atoi(bookID); err != nil {
		// 参数字符串无法转换为书号，尝试作为搜索关键词
		if bookID, err = p.FindBookID(handler, search.Keyword(bookID)); err != nil {
			handler.SendWithImageFail(`关键词“`, bookID, "”搜索错误喵！", err)
			return
		}
	}
	handler.Quote().AtLf().Text(`平台：`, p, "\n书号：", bookID).Send()
	configPath.Lock()
	defer configPath.Unlock()
	c, err := fio.LoadWithContext[book.Books](handler, configPath.Path, fio.Empty) // 报更配置
	if err != nil {
		handler.SendWithImageFail(book.ErrLoad, err)
		return
	}
	// 本书索引
	i := slices.IndexFunc(c, func(b book.Book) bool {
		return p.String() == b.Platform && bookID == b.BookID
	})
	if i == -1 {
		handler.SendWithImageFail(`本书未配置报更喵！`)
		return
	}
	c[i].Protagonists = args[2:]
	if err := c.SaveConfig(cu, configPath.Path); err != nil {
		handler.SendWithImageFail(`设置`, strings.Join(args[2:], `、`), `为主角时`, book.ErrSave, err)
		return
	}
	handler.Quote().AtLf().Text(`设置`, strings.Join(args[2:], `、`), `为主角成功喵！`).Send()
}

// 更新预览
func updatePreview(handler *msg.Handler) {
	nv, err := getNovel(handler, false) // 小说实例
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	defer novel.Pool.Put(nv)
	if r := nv.Preview; r != `` {
		handler.Quote().AtLf().Text(`《`, nv.Name, `》
`, &nv.Chapter, `
`, r).Send()
		return
	}
	handler.SendWithImageFail(`不存在的喵！`)
}

// 小说信息
func novelInfo(handler *msg.Handler, comment bool) {
	nv, err := getNovel(handler, false) // 小说实例
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	defer novel.Pool.Put(nv)
	p, err := platform.Get(nv.Platform)
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	if p == fanqie.API {
		// 还原番茄平台名称
		nv.Platform = fanqie.Platform.String()
	}
	handler.Quote().AtLf().
		Image(fio.NewPath(nv.CoverURL)).
		Text(nv)
	if comment && novel.Export.Commenter != nil {
		handler.Text(novel.Export.CommentNovel(handler, nv))
	}
	handler.Send()
}

// 添加报更
func add(handler *msg.Handler) {
	o, err := handler.Object() // 发送对象
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	configPath.Lock()
	defer configPath.Unlock()
	c, err := fio.LoadWithContext[book.Books](handler, configPath.Path, fio.Empty) // 报更配置
	if err != nil {
		handler.SendWithImageFail(book.ErrLoad, err)
		return
	}
	nv, err := getNovel(handler, false) // 小说实例
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	defer novel.Pool.Put(nv)
	// 本书索引
	if i := slices.IndexFunc(c, func(b book.Book) bool {
		return equal(nv, b)
	}); i == -1 {
		// 没有该小说，进行添加
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
			Users:    []usr.QQ{o},
		})
	} else {
		// 已经有该小说
		if slices.Contains(c[i].Users, o) {
			// 已有该用户，无需添加
			handler.SendWithImageFail(`《`, nv.Name, `》已经添加报更了喵！`)
			return
		}
		// 尚无该用户，需要添加
		c[i].Users = append(c[i].Users, o)
		slices.Sort(c[i].Users)
	}
	if err := c.SaveConfig(cu, configPath.Path); err != nil {
		handler.SendWithImageFail(`《`, nv.Name, `》添加报更时`, book.ErrSave, err)
		return
	}
	handler.Quote().AtLf().Text(`《`, nv.Name, `》添加报更成功喵！`).Send()
}

// 取消报更
func cancel(handler *msg.Handler) {
	o, err := handler.Object() // 发送对象
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	configPath.Lock()
	defer configPath.Unlock()
	c, err := fio.LoadWithContext[book.Books](handler, configPath.Path, fio.Empty) // 报更配置
	if err != nil {
		handler.SendWithImageFail(book.ErrLoad, err)
		return
	}
	if len(c) == 0 {
		handler.SendWithImageFail(book.ErrNotConfig)
		return
	}
	nv, err := getNovel(handler, false) // 小说实例
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	defer novel.Pool.Put(nv)
	// 本书索引
	i := slices.IndexFunc(c, func(b book.Book) bool {
		return equal(nv, b)
	})
	if i == -1 {
		handler.SendWithImageFail(`《`, nv.Name, `》未在追更喵！`)
		return
	}
	// 用户索引
	uid := slices.Index(c[i].Users, o)
	if uid == -1 {
		handler.SendWithImageFail(`《`, nv.Name, `》未在追更喵！`)
		return
	}
	// 移除在当前发送对象的报更
	if len(slices.Delete(c[i].Users, uid, uid+1)) == 0 {
		// 如果移除后，不再有报更对象（也可能本来就没有报更对象），则整体移除该小说
		c = slices.Delete(c, i, i+1)
	}
	if err := c.SaveConfig(cu, configPath.Path); err != nil {
		handler.SendWithImageFail(`《`, nv.Name, `》取消报更时`, book.ErrSave, err)
		return
	}
	handler.Quote().AtLf().Text(`《`, nv.Name, `》取消报更成功喵！`).Send()
}

// 查询报更
func query(handler *msg.Handler) {
	o, err := handler.Object() // 发送对象
	if err != nil {
		handler.SendWithImageFail(err)
		return
	}
	configPath.RLock()
	c, err := fio.LoadWithContext[book.Books](handler, configPath.Path, fio.Empty) // 报更配置
	configPath.RUnlock()
	if err != nil {
		handler.SendWithImageFail(book.ErrLoad, err)
		return
	}
	if len(c) == 0 {
		handler.SendWithImageFail(book.ErrNotConfig)
		return
	}
	const h = `【报更列表】`
	s := new(strings.Builder)
	s.Grow(64 * len(c))
	s.WriteString(h)
	for _, b := range c {
		if !slices.Contains(b.Users, o) {
			// 如果本书不在这里报更，则直接遍历至下一本书
			continue
		}
		fmt.Fprintf(s, "\n%s", b)
	}
	fmt.Fprintf(s, "\n检测速率：%d 本/小时", check.QueryRate())
	handler.Quote().AtLf().Text(s).Send()
}

// 获取小说
//
//	如果传入值不为书号，则先获取书号
func getNovel(handler *msg.Handler, cache bool) (*novel.Novel, error) {
	var (
		args = handler.Args()
		pkey string // 平台关键词
		bkey string // 书名关键词
	)
	n, err := fmt.Sscanln(args, &pkey, &bkey)
	if err != nil {
		return nil, fmt.Errorf(`%s %s
参数错误（识别到 %d 个）：
%w`,
			pf, ag,
			n,
			err,
		)
	}
	p, err := getPlatform(pkey)
	if err != nil {
		return nil, err
	}
	if _, err := strconv.Atoi(bkey); err == nil {
		return novel.Init(handler, p, bkey, cache)
	}
	// 获取小说时，参数字符串无法转换为书号，尝试作为搜索关键词
	nvID, err := p.FindBookID(handler, search.Keyword(bkey))
	if err != nil {
		return nil, fmt.Errorf("关键词“%s”搜索时发生错误：\n%w", bkey, err)
	}
	return novel.Init(handler, p, nvID, cache)
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
		return nil, platform.NotSupported(keyword)
	}
}

// 报更
func track() {
	// 加载报更配置文件
	configPath.RLock()
	data, err := fio.Load[book.Books](configPath.Path, fio.Empty)
	configPath.RUnlock()
	if err != nil {
		log.Error(book.ErrLoad, err)
		return
	}
	log.Infoln(`当前有`, len(data), `本小说正在报更喵！`)
	func() {
		process.GlobalInitMutex.Lock()
		defer process.GlobalInitMutex.Unlock()
	}()
	sid := usr.Self()
	if stat.Tracker = msg.New(zero.GetBot(sid.Int())); !stat.Tracker.Check(kitten.Caller) {
		log.Panic(`获取 Bot 实例失败喵！`, stat.Tracker)
	}
	// 报更
	for {
		select {
		case data = <-cu: // 接收到更新配置则使用
		default: // 未接收到更新配置则释放
		}
		data.Report(stat.Tracker, cu, configPath) // 执行报更
		stat.Tracker.Reset()                      // 重置 Bot 实例
	}
}

// 判断是否为同一本书
func equal(nv *novel.Novel, b book.Book) bool {
	return nv.Platform == b.Platform && nv.ID == b.BookID
}

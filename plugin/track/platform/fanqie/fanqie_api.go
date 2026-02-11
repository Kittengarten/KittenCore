package fanqie

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/stat"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/search"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/tidwall/gjson"
)

type (
	// FQAPI 番茄 API
	FQAPI utils.Object
)

const (
	Detail  = `detail`   // 详情
	Catalog = `catalog`  // 目录
	Content = `content`  // 内容
	Search  = `search`   // 搜索
	BookID  = `book_id`  // 书号
	ItemID  = `item_id`  // 章节号
	Query   = `query`    // 关键词
	Offset  = `offset`   // 10 * (page - 1)
	TabType = `tab_type` // 0: 不限，1：综合，2：听书，3：书籍，4：社区，5：全文，8：漫画，11：短句
)

// API 番茄 API
var API = FQAPI{}

func init() {
	platform.Register(API)
	// 注册恢复器
	novel.GlobalRestorer = API
}

// RestoreURL 恢复 URL，实现 novel.URLRestorer
func (FQAPI) RestoreURL(nv *novel.Novel) string {
	if nv.Platform == API.String() {
		// 番茄 API 链接还原为章节链接
		// 不在这里还原平台，以免影响 nv.todayReport() 计算
		nv.Chapter.URL = ChapterURL + API.ChapterID(nv.Chapter.URL)
	}
	return "\n" + nv.Chapter.URL
}

// RestorePlatform 恢复平台，实现 novel.PlatformRestorer
func (FQAPI) RestorePlatform(nv *novel.Novel) {
	if nv.Platform == API.String() {
		nv.Platform = Platform.String()
	}
}

// String 实现 fmt.Stringer，返回小说平台名称
func (FQAPI) String() string {
	if stat.APIHOST[stat.Fanqie] == `` {
		return Platform.String()
	}
	return `番茄 API`
}

// Layout 返回时间格式
func (FQAPI) Layout() string {
	if stat.APIHOST[stat.Fanqie] == `` {
		return Platform.Layout()
	}
	return times.Layout
}

// FindBookID 用关键词搜索书号
func (FQAPI) FindBookID(ctx context.Context, key search.Keyword) (string, error) {
	if stat.APIHOST[stat.Fanqie] == `` {
		return Platform.FindBookID(ctx, key)
	}
	u, err := url.Parse(stat.APIHOST[stat.Fanqie])
	if err != nil {
		return ``, err
	}
	u = u.JoinPath(Search)
	v := u.Query()
	v.Add(Query, string(key))
	const page = 1 // 默认搜索第一页
	v.Add(Offset, strconv.FormatInt(10*(page-1), 10))
	v.Add(TabType, `3`) // 默认搜索类型（3：小说）
	u.RawQuery = v.Encode()
	data, err := shttp.GETDataURLWithContext(ctx, u)
	if err != nil {
		return ``, maskError(err)
	}
	if !gjson.ValidBytes(data) {
		return ``, fmt.Errorf(`%wJSON：%s`, utils.ErrInvalidData, string(data))
	}
	bookIDs := gjson.GetBytes(data, `search_tabs.#(tab_type=3).data.#.book_id`).Array()
	if len(bookIDs) == 0 {
		return ``, key.NotFound()
	}
	const number = 1 // 默认取第一本书
	return bookIDs[number-1].String(), nil
}

// ChapterID 获取章号
func (FQAPI) ChapterID(cpURL string) string {
	return Platform.ChapterID(cpURL)
}

// Init 小说网页信息获取
func (f FQAPI) Init(ctx context.Context, cpID string, cache bool) (any, error) {
	if stat.APIHOST[stat.Fanqie] == `` {
		return Platform.Init(ctx, cpID, cache)
	}
	// 初始化小说
	nv := novel.Pool.Get().(*novel.Novel)
	*nv = novel.Novel{}
	// 初始化小说平台
	// 此处仍使用 API，以免影响后续判定，待处理完毕后还原为 Platform
	nv.Platform = API.String()
	// 向小说传入书号
	nv.ID = cpID
	// 生成链接
	nv.URL = URL + nv.ID
	u, err := url.Parse(stat.APIHOST[stat.Fanqie])
	if err != nil {
		return nv, err
	}
	u = u.JoinPath(Detail)
	v := u.Query()
	v.Add(BookID, cpID)
	u.RawQuery = v.Encode()
	// 获取小说网页，失败则返回
	data, err := shttp.GETDataURLWithContext(ctx, u)
	if err != nil {
		return nv, maskError(err)
	}
	if !gjson.ValidBytes(data) {
		log.Errorf("无效的 JSON：\n%s", data)
		err = status.ErrStatus(nv.URL, status.BookStatusException)
		return nv, err
	}
	if gjson.GetBytes(data, `message`).String() != `SUCCESS` {
		err = status.ErrStatus(nv.URL, status.BookUnreachable)
		return nv, err
	}
	result := gjson.GetBytes(data, `data`)
	nv.Right = []string{`免费`}                                   // 获取上架状态（番茄均为免费）
	nv.Name = result.Get(`book_name`).String()                  // 获取书名
	nv.Writer = result.Get(`author`).String()                   // 获取作者
	nv.HeadURL = result.Get(`author_info.user_avatar`).String() // 获取头像链接
	nv.Collection = result.Get(`all_bookshelf_count`).String()  // 获取收藏
	nv.HitNum = result.Get(`read_count_all`).String()           // 获取点击
	nv.TotalWordNum = result.Get(`word_number`).String()        // 获取小说字数
	nv.CoverURL = cmp.Or(result.Get(`expand_thumb_url`).String(),
		result.Get(`thumb_url`).String()) // 获取封面
	nv.Status = func() string {
		if v, ok := map[int64]string{
			0: `已完结`,
			1: `连载中`,
			4: `断更`,
		}[result.Get(`creation_status`).Int()]; ok {
			return v
		}
		return `未知`
	}() // 获取状态
	nv.TagList = strings.Split(result.Get(`tags`).String(), `,`) // 获取标签
	nv.Introduce = result.Get(`book_abstract_v2`).String()       // 获取简述
	ncp := result.Get(`last_chapter_item_id`).String()           // 获取新章节链接
	nv.Protagonists = func() (p []string) {
		for _, n := range gjson.Parse(
			result.Get(`roles`).String(),
		).Array() {
			p = append(p, n.String())
		}
		return
	}() // 获取主角
	// 加载新章节
	u.Path = ``
	u = u.JoinPath(Content)
	clear(v)
	v.Add(ItemID, ncp)
	u.RawQuery = v.Encode()
	nv.Chapter, err = chapter.New(ctx, f, u.String())
	return nv, err
}

// NewChapter 章节信息获取
func (FQAPI) NewChapter(ctx context.Context, cpURL string) (any, error) {
	if stat.APIHOST[stat.Fanqie] == `` {
		return Platform.NewChapter(ctx, cpURL)
	}
	// 初始化章节
	cp := chapter.Pool.Get().(*chapter.Chapter)
	*cp = chapter.Chapter{}
	// 向章节传入链接
	cp.URL = cpURL
	// 路径
	const i = `0`
	p := `novel_data`
	data, err := func() ([]byte, error) {
		// 先从 SNSSDK API 获取
		s, err := SNSSDKDetail(cp.URL)
		if err == nil {
			var b []byte
			b, err = shttp.GETDataURLWithContext(ctx, s)
			if err == nil {
				// 替换 URL
				cp.URL = s.String()
				p = i
				return b, nil
			}
		}
		// SNSSDK API 获取失败
		log.Error(err)
		// 获取章节网页，失败则返回
		return shttp.GETDataWithContext(ctx, cp.URL)
	}()
	defer func() { cp.URL = cpURL }() // 恢复 URL
	if err != nil {
		return cp, maskError(err)
	}
	if !gjson.ValidBytes(data) {
		log.Errorf("无效的 JSON：\n%s", data)
		err = status.ErrStatus(cp.URL, status.ChapterStatusException)
		return cp, err
	}
	result := gjson.GetBytes(data, `data`)
	cp.Update = time.Unix(result.Get(p+`.first_pass_time`).Int(),
		0).Local() // 获取更新时间
	cp.Title = func() string {
		if volumeName := result.Get(p + `.volume_name`).String(); volumeName != `` {
			return volumeName + `
` + result.Get(p+`.title`).String()
		}
		return result.Get(`novel_data.chapter_title`).String()
	}() // 获取章节名称
	// 获取章节字数
	if cp.WordNum = int(cmp.Or(
		result.Get(p+`.chapter_word_number`).Int(),
		result.Get(`novel_data.word_number`).Int(),
	)); cp.WordNum <= 0 {
		log.Errorf("错误的 JSON：\n%s", data)
		return cp, fmt.Errorf(`%w字数：%d`,
			status.ErrStatus(cp.URL, status.ChapterStatusException), cp.WordNum)
	}
	if p == i {
		// 正在使用 SNSSDK
		data, err := shttp.GETDataWithContext(ctx, cpURL)
		if err != nil {
			return cp, err
		}
		result = gjson.GetBytes(data, `data`)
	}
	cp.PreURL, err = getChapterURL(
		result.Get(`novel_data.pre_item_id`).String()) // 获取上一章链接
	if err != nil {
		return cp, err
	}
	cp.NextURL, err = getChapterURL(
		result.Get(`novel_data.next_item_id`).String()) // 获取下一章链接
	return cp, err
}

// 获取章节链接
func getChapterURL(id string) (string, error) {
	v := make(url.Values)
	v.Set(ItemID, id)
	u, err := url.Parse(stat.APIHOST[stat.Fanqie])
	if err != nil {
		return ``, err
	}
	u = u.JoinPath(Content)
	u.RawQuery = v.Encode()
	return u.String(), nil
}

// ShouldUpdate 书籍更新检测
func ShouldUpdate(upd, rec string) bool {
	var (
		u = Platform.ChapterID(upd)
		r = Platform.ChapterID(rec)
	)
	// 长度大的数字一定更大
	if len(u) != len(r) {
		return len(u) > len(r)
	}
	return u > r
}

// maskError 为了防止信息泄露，替换错误信息
func maskError(err error, oldnew ...string) error {
	shttpErr, ok := errors.AsType[*shttp.Error](err)
	if !ok {
		return err
	}
	var (
		mask = []string{
			stat.APIHOST[stat.Fanqie], `HOST`,
		}
		escaped = make([]string, 0, len(oldnew)+len(mask))
	)
	for i := 0; i+1 < len(oldnew); i += 2 {
		escaped = append(escaped, url.QueryEscape(oldnew[i]), oldnew[i+1])
	}
	shttpErr.URL = strings.NewReplacer(append(escaped, mask...)...).Replace(shttpErr.URL)
	return shttpErr
}

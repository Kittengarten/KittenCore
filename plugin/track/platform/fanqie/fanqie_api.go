package fanqie

import (
	"cmp"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/search"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/tidwall/gjson"
)

// 番茄 API
type fqAPI struct{}

const (
	Detail  = `detail`   // 详情
	Catalog = `catalog`  // 目录
	Content = `content`  // 内容
	Search  = `search`   // 搜索
	BookID  = `book_id`  // 书号
	ItemID  = `item_id`  // 章节号
	Query   = `query`    // 关键词
	Offset  = `offset`   // 10 * (page - 1)
	TabType = `tab_type` // 1：综合，2：听书，3：书籍，4：社区，5：全文，8：漫画，11：短句
)

var (
	// API 番茄 API
	API = fqAPI{}
	// APIHOST
	APIHOST string
)

func init() {
	platform.Platforms = append(platform.Platforms, API)
	novel.RestoreURL = func(nv *novel.Novel) string {
		if nv.Platform == API.String() {
			// 番茄 API 链接还原为章节链接
			// 不在这里还原平台，以免影响 nv.todayReport() 计算
			nv.Chapter.URL = ChapterURL + API.ChapterID(nv.Chapter.URL)
		}
		return "\n" + nv.Chapter.URL
	}
	novel.RestorePlatform = func(nv *novel.Novel) {
		if nv.Platform == API.String() {
			nv.Platform = Platform.String()
		}
	}
}

// String 实现 fmt.Stringer，返回小说平台名称
func (f fqAPI) String() string {
	if APIHOST == `` {
		return Platform.String()
	}
	return `番茄 API`
}

// Layout 返回时间格式
func (f fqAPI) Layout() string {
	if APIHOST == `` {
		return Platform.Layout()
	}
	return times.Layout
}

// FindBookID 用关键词搜索书号
func (f fqAPI) FindBookID(key search.Keyword) (string, error) {
	if APIHOST == `` {
		return Platform.FindBookID(key)
	}
	u, err := url.Parse(APIHOST)
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
	data, err := shttp.GETDataURL(u)
	if err != nil {
		return ``, err
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
func (f fqAPI) ChapterID(cpURL string) string {
	return Platform.ChapterID(cpURL)
}

// Init 小说网页信息获取
func (f fqAPI) Init(bookID string) (nv *novel.Novel, err error) {
	if APIHOST == `` {
		return Platform.Init(bookID)
	}
	// 初始化小说
	nv = novel.Pool.Get().(*novel.Novel)
	// 初始化小说平台
	// 此处仍使用 API，以免影响后续判定，待处理完毕后还原为 Platform
	nv.Platform = API.String()
	// 向小说传入书号
	nv.ID = bookID
	// 生成链接
	nv.URL = URL + nv.ID
	u, err := url.Parse(APIHOST)
	if err != nil {
		return nv, err
	}
	u = u.JoinPath(Detail)
	v := u.Query()
	v.Add(BookID, bookID)
	u.RawQuery = v.Encode()
	// 获取小说网页，失败则返回
	data, err := shttp.GETDataURL(u)
	if err != nil {
		return nv, err
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		err = status.ErrStatus(nv.URL, status.BookStatusException)
		return nv, err
	}
	if gjson.GetBytes(data, `message`).String() != `SUCCESS` {
		err = status.ErrStatus(nv.URL, status.BookUnreachable)
		return nv, err
	}
	result := gjson.GetManyBytes(data,
		`data.book_name`,               // 书名
		`data.author`,                  // 作者
		`data.author_info.user_avatar`, // 头像
		`data.all_bookshelf_count`,     // 收藏
		`data.read_count_all`,          // 点击
		`data.word_number`,             // 字数
		`data.expand_thumb_url`,        // 大封面
		`data.thumb_url`,               // 封面
		`data.creation_status`,         // 状态（0：已完结，1：连载中，4：断更）
		`data.tags`,                    // 标签（单个字符串，由半角逗号分隔）
		`data.book_abstract_v2`,        // 简述
		`data.last_chapter_item_id`,    // 最新章节
		`data.roles`,                   // 主角
	)
	nv.Right = []string{`免费`}            // 获取上架状态（番茄均为免费）
	nv.Name = result[0].String()         // 获取书名
	nv.Writer = result[1].String()       // 获取作者
	nv.HeadURL = result[2].String()      // 获取头像链接
	nv.Collection = result[3].String()   // 获取收藏
	nv.HitNum = result[4].String()       // 获取点击
	nv.TotalWordNum = result[5].String() // 获取小说字数
	nv.CoverURL = cmp.Or(result[6].String(), result[7].String())
	nv.Status = func() string {
		if v, ok := map[int64]string{
			0: `已完结`,
			1: `连载中`,
			4: `断更`,
		}[result[8].Int()]; ok {
			return v
		}
		return `未知`
	}() // 获取状态
	nv.TagList = strings.Split(result[9].String(), `,`) // 获取标签
	nv.Introduce = result[10].String()                  // 获取简述
	ncp := result[11].String()                          // 获取新章节链接
	nv.Protagonists = func() (p []string) {
		for _, n := range gjson.Parse(result[12].String()).Array() {
			p = append(p, n.String())
		}
		return
	}() // 获取主角
	// 不支持的字段
	nv.Theme = ``
	nv.Item = nil
	nv.Preview = ``
	// 加载新章节
	u.Path = ``
	u = u.JoinPath(Content)
	clear(v)
	v.Add(ItemID, ncp)
	u.RawQuery = v.Encode()
	nv.Chapter, err = f.NewChapter(u.String())
	return nv, err
}

// NewChapter 章节信息获取
func (f fqAPI) NewChapter(cpURL string) (cp *chapter.Chapter, err error) {
	if APIHOST == `` {
		return Platform.NewChapter(cpURL)
	}
	// 初始化章节
	cp = chapter.Pool.Get().(*chapter.Chapter)
	// 向章节传入链接
	cp.URL = cpURL
	// 获取章节网页，失败则返回
	data, err := shttp.GETData(cp.URL)
	if err != nil {
		return cp, err
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		err = status.ErrStatus(cp.URL, status.ChapterStatusException)
		return cp, err
	}
	result := gjson.GetManyBytes(data,
		`data.novel_data.first_pass_time`,     // 更新时间
		`data.novel_data.volume_name`,         // 分卷名称
		`data.novel_data.title`,               // 章节名称
		`data.novel_data.chapter_word_number`, // 章节字数
		`data.novel_data.pre_item_id`,         // 上章号码
		`data.novel_data.next_item_id`,        // 下章号码
	)
	cp.Time = time.Unix(result[0].Int(), 0).Local() // 获取更新时间
	cp.Title = result[1].String() + `
` + result[2].String() // 获取章节名称
	cp.WordNum = int(result[3].Int()) // 获取章节字数
	if cp.WordNum <= 0 {
		err = status.ErrStatus(cp.URL, status.ChapterStatusException)
		return cp, err
	}
	cp.PreURL, err = getChapterURL(result[4].String()) // 获取上一章链接
	if err != nil {
		return cp, err
	}
	cp.NextURL, err = getChapterURL(result[5].String()) // 获取下一章链接
	return cp, err
}

// 获取章节链接
func getChapterURL(id string) (string, error) {
	v := make(url.Values)
	v.Set(ItemID, id)
	u, err := url.Parse(APIHOST)
	if err != nil {
		return ``, err
	}
	u = u.JoinPath(Content)
	u.RawQuery = v.Encode()
	return u.String(), nil
}

// IsUpdate 书籍更新检测
func IsUpdate(upd, rec string) bool {
	return Platform.ChapterID(upd) == Platform.ChapterID(rec)
}

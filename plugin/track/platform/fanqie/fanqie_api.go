package fanqie

import (
	"cmp"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/http"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
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
	APIHOST       = `https://api.v2.sukimon.me:45554`
	DetailURL     = APIHOST + `/detail`
	APICatalogURL = APIHOST + `/catalog`
	APIContentURL = APIHOST + `/content`
	APISearchURL  = APIHOST + `/search`
	APIBookID     = `book_id`  // 书号
	APIItemID     = `item_id`  // 章节号
	APIQuery      = `query`    // 关键词
	APIOffset     = `offset`   // 10 * (page - 1)
	APITabType    = `tab_type` // 1：综合，2：听书，3：书籍，4：社区，5：全文，8：漫画，11：短句
)

// API 番茄 API
var API = fqAPI{}

func init() {
	platform.Platforms = append(platform.Platforms, API)
	novel.RestoreURL = func(nv *novel.Novel) string {
		if nv.Platform == API.String() {
			// 番茄 API 链接还原为章节链接
			// 不在这里还原平台，以免影响 nv.todayReport() 计算
			newURL, err := url.Parse(nv.Chapter.URL)
			if err != nil {
				kitten.Error(err)
				return ``
			}
			nv.Chapter.URL = ChapterURL + newURL.Query().Get(APIItemID)
		}
		return "\n" + nv.Chapter.URL
	}
	novel.RestorePlatform = func(nv *novel.Novel) {
		if nv.Platform != API.String() {
			return
		}
		nv.Platform = Platform.String()
	}
}

// String 实现 fmt.Stringer，返回小说平台名称
func (f fqAPI) String() string {
	return `番茄 API`
}

// Layout 返回时间格式
func (f fqAPI) Layout() string {
	return times.Layout
}

// FindBookID 用关键词搜索书号
func (f fqAPI) FindBookID(key search.Keyword) (string, error) {
	values := make(url.Values)
	values.Add(APIQuery, string(key))
	const page = 1 // 默认搜索第一页
	values.Add(APIOffset, strconv.FormatInt(10*(page-1), 10))
	values.Add(APITabType, `3`) // 默认搜索类型（3：小说）
	searchURL, err := url.Parse(APISearchURL)
	if err != nil {
		return ``, err
	}
	searchURL.RawQuery = values.Encode()
	data, err := http.GETData(searchURL.String())
	if err != nil {
		return ``, err
	}
	if !gjson.ValidBytes(data) {
		return ``, errors.New(`无效的 JSON：` + string(data))
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
	u, err := url.Parse(cpURL)
	if err != nil {
		return ``
	}
	return u.Query().Get(APIItemID)
}

// Init 小说网页信息获取
func (f fqAPI) Init(bookID string) (nv *novel.Novel, err error) {
	// 初始化小说
	nv = novel.Pool.Get().(*novel.Novel)
	// 初始化小说平台（API 仍然显示为番茄小说网）
	nv.Platform = Platform.String()
	// 向小说传入书号
	nv.ID = bookID
	// 生成链接
	nv.URL = URL + nv.ID
	values := make(url.Values)
	values.Add(APIBookID, bookID)
	apiURL, err := url.Parse(DetailURL)
	if err != nil {
		return
	}
	apiURL.RawQuery = values.Encode()
	// 获取小说网页，失败则返回
	data, err := http.GETData(apiURL.String())
	if err != nil {
		return
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		err = status.ErrStatus(nv.URL, status.BookStatusException)
		return
	}
	if gjson.GetBytes(data, `message`).String() != `SUCCESS` {
		err = status.ErrStatus(nv.URL, status.BookUnreachable)
		return
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
	// 不支持的字段
	nv.Theme = ``
	nv.Item = nil
	nv.Preview = ``
	// 加载新章节
	clear(values)
	values.Add(APIItemID, ncp)
	ncpURL, err := url.Parse(APIContentURL)
	if err != nil {
		return
	}
	ncpURL.RawQuery = values.Encode()
	nv.Chapter, err = f.NewChapter(ncpURL.String())
	return
}

// NewChapter 章节信息获取
func (f fqAPI) NewChapter(cpURL string) (cp *chapter.Chapter, err error) {
	// 初始化章节
	cp = chapter.Pool.Get().(*chapter.Chapter)
	// 向章节传入链接
	cp.URL = cpURL
	// 获取章节网页，失败则返回
	data, err := http.GETData(cp.URL)
	if err != nil {
		return
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		err = status.ErrStatus(cp.URL, status.ChapterStatusException)
		return
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
		return
	}
	cp.PreURL, err = getChapterURL(result[4].String()) // 获取上一章链接
	if err != nil {
		return
	}
	cp.NextURL, err = getChapterURL(result[5].String()) // 获取下一章链接
	return
}

// 获取章节链接
func getChapterURL(id string) (string, error) {
	values := make(url.Values)
	values.Set(APIItemID, id)
	preURL, err := url.Parse(APIContentURL)
	if err != nil {
		return ``, err
	}
	preURL.RawQuery = values.Encode()
	return preURL.String(), nil
}

// IsUpdate 书籍更新检测
func IsUpdate(upd, rec string) bool {
	return cmp.Or(
		API.ChapterID(upd),
		Platform.ChapterID(upd),
	) == cmp.Or(
		API.ChapterID(rec),
		Platform.ChapterID(rec),
	)
}

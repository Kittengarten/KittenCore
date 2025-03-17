package track

import (
	"cmp"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/http"

	"github.com/tidwall/gjson"
)

const (
	FQAPI           Platform = `番茄 API`
	fqAPIHOST                = `https://api.v2.sukimon.me:45554`
	fqDetailURL              = fqAPIHOST + `/detail`
	fqAPICatalogURL          = fqAPIHOST + `/catalog`
	fqAPIContentURL          = fqAPIHOST + `/content`
	fqAPISearchURL           = fqAPIHOST + `/search`
	fqAPIBookId              = `book_id`  // 书号
	fqAPIItemId              = `item_id`  // 章节号
	fqAPIQuery               = `query`    // 关键词
	fqAPIOffset              = `offset`   // 10 * (page - 1)
	fqAPITabType             = `tab_type` // 1：综合，2：听书，3：书籍，4：社区，5：全文，8：漫画，11：短句
)

// 小说网页信息获取
func (nv *novel) initFQAPI(bookID string) error {
	// 初始化小说平台
	nv.Platform = FQAPI
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = fqURL + nv.id
	values := make(url.Values)
	values.Add(fqAPIBookId, bookID)
	apiURL, err := url.Parse(fqDetailURL)
	if err != nil {
		return err
	}
	apiURL.RawQuery = values.Encode()
	// 获取小说网页，失败则返回
	data, err := http.GETData(apiURL.String())
	if err != nil {
		return err
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		return errStatus(nv.url, bookStatusException)
	}
	if gjson.GetBytes(data, `message`).String() != `SUCCESS` {
		return errStatus(nv.url, bookUnreachable)
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
	nv.right = []string{`免费`}          // 获取上架状态（番茄均为免费）
	nv.name = result[0].String()       // 获取书名
	nv.writer = result[1].String()     // 获取作者
	nv.headURL = result[2].String()    // 获取头像链接
	nv.collection = result[3].String() // 获取收藏
	nv.hitNum = result[4].String()     // 获取点击
	nv.wordNum = result[5].String()    // 获取小说字数
	nv.coverURL = cmp.Or(result[6].String(), result[7].String())
	nv.status = func() string {
		if v, ok := map[int64]string{
			0: `已完结`,
			1: `连载中`,
			4: `断更`,
		}[result[8].Int()]; ok {
			return v
		}
		return `未知`
	}() // 获取状态
	nv.tagList = strings.Split(result[9].String(), `,`) // 获取标签
	nv.introduce = result[10].String()                  // 获取简述
	newChapter := result[11].String()                   // 获取新章节链接
	// 不支持的字段
	nv.theme = ``
	nv.item = nil
	nv.preview = ``
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	clear(values)
	values.Add(fqAPIItemId, newChapter)
	newChapterURL, err := url.Parse(fqAPIContentURL)
	if err != nil {
		return err
	}
	newChapterURL.RawQuery = values.Encode()
	return nv.newChapter.init(FQAPI, newChapterURL.String())
}

// 小说章节信息获取
func (cp *chapter) initFQAPI(cURL string) error {
	// 向章节传入链接
	cp.url = cURL
	// 获取章节网页，失败则返回
	data, err := http.GETData(cp.url)
	if err != nil {
		return err
	}
	if !gjson.ValidBytes(data) {
		kitten.Debugln(`无效的 JSON：`, string(data))
		return errStatus(cp.url, chapterStatusException)
	}
	result := gjson.GetManyBytes(data,
		`data.novel_data.first_pass_time`,     // 更新时间
		`data.novel_data.volume_name`,         // 分卷名称
		`data.novel_data.title`,               // 章节名称
		`data.novel_data.chapter_word_number`, // 章节字数
		`data.novel_data.pre_item_id`,         // 上章网址
		`data.novel_data.next_item_id`,        // 下章网址
	)
	cp.Time = time.Unix(result[0].Int(), 0).Local() // 获取更新时间
	cp.title = result[1].String() + `
` + result[2].String() // 获取章节名称
	cp.wordNum = int(result[3].Int()) // 获取章节字数
	if cp.wordNum <= 0 {
		return errStatus(cp.url, chapterStatusException)
	}
	values := make(url.Values)
	values.Add(fqAPIItemId, result[4].String())
	preURL, err := url.Parse(fqAPIContentURL)
	if err != nil {
		return err
	}
	preURL.RawQuery = values.Encode()
	values.Set(fqAPIItemId, result[5].String())
	nextURL, err := url.Parse(fqAPIContentURL)
	if err != nil {
		return err
	}
	nextURL.RawQuery = values.Encode()
	cp.preURL = preURL.String()   // 获取上一章链接
	cp.nextURL = nextURL.String() // 获取下一章链接
	return nil
}

// 用关键词搜索书号
func (key keyword) findFQAPIBookID() (string, error) {
	values := make(url.Values)
	values.Add(fqAPIQuery, string(key))
	const page = 1 // 默认搜索第一页
	values.Add(fqAPIOffset, strconv.FormatInt(10*(page-1), 10))
	values.Add(fqAPITabType, `3`) // 默认搜索类型（3：小说）
	searchURL, err := url.Parse(fqAPISearchURL)
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
		return ``, notFound(key)
	}
	const number = 1 // 默认取第一本书
	return bookIDs[number-1].String(), nil
}

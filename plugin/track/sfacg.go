package track

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/http"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type stringCount = int // 用于判断的字数

const (
	SF               Platform    = `SF轻小说`
	sfHost                       = `https://book.sfacg.com`
	sfURL                        = sfHost + `/Novel/`
	sfDateTime                   = `2006/1/2 15:04:05`
	sfBookStrings    stringCount = 43
	sfChapterStrings stringCount = 38
)

// 小说网页信息获取
func (nv *novel) initSF(bookID string) error {
	// 初始化小说平台
	nv.Platform = SF
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = sfURL + nv.id + `/`
	// 获取小说网页，失败则返回
	doc, err := htmlquery.LoadURL(nv.url)
	if err != nil {
		return err
	}
	if err := mayExist(doc, nv.url, sfBookStrings); err != nil {
		return err
	}
	// 获取书名
	nv.name = strings.TrimSpace(http.InnerText(doc, `//h1[@class="title"]/span/text()`))
	// 获取小说版权状态与项目
	nv.getNovelRightItem(doc)
	// 获取作者
	nv.writer = http.InnerText(doc, `//div[@class="author-name"]/span`)
	// 获取头像链接
	nv.headURL = http.InnerText(doc, `//div[@class="author-mask"]//img/@src`)
	// 小说详细信息
	textRow := htmlquery.Find(doc, `//div[@class="text-row"]/span`)
	// 获取类型
	nv.theme = strings.TrimPrefix(htmlquery.InnerText(textRow[0]), `类型：`)
	// 获取小说字数信息
	nv.wordNum = str.Mid(`字数：`, `字[`, htmlquery.InnerText(textRow[1]))
	// 获取状态
	nv.status = str.Mid(`[`, `]`, htmlquery.InnerText(textRow[1]))
	// 获取点击
	nv.hitNum = strings.TrimPrefix(htmlquery.InnerText(textRow[2]), `点击：`)
	// 获取简述
	nv.introduce = http.InnerText(doc, `//p[@class="introduce"]`)
	// 获取移动版简述
	if introduceMobile, err := nv.getIntroduce(); err == nil &&
		len(introduceMobile) >= len(nv.introduce) {
		nv.introduce = introduceMobile
	}
	// 获取收藏
	nv.collection = strings.TrimPrefix(
		http.InnerText(doc, `//div[@id="BasicOperation"]/a[3]`), `收藏 `)
	// 获取标签
	nv.tagList = utils.ConvertSlice(
		htmlquery.Find(doc, `//li[starts-with(@class,"tag")]/a/span[@class="text"]`),
		func(n *html.Node) string {
			return str.CleanAll(htmlquery.InnerText(n), false)
		},
	)
	// 获取封面链接
	nv.coverURL = http.InnerText(doc, `//div[@class="figure"]//img/@src`)
	// 获取预览
	nv.preview = strings.TrimPrefix(str.CleanAll(strings.ReplaceAll(http.InnerText(
		doc, `//div[@class="chapter-info"]/p`), `　　`, "\n"), true), "\n")
	// 获取新章节链接
	newChapter := htmlquery.FindOne(doc, `//div[@class="chapter-info"]/h3/a/@href`)
	if newChapter == nil {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新章节链接错误：%w`, errStatus(nv.url, noChapterURL))
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	if err := nv.newChapter.init(SF, sfHost+htmlquery.InnerText(newChapter)); err != nil {
		return err
	}
	// 如果不是 VIP 书籍，直接返回
	if !slices.Contains(nv.right, `VIP`) {
		return nil
	}
	// 检查是否存在最新章节
	return nv.checkUpdateSF(doc)
}

// 检查是否存在最新章节
func (nv *novel) checkUpdateSF(doc *html.Node) error {
	// 尝试获取新公众章节链接
	newChapterPublicNode := htmlquery.FindOne(doc, `//div[@class="chapter-info"]/div/a/@href`)
	if newChapterPublicNode == nil {
		// 如果新公众章节不存在，直接返回
		return nil
	}
	newChapterPublicURL := htmlquery.InnerText(newChapterPublicNode)
	if newChapterPublicURL == `` {
		// 如果新公众章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新公众章节链接错误：%w`, errStatus(nv.url, noChapterURL))
	}
	// 从章节池初始化章节，向章节传入本书链接
	newChapterFree := *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&newChapterFree)
	newChapterFree.bookURL = nv.url
	// 加载最新公众章节
	if err := newChapterFree.initSF(sfHost + newChapterPublicURL); err != nil {
		return err
	}
	// 如果最新公众章节比最新章节新，则以最新公众章节为准
	if newChapterFree.After(nv.newChapter.Time) {
		nv.newChapter = newChapterFree
	}
	return nil
}

// 章节信息获取
func (cp *chapter) initSF(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url+`/` == cp.bookURL {
		return errStatus(url, chapterURLException)
	}
	// 向章节传入链接
	cp.url = url
	// 获取章节网页，失败则返回
	doc, err := htmlquery.LoadURL(cp.url)
	if err != nil {
		return err
	}
	if err := mayExist(doc, cp.url, sfChapterStrings); err != nil {
		return err
	}
	// 获取章节标题
	cp.title = http.InnerText(doc, `//h1[@class="article-title"]`)
	// 获取更新时间
	cp.Time, err = SF.ParseTime(strings.TrimPrefix(
		http.InnerText(doc, `//div[@class="article-desc"]/span[2]`), `更新时间：`))
	if err != nil {
		return err
	}
	// 获取章节字数
	wordNum := http.InnerText(doc, `//div[@class="article-desc"]/span[3]`)
	cp.wordNum, err = strconv.Atoi(strings.TrimPrefix(wordNum, `字数：`))
	if err != nil {
		return fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, url, err)
	}
	// 获取上一章链接
	cp.preURL = sfHost +
		http.InnerText(doc, `//div[@id="article"]/div[@class="fn-btn"]/a[1]/@href`)
	// 获取下一章链接
	cp.nextURL = sfHost +
		http.InnerText(doc,
			`//div[@id="article"]/div[@class="fn-btn"]/a[2]/@href`)
	// 获取付费状态
	cp.isVIP = strings.Contains(url, `vip`)
	return nil
}

// 用关键词搜索书号
func (key keyword) findSFBookID() (string, error) {
	doc, err := htmlquery.LoadURL(fmt.Sprint(`http://s.sfacg.com/?Key=`, key, `&S=1&SS=0`))
	if err != nil {
		return ``, err
	}
	url, err := http.InnerText(doc, `//a[@id="SearchResultList1___ResultList_LinkInfo_0"]/@href`), nil
	if url == `` {
		err = notFound(key)
	}
	return strings.TrimPrefix(url, sfURL), err
}

// 获取移动版简述
func (nv *novel) getIntroduce() (string, error) {
	doc, err := htmlquery.LoadURL(`https://m.sfacg.com/b/` + nv.id + `/`)
	if err != nil {
		return ``, err
	}
	return str.Compose(nil, http.InnerText(doc,
		`//ul[@class="book_profile"]/li[@class="book_bk_qs1"]`),
	), mayExist(doc, nv.url, sfBookStrings)
}

// 判断小说或章节是否可能存在
func mayExist(doc *html.Node, url string, count stringCount) error {
	if len(http.InnerText(doc, `//title`)) >= count {
		return nil
	}
	return errStatus(url, bookUnreachable)
}

// 获取小说版权状态与项目
func (nv *novel) getNovelRightItem(doc *html.Node) {
	for _, tt := range htmlquery.Find(doc,
		`//h1[@class="title"]//span[starts-with(@class,"tag")]`) {
		switch b, y, g := http.InnerText(tt, `.[contains(@class,"blue")]`),
			http.InnerText(tt, `.[contains(@class,"yellow")]`),
			http.InnerText(tt, `.[contains(@class,"green")]`); {
		case b != ``:
			// 获取版权状态
			nv.right = append(nv.right, b)
		case y != ``:
			// 获取版权状态
			nv.right = append(nv.right, y)
		case g != ``:
			// 获取项目
			nv.item = append(nv.item, g)
		}
	}
}

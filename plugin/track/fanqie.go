package track

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core"

	"github.com/antchfx/htmlquery"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

const (
	FQ            Platform    = `番茄小说网`
	fqHOST                    = `https://fanqienovel.com`
	fqURL                     = fqHOST + `/page/`
	fqChapterURL              = fqHOST + `/reader/`
	fqSearchURL               = fqHOST + `/search/`
	fqDateTime                = `2006-01-02T15:04:05`
	fqBookStrings stringCount = 38
)

// 小说网页信息获取
func (nv *novel) initFQ(bookID string) error {
	// 初始化小说平台
	nv.Platform = FQ
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = fqURL + nv.id
	// 获取小说网页，失败则返回
	core.SetUserAgent(core.RandomUserAgent())
	defer core.SetUserAgent(core.UserAgent)
	doc, err := htmlquery.LoadURL(nv.url)
	if err != nil {
		return err
	}
	if !strings.Contains(core.InnerText(doc, `//title`), `在线免费阅读`) {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取封面
	nv.coverURL = gjson.Get(core.InnerText(doc,
		`//script[@type="application/ld+json"]`), `image.0`).String()
	// 获取书名
	nv.name = core.InnerText(doc, `//div[@class="info-name"]`)
	// 获取小说信息
	bookInfo := htmlquery.FindOne(doc, `//div[@class="info-label"]`)
	// 获取小说状态
	nv.status = core.InnerText(bookInfo, `//span[@class="info-label-yellow"]`)
	// 获取标签
	nv.tagList = core.ConvertSlice(
		htmlquery.Find(bookInfo, `//span[@class="info-label-grey"]`),
		func(n *html.Node) string {
			return htmlquery.InnerText(n)
		},
	)
	// 获取小说字数
	nv.wordNum = func(wordNum *html.Node) (s string) {
		s = core.InnerText(wordNum, `/span[@class="detail"]`)
		if core.InnerText(wordNum, `/span[@class="text"]`) == `万字` {
			s += `万`
		}
		return
	}(htmlquery.FindOne(doc, `//div[@class="info-count-word"]`))
	// 获取头像链接
	nv.headURL = core.InnerText(doc, `//img[@class="author-img"]/@src`)
	// 获取作者
	nv.writer = core.InnerText(doc, `//span[@class="author-name-text"]`)
	// 获取简述
	nv.introduce = core.InnerText(doc, `//div[@class="page-abstract-content"]/p`)
	// 获取新章节链接
	newChapter := core.InnerText(doc, `//div[@class="info-last"]/a[@class="chapter-item-title"]/@href`)
	// 获取上架状态（番茄均为免费）
	nv.right = []string{`免费`}
	// 不支持的字段
	nv.theme = ``
	nv.item = nil
	nv.collection = ``
	nv.hitNum = ``
	nv.preview = ``
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	return nv.newChapter.init(FQ, fqHOST+newChapter)
}

// 章节信息获取
func (cp *chapter) initFQ(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url == `` {
		return errStatus(url, chapterURLException)
	}
	// 向章节传入链接
	cp.url = url
	// 获取章节网页，失败则返回
	core.SetUserAgent(core.RandomUserAgent())
	defer core.SetUserAgent(core.UserAgent)
	doc, err := htmlquery.LoadURL(cp.url)
	if err != nil {
		return err
	}
	if !strings.Contains(core.InnerText(doc, `//title`), `在线免费阅读`) {
		return errStatus(cp.url, chapterUnreachable)
	}
	// 获取章节标题
	cp.title = core.InnerText(doc, `//h1[@class="muye-reader-title"]`)
	// 获取章节字数
	wordNum := core.InnerText(doc, `//span[@class="desc-item"]`)
	cp.wordNum, err = strconv.Atoi(core.MidText(`本章字数：`, `字`, wordNum))
	if err != nil {
		return fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, url, err)
	}
	js := func() string {
		for _, t := range htmlquery.Find(doc, `//script`) {
			const pre = `window.__INITIAL_STATE__=`
			s := htmlquery.InnerText(t)
			if !strings.Contains(s, pre) {
				continue
			}
			return core.MidText(pre, `;`, s)
		}
		return ``
	}()
	// 获取更新时间
	cp.Time, err = FQ.ParseTime(gjson.Get(core.InnerText(doc,
		`//script[@type="application/ld+json"]`), `dateModified`).String())
	if err != nil {
		return err
	}
	// 章节数据
	cpData := gjson.Get(js, `reader.chapterData`)
	// 获取上一章链接
	cp.preURL = fqChapterURL + cpData.Get(`preItemId`).String()
	// 获取下一章链接
	cp.nextURL = fqChapterURL + cpData.Get(`nextItemId`).String()
	return nil
}

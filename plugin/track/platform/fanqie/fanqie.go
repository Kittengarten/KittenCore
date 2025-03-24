package fanqie

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/htmls"
	"github.com/Kittengarten/KittenCore/kitten/core/shttp"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
	"github.com/Kittengarten/KittenCore/plugin/track/search"
	"github.com/Kittengarten/KittenCore/plugin/track/status"

	"github.com/antchfx/htmlquery"
	"github.com/tidwall/gjson"
	"golang.org/x/net/html"
)

// 番茄小说网
type fq struct{}

const (
	HOST       = `https://fanqienovel.com`
	URL        = HOST + `/page/`
	ChapterURL = HOST + `/reader/`
)

// Platform  番茄小说网
var Platform = fq{}

func init() {
	platform.Platforms = append(platform.Platforms, Platform)
}

// String 实现 fmt.Stringer，返回小说平台名称
func (f fq) String() string {
	return `番茄小说网`
}

// Layout 返回时间格式
func (f fq) Layout() string {
	return `2006-01-02T15:04:05`
}

// FindBookID 番茄网页暂不支持用关键词搜索书号
func (f fq) FindBookID(key search.Keyword) (string, error) {
	return ``, platform.NotSupported(f)
}

// ChapterID 获取章号
func (f fq) ChapterID(cpURL string) string {
	u, err := url.Parse(cpURL)
	if err != nil {
		return ``
	}
	return path.Base(u.Path)
}

// Init 小说网页信息获取
func (f fq) Init(bookID string) (nv *novel.Novel, err error) {
	// 初始化小说
	nv = novel.Pool.Get().(*novel.Novel)
	// 初始化小说平台
	nv.Platform = f.String()
	// 向小说传入书号
	nv.ID = bookID
	// 生成链接
	nv.URL = URL + nv.ID
	// 获取小说网页，失败则返回
	shttp.SetUserAgent(shttp.RandomUserAgent())
	defer shttp.SetUserAgent(shttp.UserAgent)
	doc, err := htmlquery.LoadURL(nv.URL)
	if err != nil {
		return
	}
	if !strings.Contains(htmls.InnerText(doc, `//title`), `在线免费阅读`) {
		err = status.ErrStatus(nv.URL, status.BookUnreachable)
		return
	}
	// 获取封面
	nv.CoverURL = gjson.Get(htmls.InnerText(doc,
		`//script[@type="application/ld+json"]`), `image.0`).String()
	// 获取书名
	nv.Name = htmls.InnerText(doc, `//div[@class="info-name"]`)
	// 获取小说信息
	bookInfo := htmlquery.FindOne(doc, `//div[@class="info-label"]`)
	// 获取小说状态
	nv.Status = htmls.InnerText(bookInfo, `//span[@class="info-label-yellow"]`)
	// 获取标签
	nv.TagList = utils.ConvertSlice(
		htmlquery.Find(bookInfo, `//span[@class="info-label-grey"]`),
		htmlquery.InnerText,
	)
	// 获取小说字数
	nv.TotalWordNum = func(wordNum *html.Node) (s string) {
		s = htmls.InnerText(wordNum, `/span[@class="detail"]`)
		if htmls.InnerText(wordNum, `/span[@class="text"]`) == `万字` {
			s += `万`
		}
		return
	}(htmlquery.FindOne(doc, `//div[@class="info-count-word"]`))
	// 获取头像链接
	nv.HeadURL = htmls.InnerText(doc, `//img[@class="author-img"]/@src`)
	// 获取作者
	nv.Writer = htmls.InnerText(doc, `//span[@class="author-name-text"]`)
	// 获取简述
	nv.Introduce = htmls.InnerText(doc, `//div[@class="page-abstract-content"]/p`)
	// 获取新章节链接
	ncp := htmls.InnerText(doc, `//div[@class="info-last"]/a[@class="chapter-item-title"]/@href`)
	// 获取上架状态（番茄均为免费）
	nv.Right = []string{`免费`}
	// 不支持的字段
	nv.Theme = ``
	nv.Item = nil
	nv.Collection = ``
	nv.HitNum = ``
	nv.Preview = ``
	ncpURL := HOST + ncp
	// 防止章节炸了导致获取章节跳转引发 panic
	if ncpURL+`/` == nv.URL {
		err = status.ErrStatus(ncpURL, status.ChapterURLException)
		return
	}
	// 加载新章节
	nv.Chapter, err = f.NewChapter(ncpURL)
	return
}

// NewChapter 章节信息获取
func (f fq) NewChapter(cpURL string) (cp *chapter.Chapter, err error) {
	// 初始化章节
	cp = chapter.Pool.Get().(*chapter.Chapter)
	// 向章节传入链接
	cp.URL = cpURL
	// 获取章节网页，失败则返回
	shttp.SetUserAgent(shttp.RandomUserAgent())
	defer shttp.SetUserAgent(shttp.UserAgent)
	doc, err := htmlquery.LoadURL(cp.URL)
	if err != nil {
		return
	}
	if !strings.Contains(htmls.InnerText(doc, `//title`), `在线免费阅读`) {
		err = status.ErrStatus(cp.URL, status.ChapterUnreachable)
		return
	}
	// 获取章节标题
	cp.Title = htmls.InnerText(doc, `//h1[@class="muye-reader-title"]`)
	// 获取章节字数
	wordNum := htmls.InnerText(doc, `//span[@class="desc-item"]`)
	cp.WordNum, err = strconv.Atoi(str.Mid(`本章字数：`, `字`, wordNum))
	if err != nil {
		err = fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, cp.URL, err)
		return
	}
	js := func() string {
		for _, t := range htmlquery.Find(doc, `//script`) {
			const pre = `window.__INITIAL_STATE__=`
			s := htmlquery.InnerText(t)
			if !strings.Contains(s, pre) {
				continue
			}
			return str.Mid(pre, `;`, s)
		}
		return ``
	}()
	// 获取更新时间
	cp.Time, err = platform.ParseTime(f, gjson.Get(htmls.InnerText(doc,
		`//script[@type="application/ld+json"]`), `dateModified`).String())
	if err != nil {
		return
	}
	// 章节数据
	cpData := gjson.Get(js, `reader.chapterData`)
	// 获取上一章链接
	cp.PreURL = ChapterURL + cpData.Get(`preItemId`).String()
	// 获取下一章链接
	cp.NextURL = ChapterURL + cpData.Get(`nextItemId`).String()
	return
}

// Package sfacg SF 轻小说
package sfacg

import (
	"context"
	"fmt"
	"net/url"
	"slices"
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
	"golang.org/x/net/html"
)

type (
	// SF 轻小说
	SF utils.Object
	// 用于判断的字数
	minLen = int
)

const (
	// Host 网站
	Host = `https://book.sfacg.com`
	// URL 链接
	URL = Host + `/Novel/`
	// BookStrings 书名长度
	BookStrings minLen = 43
	// ChapterStrings 章名长度
	ChapterStrings minLen = 38
)

// Platform SF 轻小说
var Platform = SF{}

func init() {
	platform.Register(Platform)
}

// String 实现 fmt.Stringer，返回小说平台名称
func (s SF) String() string {
	return `SF轻小说`
}

// Layout 返回时间格式
func (s SF) Layout() string {
	return `2006/1/2 15:04:05`
}

// FindBookID 用关键词搜索书号
func (s SF) FindBookID(ctx context.Context, key search.Keyword) (string, error) {
	doc, err := shttp.LoadURLWithContext(
		ctx, fmt.Sprint(`http://s.sfacg.com/?Key=`, key, `&S=1&SS=0`))
	if err != nil {
		return ``, err
	}
	url, err := htmls.InnerText(
		doc,
		`//a[@id="SearchResultList1___ResultList_LinkInfo_0"]/@href`,
	), nil
	if url == `` {
		err = key.NotFound()
	}
	return strings.TrimPrefix(url, URL), err
}

// ChapterID 获取章号
func (s SF) ChapterID(cpURL string) string {
	u, err := url.Parse(cpURL)
	if err != nil {
		return ``
	}
	return u.Path
}

// Init 小说网页信息获取
func (s SF) Init(ctx context.Context, cpID string, cache bool) (any, error) {
	// 初始化小说
	nv := novel.Pool.Get().(*novel.Novel)
	*nv = novel.Novel{}
	// 初始化小说平台
	nv.Platform = s.String()
	// 向小说传入书号
	nv.ID = cpID
	// 生成链接
	nv.URL = URL + nv.ID + `/`
	// 获取小说网页，失败则返回
	doc, err := platform.Doc(ctx, nv.URL, cache)
	if err != nil {
		return nv, err
	}
	if err = mayExist(doc, nv.URL, BookStrings); err != nil {
		return nv, err
	}
	// 获取书名
	nv.Name = strings.TrimSpace(htmls.InnerText(doc, `//h1[@class="title"]/span/text()`))
	// 获取小说版权状态与项目
	getNovelRightItem(nv, doc)
	// 获取作者
	nv.Writer = htmls.InnerText(doc, `//div[@class="author-name"]/span`)
	// 获取头像链接
	nv.HeadURL = htmls.InnerText(doc, `//div[@class="author-mask"]//img/@src`)
	// 小说详细信息
	textRow := htmlquery.Find(doc, `//div[@class="text-row"]/span`)
	novel.CheckComplete(`详细信息`, textRow, 3, func() {
		// 获取类型
		nv.Theme = strings.TrimPrefix(htmlquery.InnerText(textRow[0]), `类型：`)
		textRow1 := htmlquery.InnerText(textRow[1])
		// 获取小说字数信息
		nv.TotalWordNum = str.Mid(`字数：`, `字[`, textRow1)
		// 获取状态
		nv.Status = str.Mid(`[`, `]`, textRow1)
		// 获取点击
		nv.HitNum = strings.TrimPrefix(htmlquery.InnerText(textRow[2]), `点击：`)
	})
	// 获取简述
	nv.Introduce = htmls.InnerText(doc, `//p[@class="introduce"]`)
	// 获取移动版简述
	if introduceMobile, err := getIntroduce(ctx, nv); err == nil &&
		len(introduceMobile) >= len(nv.Introduce) {
		nv.Introduce = strings.TrimSpace(introduceMobile)
	}
	// 获取收藏
	nv.Collection = strings.TrimPrefix(
		htmls.InnerText(doc, `//div[@id="BasicOperation"]/a[3]`), `收藏 `)
	// 获取标签
	nv.TagList = utils.ConvertSlice(
		htmlquery.Find(doc, `//li[starts-with(@class,"tag")]/a/span[@class="text"]`),
		func(n *html.Node) string {
			return str.Clean(htmlquery.InnerText(n), false)
		},
	)
	// 获取封面链接
	nv.CoverURL = htmls.InnerText(doc, `//div[@class="figure"]//img/@src`)
	// 获取预览
	nv.Preview = strings.TrimPrefix(str.Clean(strings.ReplaceAll(htmls.InnerText(
		doc, `//div[@class="chapter-info"]/p`), `　　`, "\n"), true), "\n")
	// 获取新章节链接
	ncp := htmlquery.FindOne(doc, `//div[@class="chapter-info"]/h3/a/@href`)
	if ncp == nil {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		err = fmt.Errorf(`新章节链接错误：%w`, status.ErrStatus(nv.URL, status.NoChapterURL))
		return nv, err
	}
	ncpURL := Host + htmlquery.InnerText(ncp)
	// 防止章节炸了导致获取章节跳转引发 panic
	if ncpURL+`/` == nv.URL {
		err = status.ErrStatus(ncpURL, status.ChapterURLException)
		return nv, err
	}
	// 加载新章节
	nv.Chapter, err = chapter.New(ctx, s, ncpURL)
	if err != nil {
		return nv, err
	}
	// 如果不是 VIP 书籍，直接返回
	if !slices.Contains(nv.Right, `VIP`) {
		return nv, err
	}
	// 是 VIP 书籍，检查是否存在最新章节
	err = s.checkUpdate(ctx, nv, doc)
	return nv, err
}

// 获取小说版权状态与项目
func getNovelRightItem(nv *novel.Novel, doc *html.Node) {
	for _, tt := range htmlquery.Find(doc,
		`//h1[@class="title"]//span[starts-with(@class,"tag")]`) {
		switch b, y, g := htmls.InnerText(tt, `.[contains(@class,"blue")]`),
			htmls.InnerText(tt, `.[contains(@class,"yellow")]`),
			htmls.InnerText(tt, `.[contains(@class,"green")]`); {
		case b != ``:
			// 获取版权状态
			nv.Right = append(nv.Right, b)
		case y != ``:
			// 获取版权状态
			nv.Right = append(nv.Right, y)
		case g != ``:
			// 获取项目
			nv.Item = append(nv.Item, g)
		}
	}
}

// 获取移动版简述
func getIntroduce(ctx context.Context, nv *novel.Novel) (string, error) {
	doc, err := shttp.LoadURLWithContext(ctx, `https://m.sfacg.com/b/`+nv.ID+`/`)
	if err != nil {
		return ``, err
	}
	return str.ComposeAuto(htmls.InnerText(doc,
		`//ul[@class="book_profile"]/li[@class="book_bk_qs1"]`),
	), mayExist(doc, nv.URL, BookStrings)
}

// 检查是否存在最新章节
func (s SF) checkUpdate(ctx context.Context, nv *novel.Novel, doc *html.Node) error {
	// 尝试获取新公众章节链接
	ncpPublicNode := htmlquery.FindOne(doc, `//div[@class="chapter-info"]/div/a/@href`)
	if ncpPublicNode == nil {
		// 如果新公众章节不存在，直接返回
		return nil
	}
	ncpPublicURL := htmlquery.InnerText(ncpPublicNode)
	if ncpPublicURL == `` {
		// 如果新公众章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新公众章节链接错误：%w`, status.ErrStatus(nv.URL, status.NoChapterURL))
	}
	ncpPublicURL = Host + ncpPublicURL
	// 防止章节炸了导致获取章节跳转引发 panic
	if ncpPublicURL+`/` == nv.URL {
		return status.ErrStatus(ncpPublicURL, status.ChapterURLException)
	}
	// 加载最新公众章节
	ncpFree, err := chapter.New(ctx, s, ncpPublicURL)
	if err != nil {
		return err
	}
	// 如果最新公众章节比最新章节新，则以最新公众章节为准
	if ncpFree.Update.After(nv.Chapter.Update) {
		nv.Chapter = ncpFree
	}
	return nil
}

// NewChapter 章节信息获取
func (s SF) NewChapter(ctx context.Context, cpURL string) (any, error) {
	// 初始化章节
	cp := chapter.Pool.Get().(*chapter.Chapter)
	*cp = chapter.Chapter{}
	// 向章节传入链接
	cp.URL = cpURL
	// 获取章节网页，失败则返回
	doc, err := shttp.LoadURLWithContext(ctx, cp.URL)
	if err != nil {
		return cp, err
	}
	if err = mayExist(doc, cp.URL, ChapterStrings); err != nil {
		return cp, err
	}
	// 获取章节标题
	cp.Title = htmls.InnerText(doc, `//h1[@class="article-title"]`)
	// 获取更新时间
	cp.Update, err = platform.ParseTime(s, strings.TrimPrefix(
		htmls.InnerText(doc, `//div[@class="article-desc"]/span[2]`), `更新时间：`))
	if err != nil {
		return cp, err
	}
	// 获取章节字数
	wordNum := htmls.InnerText(doc, `//div[@class="article-desc"]/span[3]`)
	cp.WordNum, err = strconv.Atoi(strings.TrimPrefix(wordNum, `字数：`))
	if err != nil {
		err = fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, cp.URL, err)
		return cp, err
	}
	// 获取上一章链接
	cp.PreURL = Host + htmls.InnerText(doc, `//div[@id="article"]/div[@class="fn-btn"]/a[1]/@href`)
	// 获取下一章链接
	cp.NextURL = Host + htmls.InnerText(doc, `//div[@id="article"]/div[@class="fn-btn"]/a[2]/@href`)
	// 获取付费状态
	cp.IsVIP = strings.Contains(cp.URL, `vip`)
	return cp, err
}

// 判断小说或章节是否可能存在
func mayExist(doc *html.Node, url string, n minLen) error {
	if len(htmls.InnerText(doc, `//title`)) >= n {
		return nil
	}
	return status.ErrStatus(url, status.BookUnreachable)
}

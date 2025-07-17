// Package ciweimao 刺猬猫阅读
package ciweimao

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/htmls"
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

// CWM 刺猬猫阅读
type CWM struct{}

const (
	// Host 网站
	Host = `https://www.ciweimao.com`
	// URL 链接
	URL = Host + `/book/`
)

// Platform 刺猬猫阅读
var Platform = CWM{}

func init() {
	platform.Register(Platform)
}

// String 实现 fmt.Stringer，返回小说平台名称
func (c CWM) String() string {
	return `刺猬猫阅读`
}

// Layout 返回时间格式
func (c CWM) Layout() string {
	return time.DateTime
}

// FindBookID 用关键词搜索书号
func (c CWM) FindBookID(key search.Keyword) (string, error) {
	doc, err := htmlquery.LoadURL(
		fmt.Sprint(Host, `/get-search-book-list/0-0-0-0-0-0/全部/`, key, `/1`),
	)
	if err != nil {
		return ``, err
	}
	href := htmlquery.FindOne(doc, `//a[@class="cover"]/@href`)
	if href == nil {
		return ``, key.NotFound()
	}
	return strings.TrimPrefix(htmlquery.InnerText(href), URL), nil
}

// ChapterID 获取章号
func (c CWM) ChapterID(cpURL string) string {
	u, err := url.Parse(cpURL)
	if err != nil {
		return ``
	}
	return path.Base(u.Path)
}

// Init 小说网页信息获取
func (c CWM) Init(cpID string) (any, error) {
	// 初始化小说
	nv := novel.Pool.Get().(*novel.Novel)
	*nv = novel.Novel{}
	// 初始化小说平台
	nv.Platform = c.String()
	// 向小说传入书号
	nv.ID = cpID
	// 生成链接
	nv.URL = URL + nv.ID
	// 获取小说网页，失败则返回
	doc, err := htmlquery.LoadURL(nv.URL)
	if err != nil {
		return nv, err
	}
	if htmls.InnerText(doc, `//title`) == `刺猬猫` {
		err = status.ErrStatus(nv.URL, status.BookUnreachable)
		return nv, err
	}
	// 获取小说信息
	bookInfo := htmlquery.FindOne(doc, `//div[@class="book-info"]`)
	if bookInfo != nil {
		// 获取书名
		nv.Name = htmls.InnerText(bookInfo, `/h1[@class="title"]/text()`)
		// 获取作者
		nv.Writer = htmls.InnerText(bookInfo, `/h1[@class="title"]/span/a`)
		// 获取标签
		nv.TagList = utils.ConvertSlice(
			htmlquery.Find(bookInfo, `/p/span[starts-with(@class,"label")]/a`),
			func(n *html.Node) string {
				return str.CleanAll(htmlquery.InnerText(n), false)
			},
		)
		// 获取小说状态
		nv.Status = htmls.InnerText(bookInfo, `/p[@class="update-state"]`)
		// 获取小说成绩
		bookGrade := htmlquery.Find(bookInfo, `/p[@class="book-grade"]/b`)
		if len(bookGrade) >= 3 {
			// 获取小说点击
			nv.HitNum = htmlquery.InnerText(bookGrade[0])
			// 获取小说收藏
			nv.Collection = htmlquery.InnerText(bookGrade[1])
			// 获取小说字数
			nv.TotalWordNum = htmlquery.InnerText(bookGrade[2])
		}
	}
	// 获取项目
	if item := htmlquery.FindOne(doc, `//div[starts-with(@class,"book-desc")]/p`); item != nil {
		nv.Item = append(nv.Item, str.Mid(`【`, `】`, htmlquery.InnerText(item)))
	}
	// 获取简述
	var s strings.Builder
	for _, i := range htmlquery.Find(doc, `//div[starts-with(@class,"book-desc")]/text()`) {
		s.WriteString(str.CleanAll(htmlquery.InnerText(i), false))
	}
	nv.Introduce = s.String()
	// 获取小说数据
	property := htmlquery.Find(doc, `//div[starts-with(@class,"book-property")]/span/i`)
	if len(property) < 9 {
		err = status.ErrStatus(nv.URL, status.BookStatusException)
		return nv, err
	}
	// 获取上架状态
	nv.Right = append(nv.Right, htmlquery.InnerText(property[0]))
	// 获取小说类别
	nv.Theme = htmlquery.InnerText(property[4])
	// 获取头像链接
	nv.HeadURL = htmls.InnerText(doc, `//div[@class="author-info"]//img/@data-original`)
	// 获取封面
	nv.CoverURL = htmls.InnerText(doc, `//a[@class="cover"]//img/@data-original`)
	// 获取新章节链接
	ncp := htmlquery.FindOne(doc, `//h3[@class="tit"]/a[@target]/@href[1]`)
	if ncp == nil {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		err = fmt.Errorf(`新章节链接错误：%w`, status.ErrStatus(nv.URL, status.NoChapterURL))
		return nv, err
	}
	ncpURL := htmlquery.InnerText(ncp)
	if ncpURL == nv.URL {
		err = status.ErrStatus(ncpURL, status.ChapterURLException)
		return nv, err
	}
	// 加载新章节
	nv.Chapter, err = chapter.New(c, ncpURL)
	return nv, err
}

// NewChapter 章节信息获取
func (c CWM) NewChapter(cpURL string) (any, error) {
	// 初始化章节
	cp := chapter.Pool.Get().(*chapter.Chapter)
	*cp = chapter.Chapter{}
	// 向章节传入链接
	cp.URL = cpURL
	// 获取章节网页，失败则返回
	doc, err := htmlquery.LoadURL(cp.URL)
	if err != nil {
		return cp, err
	}
	if htmls.InnerText(doc, `//title`) == `刺猬猫` {
		err = status.ErrStatus(cp.URL, status.ChapterUnreachable)
		return cp, err
	}
	// 获取章节标题
	cp.Title = htmls.InnerText(doc, `//div[@class="read-hd"]/h1[@class="chapter"]`)
	// 获取更新时间
	cp.Update, err = platform.ParseTime(c, strings.TrimPrefix(
		htmls.InnerText(doc, `//div[@class="read-hd"]/p/span[3]`), `更新时间：`))
	if err != nil {
		return cp, err
	}
	// 获取章节字数
	cp.WordNum, err = strconv.Atoi(strings.TrimPrefix(
		htmls.InnerText(doc, `//div[@class="read-hd"]/p/span[5]`), `字数：`))
	if err != nil {
		err = fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, cp.URL, err)
		return cp, err
	}
	// 获取上一章链接
	if pre := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPagePrev"]/@href`); pre != nil {
		cp.PreURL = htmlquery.InnerText(pre)
	}
	// 获取下一章链接
	if next := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPageNext"]/@href`); next != nil {
		cp.NextURL = htmlquery.InnerText(next)
	}
	// 获取付费状态
	switch htmls.InnerText(doc, `//div[@class="read-bd"]/@id`) {
	case `J_BookRead`:
		cp.IsVIP = false
	case `J_ImgRead`:
		cp.IsVIP = true
	default:
		err = status.ErrStatus(cp.URL, status.VIPChapterException)
	}
	return cp, err
}

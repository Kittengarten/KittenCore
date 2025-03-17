package track

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Kittengarten/KittenCore/kitten/core/http"
	"github.com/Kittengarten/KittenCore/kitten/core/str"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	CWM     Platform = `刺猬猫阅读`
	cwmHost          = `https://www.ciweimao.com`
	cwmURL           = cwmHost + `/book/`
)

// 小说网页信息获取
func (nv *novel) initCWM(bookID string) error {
	// 初始化小说平台
	nv.Platform = CWM
	// 向小说传入书号
	nv.id = bookID
	// 生成链接
	nv.url = cwmURL + nv.id
	// 获取小说网页，失败则返回
	doc, err := htmlquery.LoadURL(nv.url)
	if err != nil {
		return err
	}
	if http.InnerText(doc, `//title`) == `刺猬猫` {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取小说信息
	bookInfo := htmlquery.FindOne(doc, `//div[@class="book-info"]`)
	// 获取书名
	nv.name = http.InnerText(bookInfo, `/h1[@class="title"]/text()`)
	// 获取作者
	nv.writer = http.InnerText(bookInfo, `/h1[@class="title"]/span/a`)
	// 获取标签
	nv.tagList = utils.ConvertSlice(
		htmlquery.Find(bookInfo, `/p/span[starts-with(@class,"label")]/a`),
		func(n *html.Node) string {
			return str.CleanAll(htmlquery.InnerText(n), false)
		},
	)
	// 获取小说状态
	nv.status = http.InnerText(bookInfo, `/p[@class="update-state"]`)
	// 获取小说成绩
	bookGrade := htmlquery.Find(bookInfo, `/p[@class="book-grade"]/b`)
	if len(bookGrade) < 3 {
		return errStatus(nv.url, bookUnreachable)
	}
	// 获取小说点击
	nv.hitNum = htmlquery.InnerText(bookGrade[0])
	// 获取小说收藏
	nv.collection = htmlquery.InnerText(bookGrade[1])
	// 获取小说字数
	nv.wordNum = htmlquery.InnerText(bookGrade[2])
	// 获取项目
	if item := htmlquery.FindOne(doc, `//div[starts-with(@class,"book-desc")]/p`); item != nil {
		nv.item = append(nv.item, htmlquery.InnerText(item))
	}
	// 获取简述
	var s strings.Builder
	for _, i := range htmlquery.Find(doc, `//div[starts-with(@class,"book-desc")]/text()`) {
		s.WriteString(htmlquery.InnerText(i))
	}
	nv.introduce = str.CleanAll(s.String(), true)
	// 获取小说数据
	property := htmlquery.Find(doc, `//div[starts-with(@class,"book-property")]/span/i`)
	if len(property) < 9 {
		return errStatus(nv.url, bookStatusException)
	}
	// 获取上架状态
	nv.right = append(nv.right, htmlquery.InnerText(property[0]))
	// 获取小说类别
	nv.theme = htmlquery.InnerText(property[4])
	// 获取头像链接
	nv.headURL = http.InnerText(doc, `//div[@class="author-info"]//img/@data-original`)
	// 获取封面
	nv.coverURL = http.InnerText(doc, `//a[@class="cover"]//img/@data-original`)
	// 不支持的字段
	nv.preview = ``
	// 获取新章节链接
	newChapter := htmlquery.FindOne(doc, `//h3[@class="tit"]/a[@target]/@href[1]`)
	if newChapter == nil {
		// 如果新章节链接不存在，防止更新章节炸了跳转到网站首页引起程序报错
		return fmt.Errorf(`新章节链接错误：%w`, errStatus(nv.url, noChapterURL))
	}
	// 从章节池初始化章节，向章节传入本书链接
	nv.newChapter = *chapterPool.Get().(*chapter)
	defer chapterPool.Put(&nv.newChapter)
	nv.newChapter.bookURL = nv.url
	// 加载新章节
	return nv.newChapter.init(CWM, htmlquery.InnerText(newChapter))
}

// 章节信息获取
func (cp *chapter) initCWM(url string) error {
	// 防止章节炸了导致获取章节跳转引发 panic
	if url == cp.bookURL {
		return errStatus(url, chapterURLException)
	}
	// 向章节传入链接
	cp.url = url
	// 获取章节网页，失败则返回
	doc, err := htmlquery.LoadURL(cp.url)
	if err != nil {
		return err
	}
	if http.InnerText(doc, `//title`) == `刺猬猫` {
		return errStatus(cp.url, chapterUnreachable)
	}
	// 获取章节标题
	cp.title = http.InnerText(doc, `//div[@class="read-hd"]/h1[@class="chapter"]`)
	// 获取更新时间
	cp.Time, err = CWM.ParseTime(strings.TrimPrefix(
		http.InnerText(doc, `//div[@class="read-hd"]/p/span[3]`), `更新时间：`))
	if err != nil {
		return err
	}
	// 获取章节字数
	cp.wordNum, err = strconv.Atoi(strings.TrimPrefix(
		http.InnerText(doc, `//div[@class="read-hd"]/p/span[5]`), `字数：`))
	if err != nil {
		return fmt.Errorf(`章节 %s 的字数获取错误喵！%w`, url, err)
	}
	// 获取上一章链接
	if pre := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPagePrev"]/@href`); pre != nil {
		cp.preURL = htmlquery.InnerText(pre)
	}
	// 获取下一章链接
	if next := htmlquery.FindOne(doc,
		`//div[@class="book-read-page"]/a[@id="J_BtnPageNext"]/@href`); next != nil {
		cp.nextURL = htmlquery.InnerText(next)
	}
	// 获取付费状态
	switch http.InnerText(doc, `//div[@class="read-bd"]/@id`) {
	case `J_BookRead`:
		cp.isVIP = false
	case `J_ImgRead`:
		cp.isVIP = true
	default:
		return errStatus(cp.url, vipChapterException)
	}
	return nil
}

// 用关键词搜索书号
func (key keyword) findCWMBookID() (string, error) {
	doc, err := htmlquery.LoadURL(
		fmt.Sprint(cwmHost, `/get-search-book-list/0-0-0-0-0-0/全部/`, key, `/1`),
	)
	if err != nil {
		return ``, err
	}
	href := htmlquery.FindOne(doc, `//a[@class="cover"]/@href`)
	if href == nil {
		return ``, notFound(key)
	}
	return strings.TrimPrefix(htmlquery.InnerText(href), cwmURL), nil
}

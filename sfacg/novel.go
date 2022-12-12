package sfacg

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"

	"github.com/PuerkitoBio/goquery"

	log "github.com/sirupsen/logrus"
)

// 小说网页信息获取
func (nv *Novel) init(bookID string) {
	nv.IsGet = false // 初始化
	nv.ID = bookID
	nv.URL = "https://book.sfacg.com/Novel/" + bookID // 生成链接
	nv.NewChapter.BookURL = nv.URL                    // 用于向章节传入本书链接
	req, err := http.Get(nv.URL)
	if !kitten.Check(err) {
		log.Warn(fmt.Sprintf("书号 %s 获取网页失败了喵！", bookID))
	} else {
		defer req.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(req.Body) // 获取小说网页
		if strings.EqualFold(doc.Find("title").Text(), "出错了") ||
			strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") ||
			43 > len(doc.Find("title").Text()) { // 防止网页炸了导致问题
			log.Info(fmt.Sprintf("书号 %s 没有喵！", bookID))
		} else {
			nv.IsGet = true
			nv.Name = doc.Find("h1.title").Find("span.text").Text()             // 获取书名
			nv.Writer = doc.Find("div.author-name").Find("span").Text()         // 获取作者
			nv.HeadURL, _ = doc.Find("div.author-mask").Find("img").Attr("src") // 获取头像链接
			textRow := doc.Find("div.text-row").Find("span")                    // 获取详细数字
			if 9 < len(textRow.Eq(0).Text()) {
				nv.Type = textRow.Eq(0).Text()[9:] // 获取类型
			} else {
				log.Error("获取类型错误喵！")
			}
			if 9 < len(textRow.Eq(2).Text()) {
				nv.HitNum = textRow.Eq(2).Text()[9:] // 获取点击
			} else {
				log.Error("获取点击错误喵！")
			}
			nv.WordNum = textRow.Eq(1).Text()
			loc, _ := time.LoadLocation("Local")
			if 9 < len(textRow.Eq(3).Text()) {
				nv.NewChapter.Time, _ = time.ParseInLocation("2006/1/2 15:04:05", textRow.Eq(3).Text()[9:], loc)
			}
			WordNumInfo := nv.WordNum
			if 9 < len(nv.WordNum) {
				nv.WordNum = nv.WordNum[9 : len(nv.WordNum)-14] // 获取字数
			}
			if 11 < len(WordNumInfo) {
				nv.Status = WordNumInfo[len(WordNumInfo)-11:] //获取状态
			}
			nv.Introduce = doc.Find("p.introduce").Text() // 获取简述
			// doc.Find("ul.tag-list > span#text").Each(func(i int, selection *goquery.Selection) {
			// 	nv.TagList[i] = selection.Text()
			// }) // 获取标签(暂时不能用)
			nv.CoverURL, _ = doc.Find("div.figure").Find("img").Eq(0).Attr("src") // 获取封面
			if 7 < len(doc.Find("#BasicOperation").Find("a").Eq(2).Text()) {
				nv.Collection = doc.Find("#BasicOperation").Find("a").Eq(2).Text()[7:] // 获取收藏
			}

			nv.Preview = doc.Find("div.chapter-info").Find("p").Text()
			nv.Preview = strings.Replace(nv.Preview, " ", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\n", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\r", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "　", "", -1) // 获取预览

			nvNewChapterURL, _ := doc.Find("div.chapter-info").Find("h3").Find("a").Attr("href")
			nv.IsVip = strings.Contains(nvNewChapterURL, "vip")

			if nvNewChapterURL == "" {
				nv.NewChapter.IsGet = false // 防止更新章节炸了跳转到网站首页引起程序报错
				log.Warn(fmt.Sprintf("%s 获取更新链接失败了喵！", nv.URL))
			} else {
				nvNewChapterURL = "https://book.sfacg.com" + nvNewChapterURL
				nv.NewChapter.init(nvNewChapterURL) // 获取新章节链接
			}
		}
	}
}

// 新章节信息获取
func (cp *Chapter) init(URL string) {
	cp.IsGet = false // 初始化
	cp.URL = URL     // 生成链接
	if req, err := http.Get(cp.URL); !kitten.Check(err) {
		log.Warn(fmt.Sprintf("%s 获取更新网页失败了喵！", URL))
	} else {
		defer req.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(req.Body) // 获取新章节网页
		if cp.URL != cp.BookURL {
			if strings.EqualFold(doc.Find("title").Text(), "出错了") ||
				strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") {
				log.Info(fmt.Sprintf("章节【%s】没有喵！", cp.URL))
			} else {
				cp.IsGet = true
				desc := doc.Find("div.article-desc").Find("span")
				if 9 < len(desc.Eq(2).Text()) {
					cp.WordNum = kitten.Atoi(desc.Eq(2).Text()[9:]) // 获取新章节字数
				}
				if 15 < len(desc.Eq(1).Text()) {
					loc, _ := time.LoadLocation("Local")
					cp.Time, _ = time.ParseInLocation("2006/1/2 15:04:05", desc.Eq(1).Text()[15:], loc) // 获取更新时间
				}
				cp.Title = doc.Find("h1.article-title").Text() // 获取新章节标题

				cp.LastURL, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(0).Attr("href")
				cp.NextURL, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(1).Attr("href")
				cp.LastURL = "https://book.sfacg.com" + cp.LastURL // 获取上一章链接
				cp.NextURL = "https://book.sfacg.com" + cp.NextURL // 获取下一章链接
			}
		} else {
			log.Warn(fmt.Sprintf("%s更新异常喵！", URL)) // 防止章节炸了导致获取新章节跳转引发panic
		}
	}
}

// 与上次更新比较
func (nv *Novel) makeCompare() (cm Compare) {
	var this, last Chapter
	this = nv.NewChapter
	if this.IsGet {
		last.init(this.LastURL)
		cm.TimeGap = this.Time.Sub(last.Time)
		cm.Times = 1
		for kitten.IsSameDate(last.Time, this.Time) {
			this = last
			last.init(this.LastURL)
			cm.Times++
		}
	} else {
		cm.Times = 0
		cm.TimeGap = 1
	} // 防止无限得不到更新章节循环
	return
}

// 小说信息
func (nv *Novel) information() (str string) {
	//	var tags string //暂时不能用
	//	for _, v := range nv.TagList {
	//		tags += "["
	//		tags += v
	//		tags += "]"
	//	}
	str = strings.Join([]string{"书名：" + nv.Name,
		"书号：" + nv.ID,
		"作者：" + nv.Writer,
		fmt.Sprintf("【%s】", nv.Type),
		"收藏：" + nv.Collection,
		"总字数：" + nv.WordNum + nv.Status,
		"点击：" + nv.HitNum,
		"更新：" + nv.NewChapter.Time.Format("2006年01月02日 15时04分05秒"),
	}, "，") + "\n\n" + nv.Introduce
	if nv.IsGet {
		return
	}
	return fmt.Sprintf("书号%s打不开喵！", nv.ID)
}

// 用关键词搜索书号，如失败，返回值为失败信息
func findBookID(key string) (string, bool) {
	var (
		searchURL = fmt.Sprintf("http://s.sfacg.com/?Key=%s&S=1&SS=0", key)
		req, err  = http.Get(searchURL)
	)
	if !kitten.Check(err) {
		log.Warn("获取书号失败了喵！")
		return "获取书号失败了喵！", false
	}
	defer req.Body.Close()
	var (
		doc, _           = goquery.NewDocumentFromReader(req.Body)
		href, haveResult = doc.Find("#SearchResultList1___ResultList_LinkInfo_0").Attr("href")
	)
	if !haveResult {
		log.Info(key + "搜索无结果喵。")
		return fmt.Sprintf("关键词【%s】找不到小说喵！", key), false
	}
	return href[29:], true
}

// 更新信息
func (nv *Novel) update() (str string) {
	var (
		cm      = nv.makeCompare()
		wordNum = fmt.Sprintf("%d字", nv.NewChapter.WordNum)
		timeGap = cm.TimeGap.String()
	)
	if cm.TimeGap == 0 {
		log.Warning("更新异常喵！")
		return "更新异常喵！"
	}
	timeGap = strings.Replace(timeGap, "h", "小时", 1)
	timeGap = strings.Replace(timeGap, "m", "分钟", 1)
	timeGap = strings.Replace(timeGap, "s", "秒", 1)
	chapterName := nv.NewChapter.Title
	str = strings.Join([]string{fmt.Sprintf("《%s》更新了喵～", nv.Name) + chapterName,
		"更新字数：" + wordNum,
		"间隔时间：" + timeGap,
		fmt.Sprintf("当日第%d更", cm.Times),
	}, "，")
	return
}

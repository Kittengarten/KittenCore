package sfacg

import (
	"fmt"
	"kitten/kitten"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

// 小说网页信息获取
func (nv *Novel) Init(bookId string) {
	nv.Id = bookId
	nv.Url = "https://book.sfacg.com/Novel/" + bookId // 生成链接
	nv.NewChapter.BookUrl = nv.Url                    // 用于向章节传入本书链接
	req, err := http.Get(nv.Url)
	if !kitten.Check(err) {
		log.Warn(fmt.Sprintf("书号%s获取网页失败了喵！", bookId))
		nv.IsGet = false
	} else {
		defer req.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(req.Body) // 获取小说网页
		if strings.EqualFold(doc.Find("title").Text(), "出错了") ||
			strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") ||
			len(doc.Find("title").Text()) < 43 { // 防止网页炸了导致问题
			log.Info(fmt.Sprintf("书号%s没有喵！", bookId))
			nv.IsGet = false
		} else {
			nv.IsGet = true
			nv.Name = doc.Find("h1.title").Find("span.text").Text()             // 获取书名
			nv.Writer = doc.Find("div.author-name").Find("span").Text()         // 获取作者
			nv.HeadUrl, _ = doc.Find("div.author-mask").Find("img").Attr("src") // 获取头像链接

			textRow := doc.Find("div.text-row").Find("span") // 获取详细数字
			nv.Type = textRow.Eq(0).Text()[9:]               // 获取类型
			nv.HitNum = textRow.Eq(2).Text()[9:]             // 获取点击
			nv.WordNum = textRow.Eq(1).Text()
			loc, _ := time.LoadLocation("Local")
			nv.NewChapter.Time, _ =
				time.ParseInLocation("2006/1/2 15:04:05",
					textRow.Eq(3).Text()[9:],
					loc) // 防止章节炸了导致获取不到更新时间

			WordNumInfo := nv.WordNum
			nv.WordNum = nv.WordNum[9 : len(nv.WordNum)-14] // 获取字数
			nv.Status = WordNumInfo[len(WordNumInfo)-11:]   //获取状态

			nv.Introduce = doc.Find("p.introduce").Text() // 获取简述
			// doc.Find("ul.tag-list > span#text").Each(func(i int, selection *goquery.Selection) {
			// 	nv.TagList[i] = selection.Text()
			// }) // 获取标签(暂时不能用)

			nv.CoverUrl, _ = doc.Find("div.figure").Find("img").Eq(0).Attr("src")  // 获取封面
			nv.Collection = doc.Find("#BasicOperation").Find("a").Eq(2).Text()[7:] // 获取收藏

			nv.Preview = doc.Find("div.chapter-info").Find("p").Text()
			nv.Preview = strings.Replace(nv.Preview, " ", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\n", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\r", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "　", "", -1) // 获取预览

			nvNewChapterUrl, _ := doc.Find("div.chapter-info").Find("h3").Find("a").Attr("href")
			nv.IsVip = strings.Contains(nvNewChapterUrl, "vip")

			if nvNewChapterUrl == "" {
				nv.NewChapter.IsGet = false // 防止更新章节炸了跳转到网站首页引起程序报错
				log.Warn(fmt.Sprintf("%s获取更新链接失败了喵！", nv.Url))
			} else {
				nvNewChapterUrl = "https://book.sfacg.com" + nvNewChapterUrl
				nv.NewChapter.Init(nvNewChapterUrl) // 获取新章节链接
			}
		}
	}
}

// 新章节信息获取
func (cp *Chapter) Init(url string) {
	loc, _ := time.LoadLocation("Local")
	cp.Url = url // 生成链接
	req, err := http.Get(cp.Url)
	if !kitten.Check(err) {
		cp.IsGet = false
		log.Warn(fmt.Sprintf("%s获取更新网页失败了喵！", url))
	} else {
		defer req.Body.Close()
		doc, _ := goquery.NewDocumentFromReader(req.Body) // 获取新章节网页
		if cp.Url != cp.BookUrl {
			if strings.EqualFold(doc.Find("title").Text(), "出错了") ||
				strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") {
				log.Info(fmt.Sprintf("章节%s没有喵！", cp.Url))
				cp.IsGet = false // 防止奇怪的用户对不存在的书号进行更新测试，导致程序报错
			} else {
				cp.IsGet = true
				desc := doc.Find("div.article-desc").Find("span")                                   // Debug用
				cp.WordNum = kitten.Atoi(desc.Eq(2).Text()[9:])                                     // 获取新章节字数
				cp.Time, _ = time.ParseInLocation("2006/1/2 15:04:05", desc.Eq(1).Text()[15:], loc) // 获取更新时间
				cp.Title = doc.Find("h1.article-title").Text()                                      // 获取新章节标题

				cp.LastUrl, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(0).Attr("href")
				cp.NextUrl, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(1).Attr("href")
				cp.LastUrl = "https://book.sfacg.com" + cp.LastUrl // 获取上一章链接
				cp.NextUrl = "https://book.sfacg.com" + cp.NextUrl // 获取下一章链接
			}
		} else {
			cp.IsGet = false
			log.Warn(fmt.Sprintf("%s更新异常喵！", url))
		} // 防止章节炸了导致获取新章节跳转引发panic
	}
}

// 新章节更新时间获取
func FindChapterUpdateTime(bookid string) string {
	var nv Novel
	nv.Init(bookid)
	return nv.NewChapter.Time.Format("2006年01月02日 15时04分05秒")
}

// 与上次更新比较
func (nv *Novel) makeCompare() Compare {
	var cm Compare
	var this, last Chapter
	this = nv.NewChapter
	if this.IsGet {
		last.Init(this.LastUrl)
		cm.TimeGap = this.Time.Sub(last.Time)
		cm.Times = 1
		for kitten.IsSameDate(last.Time, this.Time) {
			this = last
			last.Init(this.LastUrl)
			cm.Times++
		}
	} else {
		cm.Times = 0
		cm.TimeGap = 1
	} // 防止无限得不到更新章节循环
	return cm
}

// 小说信息
func (nv *Novel) Information() string {
	//	var tags string //暂时不能用
	//	for _, v := range nv.TagList {
	//		tags += "["
	//		tags += v
	//		tags += "]"
	//	}
	str := strings.Join([]string{"书名：" + nv.Name,
		"书号：" + nv.Id,
		"作者：" + nv.Writer,
		fmt.Sprintf("【%s】", nv.Type),
		"收藏：" + nv.Collection,
		"总字数：" + nv.WordNum + nv.Status,
		"点击：" + nv.HitNum,
		"更新：" + nv.NewChapter.Time.Format("2006年01月02日 15时04分05秒"),
	}, "，") + "\n\n" + nv.Introduce
	if nv.IsGet {
		return str
	} else {
		return fmt.Sprintf("书号%s打不开喵！", nv.Id)
	}
}

// 搜索
func FindBookID(keyword string) (string, bool) {
	searchUrl := "http://s.sfacg.com/?Key=" + keyword + "&S=1&SS=0"
	req, err := http.Get(searchUrl)
	if !kitten.Check(err) {
		log.Warn("获取书号失败了喵！")
		return "获取书号失败了喵！", false
	}
	defer req.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(req.Body)

	href, haveResult := doc.Find("#SearchResultList1___ResultList_LinkInfo_0").Attr("href")
	if !haveResult {
		log.Info(keyword + "搜索无结果喵。")
		return fmt.Sprintf("关键词【%s】找不到小说喵！", keyword), false
	}
	return href[29:], true
}

// 更新信息
func (nv *Novel) Update() string {
	var cm = nv.makeCompare()

	wordNum := fmt.Sprintf("%d字", nv.NewChapter.WordNum)

	timeGap := cm.TimeGap.String()
	if cm.TimeGap == 0 {
		log.Warning("更新异常喵！")
		return "更新异常喵！"
	} else {
		timeGap = strings.Replace(timeGap, "h", "小时", 1)
		timeGap = strings.Replace(timeGap, "m", "分钟", 1)
		timeGap = strings.Replace(timeGap, "s", "秒", 1)

		chapterName := nv.NewChapter.Title

		str := strings.Join([]string{fmt.Sprintf("《%s》更新了喵～", nv.Name) + chapterName,
			"更新字数：" + wordNum,
			"间隔时间：" + timeGap,
			fmt.Sprintf("当日第%d更", cm.Times),
		}, "，")
		return str
	}
}

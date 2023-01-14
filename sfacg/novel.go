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

const imagePath kitten.Path = "image/path.txt" // 保存图片路径的文件

// 小说网页信息获取
func (nv *Novel) init(bookID string) {
	// 初始化
	nv.IsGet = false
	// 向小说传入书号
	nv.ID = bookID
	// 生成链接
	nv.URL = "https://book.sfacg.com/Novel/" + bookID
	// 向章节传入本书链接
	nv.NewChapter.BookURL = nv.URL
	req, err := http.Get(nv.URL)
	if kitten.Check(err) {
		defer req.Body.Close()
		// 获取小说网页
		if doc, err := goquery.NewDocumentFromReader(req.Body); strings.EqualFold(doc.Find("title").Text(), "出错了") ||
			strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") ||
			43 > len(doc.Find("title").Text()) ||
			kitten.Check(err) == false {
			// 防止网页炸了导致问题
			log.Infof("书号 %s 没有喵！", bookID)
		} else {
			nv.IsGet = true

			// 获取书名
			nv.Name = doc.Find("h1.title").Find("span.text").Text()

			// 获取作者
			nv.Writer = doc.Find("div.author-name").Find("span").Text()

			// 头像链接是否存在
			var he bool
			// 获取头像链接
			nv.HeadURL, he = doc.Find("div.author-mask").Find("img").Attr("src")
			// 头像链接获取失败时使用报错图片
			if !he {
				nv.HeadURL = string(imagePath.LoadPath()) + "no.png"
			}

			// 获取详细数字
			textRow := doc.Find("div.text-row").Find("span")
			// 获取类型
			if 9 < len(textRow.Eq(0).Text()) {
				nv.Type = textRow.Eq(0).Text()[9:]
			} else {
				log.Error("获取类型错误喵！")
			}
			// 获取点击
			if 9 < len(textRow.Eq(2).Text()) {
				nv.HitNum = textRow.Eq(2).Text()[9:]
			} else {
				log.Error("获取点击错误喵！")
			}

			// 获取更新时间
			loc, _ := time.LoadLocation("Local")
			if 9 < len(textRow.Eq(3).Text()) {
				var errT error
				nv.NewChapter.Time, errT = time.ParseInLocation("2006/1/2 15:04:05", textRow.Eq(3).Text()[9:], loc)
				if !kitten.Check(errT) {
					log.Errorf("时间转换出错喵！\n%v", errT)
				}
			}

			// 获取小说字数信息
			nv.WordNum = textRow.Eq(1).Text()
			WordNumInfo := nv.WordNum
			// 获取字数
			if 9 < len(nv.WordNum) {
				nv.WordNum = nv.WordNum[9 : len(nv.WordNum)-14]
			}
			//获取状态
			if 11 < len(WordNumInfo) {
				nv.Status = WordNumInfo[len(WordNumInfo)-11:]
			}

			// 获取简述
			nv.Introduce = doc.Find("p.introduce").Text()

			// 获取标签(暂时不能用)
			// doc.Find("ul.tag-list > span#text").Each(func(i int, selection *goquery.Selection) {
			// 	nv.TagList[i] = selection.Text()
			// })

			// 封面链接是否存在
			var ce bool
			// 获取封面
			nv.CoverURL, ce = doc.Find("div.figure").Find("img").Eq(0).Attr("src")
			// 封面链接获取失败时使用报错图片
			if !ce {
				nv.CoverURL = string(imagePath.LoadPath()) + "no.png"
			}

			// 获取收藏
			if 7 < len(doc.Find("#BasicOperation").Find("a").Eq(2).Text()) {
				nv.Collection = doc.Find("#BasicOperation").Find("a").Eq(2).Text()[7:]
			}

			// 获取预览
			nv.Preview = doc.Find("div.chapter-info").Find("p").Text()
			nv.Preview = strings.Replace(nv.Preview, " ", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\n", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "\r", "", -1)
			nv.Preview = strings.Replace(nv.Preview, "　", "", -1)

			// 获取新章节链接
			nvNewChapterURL, eC := doc.Find("div.chapter-info").Find("h3").Find("a").Attr("href")
			nv.IsVip = strings.Contains(nvNewChapterURL, "vip")
			// 如果新章节链接存在
			if eC {
				// 构造新章节链接
				nvNewChapterURL = "https://book.sfacg.com" + nvNewChapterURL
				nv.NewChapter.init(nvNewChapterURL)
			} else {
				// 防止更新章节炸了跳转到网站首页引起程序报错
				nv.NewChapter.IsGet = false
				log.Warnf("%s 获取更新链接失败了喵！\n", nv.URL)
			}
		}
	} else {
		log.Warnf("书号 %s 获取网页失败了喵！", bookID)
	}
}

// 新章节信息获取
func (cp *Chapter) init(URL string) {
	// 初始化
	cp.IsGet = false
	// 向章节传入链接
	cp.URL = URL
	if req, err := http.Get(cp.URL); kitten.Check(err) {
		defer req.Body.Close()
		// 获取新章节网页
		doc, err := goquery.NewDocumentFromReader(req.Body)
		// 防止章节炸了导致获取新章节跳转引发 panic
		if cp.URL != cp.BookURL {
			if strings.EqualFold(doc.Find("title").Text(), "出错了") ||
				strings.EqualFold(doc.Find("title").Text(), "糟糕,页面找不到了") ||
				!kitten.Check(err) {
				// 防止网页炸了导致问题
				log.Infof("章节 %s 没有喵！\n%v", cp.URL, err)
			} else {
				cp.IsGet = true
				desc := doc.Find("div.article-desc").Find("span")
				// 获取新章节字数
				if 9 < len(desc.Eq(2).Text()) {
					cp.WordNum = kitten.IntString(desc.Eq(2).Text()[9:]).Int()
				}
				// 获取更新时间
				if 15 < len(desc.Eq(1).Text()) {
					loc, _ := time.LoadLocation("Local")
					var errT error
					cp.Time, errT = time.ParseInLocation("2006/1/2 15:04:05", desc.Eq(1).Text()[15:], loc)
					if !kitten.Check(errT) {
						log.Errorf("时间转换出错喵！\n%v", errT)
					}
				}
				// 获取新章节标题
				cp.Title = doc.Find("h1.article-title").Text()
				// 获取上一章链接
				cp.LastURL, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(0).Attr("href")
				cp.LastURL = "https://book.sfacg.com" + cp.LastURL
				// 获取下一章链接
				cp.NextURL, _ = doc.Find("div.fn-btn").Eq(-1).Find("a").Eq(1).Attr("href")
				cp.NextURL = "https://book.sfacg.com" + cp.NextURL
			}
		} else {
			log.Warnf("%s 更新异常喵！", URL)
		}
	} else {
		log.Warnf("%s 获取更新网页失败了喵！", URL)
	}
}

// 与上次更新比较
func (nv *Novel) makeCompare() (cm Compare) {
	var this, last Chapter
	if this = nv.NewChapter; this.IsGet {
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
	//	var tags string // 暂时不能用
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
func (key keyWord) findBookID() (string, bool) {
	var (
		searchURL = fmt.Sprintf("http://s.sfacg.com/?Key=%s&S=1&SS=0", key)
		req, err  = http.Get(searchURL)
	)
	if !kitten.Check(err) {
		log.Warnf("获取书号失败了喵！\n错误：%v", err)
		return "获取书号失败了喵！", false
	}
	defer req.Body.Close()
	var (
		doc, errR        = goquery.NewDocumentFromReader(req.Body)
		href, haveResult = doc.Find("#SearchResultList1___ResultList_LinkInfo_0").Attr("href")
	)
	if !kitten.Check(errR) {
		log.Errorf("转换出错喵！\n%v", errR)
	} else if !haveResult {
		log.Info(key + "搜索无结果喵。")
		return fmt.Sprintf("关键词【%s】找不到小说喵！", key), false
	}
	return href[29:], true
}

// 更新信息
func (nv *Novel) update() (str string) {
	var (
		cm      = nv.makeCompare()
		wordNum = fmt.Sprintf("%d 字", nv.NewChapter.WordNum)
		timeGap = cm.TimeGap.String()
	)
	if cm.TimeGap == 0 {
		log.Warn("更新异常喵！")
		return "更新异常喵！"
	}
	timeGap = strings.Replace(timeGap, "h", " 小时 ", 1)
	timeGap = strings.Replace(timeGap, "m", " 分钟 ", 1)
	timeGap = strings.Replace(timeGap, "s", " 秒", 1)
	chapterName := nv.NewChapter.Title
	str = strings.Join([]string{fmt.Sprintf("《%s》更新了喵～", nv.Name) + chapterName,
		"更新字数：" + wordNum,
		"间隔时间：" + timeGap,
		fmt.Sprintf("当日第 %d 更", cm.Times),
	}, "，")
	return
}

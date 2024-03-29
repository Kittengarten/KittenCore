package sfacg

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
)

var (
	// 图片路径
	imagePath kitten.Path = kitten.Path(kitten.Configs.Path + `image/`)
	// 小说池
	novelPool = sync.Pool{
		New: func() interface{} {
			return &Novel{}
		},
	}
	// 章节池
	chapterPool = sync.Pool{
		New: func() interface{} {
			return &Chapter{}
		},
	}
	// 配置池
	configPool sync.Pool
)

// 小说网页信息获取
func (nv *Novel) init(bookID string) {
	var (
		req     *http.Response     // HTTP 响应
		doc     *goquery.Document  // 网页
		err     error              // 错误
		he      bool               // 头像链接是否存在
		ce      bool               // 封面链接是否存在
		textRow *goquery.Selection // 小说详细信息

	)
	// 初始化
	nv.IsGet = false
	// 向小说传入书号
	nv.ID = bookID
	// 生成链接
	nv.URL = `https://book.sfacg.com/Novel/` + bookID
	// 从章节池初始化章节，向章节传入本书链接
	nv.NewChapter = *chapterPool.Get().(*Chapter)
	nv.NewChapter.BookURL = nv.URL
	defer chapterPool.Put(&nv.NewChapter)
	// 获取 HTTP 响应
	req, err = http.Get(nv.URL)
	if kitten.Check(err) {
		defer req.Body.Close()
		// 获取小说网页，如果网页炸了则返回
		if doc, err = goquery.NewDocumentFromReader(req.Body); strings.EqualFold(doc.Find(`title`).Text(), `出错了`) ||
			strings.EqualFold(doc.Find(`title`).Text(), `糟糕,页面找不到了`) ||
			43 > len(doc.Find(`title`).Text()) || !kitten.Check(err) {
			zap.S().Errorf(`书号 %s 没有喵！`, bookID)
			return
		}
		// 网页没炸，小说获取成功
		nv.IsGet = true
		// 获取书名
		nv.Name = doc.Find(`h1.title`).Find(`span.text`).Text()
		// 获取作者
		nv.Writer = doc.Find(`div.author-name`).Find(`span`).Text()
		// 获取头像链接，失败时使用报错图片
		if nv.HeadURL, he = doc.Find(`div.author-mask`).Find(`img`).Attr(`src`); !he {
			nv.HeadURL = imagePath.LoadPath().String() + `no.png`
			zap.S().Error(`头像链接获取失败喵！`)
		}
		// 获取小说详细信息
		textRow = doc.Find(`div.text-row`).Find(`span`)
		// 获取类型
		if 9 < len(textRow.Eq(0).Text()) {
			nv.Type = textRow.Eq(0).Text()[9:]
		} else {
			zap.S().Error(`获取类型错误喵！`)
		}
		// 获取点击
		if 9 < len(textRow.Eq(2).Text()) {
			nv.HitNum = textRow.Eq(2).Text()[9:]
		} else {
			zap.S().Error(`获取点击错误喵！`)
		}
		// 获取更新时间
		loc, err := time.LoadLocation(`Local`)
		if !kitten.Check(err) {
			zap.S().Errorf("时区获取出错喵！\n%v", err)
		}
		if 9 < len(textRow.Eq(3).Text()) {
			nv.NewChapter.Time, err = time.ParseInLocation(`2006/1/2 15:04:05`, textRow.Eq(3).Text()[9:], loc)
			if !kitten.Check(err) {
				zap.S().Errorf("时间转换出错喵！\n%v", err)
			}
		}
		// 获取小说字数信息
		nv.WordNum = textRow.Eq(1).Text()
		WordNumInfo := nv.WordNum
		// 获取字数
		if 9 < len(nv.WordNum) {
			nv.WordNum = nv.WordNum[9 : len(nv.WordNum)-14]
		}
		// 获取状态
		if 11 < len(WordNumInfo) {
			nv.Status = WordNumInfo[len(WordNumInfo)-11:]
		}
		// 获取简述
		nv.Introduce = doc.Find(`p.introduce`).Text()
		// // 获取标签
		// doc.Find(`ul.tag-list`).Find(`a`).Find(`span.text`).Each(func(i int, selection *goquery.Selection) {
		// 	nv.TagList[i] = selection.Text()
		// })
		// 获取封面，失败时使用报错图片
		if nv.CoverURL, ce = doc.Find(`div.figure`).Find(`img`).Eq(0).Attr(`src`); !ce {
			nv.CoverURL = imagePath.LoadPath().String() + `no.png`
			zap.S().Error("封面链接获取失败喵！")
		}
		// 获取收藏
		if 7 < len(doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()) {
			nv.Collection = doc.Find(`#BasicOperation`).Find(`a`).Eq(2).Text()[7:]
		}
		// 获取预览
		nv.Preview = doc.Find(`div.chapter-info`).Find(`p`).Text()
		nv.Preview = strings.Replace(nv.Preview, ` `, ``, -1)
		nv.Preview = strings.Replace(nv.Preview, "\n", ``, -1)
		nv.Preview = strings.Replace(nv.Preview, "\r", ``, -1)
		nv.Preview = strings.Replace(nv.Preview, `　`, ``, -1)
		// 获取新章节链接
		nvNewChapterURL, eC := doc.Find(`div.chapter-info`).Find(`h3`).Find(`a`).Attr(`href`)
		nv.IsVip = strings.Contains(nvNewChapterURL, `vip`)
		// 如果新章节链接存在
		if eC {
			// 构造新章节链接
			nvNewChapterURL = `https://book.sfacg.com` + nvNewChapterURL
			nv.NewChapter.init(nvNewChapterURL)
		} else {
			// 防止更新章节炸了跳转到网站首页引起程序报错
			nv.NewChapter.IsGet = false
			zap.S().Warnf("%s 获取更新链接失败了喵！\n", nv.URL)
		}
		return
	}
	zap.S().Warnf(`书号 %s 获取网页失败了喵！`, bookID)
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
			if strings.EqualFold(doc.Find(`title`).Text(), `出错了`) ||
				strings.EqualFold(doc.Find(`title`).Text(), `糟糕,页面找不到了`) ||
				!kitten.Check(err) {
				// 防止网页炸了导致问题
				zap.S().Infof("章节 %s 没有喵！\n%v", cp.URL, err)
			} else {
				cp.IsGet = true
				desc := doc.Find(`div.article-desc`).Find(`span`)
				// 获取新章节字数
				if 9 < len(desc.Eq(2).Text()) {
					cp.WordNum, err = strconv.Atoi(desc.Eq(2).Text()[9:])
				}
				if !kitten.Check(err) {
					zap.S().Warnf("转换新章节字数失败了喵！\n%v", err)
				}
				// 获取更新时间
				if 15 < len(desc.Eq(1).Text()) {
					loc, err := time.LoadLocation(`Local`)
					if !kitten.Check(err) {
						zap.S().Errorf("时区获取出错喵！\n%v", err)
					}
					var errT error
					cp.Time, errT = time.ParseInLocation(`2006/1/2 15:04:05`, desc.Eq(1).Text()[15:], loc)
					if !kitten.Check(errT) {
						zap.S().Errorf("时间转换出错喵！\n%v", errT)
					}
				}
				// 获取新章节标题
				cp.Title = doc.Find(`h1.article-title`).Text()
				// 获取上一章链接
				var ok bool
				cp.LastURL, ok = doc.Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(0).Attr(`href`)
				if !ok {
					zap.S().Warnf("%s上一章链接获取失败喵！", URL)
				}
				cp.LastURL = `https://book.sfacg.com` + cp.LastURL
				// 获取下一章链接
				cp.NextURL, ok = doc.Find(`div.fn-btn`).Eq(-1).Find(`a`).Eq(1).Attr(`href`)
				if !ok {
					zap.S().Warnf("%s上一章链接获取失败喵！", URL)
				}
				cp.NextURL = `https://book.sfacg.com` + cp.NextURL
			}
		} else {
			zap.S().Warnf(`%s 更新异常喵！`, URL)
		}
	} else {
		zap.S().Warnf(`%s 获取更新网页失败了喵！`, URL)
	}
}

// 与上次更新比较
func (nv *Novel) makeCompare() (cm Compare) {
	var this, last Chapter
	this = nv.NewChapter
	if !this.IsGet {
		// 防止无限得不到更新章节循环
		return Compare{TimeGap: 1}
	}
	last.init(this.LastURL)
	cm.TimeGap = max(time.Second, this.Time.Sub(last.Time))
	for cm.Times = 1; kitten.IsSameDate(last.Time, this.Time); cm.Times++ {
		this = last
		last.init(this.LastURL)
	}
	return
}

// 小说信息
func (nv *Novel) information() (str string) {
	// var tags string // 标签
	// for k := range nv.TagList {
	// 	tags += fmt.Sprintf(`[%s]`, nv.TagList[k])
	// }
	str = strings.Join([]string{`书名：` + nv.Name,
		`书号：` + nv.ID,
		`作者：` + nv.Writer,
		fmt.Sprintf(`【%s】`, nv.Type),
		`收藏：` + nv.Collection,
		`总字数：` + nv.WordNum + nv.Status,
		`点击：` + nv.HitNum,
		`更新：` + nv.NewChapter.Time.Format(`2006年01月02日 15时04分05秒`),
	}, "\n") + "\n\n" + nv.Introduce
	if nv.IsGet {
		return
	}
	if `` == nv.ID {
		return `获取不到书号喵！`
	}
	return fmt.Sprintf(`书号 %s 打不开喵！`, nv.ID)
}

// 用关键词搜索书号，如失败，返回原关键词和失败信息
func (key keyWord) findBookID() (string, error) {
	var (
		searchURL = fmt.Sprintf(`http://s.sfacg.com/?Key=%s&S=1&SS=0`, key)
		req, err  = http.Get(searchURL)
	)
	if !kitten.Check(err) {
		zap.S().Warnf("获取书号失败了喵！\n错误：%v", err)
		return string(key), err
	}
	defer req.Body.Close()
	var (
		doc, errR        = goquery.NewDocumentFromReader(req.Body)
		href, haveResult = doc.Find(`#SearchResultList1___ResultList_LinkInfo_0`).Attr(`href`)
	)
	if kitten.Check(errR) {
		if haveResult {
			return href[29:], nil
		}
		e := fmt.Sprintf(`关键词【%s】找不到小说喵！`, key)
		zap.S().Info(e)
		return string(key), errors.New(e)
	}
	e := fmt.Sprintf("网页转换出错喵！\n%v", errR)
	zap.S().Error(e)
	return string(key), errors.New(e)
}

// 更新信息
func (nv *Novel) update() (str string, d time.Duration) {
	var (
		cm      = nv.makeCompare()
		wordNum = fmt.Sprintf(`%d 字`, nv.NewChapter.WordNum)
		timeGap string
	)
	d = cm.TimeGap
	if d < 144*time.Hour {
		timeGap = d.String()
	} else {
		timeGap = `不明`
	}
	timeGap = strings.Replace(timeGap, `h`, ` 小时 `, 1)
	timeGap = strings.Replace(timeGap, `m`, ` 分钟 `, 1)
	timeGap = strings.Replace(timeGap, `s`, ` 秒`, 1)
	chapterName := nv.NewChapter.Title
	str = strings.Join([]string{fmt.Sprintf(`《%s》更新了喵～`, nv.Name),
		chapterName,
		`更新字数：` + wordNum,
		`间隔时间：` + timeGap,
		fmt.Sprintf(`当日第 %d 更`, cm.Times),
	}, "\n")
	return
}

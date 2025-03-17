package track

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/equal"
	"github.com/Kittengarten/KittenCore/kitten/core/times"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// 小说池
	novelPool = sync.Pool{
		New: func() any {
			return new(novel)
		},
	}
	// 章节池
	chapterPool = sync.Pool{
		New: func() any {
			return new(chapter)
		},
	}
	// 小说初始化器
	novelInitializer = map[Platform]func(*novel, string) error{
		CWM: func(nv *novel, bookID string) error {
			return nv.initCWM(bookID)
		},
		FQ: func(nv *novel, bookID string) error {
			return nv.initFQ(bookID)
		},
		FQAPI: func(nv *novel, bookID string) error {
			return nv.initFQAPI(bookID)
		},
		SF: func(nv *novel, bookID string) error {
			return nv.initSF(bookID)
		},
	}
	// 章节初始化器
	chapterInitializer = map[Platform]func(*chapter, string) error{
		CWM: func(cp *chapter, url string) error {
			return cp.initCWM(url)
		},
		FQ: func(cp *chapter, url string) error {
			return cp.initFQ(url)
		},
		FQAPI: func(cp *chapter, url string) error {
			return cp.initFQAPI(url)
		},
		SF: func(cp *chapter, url string) error {
			return cp.initSF(url)
		},
	}
	// 关键词搜索器
	searcher = map[Platform]func(keyword) (string, error){
		SF: func(key keyword) (string, error) {
			return key.findSFBookID()
		},
		CWM: func(key keyword) (string, error) {
			return key.findCWMBookID()
		},
		FQAPI: func(key keyword) (string, error) {
			return key.findFQAPIBookID()
		},
	}
)

// 小说网页信息获取
func (nv *novel) init(p Platform, bookID string) error {
	i, ok := novelInitializer[p]
	if !ok {
		return notSupported(p)
	}
	return i(nv, bookID)
}

// 章节信息获取
func (cp *chapter) init(p Platform, url string) error {
	i, ok := chapterInitializer[p]
	if !ok {
		return notSupported(p)
	}
	return i(cp, url)
}

// 获取小说平台
func (nv *novel) GetPlatform() string {
	if nv.Platform == FQAPI {
		nv.Platform = FQ
	}
	return `平台：` + string(nv.Platform)
}

// 获取小说书名
func (nv *novel) Name() string {
	return `书名：` + nv.name
}

// 获取小说书号
func (nv *novel) ID() string {
	return `书号：` + nv.id
}

// 获取小说作者
func (nv *novel) Writer() string {
	return `作者：` + nv.writer
}

// 获取小说类型（主题、项目）
func (nv *novel) Theme() string {
	return func(t string) string {
		if t == `` {
			return ``
		}
		return `【` + t + `】`
	}(nv.theme) + func(r []string) string {
		if len(r) == 0 {
			return ``
		}
		return `【` + strings.Join(r, `】【`) + `】`
	}(nv.right) + func(i []string) string {
		switch nv.Platform {
		case SF:
			if len(i) == 0 {
				return ``
			}
			return `【` + strings.Join(i, `】【`) + `】`
		case CWM:
			return strings.Join(i, ``)
		default:
			return ``
		}
	}(nv.item)
}

// 获取小说收藏
func (nv *novel) Collection() string {
	return `收藏：` + nv.collection
}

// 获取小说字数（状态）
func (nv *novel) WordNum() string {
	return `字数：` + nv.wordNum + func(i string) string {
		switch nv.Platform {
		case SF, CWM, FQAPI:
			if len(i) == 0 {
				return ``
			}
			return `（` + i + `）`
		default:
			return ``
		}
	}(nv.status)
}

// 获取小说点击
func (nv *novel) HitNum() string {
	return `点击：` + nv.hitNum
}

// 获取小说更新、简介
func (nv *novel) UpdateAndIntroduce() string {
	return `更新：` + nv.newChapter.Format(times.Layout) + func() string {
		switch nv.Platform {
		case CWM:
			return ``
		default:
			return "\n\n"
		}
	}() + nv.introduce
}

// String 实现 fmt.Stringer
func (nv *novel) String() string {
	if nv.id == `` {
		return `获取不到书号喵！`
	}
	var tags strings.Builder // 标签
	tags.Grow(8 * len(nv.tagList))
	for _, t := range nv.tagList {
		fmt.Fprint(&tags, `[`, t, `]`)
	}
	return strings.Join([]string{
		nv.GetPlatform(),
		nv.Name(),
		nv.ID(),
		nv.Writer(),
		nv.url,
		nv.Theme(),
		tags.String(),
		nv.Collection(),
		nv.WordNum(),
		nv.HitNum(),
		nv.UpdateAndIntroduce(),
	}, "\n")
}

// String 实现 fmt.Stringer
func (cp *chapter) String() string {
	return cp.title + "\n" + cp.url
}

// String 实现 fmt.Stringer
func (b book) String() string {
	return `《` + b.BookName + `》` + `
作者：　　	` + b.Writer + `
平台：　　	` + string(b.Platform) + `
书号：　　	` + b.BookID + `
上次更新：	` + func() string {
		if b.UpdateTime.IsZero() {
			return unknown
		}
		return b.UpdateTime.Format(times.Layout)
	}()
}

// 用关键词搜索书号
func (key keyword) findBookID(p Platform) (string, error) {
	s, ok := searcher[p]
	if !ok {
		return ``, notSupported(p)
	}
	return s(key)
}

// 与上次更新比较
func (nv *novel) makeCompare() error {
	var this, pre chapter
	this = nv.newChapter
	if this.preURL == `` || this.preURL == nv.url {
		return errStatus(nv.url, onlyAChapter)
	}
	if err := pre.init(nv.Platform, this.preURL); err != nil {
		return err
	}
	nv.todayWordNum = this.wordNum
	nv.Duration = max(time.Second, this.Sub(pre.Time))
	for nv.times = 1; equal.IsSameDate(pre.Time, this.Time) &&
		pre.preURL != nv.url; nv.times++ {
		times.RandomDelayRange(time.Second, 2*time.Second)
		this = pre
		nv.todayWordNum += this.wordNum
		if err := pre.init(nv.Platform, this.preURL); err != nil {
			return err
		}
	}
	return nil
}

// CommentUpdate 评论更新
var CommentUpdate = func(p Platform, nv fmt.Stringer, bookID string, chapterID string) string {
	// 默认为空实现
	return ``
}

// 尝试评论更新
func tryCommentUpdate(
	msgr *kitten.Messager,
	msgID []message.ID,
	users []kitten.QQ,
	nv novel,
	chapterID string,
) {
	const tryCount = 5 // 重试最多 5 次
	for range tryCount {
		s := CommentUpdate(nv.Platform, &nv, nv.id, chapterID)
		if s == `` {
			continue
		}
		for i, user := range users {
			msgr.Reply(msgID[i]).Text(s).Send(user)
		}
		return
	}
	kitten.Error(`评论更新失败喵！`)
}

// 获取章节 ID
func getChapterID(p Platform, chapterURL string) string {
	u, err := url.Parse(chapterURL)
	if err != nil {
		return ``
	}
	switch p {
	case SF:
		return u.Path
	case FQ, CWM:
		return path.Base(u.Path)
	case FQAPI:
		return u.Query().Get(fqAPIItemId)
	default:
		// 不支持的平台
		return ``
	}
}

// 更新信息
func (nv *novel) update() string {
	if nv.Platform == FQAPI {
		defer func() {
			// 番茄 API 还原为番茄平台
			nv.Platform = FQ
		}()
	}
	return fmt.Sprintf(`《%s》更新了喵～
%s%s
更新字数：%d 字（%s）%s`,
		nv.name,
		nv.newChapter.title,
		func(p Platform) string {
			if p == FQAPI {
				// 番茄 API 链接还原为章节链接
				// 不在这里还原平台，以免影响 nv.todayReport() 计算
				newURL, err := url.Parse(nv.newChapter.url)
				if err != nil {
					kitten.Error(err)
					return ``
				}
				nv.newChapter.url = fqChapterURL + newURL.Query().Get(fqAPIItemId)
			}
			return "\n" + nv.newChapter.url
		}(nv.Platform),
		nv.newChapter.wordNum, func(v bool) string {
			if v {
				return `付费`
			}
			return `免费`
		}(nv.newChapter.isVIP),
		func() string {
			tr, err := nv.todayReport()
			if err != nil {
				kitten.Error(err)
				return ``
			}
			return "\n间隔时间：" + tr
		}(),
	)
}

// 今日报更
func (nv *novel) todayReport() (string, error) {
	if err := nv.makeCompare(); err != nil {
		return ``, err
	}
	s, err := nv.DurationConvert()
	if err != nil {
		return ``, err
	}
	return s.String() + "\n" + nv.todayUpdate(), nil
}

// 距上次更新时间的时间差转换为时间间隔的结构体
func (nv *novel) DurationConvert() (times.TimeDuration, error) {
	// 如果时间早于 2006.1.2 15:04:05
	if s, _ := Platform(``).ParseTime(times.Layout); nv.Duration > time.Since(s) {
		return times.TimeDuration{}, errStatus(nv.url, timeException)
	}
	return times.ConvertTimeDuration(nv.Duration), nil
}

// 今日更新信息
func (nv *novel) todayUpdate() string {
	switch nv.times {
	case 0:
		return ``
	case 1:
		return `当日第 1 更`
	default:
		return fmt.Sprint(`当日第`, nv.times, `更，日更`, nv.todayWordNum, `字`)
	}
}

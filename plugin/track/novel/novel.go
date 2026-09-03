package novel

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Kittengarten/KittenCore/kitten/core/log"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/kitten/core/utils"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Export 导出点评接口
var Export Commenter // Commenter 小说点评者

// 小说池
var pool = sync.Pool{
	New: func() any {
		return new(Novel)
	},
}

// 获取小说平台
func (nv *Novel) platform() string {
	return `平台：` + nv.Platform
}

// 获取小说书名
func (nv *Novel) name() string {
	return `书名：` + nv.Name
}

// BookID 获取小说书号
func (nv *Novel) BookID() string {
	return nv.ID
}

// 获取小说书号
func (nv *Novel) id() string {
	return `书号：` + nv.ID
}

// 获取小说作者
func (nv *Novel) writer() string {
	return `作者：` + nv.Writer
}

// 获取小说主角
func (nv *Novel) protagonists() string {
	if len(nv.Protagonists) == 0 {
		return ``
	}
	return `主角：` + strings.Join(nv.Protagonists, `、`)
}

// 获取小说主题
func (nv *Novel) theme() string {
	if nv.Theme == `` {
		return ``
	}
	return `【` + nv.Theme + `】`
}

// 获取小说版权
func (nv *Novel) right() string {
	if len(nv.Right) == 0 {
		return ``
	}
	return `【` + strings.Join(nv.Right, `】【`) + `】`
}

// 获取小说类型（主题、项目）
func (nv *Novel) themes() string {
	return nv.theme() + nv.right() + nv.item()
}

// 获取小说项目
func (nv *Novel) item() string {
	if len(nv.Item) == 0 {
		return ``
	}
	return `【` + strings.Join(nv.Item, `】【`) + `】`
}

// 获取小说标签
func (nv *Novel) tags() string {
	s := new(strings.Builder)
	s.Grow(8 * len(nv.TagList))
	for _, t := range nv.TagList {
		fmt.Fprint(s, `[`, t, `]`)
	}
	return s.String()
}

// 计算小说数据评分
func (nv *Novel) score() string {
	n := max(20_000, float64(nv.TotalWordNum))
	return fmt.Sprintf(`评分：%.2f / 10`,
		10+max(0, min(4, math.Log10(float64(nv.DailyWordNum))))/2-
			max(0, math.Log10(float64(time.Since(nv.Chapter.Update)/times.Year)))+
			2*math.Log10(float64(nv.Collection)*float64(nv.HitNum)/math.Pow(n, 2)))
}

// 获取小说收藏
func (nv *Novel) collection() string {
	return `收藏：` + strconv.FormatUint(nv.Collection, 10)
}

// 获取小说状态
func (nv *Novel) status() string {
	if nv.Status == `` {
		return ``
	}
	return `（` + nv.Status.String() + `）`
}

// 获取小说字数（状态）
func (nv *Novel) wordNum() string {
	return `字数：` + strconv.FormatUint(nv.TotalWordNum, 10) + nv.status()
}

// 获取小说点击
func (nv *Novel) hitNum() string {
	return `点击：` + strconv.FormatUint(nv.HitNum, 10)
}

// 获取小说更新时间
func (nv *Novel) update() string {
	return fmt.Sprintf(`更新：%s%s`,
		nv.Chapter.Update.Format(times.LayoutHeart), nv.UpdateData)
}

// 获取小说简介
func (nv *Novel) introduce() string {
	return "简介：\n" + nv.Introduce
}

// ChapterIDSource 小说更新章号源
type ChapterIDSource interface {
	// ChapterID 获取小说更新章号
	ChapterID(chapURL string) string
}

// GetChapterIDSource 获取小说更新章号源
var GetChapterIDSource func(platform string) (ChapterIDSource, error)

// ChapterID 获取小说更新章号
func (nv *Novel) ChapterID() string {
	p, err := GetChapterIDSource(nv.Platform)
	if err != nil {
		log.Error(err)
		return ``
	}
	return p.ChapterID(nv.Chapter.URL)
}

// Format 实现 fmt.Formatter
func (nv *Novel) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') || state.Flag('#') {
			nv.format(state, verb)
			return
		}
		_, _ = fmt.Fprint(state, nv.String())
	case 's':
		_, _ = fmt.Fprint(state, nv.String())
	default:
		nv.format(state, verb)
	}
}

func (nv *Novel) format(state fmt.State, verb rune) {
	type raw *Novel
	_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), raw(nv))
}

// String 实现 fmt.Stringer
func (nv *Novel) String() string {
	if nv.ID == `` {
		return `获取不到小说喵！`
	}
	return strings.Join(slices.DeleteFunc([]string{
		nv.platform(),
		nv.name(),
		nv.id(),
		nv.writer(),
		nv.protagonists(),
		nv.URL,
		nv.themes(),
		nv.tags(),
		nv.score(),
		nv.collection(),
		nv.wordNum(),
		nv.hitNum(),
		nv.update(),
		nv.introduce(),
	}, func(s string) bool { return s == `` }), "\n")
}

// Source 小说源
type Source interface {
	// Init 初始化小说
	Init(ctx context.Context, bookID string) (*Novel, error)
}

// CheckComplete 检查数据长度，判断是否完整
func CheckComplete(name string, data []*html.Node, n int, f func()) {
	if len(data) >= n {
		f()
		return
	}
	log.Warnf(`小说%s不完整喵！
长度：%d
内容：
%s`,
		name,
		len(data),
		strings.Join(utils.ConvertSlice(data,
			func(n *html.Node) string { return htmlquery.InnerText(n) },
		), "\n"))
}

// Get 获取小说
func Get() *Novel {
	return pool.Get().(*Novel)
}

// Put 回收小说
func (nv *Novel) Put() {
	if nv == nil {
		return
	}
	if nv.Chapter != nil {
		nv.Chapter.Put()
	}
	*nv = Novel{}
	pool.Put(nv)
}

// Format 实现 fmt.Formatter，返回日均更新数据
//
//	%s 完整字符串
//	%c 省略过的字符串
func (d UpdateData) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') || state.Flag('#') {
			d.format(state, verb)
			return
		}
		if d.Status == Completed {
			// 完结书籍不显示该数据
			return
		}
		_, _ = fmt.Fprint(state, d.String())
	case 's': // 精简
		if d.Status == Completed {
			// 完结书籍不显示该数据
			return
		}
		_, _ = fmt.Fprint(state, d.Str())
	case 'c': // 完整
		if d.Status == Completed {
			// 完结书籍不显示该数据
			return
		}
		_, _ = fmt.Fprint(state, d.String())
	default:
		d.format(state, verb)
	}
}

func (d UpdateData) format(state fmt.State, verb rune) {
	type raw UpdateData
	_, _ = fmt.Fprintf(state, fmt.FormatString(state, verb), raw(d))
}

// Str 返回省略过的字符串
func (d UpdateData) Str() string {
	return fmt.Sprintf(`
日更：%d`, d.DailyWordNum)
}

// String 实现 fmt.Stringer
func (d UpdateData) String() string {
	return fmt.Sprintf(`
一周日均：%d 字
一月日均：%d 字
全书日均：%d 字`,
		d.WeekDailyWordNum,
		d.MonthDailyWordNum,
		d.DailyWordNum,
	)
}

// String 实现 fmt.Stringer
func (s Status) String() string {
	return string(s)
}

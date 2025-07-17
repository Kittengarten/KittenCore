package novel

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/Kittengarten/KittenCore/kitten"
	"github.com/Kittengarten/KittenCore/kitten/core/times"
	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/platform"
)

// Export 导出点评接口
var Export struct {
	Commenter // Commenter 小说点评者
}

// Pool 小说池
var Pool = sync.Pool{
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
func (nv *Novel) tags() *strings.Builder {
	var s strings.Builder
	s.Grow(8 * len(nv.TagList))
	for _, t := range nv.TagList {
		fmt.Fprint(&s, `[`, t, `]`)
	}
	return &s
}

// 获取小说收藏
func (nv *Novel) collection() string {
	return `收藏：` + nv.Collection
}

// 获取小说状态
func (nv *Novel) status() string {
	if nv.Status == `` {
		return ``
	}
	return `（` + nv.Status + `）`
}

// 获取小说字数（状态）
func (nv *Novel) wordNum() string {
	return `字数：` + nv.TotalWordNum + nv.status()
}

// 获取小说点击
func (nv *Novel) hitNum() string {
	return `点击：` + nv.HitNum
}

// 获取小说更新时间
func (nv *Novel) update() string {
	return `更新：` + nv.Chapter.Update.Format(times.LayoutHeart)
}

// 获取小说简介
func (nv *Novel) introduce() string {
	return "简介：\n" + nv.Introduce
}

// ChapterID 获取小说更新章号
func (nv *Novel) ChapterID() string {
	p, err := platform.Get(nv.Platform)
	if err != nil {
		kitten.Error(err)
		return ``
	}
	return p.ChapterID(nv.Chapter.URL)
}

// String 实现 fmt.Stringer
func (nv *Novel) String() string {
	if nv.ID == `` {
		return `获取不到书号喵！`
	}
	return strings.Join(slices.DeleteFunc([]string{
		nv.platform(),
		nv.name(),
		nv.id(),
		nv.writer(),
		nv.protagonists(),
		nv.URL,
		nv.themes(),
		nv.tags().String(),
		nv.collection(),
		nv.wordNum(),
		nv.hitNum(),
		nv.update(),
		nv.introduce(),
	}, func(s string) bool { return s == `` }), "\n")
}

// Init 初始化小说
func Init(p platform.Platform, nvID string) (*Novel, error) {
	nva, err := p.Init(nvID)
	if err != nil {
		return nil, err
	}
	return Assert(nva), nil
}

// Assert 断言为小说，不是小说时返回空小说
func Assert(a any) *Novel {
	if nv, ok := a.(*Novel); ok {
		return nv
	}
	return &Novel{Chapter: &chapter.Chapter{}}
}

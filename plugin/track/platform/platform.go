package platform

import (
	"fmt"
	"time"

	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/search"
)

type (
	// Platform 平台
	Platform interface {
		// String 实现 fmt.Stringer，获得平台名称
		String() string
		// Layout 时间格式
		Layout() string
		// FindBookID 搜索书号
		FindBookID(search.Keyword) (string, error)
		// ChapterID 获取章号
		ChapterID(cpURL string) string
		// Init 初始化小说
		Init(bookID string) (*novel.Novel, error)
		// NewChapter 初始化章节
		NewChapter(cpURL string) (*chapter.Chapter, error)
	}
	// 不支持的平台
	notSupportedError struct {
		Platform // 平台
	}
)

var (
	// Platforms 小说平台，用于各平台的实现导入
	Platforms []Platform
	// CommentUpdate 评论更新
	CommentUpdate = func(p Platform, nv fmt.Stringer, bookID string, cpID string) string {
		// 默认为空实现
		return ``
	}
)

func init() {
	novel.NewChapter = func(nv *novel.Novel, cpURL string) (*chapter.Chapter, error) {
		return Get(nv.Platform).NewChapter(cpURL)
	}
	novel.CommentUpdate = func(nv *novel.Novel, cpID string) string {
		return CommentUpdate(Get(nv.Platform), nv, nv.ID, cpID)
	}
}

// Get 获取小说平台
func Get(platform string) Platform {
	for _, p := range Platforms {
		if p.String() == platform {
			return p
		}
	}
	return nil
}

// NotSupported *notSupportedErr 的构造函数，不支持的平台
func NotSupported(p Platform) *notSupportedError {
	return &notSupportedError{
		Platform: p,
	}
}

// Error 实现 error
func (e *notSupportedError) Error() string {
	return e.Platform.String() + `不是受支持的小说平台喵！`
}

// ParseTime 解析时间
func ParseTime(p Platform, str string) (time.Time, error) {
	return time.Parse(p.Layout(), str)
}

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

	// NotSupportedError 不支持的平台
	NotSupportedError struct {
		platform Platform // 平台
	}

	// IDStringer 实现 fmt.Stringer，获得书号、更新章号
	BookChapterIDStringer interface {
		BookID() string
		ChapterID() (string, error)
		fmt.Stringer
	}
)

var (
	// Platforms 小说平台，用于各平台的实现导入
	Platforms []Platform
	// CommentUpdate (p Platform, nv BookChapterIDStringer) (string, error)
	// 评论更新
	CommentUpdate = func(_ Platform, _ BookChapterIDStringer) (string, error) {
		// 默认为空实现
		return ``, nil
	}
)

func init() {
	novel.NewChapter = func(nv *novel.Novel, cpURL string) (*chapter.Chapter, error) {
		p, err := Get(nv.Platform)
		if err != nil {
			return nil, err
		}
		return p.NewChapter(cpURL)
	}
	novel.CommentUpdate = func(nv *novel.Novel) (string, error) {
		p, err := Get(nv.Platform)
		if err != nil {
			return ``, err
		}
		return CommentUpdate(p, nv)
	}
	novel.ChapterID = func(nv *novel.Novel) (string, error) {
		p, err := Get(nv.Platform)
		if err != nil {
			return ``, err
		}
		return p.ChapterID(nv.Chapter.URL), nil
	}
}

// Get 获取小说平台
func Get(platform string) (Platform, error) {
	for _, p := range Platforms {
		if p.String() == platform {
			return p, nil
		}
	}
	return nil, NotSupported(nil)
}

// NotSupported *NotSupportedError 的构造函数，不支持的平台
func NotSupported(p Platform) *NotSupportedError {
	return &NotSupportedError{
		platform: p,
	}
}

// Error 实现 error
func (e *NotSupportedError) Error() string {
	if p := e.platform; p != nil {
		return p.String() + `不是受支持的小说平台喵！`
	}
	return `小说平台无法识别喵！`
}

// ParseTime 解析时间
func ParseTime(p Platform, str string) (time.Time, error) {
	return time.Parse(p.Layout(), str)
}

package platform

import (
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
		Init(nvID string) (any, error)
		// NewChapter 初始化章节
		NewChapter(cpURL string) (any, error)
	}

	// NotSupportedError 不支持的平台
	NotSupportedError struct {
		platform string // 平台
	}
)

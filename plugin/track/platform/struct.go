package platform

import (
	"context"

	"github.com/Kittengarten/KittenCore/plugin/track/search"
)

// Platform 平台
type Platform interface {
	// String 实现 fmt.Stringer，获得平台名称
	String() string
	// Layout 时间格式
	Layout() string
	// FindBookID 搜索书号
	FindBookID(context.Context, search.Keyword) (string, error)
	// ChapterID 获取章号
	ChapterID(cpURL string) string
	// Init 初始化小说
	Init(ctx context.Context, nvID string) (any, error)
	// NewChapter 初始化章节
	NewChapter(ctx context.Context, cpURL string) (any, error)
}

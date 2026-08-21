package platform

import (
	"context"
	"fmt"

	"github.com/Kittengarten/KittenCore/plugin/track/chapter"
	"github.com/Kittengarten/KittenCore/plugin/track/novel"
	"github.com/Kittengarten/KittenCore/plugin/track/search"
)

// Platform 平台
type Platform interface {
	// Stringer 获得平台名称
	fmt.Stringer
	// Layout 时间格式
	Layout() string
	// FindBookID 搜索书号
	FindBookID(context.Context, search.Keyword) (string, error)
	// ChapterIDSource 小说更新章号源
	novel.ChapterIDSource
	// Source 小说源
	novel.Source
	// Source 章节源
	chapter.Source
}

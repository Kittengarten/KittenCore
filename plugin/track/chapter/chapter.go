package chapter

import (
	"context"
	"sync"
)

// 章节池
var pool = sync.Pool{
	New: func() any {
		return new(Chapter)
	},
}

// String 实现 fmt.Stringer
func (cp *Chapter) String() string {
	return cp.Title + "\n" + cp.URL
}

// Source 章节源
type Source interface {
	// NewChapter 初始化章节
	NewChapter(ctx context.Context, chapURL, volumeName string) (*Chapter, error)
}

// GetChapterSource 获取章节源
var GetChapterSource func(platform string) (Source, error)

// Get 获取章节
func Get() *Chapter {
	return pool.Get().(*Chapter)
}

// Put 回收章节
func (cp *Chapter) Put() {
	if cp == nil {
		return
	}
	*cp = Chapter{}
	pool.Put(cp)
}
